package solanaswapgo

import (
	"bytes"
	"fmt"

	ag_binary "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/mr-tron/base58"
)

// Raydium LaunchLab 事件判别器
var (
	RaydiumLaunchLabBuyEventDiscriminator  = [8]byte{102, 230, 103, 87, 137, 106, 134, 187} // buy_exact_in 指令的判别器
	RaydiumLaunchLabSellEventDiscriminator = [8]byte{149, 39, 222, 155, 211, 124, 152, 26}  // sell_exact_in 指令的判别器
)

// RaydiumLaunchLabBuyEvent 表示 Raydium LaunchLab 的买入事件
type RaydiumLaunchLabBuyEvent struct {
	PoolState        solana.PublicKey
	TotalBaseSell    uint64
	VirtualBase      uint64
	VirtualQuote     uint64
	RealBaseBefore   uint64
	RealQuoteBefore  uint64
	RealBaseAfter    uint64
	RealQuoteAfter   uint64
	AmountIn         uint64
	AmountOut        uint64
	MinimumAmountOut uint64
	ProtocolFee      uint64
	PlatformFee      uint64
	ShareFee         uint64
	TradeDirection   TradeDirection
	PoolStatus       PoolStatus
	TokenMint        solana.PublicKey
	TokenDecimals    uint8
	IsBuy            bool // 标识是买入还是卖出
}

// TradeDirection 交易方向
type TradeDirection struct {
	IsBuy bool
}

// PoolStatus 池状态
type PoolStatus struct {
	IsFund bool
}

// RaydiumLaunchLabInstructionData 指令数据结构
type RaydiumLaunchLabInstructionData struct {
	AmountIn         uint64
	MinimumAmountOut uint64
	ShareFeeRate     uint64
}

// processRaydiumLaunchLabSwaps 处理 Raydium LaunchLab 交换
func (p *Parser) processRaydiumLaunchLabSwaps(instructionIndex int) []SwapData {
	var swaps []SwapData

	// 首先尝试解析主指令数据
	mainInstruction := p.txInfo.Message.Instructions[instructionIndex]
	instructionData, isBuy, err := p.parseRaydiumLaunchLabInstruction(mainInstruction)
	if err == nil && instructionData != nil {
		// 创建基于指令数据的事件
		event := &RaydiumLaunchLabBuyEvent{
			AmountIn:         instructionData.AmountIn,
			MinimumAmountOut: instructionData.MinimumAmountOut,
			ShareFee:         instructionData.ShareFeeRate,
			IsBuy:            isBuy,
		}

		// 从转账记录中获取实际的输出金额和代币信息
		p.enrichRaydiumLaunchLabEventFromTransfers(event, instructionIndex)

		swaps = append(swaps, SwapData{Type: RAYDIUM_LAUNCHLAB, Data: event})
		return swaps
	}

	// 如果主指令解析失败，尝试从内部指令解析事件
	for _, innerInstructionSet := range p.txMeta.InnerInstructions {
		if innerInstructionSet.Index == uint16(instructionIndex) {
			for _, innerInstruction := range innerInstructionSet.Instructions {
				if p.isRaydiumLaunchLabEvent(innerInstruction) {
					eventData, err := p.parseRaydiumLaunchLabEvent(innerInstruction)
					if err != nil {
						p.Log.Errorf("error processing Raydium LaunchLab event: %s", err)
						continue
					}
					if eventData != nil {
						// 获取代币信息并设置到事件中
						p.enrichRaydiumLaunchLabEvent(eventData, instructionIndex)
						swaps = append(swaps, SwapData{Type: RAYDIUM_LAUNCHLAB, Data: eventData})
					}
				}
			}
		}
	}

	// 如果没有找到事件数据，尝试使用转账数据进行解析
	if len(swaps) == 0 {
		return p.processRaydiumLaunchLabTransfers(instructionIndex)
	}

	return swaps
}

// enrichRaydiumLaunchLabEvent 从交易中获取代币信息并丰富事件数据
func (p *Parser) enrichRaydiumLaunchLabEvent(event *RaydiumLaunchLabBuyEvent, instructionIndex int) {
	// 从转账记录中获取代币信息
	for _, innerInstructionSet := range p.txMeta.InnerInstructions {
		if innerInstructionSet.Index == uint16(instructionIndex) {
			for _, innerInstruction := range innerInstructionSet.Instructions {
				if p.isTransferCheck(innerInstruction) {
					// 解析转账检查指令以获取代币 mint
					if len(innerInstruction.Accounts) >= 3 {
						mintIndex := innerInstruction.Accounts[2]
						if int(mintIndex) < len(p.allAccountKeys) {
							mint := p.allAccountKeys[mintIndex]
							// 如果不是 SOL，则这应该是输出代币
							if !mint.Equals(NATIVE_SOL_MINT_PROGRAM_ID) {
								event.TokenMint = mint
								if decimals, exists := p.splDecimalsMap[mint.String()]; exists {
									event.TokenDecimals = decimals
								} else {
									event.TokenDecimals = 6 // 默认值
								}
								return
							}
						}
					}
				}
			}
		}
	}
}

// enrichRaydiumLaunchLabEventFromTransfers 从转账记录中获取实际的交换信息
func (p *Parser) enrichRaydiumLaunchLabEventFromTransfers(event *RaydiumLaunchLabBuyEvent, instructionIndex int) {
	var solAmount uint64
	var tokenAmount uint64
	var tokenMint solana.PublicKey
	var tokenDecimals uint8 = 6

	// 分析内部指令中的转账记录
	for _, innerInstructionSet := range p.txMeta.InnerInstructions {
		if innerInstructionSet.Index == uint16(instructionIndex) {
			for _, innerInstruction := range innerInstructionSet.Instructions {
				if p.isTransferCheck(innerInstruction) {
					transfer := p.processTransferCheck(innerInstruction)
					if transfer != nil {
						// 检查是否是 SOL 转账
						if transfer.Info.Mint == NATIVE_SOL_MINT_PROGRAM_ID.String() {
							solAmount, _ = parseUint64(transfer.Info.TokenAmount.Amount)
						} else if transfer.Info.Mint != "Unknown" && transfer.Info.Mint != "" {
							// 这应该是代币转账
							tokenAmount, _ = parseUint64(transfer.Info.TokenAmount.Amount)
							tokenMint = solana.MustPublicKeyFromBase58(transfer.Info.Mint)
							tokenDecimals = transfer.Info.TokenAmount.Decimals
						}
					}
				}
			}
		}
	}

	// 更新事件数据 - 根据买入/卖出方向正确设置输入输出
	if event.IsBuy {
		// 买入：SOL -> Token
		if solAmount > 0 {
			event.AmountIn = solAmount
		}
		if tokenAmount > 0 {
			event.AmountOut = tokenAmount
		}
	} else {
		// 卖出：Token -> SOL
		if tokenAmount > 0 {
			event.AmountIn = tokenAmount
		}
		if solAmount > 0 {
			event.AmountOut = solAmount
		}
	}

	if !tokenMint.IsZero() {
		event.TokenMint = tokenMint
		event.TokenDecimals = tokenDecimals
	}
}

// parseUint64 辅助函数，将字符串转换为 uint64
func parseUint64(s string) (uint64, error) {
	var result uint64
	for _, c := range s {
		if c < '0' || c > '9' {
			break
		}
		result = result*10 + uint64(c-'0')
	}
	return result, nil
}

// processRaydiumLaunchLabTransfers 通过分析转账记录解析交换信息
func (p *Parser) processRaydiumLaunchLabTransfers(instructionIndex int) []SwapData {
	var swaps []SwapData

	for _, innerInstructionSet := range p.txMeta.InnerInstructions {
		if innerInstructionSet.Index == uint16(instructionIndex) {
			for _, innerInstruction := range innerInstructionSet.Instructions {
				switch {
				case p.isTransferCheck(innerInstruction):
					transfer := p.processTransferCheck(innerInstruction)
					if transfer != nil {
						swaps = append(swaps, SwapData{Type: RAYDIUM_LAUNCHLAB, Data: transfer})
					}
				case p.isTransfer(innerInstruction):
					transfer := p.processTransfer(innerInstruction)
					if transfer != nil {
						swaps = append(swaps, SwapData{Type: RAYDIUM_LAUNCHLAB, Data: transfer})
					}
				}
			}
		}
	}
	return swaps
}

// isRaydiumLaunchLabEvent 检查是否为 Raydium LaunchLab 事件指令
func (p *Parser) isRaydiumLaunchLabEvent(inst solana.CompiledInstruction) bool {
	if !p.allAccountKeys[inst.ProgramIDIndex].Equals(RAYDIUM_LAUNCHLAB_PROGRAM_ID) {
		return false
	}

	if len(inst.Data) < 8 {
		return false
	}

	decodedBytes, err := base58.Decode(inst.Data.String())
	if err != nil {
		return false
	}

	// 检查是否匹配已知的事件判别器
	return bytes.Equal(decodedBytes[:8], RaydiumLaunchLabBuyEventDiscriminator[:]) || bytes.Equal(decodedBytes[:8], RaydiumLaunchLabSellEventDiscriminator[:])
}

// parseRaydiumLaunchLabEvent 解析 Raydium LaunchLab 事件
func (p *Parser) parseRaydiumLaunchLabEvent(instruction solana.CompiledInstruction) (*RaydiumLaunchLabBuyEvent, error) {
	decodedBytes, err := base58.Decode(instruction.Data.String())
	if err != nil {
		return nil, fmt.Errorf("error decoding instruction data: %s", err)
	}

	// 跳过判别器，解析事件数据
	decoder := ag_binary.NewBorshDecoder(decodedBytes[8:])

	var event RaydiumLaunchLabBuyEvent
	if err := decoder.Decode(&event); err != nil {
		return nil, fmt.Errorf("error unmarshaling RaydiumLaunchLabBuyEvent: %s", err)
	}

	return &event, nil
}

// parseRaydiumLaunchLabInstruction 解析 Raydium LaunchLab 主指令数据
func (p *Parser) parseRaydiumLaunchLabInstruction(instruction solana.CompiledInstruction) (*RaydiumLaunchLabInstructionData, bool, error) {
	decodedBytes, err := base58.Decode(instruction.Data.String())
	if err != nil {
		return nil, false, fmt.Errorf("error decoding instruction data: %s", err)
	}

	if len(decodedBytes) < 32 {
		return nil, false, fmt.Errorf("instruction data too short")
	}

	// 检查指令类型
	discriminator := decodedBytes[:8]
	isBuy := bytes.Equal(discriminator, RaydiumLaunchLabBuyEventDiscriminator[:])
	isSell := bytes.Equal(discriminator, RaydiumLaunchLabSellEventDiscriminator[:])

	if !isBuy && !isSell {
		return nil, false, fmt.Errorf("unknown Raydium LaunchLab instruction discriminator: %v", discriminator)
	}

	// 跳过判别器，解析指令参数
	decoder := ag_binary.NewBorshDecoder(decodedBytes[8:])

	var instructionData RaydiumLaunchLabInstructionData
	if err := decoder.Decode(&instructionData); err != nil {
		return nil, false, fmt.Errorf("error unmarshaling instruction data: %s", err)
	}

	return &instructionData, isBuy, nil
}
