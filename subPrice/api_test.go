package main

import (
	"encoding/json"
	"fmt"
	"go-solana-bot/common"
	"go-solana-bot/utils"
	"testing"

	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/jsonrpc"
)

func TestUpdateCurrentPrice(t *testing.T) {
	config, _ := utils.LoadConfig()

	db, err := InitMysql(config.MySqlUrl)
	if err != nil {
		return
	}

	err = UpdateCurrentPrice(db, config)
	if err != nil {
		return
	}

}

func TestUpdateAmount(t *testing.T) {
	config, _ := utils.LoadConfig()
	db, err := InitMysql(config.MySqlUrl)
	if err != nil {
		return
	}
	err = UpdateAmount(db, config)
	if err != nil {
		return
	}
}

func TestGetTokenAccounts(t *testing.T) {
	pubkey := "7e8wmEQ9UDSE1ruLJijBPDLLp3YiFSYAjtz787Aj4swr"

	reqBody := fmt.Sprintf(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "getTokenAccountsByOwner",
		"params": [
		  "%s",
		  {
		    "programId": "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"
		  },
		  {
		    "commitment": "finalized",
		    "encoding": "jsonParsed"
          }
		]
	}`, pubkey)

	resp, err := utils.HttpProxyPost(common.MainNetEndPoint, []byte(reqBody), nil)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(resp))
	var rpcResponse jsonrpc.RPCResponse
	var result rpc.GetTokenAccountsResult
	err = json.Unmarshal(resp, &rpcResponse)
	if err != nil {
		panic(err)
	}
	err = rpcResponse.GetObject(&result)
	if err != nil {
		panic(err)
	}
	var amountInfos = make([]common.TokenAccount, 0)
	for _, v := range result.Value {
		data := v.Account.Data.GetRawJSON()
		//fmt.Println(string(data))
		var tokenAccount common.TokenAccount
		_ = json.Unmarshal(data, &tokenAccount)
		//fmt.Println(tokenAccount.Parsed.Info.Mint)
		amountInfos = append(amountInfos, tokenAccount)
	}
}
