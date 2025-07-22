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

func TestRaydiumV4Transaction(t *testing.T) {
	rpcClient := rpc.New("https://mainnet.helius-rpc.com/?api-key=b69296cd-290c-43f7-bf49-9e9ded12b5e0")
	txSig := solana.MustSignatureFromBase58("5kaAWK5X9DdMmsWm6skaUXLd6prFisuYJavd9B62A941nRGcrmwvncg3tRtUfn7TcMLsrrmjCChdEjK3sjxS6YG9")

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

	fmt.Printf("=== Raydium V4 交易解析结果 ===\n")
	marshalledSwapData, _ := json.MarshalIndent(swapInfo, "", "  ")
	fmt.Println(string(marshalledSwapData))

	// 验证解析结果
	if len(swapInfo.AMMs) == 0 {
		t.Error("应该解析出 AMM 信息")
	}
}

func TestRaydiumRoutingTransaction(t *testing.T) {
	rpcClient := rpc.New("https://mainnet.helius-rpc.com/?api-key=b69296cd-290c-43f7-bf49-9e9ded12b5e0")
	txSig := solana.MustSignatureFromBase58("51nj5GtAmDC23QkeyfCNfTJ6Pdgwx7eq4BARfq1sMmeEaPeLsx9stFA3Dzt9MeLV5xFujBgvghLGcayC3ZevaQYi")

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

	fmt.Printf("=== Raydium Routing 交易解析结果 ===\n")
	marshalledSwapData, _ := json.MarshalIndent(swapInfo, "", "  ")
	fmt.Println(string(marshalledSwapData))

	// 验证解析结果
	if len(swapInfo.AMMs) == 0 {
		t.Error("应该解析出 AMM 信息")
	}
}

func TestRaydiumCPMMTransaction(t *testing.T) {
	rpcClient := rpc.New("https://mainnet.helius-rpc.com/?api-key=b69296cd-290c-43f7-bf49-9e9ded12b5e0")
	txSig := solana.MustSignatureFromBase58("afUCiFQ6amxuxx2AAwsghLt7Q9GYqHfZiF4u3AHhAzs8p1ThzmrtSUFMbcdJy8UnQNTa35Fb1YqxR6F9JMZynYp")

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

	fmt.Printf("=== Raydium CPMM 交易解析结果 ===\n")
	marshalledSwapData, _ := json.MarshalIndent(swapInfo, "", "  ")
	fmt.Println(string(marshalledSwapData))

	// 验证解析结果
	if len(swapInfo.AMMs) == 0 {
		t.Error("应该解析出 AMM 信息")
	}
}

func TestRaydiumConcentratedLiquiditySwapV2Transaction(t *testing.T) {
	rpcClient := rpc.New("https://mainnet.helius-rpc.com/?api-key=b69296cd-290c-43f7-bf49-9e9ded12b5e0")
	txSig := solana.MustSignatureFromBase58("2durZHGFkK4vjpWFGc5GWh5miDs8ke8nWkuee8AUYJA8F9qqT2Um76Q5jGsbK3w2MMgqwZKbnENTLWZoi3d6o2Ds")

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

	fmt.Printf("=== Raydium Concentrated Liquidity SwapV2 交易解析结果 ===\n")
	marshalledSwapData, _ := json.MarshalIndent(swapInfo, "", "  ")
	fmt.Println(string(marshalledSwapData))

	// 验证解析结果
	if len(swapInfo.AMMs) == 0 {
		t.Error("应该解析出 AMM 信息")
	}
}

func TestRaydiumConcentratedLiquiditySwapTransaction(t *testing.T) {
	rpcClient := rpc.New("https://mainnet.helius-rpc.com/?api-key=b69296cd-290c-43f7-bf49-9e9ded12b5e0")
	txSig := solana.MustSignatureFromBase58("4MSVpVBwxnYTQSF3bSrAB99a3pVr6P6bgoCRDsrBbDMA77WeQqoBDDDXqEh8WpnUy5U4GeotdCG9xyExjNTjYE1u")

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

	fmt.Printf("=== Raydium Concentrated Liquidity Swap 交易解析结果 ===\n")
	marshalledSwapData, _ := json.MarshalIndent(swapInfo, "", "  ")
	fmt.Println(string(marshalledSwapData))

	// 验证解析结果
	if len(swapInfo.AMMs) == 0 {
		t.Error("应该解析出 AMM 信息")
	}
}

func TestRaydiumLaunchLabBuyTransaction(t *testing.T) {
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

	fmt.Printf("=== Raydium LaunchLab Buy 交易解析结果 ===\n")
	marshalledSwapData, _ := json.MarshalIndent(swapInfo, "", "  ")
	fmt.Println(string(marshalledSwapData))

	// 验证解析结果
	if len(swapInfo.AMMs) == 0 {
		t.Error("应该解析出 AMM 信息")
	}
}
