package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

func main() {
	input := flag.String("i", "input/zz500.csv", "input file(*.csv)")
	batch_size := flag.Int("n", 1, "batch_size")
	flag.Parse()

	client := TradeClient{ConfigFilename: "clients/test/dev.cfg"}
	client.Start()

	fmt.Println("cmds: e:exit, b:buy, s:sell, c:cancel_all, l:order_list")
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
		case "l": // for test
			client.SendOrderList()
		case "c":
			client.CancelAll()
		}

		time.Sleep(time.Second)
		fmt.Println("cmds: e:exit, b:buy, s:sell, c:cancel_all, l:order_list")
	}
}
