package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
)

func SimpleTest() {
	input := flag.String("i", "input/single.csv", "input file(*.csv)")
	batchSize := flag.Int("n", 1, "the total number of batches")
	cfg := flag.String("c", "clients/test/swap.cfg", "client config file")
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

func MultiClient() {
	// input := flag.String("i", "input/zz500.csv", "input file(*.csv)")
	// batchSize := flag.Int("n", 1, "the total number of batches")
	configsDir := flag.String("c", "clients/multi", "client config files")
	flag.Parse()

	entries, err := os.ReadDir(*configsDir)
	if err != nil {
		panic(err)
	}

	clientCounts := len(entries)
	clients := make(chan TradeClient, clientCounts)
	for _, e := range entries {
		filename := fmt.Sprintf("%s/%s", *configsDir, e.Name())
		clients <- TradeClient{ConfigFilename: filename}
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		switch strings.ToLower(scanner.Text()) {
		case "e": // exit
			os.Exit(0)
		case "in":
			// login
			for i := 0; i < clientCounts; i++ {
				go func(clis chan TradeClient) {
					client := <-clis
					client.Start()
				}(clients)
			}
		}
	}
}

func main() {
	// SimpleTest()
	MultiClient()
}
