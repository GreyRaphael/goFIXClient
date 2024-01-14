package stock_utils

import (
	"encoding/csv"
	"os"
	"strconv"
)

type Stock struct {
	code  string
	vol   int
	price float64
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
			code:  record[0],
			vol:   vol,
			price: price,
		}
		stocks = append(stocks, stock)
	}
	return stocks
}
