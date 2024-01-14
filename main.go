package main

import (
	"bufio"
	"fmt"
	"os"
	"time"
)

func main() {
	client := TradeClient{ConfigFilename: "clients/test/dev.cfg"}
	client.Start()
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("cmds: exit, o:order, s:strategy, b:basket, l:orderlist")
	for scanner.Scan() {
		switch scanner.Text() {
		case "exit":
			client.Stop()
			os.Exit(0)
		case "o":
			orderid := client.SendOrder("1", "600000.SS", 100, 6.0)
			client.CancelOrder(orderid)
		case "s":
			for i := 0; i < 500; i++ {
				client.SendOrder("1", "600000.SS", 100, 6.0)
			}
		case "b":
			client.SendBasket("1", "input/zz500.csv", 1)
			time.Sleep(time.Second)
			client.CancelAll()
		case "l":
			client.SendOrderList("order_list1")
		}
		time.Sleep(time.Second * 3)
		fmt.Println("cmds: exit, o:order, s:strategy, b:basket, l:orderlist")
	}
}
