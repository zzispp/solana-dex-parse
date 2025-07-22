package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/mr-tron/base58"
	solanaswapgo "github.com/zzispp/solana-dex-parse/dex-parse"
)

/*
Example Transactions:
- Orca: 2kAW5GAhPZjM3NoSrhJVHdEpwjmq9neWtckWnjopCfsmCGB27e3v2ZyMM79FdsL4VWGEtYSFi1sF1Zhs7bqdoaVT
- Pumpfun: 4Cod1cNGv6RboJ7rSB79yeVCR4Lfd25rFgLY3eiPJfTJjTGyYP1r2i1upAYZHQsWDqUbGd1bhTRm1bpSQcpWMnEz
- Pumpfun AMM (Pumpswap): 23QJ6qbKcwzA76TX2uSaEb3EtBorKYty9phGYUueMyGoazopvyyZfPfGmGgGzmdt5CPW9nEuB72nnBfaGnydUa6D
- Banana Gun: oXUd22GQ1d45a6XNzfdpHAX6NfFEfFa9o2Awn2oimY89Rms3PmXL1uBJx3CnTYjULJw6uim174b3PLBFkaAxKzK
- Jupiter: DBctXdTTtvn7Rr4ikeJFCBz4AtHmJRyjHGQFpE59LuY3Shb7UcRJThAXC7TGRXXskXuu9LEm9RqtU6mWxe5cjPF
- Jupiter DCA: 4mxr44yo5Qi7Rabwbknkh8MNUEWAMKmzFQEmqUVdx5JpHEEuh59TrqiMCjZ7mgZMozRK1zW8me34w8Myi8Qi1tWP
- Meteora DLMM: 125MRda3h1pwGZpPRwSRdesTPiETaKvy4gdiizyc3SWAik4cECqKGw2gggwyA1sb2uekQVkupA2X9S4vKjbstxx3
- Rayd V4: 5kaAWK5X9DdMmsWm6skaUXLd6prFisuYJavd9B62A941nRGcrmwvncg3tRtUfn7TcMLsrrmjCChdEjK3sjxS6YG9
- Rayd Routing: 51nj5GtAmDC23QkeyfCNfTJ6Pdgwx7eq4BARfq1sMmeEaPeLsx9stFA3Dzt9MeLV5xFujBgvghLGcayC3ZevaQYi
- Rayd CPMM: afUCiFQ6amxuxx2AAwsghLt7Q9GYqHfZiF4u3AHhAzs8p1ThzmrtSUFMbcdJy8UnQNTa35Fb1YqxR6F9JMZynYp
- Rayd Concentrated Liquidity SwapV2: 2durZHGFkK4vjpWFGc5GWh5miDs8ke8nWkuee8AUYJA8F9qqT2Um76Q5jGsbK3w2MMgqwZKbnENTLWZoi3d6o2Ds
- Rayd Concentrated Liquidity Swap: 4MSVpVBwxnYTQSF3bSrAB99a3pVr6P6bgoCRDsrBbDMA77WeQqoBDDDXqEh8WpnUy5U4GeotdCG9xyExjNTjYE1u
- Maestro: mWaH4FELcPj4zeY4Cgk5gxUirQDM7yE54VgMEVaqiUDQjStyzwNrxLx4FMEaKEHQoYsgCRhc1YdmBvhGDRVgRrq
- Meteora Pools Program: 4uuw76SPksFw6PvxLFkG9jRyReV1F4EyPYNc3DdSECip8tM22ewqGWJUaRZ1SJEZpuLJz1qPTEPb2es8Zuegng9Z
- Meteora DLMM: 5PC8qXvzyeqjiTuYkNKyKRShutvVUt7hXySvg6Ux98oa9xuGT6DpTaYoEJKaq5b3tL4XFtJMxZW8SreujL2YkyPg
- Meteora DAMM v2: 3DBswgW6BS4iBsjA3QRJgXwUCPuv68n4HVYvh7cG5T6XA5wz71xtwo7P2XHdfyT4LPmhvWpzhzaRroWoEN81czLV
- Moonshot: AhiFQX1Z3VYbkKQH64ryPDRwxUv8oEPzQVjSvT7zY58UYDm4Yvkkt2Ee9VtSXtF6fJz8fXmb5j3xYVDF17Gr9CG (Buy)
- Moonshot: 2XYu86VrUXiwNNj8WvngcXGytrCsSrpay69Rt3XBz9YZvCQcZJLjvDfh9UWETFtFW47vi4xG2CkiarRJwSe6VekE (Sell)
- Multiple AMMs: 46Jp5EEUrmdCVcE3jeewqUmsMHhqiWWtj243UZNDFZ3mmma6h2DF4AkgPE9ToRYVLVrfKQCJphrvxbNk68Lub9vw //! not supported yet
- OKX: 5xaT2SXQUyvyLGsnyyoKMwsDoHrx1enCKofkdRMdNaL5MW26gjQBM3AWebwjTJ49uqEqnFu5d9nXJek6gUSGCqbL
- Raydium LaunchLab: 4S9AT3Qc5auU62fYPDdUWCtNb6EDiGXEBAhMjWCRs4ESfqHuYuFyJNXiodTBEjyvPM68prij3a7YKgd1YuL26DPV
*/

func TestParseTransaction(t *testing.T) {
	rpcClient := rpc.New("https://mainnet.helius-rpc.com/?api-key=b69296cd-290c-43f7-bf49-9e9ded12b5e0")
	txSig := solana.MustSignatureFromBase58("3DBswgW6BS4iBsjA3QRJgXwUCPuv68n4HVYvh7cG5T6XA5wz71xtwo7P2XHdfyT4LPmhvWpzhzaRroWoEN81czLV")

	var maxTxVersion uint64 = 0
	tx, err := rpcClient.GetTransaction(
		context.TODO(),
		txSig,
		&rpc.GetTransactionOpts{
			Commitment:                     rpc.CommitmentConfirmed,
			MaxSupportedTransactionVersion: &maxTxVersion,
		},
	)
	if err != nil {
		log.Fatalf("error getting tx: %s", err)
	}

	parser, err := solanaswapgo.NewTransactionParser(tx)
	if err != nil {
		log.Fatalf("error creating parser: %s", err)
	}

	transactionData, err := parser.ParseTransaction()
	if err != nil {
		log.Fatalf("error parsing transaction: %s", err)
	}

	swapInfo, err := parser.ProcessSwapData(transactionData)
	if err != nil {
		log.Fatalf("error processing swap data: %s", err)
	}

	marshalledSwapData, _ := json.MarshalIndent(swapInfo, "", "  ")
	fmt.Println(string(marshalledSwapData))
}

func TestRaydiumLaunchLabTransaction(t *testing.T) {
	rpcClient := rpc.New("https://mainnet.helius-rpc.com/?api-key=b69296cd-290c-43f7-bf49-9e9ded12b5e0")
	txSig := solana.MustSignatureFromBase58("4S9AT3Qc5auU62fYPDdUWCtNb6EDiGXEBAhMjWCRs4ESfqHuYuFyJNXiodTBEjyvPM68prij3a7YKgd1YuL26DPV")

	var maxTxVersion uint64 = 0
	tx, err := rpcClient.GetTransaction(
		context.TODO(),
		txSig,
		&rpc.GetTransactionOpts{
			Commitment:                     rpc.CommitmentConfirmed,
			MaxSupportedTransactionVersion: &maxTxVersion,
		},
	)
	if err != nil {
		log.Fatalf("error getting tx: %s", err)
	}

	// 打印交易的所有程序 ID 用于分析
	fmt.Println("=== 分析 Raydium LaunchLab 交易 ===")

	txInfo, err := tx.Transaction.GetTransaction()
	if err != nil {
		log.Fatalf("failed to get transaction: %s", err)
	}

	allAccountKeys := append(txInfo.Message.AccountKeys, tx.Meta.LoadedAddresses.Writable...)
	allAccountKeys = append(allAccountKeys, tx.Meta.LoadedAddresses.ReadOnly...)

	fmt.Printf("总指令数: %d\n", len(txInfo.Message.Instructions))
	for i, instruction := range txInfo.Message.Instructions {
		programID := allAccountKeys[instruction.ProgramIDIndex]
		fmt.Printf("指令 %d: 程序ID = %s\n", i, programID.String())
		fmt.Printf("  账户数: %d\n", len(instruction.Accounts))
		fmt.Printf("  数据长度: %d\n", len(instruction.Data))
	}

	// 尝试解析
	parser, err := solanaswapgo.NewTransactionParser(tx)
	if err != nil {
		log.Fatalf("error creating parser: %s", err)
	}

	transactionData, err := parser.ParseTransaction()
	if err != nil {
		log.Fatalf("error parsing transaction: %s", err)
	}

	fmt.Printf("\n解析出的交换数据数量: %d\n", len(transactionData))
	for i, swap := range transactionData {
		fmt.Printf("交换 %d: 类型 = %s\n", i, swap.Type)
	}

	if len(transactionData) > 0 {
		swapInfo, err := parser.ProcessSwapData(transactionData)
		if err != nil {
			log.Printf("error processing swap data: %s", err)
		} else {
			marshalledSwapData, _ := json.MarshalIndent(swapInfo, "", "  ")
			fmt.Println("\n最终交换信息:")
			fmt.Println(string(marshalledSwapData))
		}
	}
}

func TestRaydiumLaunchLabSellTransaction(t *testing.T) {
	rpcClient := rpc.New("https://mainnet.helius-rpc.com/?api-key=b69296cd-290c-43f7-bf49-9e9ded12b5e0")
	txSig := solana.MustSignatureFromBase58("4DvBxPsGWWTXybZsUC7g2Cxzweuu4VaNoqvBgtFLk1gSgrQSiXXCteH6wSHkfuFMyaBC3aA56nPaqhbZRzSv5sEz")

	var maxTxVersion uint64 = 0
	tx, err := rpcClient.GetTransaction(
		context.TODO(),
		txSig,
		&rpc.GetTransactionOpts{
			Commitment:                     rpc.CommitmentConfirmed,
			MaxSupportedTransactionVersion: &maxTxVersion,
		},
	)
	if err != nil {
		log.Fatalf("error getting tx: %s", err)
	}

	// 打印交易的所有程序 ID 用于分析
	fmt.Println("=== 分析 Raydium LaunchLab 卖出交易 ===")

	txInfo, err := tx.Transaction.GetTransaction()
	if err != nil {
		log.Fatalf("failed to get transaction: %s", err)
	}

	allAccountKeys := append(txInfo.Message.AccountKeys, tx.Meta.LoadedAddresses.Writable...)
	allAccountKeys = append(allAccountKeys, tx.Meta.LoadedAddresses.ReadOnly...)

	fmt.Printf("总指令数: %d\n", len(txInfo.Message.Instructions))
	for i, instruction := range txInfo.Message.Instructions {
		programID := allAccountKeys[instruction.ProgramIDIndex]
		fmt.Printf("指令 %d: 程序ID = %s\n", i, programID.String())
		fmt.Printf("  账户数: %d\n", len(instruction.Accounts))
		fmt.Printf("  数据长度: %d\n", len(instruction.Data))

		// 如果是 Raydium LaunchLab 指令，打印前8字节作为判别器
		if programID.String() == "LanMV9sAd7wArD4vJFi2qDdfnVhFxYSUg6eADduJ3uj" {
			decodedBytes, err := base58.Decode(instruction.Data.String())
			if err == nil && len(decodedBytes) >= 8 {
				fmt.Printf("  判别器: %v\n", decodedBytes[:8])
			}
		}
	}

	// 尝试解析
	parser, err := solanaswapgo.NewTransactionParser(tx)
	if err != nil {
		log.Fatalf("error creating parser: %s", err)
	}

	transactionData, err := parser.ParseTransaction()
	if err != nil {
		log.Fatalf("error parsing transaction: %s", err)
	}

	fmt.Printf("\n解析出的交换数据数量: %d\n", len(transactionData))
	for i, swap := range transactionData {
		fmt.Printf("交换 %d: 类型 = %s\n", i, swap.Type)
	}

	if len(transactionData) > 0 {
		swapInfo, err := parser.ProcessSwapData(transactionData)
		if err != nil {
			log.Printf("error processing swap data: %s", err)
		} else {
			marshalledSwapData, _ := json.MarshalIndent(swapInfo, "", "  ")
			fmt.Println("\n最终交换信息:")
			fmt.Println(string(marshalledSwapData))
		}
	}
}

func TestJupiterTransactionTimestamp(t *testing.T) {
	rpcClient := rpc.New("https://mainnet.helius-rpc.com/?api-key=b69296cd-290c-43f7-bf49-9e9ded12b5e0")

	// 使用一个已知的 Jupiter 交易签名
	txSig := solana.MustSignatureFromBase58("87RZvR1MT7VpjT2YuHuFGZvQ63u2YXYsvE7WqVbcNm51JQo43sUgm8DEa6wnjpoodWBWuh1YPHJMmcZ45qehVgu")

	var maxTxVersion uint64 = 0
	tx, err := rpcClient.GetTransaction(
		context.TODO(),
		txSig,
		&rpc.GetTransactionOpts{
			Commitment:                     rpc.CommitmentConfirmed,
			MaxSupportedTransactionVersion: &maxTxVersion,
		},
	)
	if err != nil {
		log.Fatalf("error getting tx: %s", err)
	}

	parser, err := solanaswapgo.NewTransactionParser(tx)
	if err != nil {
		log.Fatalf("error creating parser: %s", err)
	}

	transactionData, err := parser.ParseTransaction()
	if err != nil {
		log.Fatalf("error parsing transaction: %s", err)
	}

	swapInfo, err := parser.ProcessSwapData(transactionData)
	if err != nil {
		log.Fatalf("error processing swap data: %s", err)
	}

	fmt.Printf("Jupiter 交易时间戳: %s\n", swapInfo.Timestamp.Format(time.RFC3339))
	fmt.Printf("区块时间: %v\n", tx.BlockTime)

	// 验证时间戳不是零值
	if swapInfo.Timestamp.IsZero() {
		t.Error("时间戳不应该是零值")
	}

	// 验证时间戳不是"现在"（应该是历史时间）
	now := time.Now()
	if swapInfo.Timestamp.After(now.Add(-time.Minute)) {
		t.Error("时间戳应该是历史时间，不应该接近当前时间")
	}

	marshalledSwapData, _ := json.MarshalIndent(swapInfo, "", "  ")
	fmt.Println(string(marshalledSwapData))
}

func TestMeteoraDAMMv2Transaction(t *testing.T) {
	rpcClient := rpc.New("https://mainnet.helius-rpc.com/?api-key=b69296cd-290c-43f7-bf49-9e9ded12b5e0")
	txSig := solana.MustSignatureFromBase58("3DBswgW6BS4iBsjA3QRJgXwUCPuv68n4HVYvh7cG5T6XA5wz71xtwo7P2XHdfyT4LPmhvWpzhzaRroWoEN81czLV")

	var maxTxVersion uint64 = 0
	tx, err := rpcClient.GetTransaction(
		context.TODO(),
		txSig,
		&rpc.GetTransactionOpts{
			Commitment:                     rpc.CommitmentConfirmed,
			MaxSupportedTransactionVersion: &maxTxVersion,
		},
	)
	if err != nil {
		log.Fatalf("error getting tx: %s", err)
	}

	fmt.Println("=== 分析 Meteora DAMM v2 交易 ===")

	txInfo, err := tx.Transaction.GetTransaction()
	if err != nil {
		log.Fatalf("failed to get transaction: %s", err)
	}

	allAccountKeys := append(txInfo.Message.AccountKeys, tx.Meta.LoadedAddresses.Writable...)
	allAccountKeys = append(allAccountKeys, tx.Meta.LoadedAddresses.ReadOnly...)

	fmt.Printf("总指令数: %d\n", len(txInfo.Message.Instructions))
	for i, instruction := range txInfo.Message.Instructions {
		programID := allAccountKeys[instruction.ProgramIDIndex]
		fmt.Printf("指令 %d: 程序ID = %s\n", i, programID.String())
		fmt.Printf("  账户数: %d\n", len(instruction.Accounts))
		fmt.Printf("  数据长度: %d\n", len(instruction.Data))

		// 如果是 Meteora DAMM v2 指令，特别标记
		if programID.String() == "cpamdpZCGKUy5JxQXB4dcpGPiikHawvSWAd6mEn1sGG" {
			fmt.Printf("  >>> 这是 Meteora DAMM v2 指令!\n")
			decodedBytes, err := base58.Decode(instruction.Data.String())
			if err == nil && len(decodedBytes) >= 8 {
				fmt.Printf("  判别器: %v\n", decodedBytes[:8])
			}
		}
	}

	parser, err := solanaswapgo.NewTransactionParser(tx)
	if err != nil {
		log.Fatalf("error creating parser: %s", err)
	}

	transactionData, err := parser.ParseTransaction()
	if err != nil {
		log.Fatalf("error parsing transaction: %s", err)
	}

	fmt.Printf("\n解析出的交换数据数量: %d\n", len(transactionData))
	for i, swap := range transactionData {
		fmt.Printf("交换 %d: 类型 = %s\n", i, swap.Type)
		// 检查数据类型来确定使用了哪种解析方式
		switch data := swap.Data.(type) {
		case *solanaswapgo.MeteoraDAMMv2SwapEvent:
			fmt.Printf("  >>> 使用指令解析! AmountIn=%d, ActualAmountOut=%d\n", data.AmountIn, data.ActualAmountOut)
		case *solanaswapgo.TransferCheck:
			fmt.Printf("  >>> 使用转账解析 (保底机制)\n")
		default:
			fmt.Printf("  >>> 其他类型: %T\n", data)
		}
	}

	if len(transactionData) > 0 {
		swapInfo, err := parser.ProcessSwapData(transactionData)
		if err != nil {
			log.Printf("error processing swap data: %s", err)
		} else {
			marshalledSwapData, _ := json.MarshalIndent(swapInfo, "", "  ")
			fmt.Println("\n最终交换信息:")
			fmt.Println(string(marshalledSwapData))
		}
	} else {
		fmt.Println("没有找到交换数据，可能需要进一步调试...")
	}
}
