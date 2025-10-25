package main

import (
	"encoding/json"
	"fmt"
	"go-solana-bot/common"
	"go-solana-bot/utils"
	"testing"
	"time"
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
	sign := "51TYWL98dtzpbkEMnQebapGBhT9CxuG4oEeKjHFHLoTt9SQWZ2Rb7rtuvHttjUxAz3hvQG9cnveaDbSDheYHGkP8"

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
	out, err := utils.HttpGet(config.STransactionsUrl+"?network=mainnet-beta&txn_signature="+
		"4rYgW8xHE6G22ZEXTfbEo8odTHmQooRHpa223Avj8f6YMvr5gYB7dY8n9Yej77BDozDEUZWBcumY2en6LnND9FYb",
		map[string]string{"x-api-key": config.STransactionsApiKey})
	if err != nil {
		panic(err)
	}
	fmt.Println(string(out))
	var transaction common.STransaction
	err = json.Unmarshal(out, &transaction)
	if err != nil {
		panic(err)
	}
	fmt.Println(transaction.Result.TokenBalanceChanges)
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
