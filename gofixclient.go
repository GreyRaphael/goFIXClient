package main

import (
	"bytes"
	"fmt"
	"gofix/stock_utils"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/pelletier/go-toml/v2"
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
	ConfigFilename   string
	initiator        *quickfix.Initiator
	isLogon          chan bool
	orderContainer   map[string]bool
	requestId        int32
	logonSeqNum      int
	senderAccountMap map[string]string
	sessAccountMap   map[quickfix.SessionID]string
}

// OnCreate implemented as part of Application interface
func (e *TradeClient) OnCreate(sessionID quickfix.SessionID) {
	e.orderContainer = make(map[string]bool)
	e.sessAccountMap = make(map[quickfix.SessionID]string)
}

// OnLogon implemented as part of Application interface
func (e *TradeClient) OnLogon(sessionID quickfix.SessionID) {
	fmt.Printf("logon, SessionID=%s\n", sessionID)

	for senderId, accountId := range e.senderAccountMap {
		if strings.Contains(sessionID.String(), senderId) {
			e.sessAccountMap[sessionID] = accountId
		}
	}
	e.isLogon <- true
}

// OnLogout implemented as part of Application interface
func (e *TradeClient) OnLogout(sessionID quickfix.SessionID) {
	fmt.Printf("logout, SessionID=%s\n", sessionID)
}

// FromAdmin implemented as part of Application interface
func (e *TradeClient) FromAdmin(msg *quickfix.Message, sessionID quickfix.SessionID) (reject quickfix.MessageRejectError) {
	var msgType field.MsgTypeField
	msg.Header.Get(&msgType)
	if msgType.String() == "A" { // "A" is logon
		var logonMsgSeqNum field.MsgSeqNumField
		msg.Header.Get(&logonMsgSeqNum)
		e.logonSeqNum = logonMsgSeqNum.Int()
	}
	return nil
}

// ToAdmin implemented as part of Application interface
func (e *TradeClient) ToAdmin(msg *quickfix.Message, sessionID quickfix.SessionID) {}

// ToApp implemented as part of Application interface
func (e *TradeClient) ToApp(msg *quickfix.Message, sessionID quickfix.SessionID) (err error) {
	var msgType field.MsgTypeField
	msg.Header.Get(&msgType)
	if msgType.String() == "D" {
		fmt.Println("APP ORDER:", strings.ReplaceAll(msg.String(), string(rune(1)), "|"))

		var order_id field.ClOrdIDField
		msg.Body.Get(&order_id)
		e.orderContainer[order_id.String()] = true
	} else if msgType.String() == "F" {
		fmt.Println("APP CANCEL:", strings.ReplaceAll(msg.String(), string(rune(1)), "|"))

		var orig_order_id field.OrigClOrdIDField
		msg.Body.Get(&orig_order_id)
		delete(e.orderContainer, orig_order_id.String())
	} else {
		fmt.Println("APP SEND:", strings.ReplaceAll(msg.String(), string(rune(1)), "|"))
	}
	// fmt.Println("===>available order ids:", e.orderContainer)
	return
}

// FromApp implemented as part of Application interface. This is the callback for all Application level messages from the counter party.
func (e *TradeClient) FromApp(msg *quickfix.Message, sessionID quickfix.SessionID) (reject quickfix.MessageRejectError) {
	fmt.Println("APP RECV: ", strings.ReplaceAll(msg.String(), string(rune(1)), "|"))
	return
}

func (e *TradeClient) SendOrder(direction string, secucode string, volume int32, price float64, orderType string) {
	e.requestId++
	codeinfo := strings.Split(secucode, ".")
	orderid := fmt.Sprintf("%d.%d", e.logonSeqNum, e.requestId)

	ClOrdID := field.NewClOrdID(orderid)                      // time as orderid
	HandInst := field.NewHandlInst(enum.HandlInst(orderType)) // "1":DMA; "2":DMA2; "3":CARE
	Symbol := field.NewSymbol(codeinfo[0])
	Side := field.NewSide(enum.Side(direction)) // 1 buy, 2 sell
	TransactionTime := field.NewTransactTime(time.Now())
	OrdType := field.NewOrdType(enum.OrdType_LIMIT) // "2"

	order := newordersingle.New(ClOrdID, HandInst, Symbol, Side, TransactionTime, OrdType)
	order.SetOrderQty(decimal.NewFromInt32(volume), 0)
	order.SetPrice(decimal.NewFromFloat(price), 2) // scale小数点后2位
	order.SetCurrency("CNY")
	order.SetSecurityType(enum.SecurityType_COMMON_STOCK) // "CS"

	order.SetSecurityExchange(codeinfo[1])

	for sessId, accountId := range e.sessAccountMap {
		order.SetAccount(accountId)
		msg := order.ToMessage()
		quickfix.SendToTarget(msg, sessId)
	}
}

type DSAConfig struct {
	Name           string
	Duration       int
	MaxMarketShare float64
	TradeStyle     int
	PriceType      int
	Unit           int
	Change         float64
}

func (e *TradeClient) SendDSA(direction string, secucode string, volume int32, price float64, orderType string) {
	e.requestId++
	codeinfo := strings.Split(secucode, ".")
	orderid := fmt.Sprintf("%d.%d", e.logonSeqNum, e.requestId)

	ClOrdID := field.NewClOrdID(orderid)                                       // time as orderid
	HandInst := field.NewHandlInst(enum.HandlInst_MANUAL_ORDER_BEST_EXECUTION) // use MANUAL order
	Symbol := field.NewSymbol(codeinfo[0])
	Side := field.NewSide(enum.Side(direction)) // 1 buy, 2 sell
	TransactionTime := field.NewTransactTime(time.Now())
	OrdType := field.NewOrdType(enum.OrdType_LIMIT) // "2"

	order := newordersingle.New(ClOrdID, HandInst, Symbol, Side, TransactionTime, OrdType)
	order.SetOrderQty(decimal.NewFromInt32(volume), 0)
	order.SetPrice(decimal.NewFromFloat(price), 2) // scale小数点后2位
	order.SetCurrency("CNY")
	order.SetSecurityType(enum.SecurityType_COMMON_STOCK) // "CS"

	order.SetSecurityExchange(codeinfo[1])

	// parse algo config file
	algoBytes, _ := os.ReadFile("input/dsa.toml")
	var algoCfg DSAConfig
	err := toml.Unmarshal(algoBytes, &algoCfg)
	if err != nil {
		panic(err)
	}
	// algo parameters
	order.SetField(6061, quickfix.FIXString(algoCfg.Name))
	precision := quickfix.TimestampPrecision(time.Second)
	order.SetField(6062, quickfix.FIXUTCTimestamp{Time: time.Now(), Precision: precision})
	order.SetField(6063, quickfix.FIXUTCTimestamp{Time: time.Now().Add(time.Minute * time.Duration(algoCfg.Duration)), Precision: precision})
	order.SetField(6064, quickfix.FIXFloat(algoCfg.MaxMarketShare))
	order.SetField(6065, quickfix.FIXInt(algoCfg.TradeStyle))
	order.SetField(6302, quickfix.FIXInt(algoCfg.PriceType))
	order.SetField(6303, quickfix.FIXInt(algoCfg.Unit))
	order.SetField(6304, quickfix.FIXFloat(algoCfg.Change))

	for sessId, accountId := range e.sessAccountMap {
		order.SetAccount(accountId)
		msg := order.ToMessage()
		quickfix.SendToTarget(msg, sessId)
	}
}

func (e *TradeClient) SendOrderList() {
	// useless function
	list_id := time.Now().Format("150412.999999")
	orders := neworderlist.New(field.NewListID(list_id), field.NewBidType(enum.BidType_NO_BIDDING_PROCESS), field.NewTotNoOrders(10))

	gp := neworderlist.NewNoOrdersRepeatingGroup()
	for i := 0; i < 10; i++ {
		noorders := gp.Add()
		e.requestId++
		orderid := fmt.Sprintf("%d.%d", e.logonSeqNum, e.requestId)

		noorders.SetClOrdID(orderid)
		noorders.SetHandlInst(enum.HandlInst_AUTOMATED_EXECUTION_ORDER_PRIVATE_NO_BROKER_INTERVENTION)
		noorders.SetSymbol("688009")
		noorders.SetSide(enum.Side("1"))
		noorders.SetTransactTime(time.Now())
		noorders.SetOrdType(enum.OrdType_LIMIT)
		noorders.SetOrderQty(decimal.NewFromInt32(100), 0)
		noorders.SetPrice(decimal.NewFromFloat(100.12), 2)
		// noorders.SetAccount("xxxxxx")
		noorders.SetCurrency("CNY")
		noorders.SetSecurityType(enum.SecurityType_COMMON_STOCK)
		noorders.SetSecurityExchange("SS")
	}
	orders.SetNoOrders(gp)
	msg := orders.ToMessage()

	for sessId := range e.sessAccountMap {
		quickfix.SendToTarget(msg, sessId)
	}
}

func (e *TradeClient) SendBasket(direction string, filename string, batch_size int, orderType string) {
	stocks := stock_utils.ReadCsv(filename, ',')
	for i := 0; i < batch_size; i++ {
		for _, stock := range stocks {
			if orderType != "4" {
				e.SendOrder(direction, stock.Code, stock.Vol, stock.Price, orderType)
			} else {
				e.SendDSA(direction, stock.Code, stock.Vol, stock.Price, orderType)
			}
		}
	}
}

func (e *TradeClient) CancelOrder(origid string) {
	e.requestId++
	origclordid := field.NewOrigClOrdID(origid)
	orderid := fmt.Sprintf("%d.%d", e.logonSeqNum, e.requestId)
	clordid := field.NewClOrdID(orderid)
	cancel_req := ordercancelrequest.New(origclordid, clordid, field.NewSymbol("000001"), field.NewSide(enum.Side_BUY), field.NewTransactTime(time.Now()))

	cancel_req.SetSecurityExchange("SS")
	cancel_req.SetOrderQty(decimal.NewFromInt32(100), 0)
	// // useless tags
	// cancel_req.SetField(quickfix.Tag(40), quickfix.FIXString("2")) // OrdType is "2"
	// cancel_req.SetField(quickfix.Tag(44), quickfix.FIXFloat(100.12)) // Price is "2"
	// cancel_req.SetAccount("xxxxxx")

	for sessId, accountId := range e.sessAccountMap {
		cancel_req.SetAccount(accountId)
		msg := cancel_req.ToMessage()
		quickfix.SendToTarget(msg, sessId)
	}
}

func (e *TradeClient) CancelAll() {
	for orderid := range e.orderContainer {
		e.CancelOrder(orderid)
	}
}

func (e *TradeClient) Start() {
	// read .cfg file
	conf_bytes, _ := os.ReadFile(e.ConfigFilename)

	// find all AccountID
	accountRegex := regexp.MustCompile(`AccountID=(\d+)`)
	accountParts := accountRegex.FindAllSubmatch(conf_bytes, -1)
	// find all SenderCompID
	senderRegex := regexp.MustCompile(`SenderCompID=(\w+)`)
	senderParts := senderRegex.FindAllSubmatch(conf_bytes, -1)
	e.senderAccountMap = make(map[string]string)
	for i, v := range senderParts {
		SenderCompID := string(v[1])
		AccountID := string(accountParts[i][1])
		e.senderAccountMap[SenderCompID] = AccountID
	}
	// fmt.Println(e.senderAccountMap)

	// init session number
	sessionNum := len(senderParts)
	e.isLogon = make(chan bool, sessionNum)

	// init settings, log_factory
	settings, _ := quickfix.ParseSettings(bytes.NewReader(conf_bytes))
	log_factory, _ := quickfix.NewFileLogFactory(settings)

	// init initiator
	e.initiator, _ = quickfix.NewInitiator(e, quickfix.NewMemoryStoreFactory(), settings, log_factory)
	e.initiator.Start()

	// wait all sesssion logon
	for i := 0; i < sessionNum; i++ {
		<-e.isLogon
	}
	close(e.isLogon)
}

func (e *TradeClient) SendAlgo(direction string, filename string, batchSize int, orderType string) {
	stocks := stock_utils.ReadCsv(filename, ',')
	for i := 0; i < batchSize; i++ {
		for _, stock := range stocks {
			e.SendOrder(direction, stock.Code, stock.Vol, stock.Price, orderType)
			time.Sleep(1 * time.Second)
		}
		e.CancelAll()
		time.Sleep(3 * time.Second)
	}
}

func (e *TradeClient) Stop() {
	e.initiator.Stop()
}
