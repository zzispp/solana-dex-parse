package solanaswapgo

import (
	"bytes"
	"fmt"

	ag_binary "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/mr-tron/base58"
)

// Boop.fun 指令判别器
var (
	BoopFunBuyTokenDiscriminator = [8]byte{138, 127, 14, 91, 38, 87, 115, 105} // buy_token 指令的判别器
)

// BoopFunSwapEvent 表示 Boop.fun 的交换事件
type BoopFunSwapEvent struct {
	BuyAmount     uint64 // SOL 输入金额 (lamports)
	AmountOutMin  uint64 // 最小输出代币金额
	TokenOut      uint64 // 实际输出代币金额
	TokenMint     solana.PublicKey
	TokenDecimals uint8
	IsBuy         bool // true 表示买入，false 表示卖出
}

// BoopFunInstructionData 指令数据结构
type BoopFunInstructionData struct {
	BuyAmount    uint64 `borsh:"buy_amount"`
	AmountOutMin uint64 `borsh:"amount_out_min"`
}

// processBoopFunSwaps 处理 Boop.fun 交换
func (p *Parser) processBoopFunSwaps(instructionIndex int) []SwapData {
	var swaps []SwapData

	// 首先尝试解析主指令数据
	mainInstruction := p.txInfo.Message.Instructions[instructionIndex]
	programID := p.allAccountKeys[mainInstruction.ProgramIDIndex]

	// 检查是否是 Boop.fun 程序
	if programID.Equals(BOOPFUN_PROGRAM_ID) {
		// 尝试解析指令数据
		instructionData, err := p.parseBoopFunInstruction(mainInstruction)
		if err == nil && instructionData != nil {
			// 创建基于指令数据的事件
			event := &BoopFunSwapEvent{
				BuyAmount:    instructionData.BuyAmount,
				AmountOutMin: instructionData.AmountOutMin,
				IsBuy:        true, // Boop.fun 的 buy_token 指令
			}

			// 从转账记录中获取实际的输出金额和代币信息
			p.enrichBoopFunEventFromTransfers(event, instructionIndex)

			swaps = append(swaps, SwapData{Type: BOOPFUN, Data: event})
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
						swaps = append(swaps, SwapData{Type: BOOPFUN, Data: transfer})
					}
				case p.isTransfer(innerInstruction):
					transfer := p.processTransfer(innerInstruction)
					if transfer != nil {
						swaps = append(swaps, SwapData{Type: BOOPFUN, Data: transfer})
					}
				}
			}
		}
	}

	return swaps
}

// parseBoopFunInstruction 解析 Boop.fun 指令数据
func (p *Parser) parseBoopFunInstruction(instruction solana.CompiledInstruction) (*BoopFunInstructionData, error) {
	decodedBytes, err := base58.Decode(instruction.Data.String())
	if err != nil {
		return nil, fmt.Errorf("error decoding instruction data: %s", err)
	}

	if len(decodedBytes) < 24 { // 至少需要 8 + 8 + 8 = 24 字节
		return nil, fmt.Errorf("instruction data too short")
	}

	// 检查判别器
	discriminator := decodedBytes[:8]
	if !bytes.Equal(discriminator, BoopFunBuyTokenDiscriminator[:]) {
		return nil, fmt.Errorf("unknown Boop.fun instruction discriminator: %v", discriminator)
	}

	// 跳过判别器，解析指令参数
	remainingBytes := decodedBytes[8:]

	decoder := ag_binary.NewBorshDecoder(remainingBytes)

	var instructionData BoopFunInstructionData
	if err := decoder.Decode(&instructionData); err != nil {
		return nil, fmt.Errorf("error unmarshaling instruction data: %s", err)
	}

	return &instructionData, nil
}

// enrichBoopFunEventFromTransfers 从转账记录中获取实际的交换信息
func (p *Parser) enrichBoopFunEventFromTransfers(event *BoopFunSwapEvent, instructionIndex int) {
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

	// 分析转账记录确定输出代币和金额
	// 根据交易详情，应该有一个 transferChecked 指令是代币输出
	for _, transfer := range transfers {
		// 查找非 SOL 的代币转账
		if transfer.Info.Mint != "So11111111111111111111111111111111111111112" &&
			transfer.Info.Mint != "Unknown" && transfer.Info.Mint != "" {
			event.TokenMint = solana.MustPublicKeyFromBase58(transfer.Info.Mint)
			event.TokenDecimals = transfer.Info.TokenAmount.Decimals
			if tokenOut, err := parseUint64(transfer.Info.TokenAmount.Amount); err == nil {
				event.TokenOut = tokenOut
			}
			break
		}
	}
}
