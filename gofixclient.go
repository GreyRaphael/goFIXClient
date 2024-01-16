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
	"github.com/quickfixgo/fix42/neworderlist"
	"github.com/quickfixgo/fix42/newordersingle"
	"github.com/quickfixgo/fix42/ordercancelrequest"
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
	order_sets     map[string]bool
	order_counter  int32
	msg_seq_num    int
}

// OnCreate implemented as part of Application interface
func (e *TradeClient) OnCreate(sessionID quickfix.SessionID) {
	e.is_logon = make(chan bool, 1)
	e.order_sets = make(map[string]bool)
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
	var msg_type field.MsgTypeField
	msg.Header.Get(&msg_type)
	if msg_type.String() == "A" { // "A" is logon
		var msg_seq_num field.MsgSeqNumField
		msg.Header.Get(&msg_seq_num)
		e.msg_seq_num = msg_seq_num.Int()
	}
	return nil
}

// ToAdmin implemented as part of Application interface
func (e *TradeClient) ToAdmin(msg *quickfix.Message, sessionID quickfix.SessionID) {}

// ToApp implemented as part of Application interface
func (e *TradeClient) ToApp(msg *quickfix.Message, sessionID quickfix.SessionID) (err error) {
	var msg_type field.MsgTypeField
	msg.Header.Get(&msg_type)
	if msg_type.String() == "D" {
		fmt.Println("APP SEND:", strings.ReplaceAll(msg.String(), string(rune(1)), "|"))

		var order_id field.ClOrdIDField
		msg.Body.Get(&order_id)
		e.order_sets[order_id.String()] = true
	} else if msg_type.String() == "F" {
		fmt.Println("APP CANCEL:", strings.ReplaceAll(msg.String(), string(rune(1)), "|"))

		var orig_order_id field.OrigClOrdIDField
		msg.Body.Get(&orig_order_id)
		delete(e.order_sets, orig_order_id.String())
	} else {
		fmt.Println("APP SEND:", strings.ReplaceAll(msg.String(), string(rune(1)), "|"))
	}
	// fmt.Println("===>available order ids:", e.order_sets)
	return
}

// FromApp implemented as part of Application interface. This is the callback for all Application level messages from the counter party.
func (e *TradeClient) FromApp(msg *quickfix.Message, sessionID quickfix.SessionID) (reject quickfix.MessageRejectError) {
	fmt.Println("APP RECV: ", strings.ReplaceAll(msg.String(), string(rune(1)), "|"))
	return
}

func (e *TradeClient) SendOrder(direction string, secucode string, volume int32, price float64) string {
	e.order_counter++
	codeinfo := strings.Split(secucode, ".")
	orderid := fmt.Sprintf("%d.%d", e.msg_seq_num, e.order_counter)

	ClOrdID := field.NewClOrdID(orderid)                                                                    // time as orderid
	HandInst := field.NewHandlInst(enum.HandlInst_AUTOMATED_EXECUTION_ORDER_PRIVATE_NO_BROKER_INTERVENTION) // "1"
	Symbol := field.NewSymbol(codeinfo[0])
	Side := field.NewSide(enum.Side(direction)) // 1 buy, 2 sell
	TransactionTime := field.NewTransactTime(time.Now())
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
	return orderid
}

func (e *TradeClient) SendOrderList() {
	list_id := time.Now().Format("150412.999999")
	orders := neworderlist.New(field.NewListID(list_id), field.NewBidType(enum.BidType_NO_BIDDING_PROCESS), field.NewTotNoOrders(10))

	gp := neworderlist.NewNoOrdersRepeatingGroup()
	for i := 0; i < 10; i++ {
		noorders := gp.Add()
		e.order_counter++
		orderid := fmt.Sprintf("%d.%d", e.msg_seq_num, e.order_counter)

		noorders.SetClOrdID(orderid)
		noorders.SetHandlInst(enum.HandlInst_AUTOMATED_EXECUTION_ORDER_PRIVATE_NO_BROKER_INTERVENTION)
		noorders.SetSymbol("688009")
		noorders.SetSide(enum.Side("1"))
		noorders.SetTransactTime(time.Now())
		noorders.SetOrdType(enum.OrdType_LIMIT)
		noorders.SetOrderQty(decimal.NewFromInt32(100), 0)
		noorders.SetPrice(decimal.NewFromFloat(100.12), 2)
		noorders.SetAccount(e.account_id)
		noorders.SetCurrency("CNY")
		noorders.SetSecurityType(enum.SecurityType_COMMON_STOCK)
		noorders.SetSecurityExchange("SS")
	}
	orders.SetNoOrders(gp)
	msg := orders.ToMessage()
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

func (e *TradeClient) CancelOrder(origid string) {
	e.order_counter++
	origclordid := field.NewOrigClOrdID(origid)
	orderid := fmt.Sprintf("%d.%d", e.msg_seq_num, e.order_counter)
	clordid := field.NewClOrdID(orderid)
	cancel_req := ordercancelrequest.New(origclordid, clordid, field.NewSymbol("000001"), field.NewSide(enum.Side_BUY), field.NewTransactTime(time.Now()))

	cancel_req.SetSecurityExchange("SS")
	cancel_req.SetOrderQty(decimal.NewFromInt32(100), 0)
	// // useless tags
	// cancel_req.SetField(quickfix.Tag(40), quickfix.FIXString("2")) // OrdType is "2"
	// cancel_req.SetField(quickfix.Tag(44), quickfix.FIXFloat(100.12)) // Price is "2"
	// cancel_req.SetAccount(e.account_id)

	msg := cancel_req.ToMessage()
	quickfix.SendToTarget(msg, e.session_id)
}

func (e *TradeClient) CancelAll() {
	for orderid := range e.order_sets {
		e.CancelOrder(orderid)
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
