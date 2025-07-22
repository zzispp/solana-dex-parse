package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	solanaswapgo "github.com/zzispp/solana-dex-parse/dex-parse"
)

func TestJupiterTransaction(t *testing.T) {
	rpcClient := rpc.New("https://mainnet.helius-rpc.com/?api-key=b69296cd-290c-43f7-bf49-9e9ded12b5e0")
	txSig := solana.MustSignatureFromBase58("DBctXdTTtvn7Rr4ikeJFCBz4AtHmJRyjHGQFpE59LuY3Shb7UcRJThAXC7TGRXXskXuu9LEm9RqtU6mWxe5cjPF")

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

	fmt.Printf("=== Jupiter 交易解析结果 ===\n")
	marshalledSwapData, _ := json.MarshalIndent(swapInfo, "", "  ")
	fmt.Println(string(marshalledSwapData))

	// 验证解析结果
	if len(swapInfo.AMMs) == 0 {
		t.Error("应该解析出 AMM 信息")
	}

	// 验证时间戳不是零值
	if swapInfo.Timestamp.IsZero() {
		t.Error("时间戳不应该是零值")
	}

	// 验证时间戳是历史时间
	now := time.Now()
	if swapInfo.Timestamp.After(now.Add(-time.Minute)) {
		t.Error("时间戳应该是历史时间，不应该接近当前时间")
	}
}

func TestJupiterDCATransaction(t *testing.T) {
	rpcClient := rpc.New("https://mainnet.helius-rpc.com/?api-key=b69296cd-290c-43f7-bf49-9e9ded12b5e0")
	txSig := solana.MustSignatureFromBase58("4mxr44yo5Qi7Rabwbknkh8MNUEWAMKmzFQEmqUVdx5JpHEEuh59TrqiMCjZ7mgZMozRK1zW8me34w8Myi8Qi1tWP")

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

	fmt.Printf("=== Jupiter DCA 交易解析结果 ===\n")
	marshalledSwapData, _ := json.MarshalIndent(swapInfo, "", "  ")
	fmt.Println(string(marshalledSwapData))

	// 验证解析结果
	if len(swapInfo.AMMs) == 0 {
		t.Error("应该解析出 AMM 信息")
	}
}
