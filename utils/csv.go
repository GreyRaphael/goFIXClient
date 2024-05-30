package utils

import (
	"encoding/csv"
	"os"
	"strconv"
)

type Stock struct {
	Code  string
	Vol   int32
	Price float64
}

func ReadCsv(filename string, sep rune) []Stock {
	// open csv file
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// parse csv file
	reader := csv.NewReader(file)
	reader.Comma = sep
	records, err := reader.ReadAll()
	if err != nil {
		panic(err)
	}

	// return stocks
	var stocks []Stock
	for _, record := range records[1:] {
		vol, _ := strconv.Atoi(record[1])
		price, _ := strconv.ParseFloat(record[2], 64)
		stock := Stock{
			Code:  record[0],
			Vol:   int32(vol),
			Price: price,
		}
		stocks = append(stocks, stock)
	}
	return stocks
}
