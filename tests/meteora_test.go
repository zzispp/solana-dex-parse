package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	solanaswapgo "github.com/zzispp/solana-dex-parse/dex-parse"
)

func TestMeteoraPoolsTransaction(t *testing.T) {
	rpcClient := rpc.New("https://mainnet.helius-rpc.com/?api-key=b69296cd-290c-43f7-bf49-9e9ded12b5e0")
	txSig := solana.MustSignatureFromBase58("4uuw76SPksFw6PvxLFkG9jRyReV1F4EyPYNc3DdSECip8tM22ewqGWJUaRZ1SJEZpuLJz1qPTEPb2es8Zuegng9Z")

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

	fmt.Printf("=== Meteora Pools Program 交易解析结果 ===\n")
	marshalledSwapData, _ := json.MarshalIndent(swapInfo, "", "  ")
	fmt.Println(string(marshalledSwapData))

	// 验证解析结果
	if len(swapInfo.AMMs) == 0 {
		t.Error("应该解析出 AMM 信息")
	}
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

	fmt.Printf("=== Meteora DAMM v2 交易解析结果 ===\n")
	marshalledSwapData, _ := json.MarshalIndent(swapInfo, "", "  ")
	fmt.Println(string(marshalledSwapData))

	// 验证解析结果
	if len(swapInfo.AMMs) == 0 {
		t.Error("应该解析出 AMM 信息")
	}

	// 验证是否使用了指令解析
	fmt.Printf("解析出的交换数据数量: %d\n", len(transactionData))
	for i, swap := range transactionData {
		fmt.Printf("交换 %d: 类型 = %s\n", i, swap.Type)
		switch data := swap.Data.(type) {
		case *solanaswapgo.MeteoraDAMMv2SwapEvent:
			fmt.Printf("  >>> 使用指令解析! AmountIn=%d, ActualAmountOut=%d\n", data.AmountIn, data.ActualAmountOut)
		case *solanaswapgo.TransferCheck:
			fmt.Printf("  >>> 使用转账解析 (保底机制)\n")
		default:
			fmt.Printf("  >>> 其他类型: %T\n", data)
		}
	}
}

func TestMeteoraAmountTransaction(t *testing.T) {
	rpcClient := rpc.New("https://mainnet.helius-rpc.com/?api-key=b69296cd-290c-43f7-bf49-9e9ded12b5e0")
	txSig := solana.MustSignatureFromBase58("125MRda3h1pwGZpPRwSRdesTPiETaKvy4gdiizyc3SWAik4cECqKGw2gggwyA1sb2uekQVkupA2X9S4vKjbstxx3")

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

	fmt.Printf("=== Meteora Amount 交易解析结果 ===\n")
	marshalledSwapData, _ := json.MarshalIndent(swapInfo, "", "  ")
	fmt.Println(string(marshalledSwapData))

	// 验证解析结果
	if len(swapInfo.AMMs) == 0 {
		t.Error("应该解析出 AMM 信息")
	}
}

func TestMeteoraLDMMTransaction(t *testing.T) {
	rpcClient := rpc.New("https://mainnet.helius-rpc.com/?api-key=b69296cd-290c-43f7-bf49-9e9ded12b5e0")
	txSig := solana.MustSignatureFromBase58("5PC8qXvzyeqjiTuYkNKyKRShutvVUt7hXySvg6Ux98oa9xuGT6DpTaYoEJKaq5b3tL4XFtJMxZW8SreujL2YkyPg")

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

	fmt.Printf("=== Meteora DLMM 交易解析结果 ===\n")
	marshalledSwapData, _ := json.MarshalIndent(swapInfo, "", "  ")
	fmt.Println(string(marshalledSwapData))

	// 验证解析结果
	if len(swapInfo.AMMs) == 0 {
		t.Error("应该解析出 AMM 信息")
	}
}
