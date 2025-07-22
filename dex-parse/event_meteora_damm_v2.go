package solanaswapgo

import (
	"bytes"
	"fmt"

	ag_binary "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/mr-tron/base58"
)

// Meteora DAMM v2 swap 指令判别器
var (
	MeteoraDAMMv2SwapDiscriminator = [8]byte{248, 198, 158, 145, 225, 117, 135, 200} // swap 指令的判别器
)

// MeteoraDAMMv2SwapEvent 表示 Meteora DAMM v2 的交换事件
type MeteoraDAMMv2SwapEvent struct {
	AmountIn         uint64
	MinimumAmountOut uint64
	Direction        uint8 // 0: A to B, 1: B to A
	TokenInMint      solana.PublicKey
	TokenInDecimals  uint8
	TokenOutMint     solana.PublicKey
	TokenOutDecimals uint8
	ActualAmountOut  uint64
}

// MeteoraDAMMv2InstructionData 指令数据结构 - 简化版本，参考 Raydium LaunchLab
type MeteoraDAMMv2InstructionData struct {
	Amount               uint64
	OtherAmountThreshold uint64
}

// processMeteoraDAMMv2Swaps 处理 Meteora DAMM v2 交换
func (p *Parser) processMeteoraDAMMv2Swaps(instructionIndex int) []SwapData {
	var swaps []SwapData

	// 首先尝试解析主指令数据
	mainInstruction := p.txInfo.Message.Instructions[instructionIndex]
	// Add bounds checking for ProgramIDIndex
	if int(mainInstruction.ProgramIDIndex) >= len(p.allAccountKeys) {
		return swaps
	}
	programID := p.allAccountKeys[mainInstruction.ProgramIDIndex]

	// 检查是否是 Meteora DAMM v2 程序
	if programID.Equals(METEORA_DAMM_V2_PROGRAM_ID) {
		instructionData, err := p.parseMeteoraDAMMv2Instruction(mainInstruction)
		if err == nil && instructionData != nil {
			// 创建基于指令数据的事件
			event := &MeteoraDAMMv2SwapEvent{
				AmountIn:         instructionData.Amount,
				MinimumAmountOut: instructionData.OtherAmountThreshold,
				Direction:        0, // 默认方向，会从转账中推断
			}

			// 从转账记录中获取实际的输出金额和代币信息
			p.enrichMeteoraDAMMv2EventFromTransfers(event, instructionIndex)

			swaps = append(swaps, SwapData{Type: METEORA, Data: event})
			return swaps
		}
	}

	// 如果指令解析失败，回退到转账解析作为保底机制
	for _, innerInstructionSet := range p.txMeta.InnerInstructions {
		if innerInstructionSet.Index == uint16(instructionIndex) {
			for _, innerInstruction := range innerInstructionSet.Instructions {
				switch {
				case p.isTransferCheck(innerInstruction):
					transfer := p.processTransferCheck(innerInstruction)
					if transfer != nil {
						swaps = append(swaps, SwapData{Type: METEORA, Data: transfer})
					}
				case p.isTransfer(innerInstruction):
					transfer := p.processTransfer(innerInstruction)
					if transfer != nil {
						swaps = append(swaps, SwapData{Type: METEORA, Data: transfer})
					}
				}
			}
		}
	}

	return swaps
}

// parseMeteoraDAMMv2Instruction 解析 Meteora DAMM v2 指令数据
func (p *Parser) parseMeteoraDAMMv2Instruction(instruction solana.CompiledInstruction) (*MeteoraDAMMv2InstructionData, error) {
	decodedBytes, err := base58.Decode(instruction.Data.String())
	if err != nil {
		fmt.Printf("  [DEBUG] Base58 解码失败: %s\n", err)
		return nil, fmt.Errorf("error decoding instruction data: %s", err)
	}

	fmt.Printf("  [DEBUG] 指令数据长度: %d\n", len(decodedBytes))
	if len(decodedBytes) < 8 {
		fmt.Printf("  [DEBUG] 指令数据太短\n")
		return nil, fmt.Errorf("instruction data too short")
	}

	// 检查指令类型
	discriminator := decodedBytes[:8]
	fmt.Printf("  [DEBUG] 实际判别器: %v\n", discriminator)
	fmt.Printf("  [DEBUG] 期望判别器: %v\n", MeteoraDAMMv2SwapDiscriminator[:])

	if !bytes.Equal(discriminator, MeteoraDAMMv2SwapDiscriminator[:]) {
		fmt.Printf("  [DEBUG] 判别器不匹配\n")
		return nil, fmt.Errorf("unknown Meteora DAMM v2 instruction discriminator: %v", discriminator)
	}

	// 跳过判别器，解析指令参数
	remainingBytes := decodedBytes[8:]
	fmt.Printf("  [DEBUG] 剩余字节长度: %d\n", len(remainingBytes))

	if len(remainingBytes) < 16 { // 至少需要 8 + 8 = 16 字节用于两个 uint64
		fmt.Printf("  [DEBUG] 剩余字节太少，无法解析参数\n")
		return nil, fmt.Errorf("instruction data too short for swap parameters")
	}

	decoder := ag_binary.NewBorshDecoder(remainingBytes)

	var instructionData MeteoraDAMMv2InstructionData
	if err := decoder.Decode(&instructionData); err != nil {
		fmt.Printf("  [DEBUG] Borsh 解码失败: %s\n", err)
		return nil, fmt.Errorf("error unmarshaling instruction data: %s", err)
	}

	fmt.Printf("  [DEBUG] 解码成功! Amount=%d, OtherAmountThreshold=%d\n",
		instructionData.Amount, instructionData.OtherAmountThreshold)
	return &instructionData, nil
}

// enrichMeteoraDAMMv2EventFromTransfers 从转账记录中获取实际的交换信息
func (p *Parser) enrichMeteoraDAMMv2EventFromTransfers(event *MeteoraDAMMv2SwapEvent, instructionIndex int) {
	var transfers []*TransferCheck

	// 收集所有的转账记录
	for _, innerInstructionSet := range p.txMeta.InnerInstructions {
		if innerInstructionSet.Index == uint16(instructionIndex) {
			for _, innerInstruction := range innerInstructionSet.Instructions {
				if p.isTransferCheck(innerInstruction) {
					transfer := p.processTransferCheck(innerInstruction)
					if transfer != nil {
						transfers = append(transfers, transfer)
					}
				}
			}
		}
	}

	// 分析转账记录确定输入输出代币和金额
	if len(transfers) >= 2 {
		// 通常第一个转账是输入，第二个转账是输出
		inputTransfer := transfers[0]
		outputTransfer := transfers[1]

		// 设置输入代币信息
		if inputTransfer.Info.Mint != "Unknown" && inputTransfer.Info.Mint != "" {
			event.TokenInMint = solana.MustPublicKeyFromBase58(inputTransfer.Info.Mint)
			event.TokenInDecimals = inputTransfer.Info.TokenAmount.Decimals
		}

		// 设置输出代币信息
		if outputTransfer.Info.Mint != "Unknown" && outputTransfer.Info.Mint != "" {
			event.TokenOutMint = solana.MustPublicKeyFromBase58(outputTransfer.Info.Mint)
			event.TokenOutDecimals = outputTransfer.Info.TokenAmount.Decimals
			if actualOut, err := parseUint64(outputTransfer.Info.TokenAmount.Amount); err == nil {
				event.ActualAmountOut = actualOut
			}
		}
	}
}
