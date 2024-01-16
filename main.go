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
	batch_size := flag.Int("n", 1, "batch_size")
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
			client.SendBasket("1", *input, *batch_size)
		case "s": // sell
			client.SendBasket("2", *input, *batch_size)
		case "c":
			client.CancelAll()
		}
	}
}
