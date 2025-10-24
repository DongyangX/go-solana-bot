package main

import (
	"MySolanaClient/common"
	"MySolanaClient/utils"
	"testing"
)

func TestSwap(t *testing.T) {

	amount := 0.2 * float64(common.LAMPORTS_PER_USDC)
	Swap("EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
		"3wPQhXYqy861Nhoc4bahtpf7G3e89XCLfZ67ptEfZUSA",
		uint64(amount),
		5)

}

func TestBuyMint(t *testing.T) {
	config, _ := utils.LoadConfig()

	db, err := InitMysql(config.MySqlUrl)
	if err != nil {
		panic(err)
	}
	// BuyMint(mint string, config *utils.Config, client *http.Client, db *gorm.DB)
	BuyMint("CY1P83KnKwFYostvjQcoR2HJLyEJWRBRaVQmYyyD3cR8", config, db)
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
		// SellMint(mint string, uiAmount float64, config *utils.Config, db *gorm.DB)
		SellMint(mint.Token, mint.Amount, config, db)
	}
}
