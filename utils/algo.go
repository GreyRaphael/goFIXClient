package utils

import (
	"os"

	"github.com/pelletier/go-toml/v2"
)

type DSAConfig struct {
	Name           string
	Duration       int
	MaxMarketShare float64
	TradeStyle     int
	PriceType      int
	Unit           int
	Change         float64
}

func ReadAlgoCfg(filename string) DSAConfig {
	algoBytes, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	var algoCfg DSAConfig
	err = toml.Unmarshal(algoBytes, &algoCfg)
	if err != nil {
		panic(err)
	}
	return algoCfg
}
