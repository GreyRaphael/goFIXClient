package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/quickfixgo/enum"
	"github.com/quickfixgo/field"
	"github.com/quickfixgo/fix42/newordersingle"
	"github.com/quickfixgo/quickfix"
	"github.com/shopspring/decimal"
)

// TradeClient implements the quickfix.Application interface
type TradeClient struct {
	AccountID int64
	SessionID quickfix.SessionID
}

// OnCreate implemented as part of Application interface
func (e *TradeClient) OnCreate(sessionID quickfix.SessionID) {}

// OnLogon implemented as part of Application interface
func (e *TradeClient) OnLogon(sessionID quickfix.SessionID) {
	fmt.Printf("logon, SessionID=%s\n", sessionID)
	e.SessionID = sessionID
}

// OnLogout implemented as part of Application interface
func (e *TradeClient) OnLogout(sessionID quickfix.SessionID) {
	fmt.Println("logout")
}

// FromAdmin implemented as part of Application interface
func (e *TradeClient) FromAdmin(msg *quickfix.Message, sessionID quickfix.SessionID) (reject quickfix.MessageRejectError) {
	return nil
}

// ToAdmin implemented as part of Application interface
func (e *TradeClient) ToAdmin(msg *quickfix.Message, sessionID quickfix.SessionID) {}

// ToApp implemented as part of Application interface
func (e *TradeClient) ToApp(msg *quickfix.Message, sessionID quickfix.SessionID) (err error) {
	// Response := strings.Split(msg.String(), string(rune(1)))
	// fmt.Println(Response)
	// fmt.Println("APP SEND: ", msg)
	fmt.Println("APP SEND: ", strings.ReplaceAll(msg.String(), string(rune(1)), "|"))
	return
}

// FromApp implemented as part of Application interface. This is the callback for all Application level messages from the counter party.
func (e *TradeClient) FromApp(msg *quickfix.Message, sessionID quickfix.SessionID) (reject quickfix.MessageRejectError) {
	// fmt.Printf(fmt.Sprintf("FromApp: %s", msg.String()))
	fmt.Println("APP RECV: ", strings.ReplaceAll(msg.String(), string(rune(1)), "|"))
	return
}

func (e *TradeClient) SendOrder(direction string, secucode string, volume int32, price float64) {
	codeinfo := strings.Split(secucode, ".")

	ClOrdID := field.NewClOrdID(strconv.Itoa(rand.Intn(10000)))
	HandInst := field.NewHandlInst(enum.HandlInst_AUTOMATED_EXECUTION_ORDER_PRIVATE_NO_BROKER_INTERVENTION) // "1"
	Symbol := field.NewSymbol(codeinfo[0])
	Side := field.NewSide(enum.Side(direction)) // 1 buy, 2 sell
	TransactionTime := field.NewTransactTime(time.Now())
	OrdType := field.NewOrdType(enum.OrdType_LIMIT) // "2"

	order := newordersingle.New(ClOrdID, HandInst, Symbol, Side, TransactionTime, OrdType)
	order.SetOrderQty(decimal.NewFromInt32(volume), 2)
	order.SetPrice(decimal.NewFromFloat(price), 2)
	order.SetAccount(strconv.Itoa(int(e.AccountID)))
	order.SetCurrency("CNY")
	order.SetSecurityType(enum.SecurityType_COMMON_STOCK) // "CS"

	order.SetSecurityExchange(codeinfo[1])
	msg := order.ToMessage()
	quickfix.SendToTarget(msg, e.SessionID)
}

func (e *TradeClient) SendBasket(direction string, filename string) {
	data, _ := os.ReadFile(filename)
	var stocks []Stock
	err := json.Unmarshal(data, &stocks)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	for _, stock := range stocks {
		e.SendOrder(direction, stock.Code, stock.Volume, stock.Price)
	}
}

type Stock struct {
	Code   string  `json:"code"`
	Price  float64 `json:"price"`
	Volume int32   `json:"volume"`
}

func main() {
	cfgFile, _ := os.Open("config-test/client.cfg")
	defer cfgFile.Close()
	cfgString, _ := io.ReadAll(cfgFile)
	appSettings, _ := quickfix.ParseSettings(bytes.NewReader(cfgString))
	app := TradeClient{AccountID: 20230823001}
	fileLogFactory, _ := quickfix.NewFileLogFactory(appSettings)
	initiator, _ := quickfix.NewInitiator(&app, quickfix.NewMemoryStoreFactory(), appSettings, fileLogFactory)
	initiator.Start()

	time.Sleep(time.Second)
	fmt.Println("cmd: exit, o:order, s:strategy, b:basket")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		switch scanner.Text() {
		case "exit":
			initiator.Stop()
			os.Exit(0)
		case "o":
			app.SendOrder("1", "600000.SS", 100, 6.0)
		case "s":
			for i := 0; i < 500; i++ {
				app.SendOrder("1", "600000.SS", 100, 6.0)
			}
		case "b":
			app.SendBasket("1", "stocks/zz500.json")
		}
		time.Sleep(time.Second * 3)
		fmt.Println("cmd: exit, o:order, s:strategy")
	}
}
