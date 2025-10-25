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
	rpcClient := rpc.New(config.RpcUrl)
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
		GetTransactions1(&signatures, rpcClient)
		// Option 2
		//transactions, err := GetTransactions2(signatures, config)
		if err != nil {
			panic(err)
		}

	}
}

// GetTransactions1 GetParsedTransaction
func GetTransactions1(signatures *solana.Signature, rcpClient *rpc.Client) {

	version := uint64(0)
	tx, err := rcpClient.GetParsedTransaction(context.Background(),
		*signatures,
		&rpc.GetParsedTransactionOpts{
			MaxSupportedTransactionVersion: &version,
		},
	)
	if err != nil {
		return
	}

	balances := make([]common.TokenBalanceChange, len(tx.Meta.PostTokenBalances))

	for i, balance := range tx.Meta.PostTokenBalances {
		if balance.Mint.String() != "So11111111111111111111111111111111111111112" {
			balances[i] = common.TokenBalanceChange{
				Owner:      balance.Owner.String(),
				Mint:       balance.Mint.String(),
				PostAmount: balance.UiTokenAmount.Amount,
			}
		}
	}

	for _, preBalance := range tx.Meta.PreTokenBalances {
		for i, postBalance := range balances {
			if preBalance.Owner.String() == postBalance.Owner && preBalance.Mint.String() == postBalance.Mint {
				balances[i].PreAmount = preBalance.UiTokenAmount.Amount
			}
		}
	}

	fmt.Println("show balance change")
	for _, balance := range balances {
		fmt.Printf("%s %s %s -> %s\n", balance.Owner, balance.Mint, balance.PreAmount, balance.PostAmount)
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
