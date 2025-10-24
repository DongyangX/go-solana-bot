package main

import (
	"context"
	"encoding/json"
	"fmt"
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
	rpcClient := rpc.New(config.RpcUrl)
	pubKey := solana.MustPublicKeyFromBase58(config.SubscribeWallet)

	limit := 1
	signatures, err := rpcClient.GetSignaturesForAddressWithOpts(context.TODO(), pubKey, &rpc.GetSignaturesForAddressOpts{Limit: &limit})
	if err != nil {
		fmt.Println(err)
	}
	sign := signatures[0].Signature.String()
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

func TestGetPrice(t *testing.T) {

	config, err := utils.LoadConfig()
	if err != nil {
		panic(err)
	}

	mint := "3wPQhXYqy861Nhoc4bahtpf7G3e89XCLfZ67ptEfZUSA"
	amount := "1000000"
	qUrl := config.QuoteUrl + "?inputMint=" + config.UsdcToken + "&outputMint=" + mint + "&amount=" + amount + "&slippageBps=50"

	resp, err := utils.HttpGet(qUrl)
	fmt.Println(resp)

	time.Sleep(time.Second)

	pUrl := "https://lite-api.jup.ag/price/v3?ids=" + mint
	resp, err = utils.HttpGet(pUrl)
	fmt.Println(resp)
}

func Test(t *testing.T) {
	resp := "{\"3wPQhXYqy861Nhoc4bahtpf7G3e89XCLfZ67ptEfZUSA\":{\"usdPrice\":0.004658256824996934,\"blockId\":375211667,\"decimals\":6,\"priceChange24h\":220.99637809419295}}\n"
	priceMap := make(map[string]map[string]float64)
	json.Unmarshal([]byte(resp), &priceMap)
	fmt.Println(priceMap["3wPQhXYqy861Nhoc4bahtpf7G3e89XCLfZ67ptEfZUSA"]["usdPrice"])
}
