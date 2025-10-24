package main

import (
	"MySolanaClient/common"
	"MySolanaClient/utils"
	"context"
	"encoding/json"
	"fmt"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
)

func main() {
	config, _ := utils.LoadConfig()

	ctx := context.Background()
	rpcClient := rpc.New(config.RpcUrl)
	wsClient, err := ws.Connect(context.Background(), config.WsUrl)
	if err != nil {
		panic(err)
	}

	mqUtil := utils.NewMqUtil(config.MqUrl)
	defer mqUtil.Stop()

	pubKey := solana.MustPublicKeyFromBase58(config.SubscribeWallet)

	sub, err := wsClient.AccountSubscribe(
		pubKey,
		"",
	)
	if err != nil {
		panic(err)
	}
	defer sub.Unsubscribe()

	for {
		got, err := sub.Recv(ctx)
		if err != nil {
			panic(err)
		}
		j, _ := json.Marshal(got)
		fmt.Printf("%s\n", string(j))

		limit := 1
		signatures, err := rpcClient.GetSignaturesForAddressWithOpts(context.TODO(), pubKey, &rpc.GetSignaturesForAddressOpts{Limit: &limit})
		if err != nil {
			panic(err)
		}
		//spew.Dump(signatures)

		// Option 1
		// GetTransactions1(signatures, rpcClient)
		// Option 2
		transactions, err := GetTransactions2(signatures, config)
		if err != nil {
			panic(err)
		}

		// Only need ToUserAccount's token transfers
		buyMints := make([]string, 0)
		transaction := transactions[0]
		for _, tt := range transaction.TokenTransfers {
			if tt.ToUserAccount == pubKey.String() && tt.Mint != config.SolToken {
				buyMints = append(buyMints, tt.Mint)
			}
		}
		// Remove repeated
		buyMints = RemoveRepeatedElement(buyMints)
		jsonByte, _ := json.Marshal(common.SwapMessage{SwapType: "buy", BuyMints: buyMints})
		fmt.Println("Send token buy message:", string(jsonByte))
		_, err = mqUtil.Send(config.MqTopic, jsonByte)
		if err != nil {
			panic(err)
		}
	}
}

// GetTransactions1 GetParsedTransaction
func GetTransactions1(signatures []*rpc.TransactionSignature, rpcClient *rpc.Client) {
	sign := signatures[0].Signature
	version := uint64(0)
	transaction, err := rpcClient.GetParsedTransaction(
		context.TODO(),
		sign,
		&rpc.GetParsedTransactionOpts{Commitment: "confirmed", MaxSupportedTransactionVersion: &version},
	)
	if err != nil {
		panic(err)
	}
	innerInstruction := transaction.Meta.InnerInstructions[0]
	for _, instruction := range innerInstruction.Instructions {
		if instruction.ProgramId.String() == "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA" {
			j, err := instruction.Parsed.MarshalJSON()
			if err != nil {
				panic(err)
			}
			fmt.Println(string(j))
		}
	}
}

// GetTransactions2 https://api.helius.xyz/v0/transactions
func GetTransactions2(signatures []*rpc.TransactionSignature, config *utils.Config) ([]common.Transaction, error) {
	sign := signatures[0].Signature.String()
	url := config.TransactionsUrl
	req := "{\"transactions\": [\"" + sign + "\"]}"
	query := map[string]string{}
	query["api-key"] = config.TransactionsApiKey
	query["commitment"] = "confirmed"
	resp, err := utils.HttpPost(url, []byte(req), query)
	if err != nil {
		return nil, err
	}

	var transactions = make([]common.Transaction, 0)
	err = json.Unmarshal(resp, &transactions)
	if err != nil {
		return nil, err
	}
	//spew.Dump(transactions)
	return transactions, nil
}

// RemoveRepeatedElement Remove repeated
func RemoveRepeatedElement(arr []string) (newArr []string) {
	newArr = make([]string, 0)
	for i := 0; i < len(arr); i++ {
		repeat := false
		for j := i + 1; j < len(arr); j++ {
			if arr[i] == arr[j] {
				repeat = true
				break
			}
		}
		if !repeat {
			newArr = append(newArr, arr[i])
		}
	}
	return
}
