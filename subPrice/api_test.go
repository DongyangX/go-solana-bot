package main

import (
	"MySolanaClient/utils"
	"testing"
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
