package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
)

func main() {
	input := flag.String("i", "input/single.csv", "input file(*.csv)")
	batchSize := flag.Int("n", 1, "the total number of batches")
	flag.Parse()

	client := TradeClient{ConfigFilename: "clients/test/swap.cfg"}
	client.Start()

	fmt.Println("cmds: e:exit, b:buy, s:sell, c:cancel_all")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		switch strings.ToLower(scanner.Text()) {
		case "e": // exit
			client.Stop()
			os.Exit(0)
		case "b": // buy
			client.SendBasket("1", *input, *batchSize)
		case "s": // sell
			client.SendBasket("2", *input, *batchSize)
		case "c":
			client.CancelAll()
		}
	}
}
