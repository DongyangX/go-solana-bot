package utils

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/spf13/viper"
)

type Config struct {
	PublicKey           string
	PrivateKey          string
	SubscribeWallet     string
	RpcUrl              string
	WsUrl               string
	TransactionsUrl     string
	TransactionsApiKey  string
	STransactionsUrl    string
	STransactionsApiKey string
	UseJito             bool
	JitoUrl             string
	QuoteUrl            string
	PriceUrl            string
	OneBuyUsd           float64 // 每次购买的刀乐数
	SolToken            string
	UsdcToken           string
	MqUrl               string
	MqGroup             string
	MqTopic             string
	MySqlUrl            string
	UseProxy            bool
	ProxyUrl            string
	BuySlippage         uint64  // 全局买入滑点
	SellSlippage        uint64  // 全局卖出滑点
	BuyPriorityFee      float64 // 全局买入优先级费用
	SellPriorityFee     float64 // 全局卖出优先级费用
	SellRisePercent     float64 // 止盈百分比
	SellFallPercent     float64 // 止损百分比
}

var config *Config

func LoadConfig() (*Config, error) {
	if config == nil {
		// 获取配置文件的绝对路径
		rootDir := getRootDir()

		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath(rootDir)

		err := viper.ReadInConfig()
		if err != nil {
			return nil, fmt.Errorf("fatal error config file: %w", err)
		}

		config = &Config{}
		err = viper.Unmarshal(&config)
		if err != nil {
			return nil, fmt.Errorf("fatal error config file: %w", err)
		}
	}
	return config, nil
}

// 获取项目根目录
func getRootDir() string {
	_, b, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(b), "..")
}
