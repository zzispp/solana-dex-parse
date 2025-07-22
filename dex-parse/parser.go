package solanaswapgo

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/sirupsen/logrus"
)

const (
	PROTOCOL_RAYDIUM = "raydium"
	PROTOCOL_ORCA    = "orca"
	PROTOCOL_METEORA = "meteora"
	PROTOCOL_PUMPFUN = "pumpfun"
)

type TokenTransfer struct {
	mint     string
	amount   uint64
	decimals uint8
}

type Parser struct {
	txResult        *rpc.GetTransactionResult
	txMeta          *rpc.TransactionMeta
	txInfo          *solana.Transaction
	allAccountKeys  solana.PublicKeySlice
	splTokenInfoMap map[string]TokenInfo
	splDecimalsMap  map[string]uint8
	Log             *logrus.Logger
}

func NewTransactionParser(tx *rpc.GetTransactionResult) (*Parser, error) {
	txInfo, err := tx.Transaction.GetTransaction()
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	return NewTransactionParserFromTransactionResult(tx, txInfo, tx.Meta)
}

func NewTransactionParserFromTransactionResult(txResult *rpc.GetTransactionResult, tx *solana.Transaction, txMeta *rpc.TransactionMeta) (*Parser, error) {
	allAccountKeys := append(tx.Message.AccountKeys, txMeta.LoadedAddresses.Writable...)
	allAccountKeys = append(allAccountKeys, txMeta.LoadedAddresses.ReadOnly...)

	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	})

	parser := &Parser{
		txResult:       txResult,
		txMeta:         txMeta,
		txInfo:         tx,
		allAccountKeys: allAccountKeys,
		Log:            log,
	}

	if err := parser.extractSPLTokenInfo(); err != nil {
		return nil, fmt.Errorf("failed to extract SPL Token Addresses: %w", err)
	}

	if err := parser.extractSPLDecimals(); err != nil {
		return nil, fmt.Errorf("failed to extract SPL decimals: %w", err)
	}

	return parser, nil
}

func NewTransactionParserFromTransaction(tx *solana.Transaction, txMeta *rpc.TransactionMeta) (*Parser, error) {
	allAccountKeys := append(tx.Message.AccountKeys, txMeta.LoadedAddresses.Writable...)
	allAccountKeys = append(allAccountKeys, txMeta.LoadedAddresses.ReadOnly...)

	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	})

	parser := &Parser{
		txResult:       nil, // 在这种情况下我们没有原始结果，所以设置为 nil
		txMeta:         txMeta,
		txInfo:         tx,
		allAccountKeys: allAccountKeys,
		Log:            log,
	}

	if err := parser.extractSPLTokenInfo(); err != nil {
		return nil, fmt.Errorf("failed to extract SPL Token Addresses: %w", err)
	}

	if err := parser.extractSPLDecimals(); err != nil {
		return nil, fmt.Errorf("failed to extract SPL decimals: %w", err)
	}

	return parser, nil
}

// GetBlockTime 返回区块时间戳，如果可用的话
func (p *Parser) GetBlockTime() *time.Time {
	if p.txResult != nil && p.txResult.BlockTime != nil {
		blockTime := time.Unix(int64(*p.txResult.BlockTime), 0)
		return &blockTime
	}
	return nil
}

type SwapData struct {
	Type SwapType
	Data interface{}
}

func (p *Parser) ParseTransaction() ([]SwapData, error) {
	var parsedSwaps []SwapData

	skip := false
	for i, outerInstruction := range p.txInfo.Message.Instructions {
		// Add bounds checking for ProgramIDIndex
		if int(outerInstruction.ProgramIDIndex) >= len(p.allAccountKeys) {
			p.Log.Warnf("ProgramIDIndex %d is out of range (allAccountKeys length: %d), skipping instruction %d", outerInstruction.ProgramIDIndex, len(p.allAccountKeys), i)
			continue
		}
		progID := p.allAccountKeys[outerInstruction.ProgramIDIndex]
		switch {
		case progID.Equals(JUPITER_PROGRAM_ID):
			skip = true
			parsedSwaps = append(parsedSwaps, p.processJupiterSwaps(i)...)
		case progID.Equals(MOONSHOT_PROGRAM_ID):
			skip = true
			parsedSwaps = append(parsedSwaps, p.processMoonshotSwaps()...)
		case progID.Equals(BOOPFUN_PROGRAM_ID):
			skip = true
			parsedSwaps = append(parsedSwaps, p.processBoopFunSwaps(i)...)
		case progID.Equals(BANANA_GUN_PROGRAM_ID) ||
			progID.Equals(MINTECH_PROGRAM_ID) ||
			progID.Equals(BLOOM_PROGRAM_ID) ||
			progID.Equals(NOVA_PROGRAM_ID) ||
			progID.Equals(MAESTRO_PROGRAM_ID):
			if innerSwaps := p.processRouterSwaps(i); len(innerSwaps) > 0 {
				parsedSwaps = append(parsedSwaps, innerSwaps...)
			}
		case progID.Equals(OKX_DEX_ROUTER_PROGRAM_ID):
			skip = true
			parsedSwaps = append(parsedSwaps, p.processOKXSwaps(i)...)
		}
	}
	if skip {
		return parsedSwaps, nil
	}

	for i, outerInstruction := range p.txInfo.Message.Instructions {
		// Add bounds checking for ProgramIDIndex
		if int(outerInstruction.ProgramIDIndex) >= len(p.allAccountKeys) {
			p.Log.Warnf("ProgramIDIndex %d is out of range (allAccountKeys length: %d), skipping instruction %d", outerInstruction.ProgramIDIndex, len(p.allAccountKeys), i)
			continue
		}
		progID := p.allAccountKeys[outerInstruction.ProgramIDIndex]
		switch {
		case progID.Equals(RAYDIUM_V4_PROGRAM_ID) ||
			progID.Equals(RAYDIUM_CPMM_PROGRAM_ID) ||
			progID.Equals(RAYDIUM_AMM_PROGRAM_ID) ||
			progID.Equals(RAYDIUM_CONCENTRATED_LIQUIDITY_PROGRAM_ID) ||
			progID.Equals(solana.MustPublicKeyFromBase58("AP51WLiiqTdbZfgyRMs35PsZpdmLuPDdHYmrB23pEtMU")):
			parsedSwaps = append(parsedSwaps, p.processRaydSwaps(i)...)
		case progID.Equals(RAYDIUM_LAUNCHLAB_PROGRAM_ID):
			parsedSwaps = append(parsedSwaps, p.processRaydiumLaunchLabSwaps(i)...)
		case progID.Equals(ORCA_PROGRAM_ID):
			parsedSwaps = append(parsedSwaps, p.processOrcaSwaps(i)...)
		case progID.Equals(METEORA_DAMM_V2_PROGRAM_ID):
			parsedSwaps = append(parsedSwaps, p.processMeteoraDAMMv2Swaps(i)...)
		case progID.Equals(METEORA_DBC_PROGRAM_ID):
			parsedSwaps = append(parsedSwaps, p.processMeteoraDBCSwaps(i)...)
		case progID.Equals(METEORA_PROGRAM_ID) || progID.Equals(METEORA_POOLS_PROGRAM_ID) || progID.Equals(METEORA_DLMM_PROGRAM_ID):
			parsedSwaps = append(parsedSwaps, p.processMeteoraSwaps(i)...)
		case progID.Equals(PUMPFUN_AMM_PROGRAM_ID):
			parsedSwaps = append(parsedSwaps, p.processPumpfunAMMSwaps(i)...)
		case progID.Equals(PUMP_FUN_PROGRAM_ID) ||
			progID.Equals(solana.MustPublicKeyFromBase58("BSfD6SHZigAfDWSjzD5Q41jw8LmKwtmjskPH9XW1mrRW")):
			parsedSwaps = append(parsedSwaps, p.processPumpfunSwaps(i)...)
		}
	}

	return parsedSwaps, nil
}

type SwapInfo struct {
	Signers    []solana.PublicKey
	Signatures []solana.Signature
	AMMs       []string
	Timestamp  time.Time

	TokenInMint     solana.PublicKey
	TokenInAmount   uint64
	TokenInDecimals uint8

	TokenOutMint     solana.PublicKey
	TokenOutAmount   uint64
	TokenOutDecimals uint8
}

func (p *Parser) ProcessSwapData(swapDatas []SwapData) (*SwapInfo, error) {
	if len(swapDatas) == 0 {
		return nil, fmt.Errorf("no swap data provided")
	}

	swapInfo := &SwapInfo{
		Signatures: p.txInfo.Signatures,
	}

	if p.containsDCAProgram() {
		if len(p.allAccountKeys) > 2 {
			swapInfo.Signers = []solana.PublicKey{p.allAccountKeys[2]}
		} else {
			p.Log.Warnf("Cannot access account index 2 for DCA signer (allAccountKeys length: %d)", len(p.allAccountKeys))
		}
	} else {
		if len(p.allAccountKeys) > 0 {
			swapInfo.Signers = []solana.PublicKey{p.allAccountKeys[0]}
		} else {
			p.Log.Warnf("Cannot access account index 0 for signer (allAccountKeys length: %d)", len(p.allAccountKeys))
		}
	}

	jupiterSwaps := make([]SwapData, 0)
	pumpfunSwaps := make([]SwapData, 0)
	raydiumLaunchLabSwaps := make([]SwapData, 0)
	meteoraDAMMv2Swaps := make([]SwapData, 0)
	boopFunSwaps := make([]SwapData, 0)
	otherSwaps := make([]SwapData, 0)

	for _, swapData := range swapDatas {
		switch swapData.Type {
		case JUPITER:
			jupiterSwaps = append(jupiterSwaps, swapData)
		case PUMP_FUN:
			pumpfunSwaps = append(pumpfunSwaps, swapData)
		case RAYDIUM_LAUNCHLAB:
			raydiumLaunchLabSwaps = append(raydiumLaunchLabSwaps, swapData)
		case METEORA:
			// 检查是否是 DAMM v2 类型
			if _, ok := swapData.Data.(*MeteoraDAMMv2SwapEvent); ok {
				meteoraDAMMv2Swaps = append(meteoraDAMMv2Swaps, swapData)
			} else if _, ok := swapData.Data.(*MeteoraDBCSwapEvent); ok {
				meteoraDAMMv2Swaps = append(meteoraDAMMv2Swaps, swapData) // DBC 也归类到同一处理逻辑
			} else {
				otherSwaps = append(otherSwaps, swapData)
			}
		case BOOPFUN:
			boopFunSwaps = append(boopFunSwaps, swapData)
		default:
			otherSwaps = append(otherSwaps, swapData)
		}
	}

	if len(jupiterSwaps) > 0 {
		jupiterInfo, err := parseJupiterEvents(jupiterSwaps)
		if err != nil {
			return nil, fmt.Errorf("failed to parse Jupiter events: %w", err)
		}

		swapInfo.TokenInMint = jupiterInfo.TokenInMint
		swapInfo.TokenInAmount = jupiterInfo.TokenInAmount
		swapInfo.TokenInDecimals = jupiterInfo.TokenInDecimals
		swapInfo.TokenOutMint = jupiterInfo.TokenOutMint
		swapInfo.TokenOutAmount = jupiterInfo.TokenOutAmount
		swapInfo.TokenOutDecimals = jupiterInfo.TokenOutDecimals
		swapInfo.AMMs = jupiterInfo.AMMs

		// 使用区块时间戳，如果不可用则使用当前时间
		if blockTime := p.GetBlockTime(); blockTime != nil {
			swapInfo.Timestamp = *blockTime
		} else {
			swapInfo.Timestamp = time.Now()
		}

		return swapInfo, nil
	}

	if len(pumpfunSwaps) > 0 {
		switch data := pumpfunSwaps[0].Data.(type) {
		case *PumpfunTradeEvent:
			if data.IsBuy {
				swapInfo.TokenInMint = NATIVE_SOL_MINT_PROGRAM_ID
				swapInfo.TokenInAmount = data.SolAmount
				swapInfo.TokenInDecimals = 9
				swapInfo.TokenOutMint = data.Mint
				swapInfo.TokenOutAmount = data.TokenAmount
				swapInfo.TokenOutDecimals = p.splDecimalsMap[data.Mint.String()]
			} else {
				swapInfo.TokenInMint = data.Mint
				swapInfo.TokenInAmount = data.TokenAmount
				swapInfo.TokenInDecimals = p.splDecimalsMap[data.Mint.String()]
				swapInfo.TokenOutMint = NATIVE_SOL_MINT_PROGRAM_ID
				swapInfo.TokenOutAmount = data.SolAmount
				swapInfo.TokenOutDecimals = 9
			}
			swapInfo.AMMs = append(swapInfo.AMMs, string(pumpfunSwaps[0].Type))
			swapInfo.Timestamp = time.Unix(int64(data.Timestamp), 0)
			return swapInfo, nil
		default:
			otherSwaps = append(otherSwaps, pumpfunSwaps...)
		}
	}

	if len(raydiumLaunchLabSwaps) > 0 {
		switch data := raydiumLaunchLabSwaps[0].Data.(type) {
		case *RaydiumLaunchLabBuyEvent:
			if data.IsBuy {
				// Raydium LaunchLab 买入交易：SOL -> Token
				swapInfo.TokenInMint = NATIVE_SOL_MINT_PROGRAM_ID
				swapInfo.TokenInAmount = data.AmountIn
				swapInfo.TokenInDecimals = 9
				swapInfo.TokenOutMint = data.TokenMint
				swapInfo.TokenOutAmount = data.AmountOut
				swapInfo.TokenOutDecimals = data.TokenDecimals
			} else {
				// Raydium LaunchLab 卖出交易：Token -> SOL
				swapInfo.TokenInMint = data.TokenMint
				swapInfo.TokenInAmount = data.AmountIn
				swapInfo.TokenInDecimals = data.TokenDecimals
				swapInfo.TokenOutMint = NATIVE_SOL_MINT_PROGRAM_ID
				swapInfo.TokenOutAmount = data.AmountOut
				swapInfo.TokenOutDecimals = 9
			}
			swapInfo.AMMs = append(swapInfo.AMMs, string(raydiumLaunchLabSwaps[0].Type))

			// 使用区块时间戳，如果不可用则使用当前时间
			if blockTime := p.GetBlockTime(); blockTime != nil {
				swapInfo.Timestamp = *blockTime
			} else {
				swapInfo.Timestamp = time.Now()
			}
			return swapInfo, nil
		default:
			// 如果是通过转账解析的数据，使用通用处理逻辑
			otherSwaps = append(otherSwaps, raydiumLaunchLabSwaps...)
		}
	}

	if len(boopFunSwaps) > 0 {
		switch data := boopFunSwaps[0].Data.(type) {
		case *BoopFunSwapEvent:
			// Boop.fun 交易：SOL -> Token
			swapInfo.TokenInMint = NATIVE_SOL_MINT_PROGRAM_ID
			swapInfo.TokenInAmount = data.BuyAmount
			swapInfo.TokenInDecimals = 9
			swapInfo.TokenOutMint = data.TokenMint
			swapInfo.TokenOutAmount = data.TokenOut
			swapInfo.TokenOutDecimals = data.TokenDecimals
			swapInfo.AMMs = append(swapInfo.AMMs, string(boopFunSwaps[0].Type))

			// 使用区块时间戳，如果不可用则使用当前时间
			if blockTime := p.GetBlockTime(); blockTime != nil {
				swapInfo.Timestamp = *blockTime
			} else {
				swapInfo.Timestamp = time.Now()
			}
			return swapInfo, nil
		default:
			// 如果不是预期的类型，放入 otherSwaps
			otherSwaps = append(otherSwaps, boopFunSwaps...)
		}
	}

	if len(meteoraDAMMv2Swaps) > 0 {
		switch data := meteoraDAMMv2Swaps[0].Data.(type) {
		case *MeteoraDAMMv2SwapEvent:
			swapInfo.TokenInMint = data.TokenInMint
			swapInfo.TokenInAmount = data.AmountIn
			swapInfo.TokenInDecimals = data.TokenInDecimals
			swapInfo.TokenOutMint = data.TokenOutMint
			swapInfo.TokenOutAmount = data.ActualAmountOut
			swapInfo.TokenOutDecimals = data.TokenOutDecimals
			swapInfo.AMMs = append(swapInfo.AMMs, string(meteoraDAMMv2Swaps[0].Type))

			// 使用区块时间戳，如果不可用则使用当前时间
			if blockTime := p.GetBlockTime(); blockTime != nil {
				swapInfo.Timestamp = *blockTime
			} else {
				swapInfo.Timestamp = time.Now()
			}
			return swapInfo, nil
		case *MeteoraDBCSwapEvent:
			swapInfo.TokenInMint = data.TokenInMint
			swapInfo.TokenInAmount = data.AmountIn
			swapInfo.TokenInDecimals = data.TokenInDecimals
			swapInfo.TokenOutMint = data.TokenOutMint
			swapInfo.TokenOutAmount = data.OutputAmount
			swapInfo.TokenOutDecimals = data.TokenOutDecimals
			swapInfo.AMMs = append(swapInfo.AMMs, string(meteoraDAMMv2Swaps[0].Type))

			// 使用区块时间戳，如果不可用则使用当前时间
			if blockTime := p.GetBlockTime(); blockTime != nil {
				swapInfo.Timestamp = *blockTime
			} else {
				swapInfo.Timestamp = time.Now()
			}
			return swapInfo, nil
		default:
			// 如果不是预期的类型，放入 otherSwaps
			otherSwaps = append(otherSwaps, meteoraDAMMv2Swaps...)
		}
	}

	if len(otherSwaps) > 0 {
		var uniqueTokens []TokenTransfer
		seenTokens := make(map[string]bool)

		for _, swapData := range otherSwaps {
			transfer := getTransferFromSwapData(swapData)
			if transfer != nil && !seenTokens[transfer.mint] {
				uniqueTokens = append(uniqueTokens, *transfer)
				seenTokens[transfer.mint] = true
			}
		}

		if len(uniqueTokens) >= 2 {
			inputTransfer := uniqueTokens[0]
			outputTransfer := uniqueTokens[len(uniqueTokens)-1]

			seenInputs := make(map[string]bool)
			seenOutputs := make(map[string]bool)
			var totalInputAmount uint64 = 0
			var totalOutputAmount uint64 = 0

			for _, swapData := range otherSwaps {
				transfer := getTransferFromSwapData(swapData)
				if transfer == nil {
					continue
				}

				amountStr := fmt.Sprintf("%d-%s", transfer.amount, transfer.mint)
				if transfer.mint == inputTransfer.mint && !seenInputs[amountStr] {
					totalInputAmount += transfer.amount
					seenInputs[amountStr] = true
				}
				if transfer.mint == outputTransfer.mint && !seenOutputs[amountStr] {
					totalOutputAmount += transfer.amount
					seenOutputs[amountStr] = true
				}
			}

			swapInfo.TokenInMint = solana.MustPublicKeyFromBase58(inputTransfer.mint)
			swapInfo.TokenInAmount = totalInputAmount
			swapInfo.TokenInDecimals = inputTransfer.decimals
			swapInfo.TokenOutMint = solana.MustPublicKeyFromBase58(outputTransfer.mint)
			swapInfo.TokenOutAmount = totalOutputAmount
			swapInfo.TokenOutDecimals = outputTransfer.decimals

			seenAMMs := make(map[string]bool)
			for _, swapData := range otherSwaps {
				if !seenAMMs[string(swapData.Type)] {
					swapInfo.AMMs = append(swapInfo.AMMs, string(swapData.Type))
					seenAMMs[string(swapData.Type)] = true
				}
			}

			// 使用区块时间戳，如果不可用则使用当前时间
			if blockTime := p.GetBlockTime(); blockTime != nil {
				swapInfo.Timestamp = *blockTime
			} else {
				swapInfo.Timestamp = time.Now()
			}
			return swapInfo, nil
		}
	}

	return nil, fmt.Errorf("no valid swaps found")
}

func getTransferFromSwapData(swapData SwapData) *TokenTransfer {
	switch data := swapData.Data.(type) {
	case *MeteoraDAMMv2SwapEvent:
		// 对于 Meteora DAMM v2 事件，返回输入代币信息
		return &TokenTransfer{
			mint:     data.TokenInMint.String(),
			amount:   data.AmountIn,
			decimals: data.TokenInDecimals,
		}
	case *MeteoraDBCSwapEvent:
		// 对于 Meteora DBC 事件，返回输入代币信息
		return &TokenTransfer{
			mint:     data.TokenInMint.String(),
			amount:   data.AmountIn,
			decimals: data.TokenInDecimals,
		}
	case *BoopFunSwapEvent:
		// 对于 Boop.fun 事件，返回输入 SOL 信息
		return &TokenTransfer{
			mint:     NATIVE_SOL_MINT_PROGRAM_ID.String(),
			amount:   data.BuyAmount,
			decimals: 9,
		}
	case *TransferData:
		return &TokenTransfer{
			mint:     data.Mint,
			amount:   data.Info.Amount,
			decimals: data.Decimals,
		}
	case *TransferCheck:
		amt, err := strconv.ParseUint(data.Info.TokenAmount.Amount, 10, 64)
		if err != nil {
			return nil
		}
		return &TokenTransfer{
			mint:     data.Info.Mint,
			amount:   amt,
			decimals: data.Info.TokenAmount.Decimals,
		}
	}
	return nil
}

func (p *Parser) processRouterSwaps(instructionIndex int) []SwapData {
	var swaps []SwapData

	innerInstructions := p.getInnerInstructions(instructionIndex)
	if len(innerInstructions) == 0 {
		return swaps
	}

	processedProtocols := make(map[string]bool)

	for _, inner := range innerInstructions {
		// Add bounds checking for ProgramIDIndex in inner instructions
		if int(inner.ProgramIDIndex) >= len(p.allAccountKeys) {
			p.Log.Warnf("Inner instruction ProgramIDIndex %d is out of range (allAccountKeys length: %d), skipping", inner.ProgramIDIndex, len(p.allAccountKeys))
			continue
		}
		progID := p.allAccountKeys[inner.ProgramIDIndex]

		switch {
		case (progID.Equals(RAYDIUM_V4_PROGRAM_ID) ||
			progID.Equals(RAYDIUM_CPMM_PROGRAM_ID) ||
			progID.Equals(RAYDIUM_AMM_PROGRAM_ID) ||
			progID.Equals(RAYDIUM_CONCENTRATED_LIQUIDITY_PROGRAM_ID)) && !processedProtocols[PROTOCOL_RAYDIUM]:
			processedProtocols[PROTOCOL_RAYDIUM] = true
			if raydSwaps := p.processRaydSwaps(instructionIndex); len(raydSwaps) > 0 {
				swaps = append(swaps, raydSwaps...)
			}

		case progID.Equals(ORCA_PROGRAM_ID) && !processedProtocols[PROTOCOL_ORCA]:
			processedProtocols[PROTOCOL_ORCA] = true
			if orcaSwaps := p.processOrcaSwaps(instructionIndex); len(orcaSwaps) > 0 {
				swaps = append(swaps, orcaSwaps...)
			}

		case (progID.Equals(METEORA_PROGRAM_ID) ||
			progID.Equals(METEORA_POOLS_PROGRAM_ID) ||
			progID.Equals(METEORA_DLMM_PROGRAM_ID) ||
			progID.Equals(METEORA_DAMM_V2_PROGRAM_ID) ||
			progID.Equals(METEORA_DBC_PROGRAM_ID)) && !processedProtocols[PROTOCOL_METEORA]:
			processedProtocols[PROTOCOL_METEORA] = true
			if meteoraSwaps := p.processMeteoraSwaps(instructionIndex); len(meteoraSwaps) > 0 {
				swaps = append(swaps, meteoraSwaps...)
			}

		case progID.Equals(PUMPFUN_AMM_PROGRAM_ID) && !processedProtocols[PROTOCOL_PUMPFUN]:
			processedProtocols[PROTOCOL_PUMPFUN] = true
			if pumpfunAMMSwaps := p.processPumpfunAMMSwaps(instructionIndex); len(pumpfunAMMSwaps) > 0 {
				swaps = append(swaps, pumpfunAMMSwaps...)
			}

		case (progID.Equals(PUMP_FUN_PROGRAM_ID) ||
			progID.Equals(solana.MustPublicKeyFromBase58("BSfD6SHZigAfDWSjzD5Q41jw8LmKwtmjskPH9XW1mrRW"))) && !processedProtocols[PROTOCOL_PUMPFUN]:
			processedProtocols[PROTOCOL_PUMPFUN] = true
			if pumpfunSwaps := p.processPumpfunSwaps(instructionIndex); len(pumpfunSwaps) > 0 {
				swaps = append(swaps, pumpfunSwaps...)
			}
		}
	}

	return swaps
}

func (p *Parser) getInnerInstructions(index int) []solana.CompiledInstruction {
	if p.txMeta == nil || p.txMeta.InnerInstructions == nil {
		return nil
	}

	for _, inner := range p.txMeta.InnerInstructions {
		if inner.Index == uint16(index) {
			return inner.Instructions
		}
	}

	return nil
}
