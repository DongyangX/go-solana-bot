package common

import "gorm.io/gorm"

type Transaction struct {
	Signature      string          `json:"signature"`
	Timestamp      uint64          `json:"timestamp"`
	Slot           uint64          `json:"slot"`
	TokenTransfers []TokenTransfer `json:"tokenTransfers"`
}

type TokenTransfer struct {
	FromTokenAccount string  `json:"fromTokenAccount"`
	ToTokenAccount   string  `json:"toTokenAccount"`
	FromUserAccount  string  `json:"fromUserAccount"`
	ToUserAccount    string  `json:"toUserAccount"`
	TokenAmount      float64 `json:"tokenAmount"`
	Mint             string  `json:"mint"`
}

type Position struct {
	gorm.Model
	ID           uint `gorm:"primarykey"`
	Token        string
	Symbol       string
	Amount       int64
	CostPrice    float64
	CurrentPrice float64
	Pnl          float64
	Decimals     int64
	Status       string
}

type SwapRecord struct {
	gorm.Model
	ID         uint `gorm:"primarykey"`
	Signature  string
	BuyToken   string
	BuySymbol  string
	BuyAmount  string
	BuyPrice   float64
	SellToken  string
	SellSymbol string
	SellAmount string
	SellPrice  float64
	Decimals   int64
}

type SwapMessage struct {
	SwapType  string     `json:"swapType"`
	BuyMints  []string   `json:"buyMints"`
	SellMints []SwapMint `json:"sellMints"`
}

type SwapMint struct {
	Token    string `json:"token"`
	Amount   int64  `json:"amount"`
	Decimals int64  `json:"decimals"`
}

type JitoTransactionReqBody struct {
	Id      int64       `json:"id"`
	Jsonrpc string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

type JitoTransactionRespBody struct {
	Id      int64  `json:"id"`
	Jsonrpc string `json:"jsonrpc"`
	Result  string `json:"result"`
}

type STransaction struct {
	Success bool               `json:"success"`
	Message string             `json:"message"`
	Result  STransactionResult `json:"result"`
}

type STransactionResult struct {
	TokenBalanceChanges []TokenBalanceChange `json:"token_balance_changes"`
}

type TokenBalanceChange struct {
	Mint  string `json:"mint"`
	Owner string `json:"owner"`
}

type TokenAccount struct {
	Parsed TokenAccountParsed `json:"parsed"`
}

type TokenAccountParsed struct {
	Info TokenAccountInfo `json:"info"`
}

type TokenAccountInfo struct {
	Mint        string                  `json:"mint"`
	TokenAmount TokenAccountTokenAmount `json:"tokenAmount"`
}

type TokenAccountTokenAmount struct {
	Amount         string  `json:"amount"`
	Decimals       int64   `json:"decimals"`
	UiAmount       float64 `json:"uiAmount"`
	UiAmountString string  `json:"uiAmountString"`
}
