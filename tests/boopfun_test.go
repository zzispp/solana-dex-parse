package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"testing"

	solanaswapgo "github.com/zzispp/solana-dex-parse/dex-parse"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/mr-tron/base58"
)

func TestBoopFunTransaction(t *testing.T) {
	rpcClient := rpc.New("https://mainnet.helius-rpc.com/?api-key=b69296cd-290c-43f7-bf49-9e9ded12b5e0")
	txSig := solana.MustSignatureFromBase58("3vqyV9oQxsnojjnD2DHHsV4d3BfV2i7RvvbTostEV7Du3u4HoSXbonBZFJ2qgxGEijETsGe7x3SvEdtLWjLdBya2")

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

	fmt.Println("=== 分析 Boop.fun 交易 ===")

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

		// 如果是 Boop.fun 指令，特别标记
		if programID.String() == "boop8hVGQGqehUK2iVEMEnMrL5RbjywRzHKBmBE7ry4" {
			fmt.Printf("  >>> 这是 Boop.fun 指令!\n")
			decodedBytes, err := base58.Decode(instruction.Data.String())
			if err == nil && len(decodedBytes) >= 8 {
				fmt.Printf("  判别器: %v\n", decodedBytes[:8])
				fmt.Printf("  完整数据: %v\n", decodedBytes)
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
		switch data := swap.Data.(type) {
		case *solanaswapgo.BoopFunSwapEvent:
			fmt.Printf("  >>> 使用 Boop.fun 指令解析! BuyAmount=%d, TokenOut=%d\n", data.BuyAmount, data.TokenOut)
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
			fmt.Printf("\n=== Boop.fun 交易解析结果 ===\n")
			marshalledSwapData, _ := json.MarshalIndent(swapInfo, "", "  ")
			fmt.Println(string(marshalledSwapData))
		}
	} else {
		fmt.Println("没有找到交换数据，可能需要进一步调试...")
	}

	// 验证解析结果
	if len(transactionData) == 0 {
		t.Error("应该解析出交换数据")
	}
}
