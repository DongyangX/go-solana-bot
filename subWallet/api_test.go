package main

import (
	"context"
	"encoding/json"
	"fmt"
	"go-solana-bot/common"
	"go-solana-bot/utils"
	"testing"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

func TestGetTransactions(t *testing.T) {
	config, err := utils.LoadConfig()
	if err != nil {
		panic(err)
	}

	sign := "51TYWL98dtzpbkEMnQebapGBhT9CxuG4oEeKjHFHLoTt9SQWZ2Rb7rtuvHttjUxAz3hvQG9cnveaDbSDheYHGkP8"
	req := "{\"transactions\": [\"" + sign + "\"]}"
	query := map[string]string{}
	query["api-key"] = config.TransactionsApiKey
	query["commitment"] = "confirmed"
	resp, err := utils.HttpPost(config.TransactionsUrl, []byte(req), query)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(resp))
}

func TestGetTransactions2(t *testing.T) {
	config, err := utils.LoadConfig()
	if err != nil {
		panic(err)
	}
	sign := "3JE25UxASi6dWj7a18aXd1eGTwgJBP9jj2bFNVbG3ozudkhqjewgjSeCZxndp6Dh6Lbt8rZ469tpgy6yxaXPDvbY"

	reqBody := fmt.Sprintf(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "getTransaction",
		"params": [
		  "%s",
		  {
			"encoding": "json",
			"maxSupportedTransactionVersion": 0
		  }
		]
	}`, sign)

	resp, err := utils.HttpPost(config.RpcUrl, []byte(reqBody), nil)
	if err != nil {
		return
	}
	fmt.Println(string(resp))
}

func TestGetTransactions3(t *testing.T) {
	config, err := utils.LoadConfig()
	if err != nil {
		panic(err)
	}
	rpcClient := rpc.New(config.RpcUrl)
	version := uint64(0)
	tx, err := rpcClient.GetParsedTransaction(context.Background(),
		solana.MustSignatureFromBase58("51TYWL98dtzpbkEMnQebapGBhT9CxuG4oEeKjHFHLoTt9SQWZ2Rb7rtuvHttjUxAz3hvQG9cnveaDbSDheYHGkP8"),
		&rpc.GetParsedTransactionOpts{
			MaxSupportedTransactionVersion: &version,
		},
	)
	if err != nil {
		return
	}

	balances := make([]common.TokenBalanceChange, len(tx.Meta.PostTokenBalances))

	for i, balance := range tx.Meta.PostTokenBalances {
		balances[i] = common.TokenBalanceChange{
			Owner:      balance.Owner.String(),
			Mint:       balance.Mint.String(),
			PostAmount: balance.UiTokenAmount.Amount,
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

func TestGetPrice(t *testing.T) {

	config, err := utils.LoadConfig()
	if err != nil {
		panic(err)
	}

	mint := "3wPQhXYqy861Nhoc4bahtpf7G3e89XCLfZ67ptEfZUSA"
	amount := "1000000"
	qUrl := config.QuoteUrl + "?inputMint=" + config.UsdcToken + "&outputMint=" + mint + "&amount=" + amount + "&slippageBps=50"

	resp, err := utils.HttpProxyGet(qUrl)
	fmt.Println(string(resp))

	time.Sleep(time.Second)

	pUrl := config.PriceUrl + "?ids=" + mint
	resp, err = utils.HttpProxyGet(pUrl)
	fmt.Println(string(resp))
}

func Test(t *testing.T) {
	resp := "{\"3wPQhXYqy861Nhoc4bahtpf7G3e89XCLfZ67ptEfZUSA\":{\"usdPrice\":0.004658256824996934,\"blockId\":375211667,\"decimals\":6,\"priceChange24h\":220.99637809419295}}\n"
	priceMap := make(map[string]map[string]float64)
	json.Unmarshal([]byte(resp), &priceMap)
	fmt.Println(priceMap["3wPQhXYqy861Nhoc4bahtpf7G3e89XCLfZ67ptEfZUSA"]["usdPrice"])
}
