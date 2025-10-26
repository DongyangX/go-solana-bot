package main

import (
	"encoding/json"
	"fmt"
	"go-solana-bot/common"
	"go-solana-bot/utils"
	"math"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {

	config, _ := utils.LoadConfig()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	mqUtil := utils.NewMqUtil(config.MqUrl)
	defer mqUtil.Stop()

	db, err := InitMysql(config.MySqlUrl)
	if err != nil {
		panic(err)
	}

	for range ticker.C {
		// 1 Update Current Price
		err := UpdateCurrentPrice(db, config)
		if err != nil {
			fmt.Println(err)
			continue
		}

		// 2 Find need to sell
		positions := SelectAll(db)
		sellMints := make([]common.SwapMint, 0)
		for _, position := range positions {
			pnl := position.Pnl
			if pnl >= config.SellRisePercent {
				fmt.Println("^v^ earn money, sell them all!!!")
				swapMint := common.SwapMint{
					Token:    position.Token,
					Amount:   position.Amount,
					Decimals: position.Decimals,
				}
				sellMints = append(sellMints, swapMint)
			}
			if pnl < 0 && math.Abs(pnl) >= config.SellFallPercent {
				fmt.Println("T_T lose money, sell them all!!!")
				swapMint := common.SwapMint{
					Token:    position.Token,
					Amount:   position.Amount,
					Decimals: position.Decimals,
				}
				sellMints = append(sellMints, swapMint)
			}

		}
		if len(sellMints) > 0 {
			jsonByte, _ := json.Marshal(common.SwapMessage{SwapType: "sell", SellMints: sellMints})
			fmt.Println("Send token sell message:", string(jsonByte))
			_, err := mqUtil.Send(config.MqTopic, jsonByte)
			if err != nil {
				panic(err)
			}
		}
	}
}

func UpdateCurrentPrice(db *gorm.DB, config *utils.Config) error {
	positions := SelectAll(db)

	priceMap, err := GetPrice(positions, config)
	if err != nil {
		return err
	}

	for index, position := range positions {
		token := position.Token
		currentPrice := priceMap[token]["usdPrice"]
		if currentPrice != 0.0 {
			positions[index].CurrentPrice = currentPrice
			// calculate total cost
			totalCost := float64(position.Amount) * position.CostPrice
			// calculate current price
			balance := float64(position.Amount) * currentPrice
			// calculate pnl
			pnl := ((balance - totalCost) / totalCost) * 100

			positions[index].Pnl = pnl
		}
	}
	UpdateAll(db, positions)

	return nil
}

func GetPrice(positions []common.Position, config *utils.Config) (map[string]map[string]float64, error) {
	ids := make([]string, 0)
	for _, position := range positions {
		ids = append(ids, position.Token)
	}

	// Get Price
	pUrl := config.PriceUrl + "?ids="

	// You can query up to 50 ids at once.
	priceMap := make(map[string]map[string]float64)

	count := len(ids) / 50
	remainder := len(ids) % 50
	for i := 0; i <= count; i++ {
		var idsStr string
		if count == i {
			if remainder > 0 {
				idsStr = strings.Join(ids[i*50:], ",")
				fmt.Println(ids[i*50:])
			} else {
				break
			}
		} else {
			idsStr = strings.Join(ids[i*50:(i+1)*50], ",")
		}
		pUrl = pUrl + idsStr
		resp, err := utils.HttpProxyGet(pUrl)
		if err != nil {
			return nil, err
		}
		fmt.Println("GetPrice Response:", string(resp))
		tempMap := make(map[string]map[string]float64)
		err = json.Unmarshal(resp, &tempMap)
		if err != nil {
			return nil, err
		}
		for k, v := range tempMap {
			priceMap[k] = v
		}
	}

	return priceMap, nil
}

func SelectAll(db *gorm.DB) []common.Position {
	var positions []common.Position
	db.Where("amount != ?", 0).Find(&positions)
	return positions
}

func UpdateAll(db *gorm.DB, positions []common.Position) {
	for _, position := range positions {
		db.Model(&position).Updates(common.Position{
			CurrentPrice: position.CurrentPrice,
			Pnl:          position.Pnl,
		})
	}
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
