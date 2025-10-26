package main

import (
	"go-solana-bot/common"
	"go-solana-bot/utils"
	"testing"

	"github.com/gagliardetto/solana-go"
)

func TestSwap(t *testing.T) {

	amount := 0.2 * float64(solana.LAMPORTS_PER_SOL)
	Swap("So11111111111111111111111111111111111111112",
		"EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
		uint64(amount),
		2, "buy", nil)

}

func TestBuyMint(t *testing.T) {
	config, _ := utils.LoadConfig()

	db, err := InitMysql(config.MySqlUrl)
	if err != nil {
		panic(err)
	}
	// BuyMint(mint string, config *utils.Config, client *http.Client, db *gorm.DB)
	BuyMint("GkyPYa7NnCFbduLknCfBfP7p8564X1VZhwZYJ6CZpump", config, db, nil)
}

func TestSellMint(t *testing.T) {
	config, _ := utils.LoadConfig()
	db, err := InitMysql(config.MySqlUrl)
	if err != nil {
		panic(err)
	}
	mint := common.Position{}
	tx := db.Where("token = ?", "3wPQhXYqy861Nhoc4bahtpf7G3e89XCLfZ67ptEfZUSA").First(&mint)
	if tx.Error != nil {
		panic(tx.Error)
	} else {
		// SellMint(mint string, amount int64, config *utils.Config, db *gorm.DB)
		SellMint(mint.Token, mint.Amount, config, db, nil)
	}
}
