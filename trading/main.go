package main

import (
	"context"
	"encoding/json"
	"fmt"
	"go-solana-bot/common"
	"go-solana-bot/utils"
	"os"
	"strconv"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/rlog"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {

	sig := make(chan os.Signal)

	config, _ := utils.LoadConfig()

	db, err := InitMysql(config.MySqlUrl)
	if err != nil {
		panic(err)
	}

	// 1.Get Message
	c, err := rocketmq.NewPushConsumer(
		consumer.WithGroupName(config.MqGroup),
		consumer.WithNsResolver(primitive.NewPassthroughResolver([]string{config.MqUrl})),
	)
	if err != nil {
		panic(err)
	}
	rlog.SetLogLevel("error")

	err = c.Subscribe(config.MqTopic, consumer.MessageSelector{}, func(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
		for _, msg := range msgs {
			fmt.Println("Get token swap message:", string(msg.Body))
			var swapMessage common.SwapMessage
			_ = json.Unmarshal(msg.Body, &swapMessage)
			// 2.Swap Token
			if swapMessage.SwapType == "buy" {
				buyMints := swapMessage.BuyMints
				for _, mint := range buyMints {
					// Check signature need 60 second so use goroutine
					// 60 second same token buy or sell message will come, use sql row lock
					if CheckPositionStatus(db, mint, "B") > 0 {
						go BuyMint(mint, config, db)
					}
				}
			} else {
				sellMints := swapMessage.SellMints
				for _, mint := range sellMints {
					// Check signature need 60 second so use goroutine
					// 60 second same token buy or sell message will come, use sql row lock
					if CheckPositionStatus(db, mint.Token, "S") > 0 {
						go SellMint(mint.Token, mint.Amount, config, db)
					}
				}
			}
		}
		return consumer.ConsumeSuccess, nil
	})
	if err != nil {
		panic(err)
	}

	// Note: start after subscribe
	err = c.Start()
	if err != nil {
		panic(err)
	}
	<-sig
	err = c.Shutdown()
	if err != nil {
		panic(err)
	}
}

func BuyMint(mint string, config *utils.Config, db *gorm.DB) {
	amount := config.OneBuyUsd * float64(common.LAMPORTS_PER_USDC) // Transfer to lamports
	swapRecord, err := Swap(config.UsdcToken, mint, uint64(amount), config.BuySlippage)
	if err != nil {
		fmt.Println("swap err:", err)
		return
	}
	// Get Decimals
	pUrl := config.PriceUrl + "?ids=" + mint
	resp, err := utils.HttpProxyGet(pUrl)
	fmt.Println(string(resp))

	priceMap := make(map[string]map[string]float64)
	err = json.Unmarshal(resp, &priceMap)
	if err != nil {
		fmt.Println(err)
		return
	}
	decimals := priceMap[mint]["decimals"]
	swapRecord.Decimals = int64(decimals)

	// the api price is not correct
	// Calculate price
	// price * BuyAmount = 1000000 * OneBuyUsd
	// => price = OneBuyUsd * 1000000 / BuyAmount
	buyAmount, err := strconv.ParseFloat(swapRecord.BuyAmount, 64)
	tokenUsdPrice := config.OneBuyUsd * float64(common.LAMPORTS_PER_USDC) / buyAmount
	swapRecord.BuyPrice = tokenUsdPrice

	// Update data
	UpdateDataBuy(db, swapRecord)
}

func UpdateDataBuy(db *gorm.DB, record *common.SwapRecord) {
	// 1 insert record
	InsertRecord(db, record)

	// 2 insert or update position
	token := record.BuyToken
	buyAmount, err := strconv.ParseFloat(record.BuyAmount, 64)

	position, err := SelectPositionByToken(db, token)
	if err != nil {
		position = &common.Position{
			Token:        token,
			Symbol:       "",
			Amount:       int64(buyAmount),
			CostPrice:    record.BuyPrice,
			CurrentPrice: record.BuyPrice,
			Pnl:          float64(0),
			Decimals:     record.Decimals,
			Status:       "N",
		}
		InsertPosition(db, position)
	} else {
		// already has this token, calculate update amount and avg price
		totalAmount := buyAmount + float64(position.Amount)
		avgPrice := (float64(position.Amount)*position.CostPrice + buyAmount*record.BuyPrice) / totalAmount
		position.Amount = int64(totalAmount)
		position.CostPrice = avgPrice
		position.Status = "N"
		UpdatePosition(db, position)
	}
}

func SellMint(mint string, amount int64, config *utils.Config, db *gorm.DB) {
	swapRecord, err := Swap(mint, config.UsdcToken, uint64(amount), config.SellSlippage)
	if err != nil {
		fmt.Println("swap err:", err)
		return
	}

	// Update data
	UpdateDataSell(db, swapRecord)
}

func UpdateDataSell(db *gorm.DB, record *common.SwapRecord) {
	// 1 insert record
	InsertRecord(db, record)

	// 2 update position
	token := record.SellToken

	position, err := SelectPositionByToken(db, token)
	if err != nil {
		fmt.Println(err)
	} else {
		// set zero
		position.Amount = 0
		position.CostPrice = 0
		position.CurrentPrice = 0
		position.Pnl = 0
		position.Status = "N"
		UpdatePositionZero(db, position)
	}
}

func SelectPositionByToken(db *gorm.DB, token string) (*common.Position, error) {
	var position common.Position
	tx := db.Where("token = ?", token).First(&position)
	return &position, tx.Error
}

func InsertRecord(db *gorm.DB, record *common.SwapRecord) {
	db.Create(&record)
}

func InsertPosition(db *gorm.DB, position *common.Position) {
	db.Create(&position)
}

func UpdatePosition(db *gorm.DB, position *common.Position) {
	db.Updates(position)
}

func UpdatePositionZero(db *gorm.DB, position *common.Position) {
	db.Model(&position).Updates(map[string]interface{}{
		"Amount": 0, "CostPrice": 0, "CurrentPrice": 0, "Pnl": 0, "Status": "N"})
}

// CheckPositionStatus update status to B or S,
// if status is B mean buying, return 0, do noting,
// if status is S mean selling, return 0, do noting,
// only status is N can change, return 1, this like a lock
func CheckPositionStatus(db *gorm.DB, token string, status string) int64 {
	if status == "B" { // if buy, the token will not exist
		var count int64
		db.Table("positions").Where("token = ?", token).Count(&count)
		if count > 0 {
			// not exist, buy
			return 999
		}
	}
	tx := db.Table("positions").Where("token=? and status=?",
		token,
		"N",
	).Update("status", status)
	return tx.RowsAffected
}

func InitMysql(dsn string) (*gorm.DB, error) {
	client, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Info)})
	if err != nil {
		fmt.Println("failed to connect database" + err.Error())
		return nil, err
	}
	fmt.Println("db connect success!")
	return client, nil
}
