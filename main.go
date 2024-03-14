package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
)

func SingleClient() {
	input := flag.String("i", "input/single.csv", "input file(*.csv)")
	batchSize := flag.Int("n", 1, "the total number of batches")
	cfg := flag.String("c", "clients/test/local.cfg", "client config file")
	flag.Parse()

	client := TradeClient{ConfigFilename: *cfg}
	client.Start()

	fmt.Println("cmds: e:exit, a:algo, b:buy, s:sell, c:cancel_all, x:cancel_byid")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		switch strings.ToLower(scanner.Text()) {
		case "e": // exit
			client.Stop()
			os.Exit(0)
		case "a": // algorithm
			client.SendAlgo("1", *input, *batchSize)
		case "b": // buy
			client.SendBasket("1", *input, *batchSize)
		case "s": // sell
			client.SendBasket("2", *input, *batchSize)
		case "c":
			client.CancelAll()
		case "x":
			fmt.Println("Enter origid:")
			var origid string
			fmt.Scan(&origid)
			client.CancelOrder(origid)
		}
	}
}

func main() {
	SingleClient()
}
