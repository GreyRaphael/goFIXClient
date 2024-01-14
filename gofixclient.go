package main

import (
	"bytes"
	"fmt"
	"gofix/stock_utils"
	"io"
	"os"
	"regexp"
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
	ConfigFilename string
	account_id     string
	session_id     quickfix.SessionID
	initiator      *quickfix.Initiator
	is_logon       chan bool
}

// OnCreate implemented as part of Application interface
func (e *TradeClient) OnCreate(sessionID quickfix.SessionID) {
	e.is_logon = make(chan bool, 1)
}

// OnLogon implemented as part of Application interface
func (e *TradeClient) OnLogon(sessionID quickfix.SessionID) {
	fmt.Printf("logon, SessionID=%s\n", sessionID)
	e.session_id = sessionID
	e.is_logon <- true
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
	fmt.Println("APP SEND: ", strings.ReplaceAll(msg.String(), string(rune(1)), "|"))
	return
}

// FromApp implemented as part of Application interface. This is the callback for all Application level messages from the counter party.
func (e *TradeClient) FromApp(msg *quickfix.Message, sessionID quickfix.SessionID) (reject quickfix.MessageRejectError) {
	fmt.Println("APP RECV: ", strings.ReplaceAll(msg.String(), string(rune(1)), "|"))
	return
}

func (e *TradeClient) SendOrder(direction string, secucode string, volume int32, price float64) {
	codeinfo := strings.Split(secucode, ".")
	now := time.Now()

	ClOrdID := field.NewClOrdID(now.Format("235959.999999"))                                                // time as orderid
	HandInst := field.NewHandlInst(enum.HandlInst_AUTOMATED_EXECUTION_ORDER_PRIVATE_NO_BROKER_INTERVENTION) // "1"
	Symbol := field.NewSymbol(codeinfo[0])
	Side := field.NewSide(enum.Side(direction)) // 1 buy, 2 sell
	TransactionTime := field.NewTransactTime(now)
	OrdType := field.NewOrdType(enum.OrdType_LIMIT) // "2"

	order := newordersingle.New(ClOrdID, HandInst, Symbol, Side, TransactionTime, OrdType)
	order.SetOrderQty(decimal.NewFromInt32(volume), 0)
	order.SetPrice(decimal.NewFromFloat(price), 2) // scale小数点后2位
	order.SetAccount(e.account_id)
	order.SetCurrency("CNY")
	order.SetSecurityType(enum.SecurityType_COMMON_STOCK) // "CS"

	order.SetSecurityExchange(codeinfo[1])
	msg := order.ToMessage()
	quickfix.SendToTarget(msg, e.session_id)
}

func (e *TradeClient) SendBasket(direction string, filename string, batch_size int) {
	stocks := stock_utils.ReadCsv(filename, ',')
	for i := 0; i < batch_size; i++ {
		for _, stock := range stocks {
			e.SendOrder(direction, stock.Code, stock.Vol, stock.Price)
		}
	}
}

func (e *TradeClient) Start() {
	// read .cfg file
	conf_file, err := os.Open(e.ConfigFilename)
	if err != nil {
		panic(err)
	}
	defer conf_file.Close()

	// init settings, log_factory
	conf_bytes, _ := io.ReadAll(conf_file)

	reg_expr := regexp.MustCompile(`AccountID=(.*)`)
	parts := reg_expr.FindSubmatch(conf_bytes)
	e.account_id = string(parts[1])

	settings, _ := quickfix.ParseSettings(bytes.NewReader(conf_bytes))
	log_factory, _ := quickfix.NewFileLogFactory(settings)

	// init initiator
	e.initiator, _ = quickfix.NewInitiator(e, quickfix.NewMemoryStoreFactory(), settings, log_factory)
	e.initiator.Start()
	<-e.is_logon
}

func (e *TradeClient) Stop() {
	e.initiator.Stop()
}
