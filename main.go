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
	handlInst := orderCmd.String("t", "1", "HandlInst: 1,DMA; 2,DMA2; 3,CARE; 4:DSA")
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
		fmt.Println(strings.Repeat("=", 60))
		switch args[0] {
		case "order", "o":
			err := orderCmd.Parse(args[1:])
			if err == nil {
				if *algo == "direct" {
					client.SendBasket(*direction, *input, *batNum, *handlInst)
				} else if *algo == "latency" {
					client.SendAlgo(*direction, *input, *batNum, *handlInst)
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
		fmt.Println(strings.Repeat("=", 60))
		fmt.Println("CMDs: order(o), cancel(c), exit(e) | Helps: order -h; cancel -h;")
	}
}

func main() {
	SingleClient()
}
