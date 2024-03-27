package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
)

func SingleClient() {
	cfg := flag.String("c", "clients/test/local.cfg", "client config file")
	flag.Parse()

	client := TradeClient{ConfigFilename: *cfg}
	client.Start()

	// subcommands: order
	orderCmd := flag.NewFlagSet("order", flag.ContinueOnError)
	direction := orderCmd.String("d", "1", "direction: 1,buy; 2,sell")
	hsOrdType := orderCmd.String("t", "dma", "hsOrdType: dma; dma2; care; dsa")
	input := orderCmd.String("i", "input/single.csv", "stock info file(*.csv)")
	batNum := orderCmd.Int("n", 1, "batch number")
	algo := orderCmd.String("a", "1", "algo: 1,direct; 2,latency")

	// subcommands: cancel
	cancelCmd := flag.NewFlagSet("cancel", flag.ContinueOnError)
	origOrdId := cancelCmd.String("o", "-1", "cancel order: -1, cancel all; other, cancel specified")

	fmt.Println("CMDs: order(o), cancel(c), exit(e) | Helps: order -h; cancel -h;")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := strings.ToLower(scanner.Text())
		args := strings.Fields(line)
		fmt.Println(strings.Repeat("=", 65))
		switch args[0] {
		case "order", "o":
			err := orderCmd.Parse(args[1:])
			if err == nil {
				if *algo == "1" {
					client.SendBasket(*direction, *input, *batNum, *hsOrdType)
				} else if *algo == "2" {
					client.SendAlgo(*direction, *input, *batNum)
				}
			}
		case "cancel", "c":
			err := cancelCmd.Parse(args[1:])
			if err == nil {
				if *origOrdId == "-1" {
					client.CancelAll()
				} else {
					client.CancelOrder(*origOrdId)
				}
			}
		case "exit", "e":
			client.Stop()
			os.Exit(0)
		default:
			fmt.Println("unknown command")
		}
		// reset variable value
		*direction = "1"
		*hsOrdType = "dma"
		*batNum = 1
		*algo = "1"
		*origOrdId = "-1"
		// print tip
		fmt.Println(strings.Repeat("=", 65))
		fmt.Println("CMDs: order(o), cancel(c), exit(e) | Helps: order -h; cancel -h;")
	}
}

func main() {
	SingleClient()
}
