package main

import (
	"context"
	"encoding/json"
	"fmt"
	"go-solana-bot/common"
	"go-solana-bot/utils"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
)

func main() {
	config, _ := utils.LoadConfig()

	ctx := context.Background()
	//rpcClient := rpc.New(config.RpcUrl)
	wsClient, err := ws.Connect(context.Background(), config.WsUrl)
	if err != nil {
		panic(err)
	}

	mqUtil := utils.NewMqUtil(config.MqUrl)
	defer mqUtil.Stop()

	//pubKey := solana.MustPublicKeyFromBase58(config.SubscribeWallet)

	sub, err := wsClient.LogsSubscribe(
		ws.LogsSubscribeFilterAll,
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

		signatures := got.Value.Signature

		// Option 1
		GetTransactions1(&signatures, config)
		// Option 2
		//transactions, err := GetTransactions2(signatures, config)
		// TODO analyze transactions and send message to buy token?
	}
}

// GetTransactions1 https://api.shyft.to/sol/v1/transaction/parsed
func GetTransactions1(signatures *solana.Signature, config *utils.Config) {
	out, err := utils.HttpGet(config.STransactionsUrl+"?network=mainnet-beta&txn_signature="+signatures.String(),
		map[string]string{"x-api-key": config.STransactionsApiKey})
	if err != nil {
		panic(err)
	}
	fmt.Println(string(out))
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
