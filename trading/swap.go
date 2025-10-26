package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"go-solana-bot/common"
	"go-solana-bot/utils"
	"time"

	"github.com/gagliardetto/solana-go/rpc"
	"github.com/ilkamo/jupiter-go/jupiter"
	"github.com/ilkamo/jupiter-go/solana"
)

func Swap(inputMint string, outputMint string, amount uint64, slippageBps uint64) (*common.SwapRecord, error) {

	config, err := utils.LoadConfig()
	if err != nil {
		return nil, err
	}

	jupClient, err := jupiter.NewClientWithResponses(jupiter.DefaultAPIURL, jupiter.WithHTTPClient(utils.GetProxyClient()))
	if err != nil {
		return nil, err
	}

	ctx := context.TODO()

	// Get the current quote for a swap.
	// Ensure that the input and output mints are valid.
	// The amount is the smallest unit of the input token.
	quoteResponse, err := jupClient.QuoteGetWithResponse(ctx, &jupiter.QuoteGetParams{
		InputMint:   inputMint,
		OutputMint:  outputMint,
		Amount:      amount,
		SlippageBps: &slippageBps,
	})
	if err != nil {
		return nil, err
	}

	if quoteResponse.JSON200 == nil {
		fmt.Println("GetQuoteWithResponse Error:", string(quoteResponse.Body))
		return nil, fmt.Errorf("invalid GetQuoteWithResponse response")
	}

	quote := quoteResponse.JSON200
	j, _ := json.Marshal(quote)
	fmt.Printf("%s\n", string(j))

	dynamicComputeUnitLimit := true

	// Define the prioritization fee in lamports.
	prioritizationFeeLamports := &struct {
		JitoTipLamports              *uint64 `json:"jitoTipLamports,omitempty"`
		PriorityLevelWithMaxLamports *struct {
			MaxLamports   *uint64                                                                                `json:"maxLamports,omitempty"`
			PriorityLevel *jupiter.SwapRequestPrioritizationFeeLamportsPriorityLevelWithMaxLamportsPriorityLevel `json:"priorityLevel,omitempty"`
		} `json:"priorityLevelWithMaxLamports,omitempty"`
	}{
		PriorityLevelWithMaxLamports: &struct {
			MaxLamports   *uint64                                                                                `json:"maxLamports,omitempty"`
			PriorityLevel *jupiter.SwapRequestPrioritizationFeeLamportsPriorityLevelWithMaxLamportsPriorityLevel `json:"priorityLevel,omitempty"`
		}{
			MaxLamports:   new(uint64),
			PriorityLevel: new(jupiter.SwapRequestPrioritizationFeeLamportsPriorityLevelWithMaxLamportsPriorityLevel),
		},
	}

	*prioritizationFeeLamports.PriorityLevelWithMaxLamports.MaxLamports = 10000
	*prioritizationFeeLamports.PriorityLevelWithMaxLamports.PriorityLevel = jupiter.High

	// If you prefer to set a Jito tip, you can use the following line instead of the above block.
	// *prioritizationFeeLamports.JitoTipLamports = 1000

	// Get instructions for a swap.
	// Ensure your public key is valid.
	swapResponse, err := jupClient.SwapPostWithResponse(ctx, jupiter.SwapPostJSONRequestBody{
		PrioritizationFeeLamports: prioritizationFeeLamports,
		QuoteResponse:             *quote,
		UserPublicKey:             config.PublicKey,
		DynamicComputeUnitLimit:   &dynamicComputeUnitLimit,
	})
	if err != nil {
		return nil, err
	}

	if swapResponse.JSON200 == nil {
		fmt.Println("PostSwapWithResponse Error:", string(quoteResponse.Body))
		return nil, fmt.Errorf("invalid PostSwapWithResponse response")
	}

	swap := swapResponse.JSON200
	j, _ = json.Marshal(swap)
	fmt.Printf("%s\n", string(j))

	// Create a wallet from private key.
	walletPrivateKey := config.PrivateKey
	wallet, err := solana.NewWalletFromPrivateKeyBase58(walletPrivateKey)
	if err != nil {
		return nil, err
	}

	// Create a Solana client. Change the URL to the desired Solana node.
	solanaClient, err := solana.NewClient(wallet, config.RpcUrl)
	if err != nil {
		return nil, err
	}

	var signedTx solana.TxID
	if config.UseJito {
		// Sign and send the transaction with jito
		signedTx, err = SendTransactionWithJito(ctx, swap.SwapTransaction, wallet, config)
		if err != nil {
			return nil, err
		}
	} else {
		// Sign and send the transaction.
		signedTx, err = solanaClient.SendTransactionOnChain(ctx, swap.SwapTransaction)
		if err != nil {
			return nil, err
		}
	}
	j, _ = json.Marshal(signedTx)
	fmt.Printf("Signature: %s\n", string(j))

	// Wait a bit to let the transaction propagate to the network.
	// This is just an example and not a best practice.
	// You could use a ticker or wait until we implement the WebSocket monitoring ;)
	//time.Sleep(20 * time.Second)
	//
	//// Get the status of the transaction (pull the status from the blockchain at intervals
	//// until the transaction is confirmed)
	//_, err = solanaClient.CheckSignature(ctx, signedTx)
	//if err != nil {
	//	panic(err)
	//}

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	// Check 6 Times
	count := 0
	for count < 6 {
		select {
		case <-ticker.C:
			count++
			txSuccess, err := solanaClient.CheckSignature(ctx, signedTx)
			if err != nil {
				fmt.Println("CheckSignature err:", err)
			} else {
				if txSuccess {
					record := common.SwapRecord{
						Signature:  string(signedTx),
						BuyToken:   quote.OutputMint,
						BuyAmount:  quote.OutAmount,
						SellToken:  quote.InputMint,
						SellAmount: quote.InAmount,
					}
					return &record, nil
				}
			}
			if count >= 6 {
				break
			}
		}
	}

	return nil, fmt.Errorf("check signature timeout")
}

func SendTransactionWithJito(ctx context.Context, txBase64 string, wallet solana.Wallet, config *utils.Config) (solana.TxID, error) {
	rpcClient := rpc.New(config.RpcUrl)
	latestBlockhash, err := rpcClient.GetLatestBlockhash(ctx, "")
	if err != nil {
		return "", fmt.Errorf("could not get latest blockhash: %w", err)
	}
	tx, err := solana.NewTransactionFromBase64(txBase64)
	if err != nil {
		return "", fmt.Errorf("could not deserialize swap transaction: %w", err)
	}
	tx.Message.RecentBlockhash = latestBlockhash.Value.Blockhash
	tx, err = wallet.SignTransaction(tx)
	if err != nil {
		return "", fmt.Errorf("could not sign swap transaction: %w", err)
	}
	txData, err := tx.MarshalBinary()
	if err != nil {
		return "", fmt.Errorf("could not encode transaction: %w", err)
	}
	encodedTx := base64.StdEncoding.EncodeToString(txData)

	encodingMap := make(map[string]interface{})
	encodingMap["encoding"] = "base64"
	params := []interface{}{encodedTx, encodingMap}
	reqBody := common.JitoTransactionReqBody{Id: 1, Jsonrpc: "2.0", Method: "sendTransaction", Params: params}

	reqBodyJson, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}
	fmt.Printf("jito request: %s\n", string(reqBodyJson))
	resp, err := utils.HttpPost(config.JitoUrl, reqBodyJson, nil)
	if err != nil {
		return "", err
	}
	fmt.Printf("jito response: %s\n", string(resp))

	var respBody common.JitoTransactionRespBody
	err = json.Unmarshal(resp, &respBody)
	sig := respBody.Result
	return solana.TxID(sig), nil
}
