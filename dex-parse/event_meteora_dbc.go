package solanaswapgo

import (
	"bytes"
	"fmt"

	ag_binary "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/mr-tron/base58"
)

// Meteora DBC swap 指令判别器
var (
	MeteoraDBCSwapDiscriminator = [8]byte{248, 198, 158, 145, 225, 117, 135, 200} // swap 指令的判别器
)

// MeteoraDBCSwapEvent 表示 Meteora DBC 的交换事件
type MeteoraDBCSwapEvent struct {
	Pool              solana.PublicKey
	Config            solana.PublicKey
	TradeDirection    uint8
	HasReferral       bool
	AmountIn          uint64
	MinimumAmountOut  uint64
	ActualInputAmount uint64
	OutputAmount      uint64
	NextSqrtPrice     uint64
	TradingFee        uint64
	ProtocolFee       uint64
	ReferralFee       uint64
	CurrentTimestamp  uint64
	TokenInMint       solana.PublicKey
	TokenInDecimals   uint8
	TokenOutMint      solana.PublicKey
	TokenOutDecimals  uint8
}

// MeteoraDBCInstructionData 指令数据结构
type MeteoraDBCInstructionData struct {
	AmountIn         uint64
	MinimumAmountOut uint64
}

// processMeteoraDBCSwaps 处理 Meteora DBC 交换
func (p *Parser) processMeteoraDBCSwaps(instructionIndex int) []SwapData {
	var swaps []SwapData

	// 首先尝试解析主指令数据
	mainInstruction := p.txInfo.Message.Instructions[instructionIndex]
	// Add bounds checking for ProgramIDIndex
	if int(mainInstruction.ProgramIDIndex) >= len(p.allAccountKeys) {
		p.Log.Warnf("ProgramIDIndex %d is out of range (allAccountKeys length: %d) in Meteora DBC processing", mainInstruction.ProgramIDIndex, len(p.allAccountKeys))
		return swaps
	}
	programID := p.allAccountKeys[mainInstruction.ProgramIDIndex]

	// 检查是否是 Meteora DBC 程序
	if programID.Equals(METEORA_DBC_PROGRAM_ID) {
		// 先尝试从内部事件解析
		event := p.parseMeteoraDBCEvent(instructionIndex)
		if event != nil {
			// 从转账记录中获取代币信息
			p.enrichMeteoraDBCEventFromTransfers(event, instructionIndex)
			swaps = append(swaps, SwapData{Type: METEORA, Data: event})
			return swaps
		}

		// 如果事件解析失败，尝试解析指令数据
		instructionData, err := p.parseMeteoraDBCInstruction(mainInstruction)
		if err == nil && instructionData != nil {
			// 创建基于指令数据的事件
			event := &MeteoraDBCSwapEvent{
				AmountIn:         instructionData.AmountIn,
				MinimumAmountOut: instructionData.MinimumAmountOut,
			}

			// 从转账记录中获取实际的输出金额和代币信息
			p.enrichMeteoraDBCEventFromTransfers(event, instructionIndex)

			swaps = append(swaps, SwapData{Type: METEORA, Data: event})
			return swaps
		}
	}

	// 如果都失败，回退到转账解析作为保底机制
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

// parseMeteoraDBCEvent 从内部指令中解析 Meteora DBC 事件
func (p *Parser) parseMeteoraDBCEvent(instructionIndex int) *MeteoraDBCSwapEvent {
	for _, innerInstructionSet := range p.txMeta.InnerInstructions {
		if innerInstructionSet.Index == uint16(instructionIndex) {
			for _, innerInstruction := range innerInstructionSet.Instructions {
				// 查找 Meteora DBC 的内部事件指令
				if p.isMeteoraDBCEventInstruction(innerInstruction) {
					return p.parseMeteoraDBCEventInstruction(innerInstruction)
				}
			}
		}
	}
	return nil
}

// isMeteoraDBCEventInstruction 检查是否是 Meteora DBC 事件指令
func (p *Parser) isMeteoraDBCEventInstruction(instruction solana.CompiledInstruction) bool {
	// Add bounds checking for ProgramIDIndex
	if int(instruction.ProgramIDIndex) >= len(p.allAccountKeys) {
		p.Log.Warnf("ProgramIDIndex %d is out of range (allAccountKeys length: %d) in isMeteoraDBCEventInstruction", instruction.ProgramIDIndex, len(p.allAccountKeys))
		return false
	}
	programID := p.allAccountKeys[instruction.ProgramIDIndex]
	return programID.Equals(METEORA_DBC_PROGRAM_ID)
}

// parseMeteoraDBCEventInstruction 解析 Meteora DBC 事件指令
func (p *Parser) parseMeteoraDBCEventInstruction(instruction solana.CompiledInstruction) *MeteoraDBCSwapEvent {
	// 这里需要解析事件数据，但由于我们没有具体的事件格式
	// 暂时返回 nil，依赖转账解析
	return nil
}

// parseMeteoraDBCInstruction 解析 Meteora DBC 指令数据
func (p *Parser) parseMeteoraDBCInstruction(instruction solana.CompiledInstruction) (*MeteoraDBCInstructionData, error) {
	decodedBytes, err := base58.Decode(instruction.Data.String())
	if err != nil {
		return nil, fmt.Errorf("error decoding instruction data: %s", err)
	}

	if len(decodedBytes) < 24 { // 至少需要 8 + 8 + 8 = 24 字节
		return nil, fmt.Errorf("instruction data too short")
	}

	// 检查判别器
	discriminator := decodedBytes[:8]
	if !bytes.Equal(discriminator, MeteoraDBCSwapDiscriminator[:]) {
		return nil, fmt.Errorf("unknown Meteora DBC instruction discriminator: %v", discriminator)
	}

	// 跳过判别器，解析指令参数
	remainingBytes := decodedBytes[8:]

	// 从完整数据可以看到: [248 198 158 145 225 117 135 200 35 51 152 219 223 32 0 0 167 133 67 0 0 0 0 0]
	// 前8字节是判别器，后面16字节是两个 uint64 参数
	decoder := ag_binary.NewBorshDecoder(remainingBytes)

	var instructionData MeteoraDBCInstructionData
	if err := decoder.Decode(&instructionData); err != nil {
		return nil, fmt.Errorf("error unmarshaling instruction data: %s", err)
	}

	return &instructionData, nil
}

// enrichMeteoraDBCEventFromTransfers 从转账记录中获取实际的交换信息
func (p *Parser) enrichMeteoraDBCEventFromTransfers(event *MeteoraDBCSwapEvent, instructionIndex int) {
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
		// 根据交易详情，第一个转账是输入 (VERID)，第二个转账是输出 (WSOL)
		inputTransfer := transfers[0]
		outputTransfer := transfers[1]

		// 设置输入代币信息
		if inputTransfer.Info.Mint != "Unknown" && inputTransfer.Info.Mint != "" {
			event.TokenInMint = solana.MustPublicKeyFromBase58(inputTransfer.Info.Mint)
			event.TokenInDecimals = inputTransfer.Info.TokenAmount.Decimals
			if amountIn, err := parseUint64(inputTransfer.Info.TokenAmount.Amount); err == nil {
				event.AmountIn = amountIn
				event.ActualInputAmount = amountIn
			}
		}

		// 设置输出代币信息
		if outputTransfer.Info.Mint != "Unknown" && outputTransfer.Info.Mint != "" {
			event.TokenOutMint = solana.MustPublicKeyFromBase58(outputTransfer.Info.Mint)
			event.TokenOutDecimals = outputTransfer.Info.TokenAmount.Decimals
			if amountOut, err := parseUint64(outputTransfer.Info.TokenAmount.Amount); err == nil {
				event.OutputAmount = amountOut
			}
		}
	}
}
