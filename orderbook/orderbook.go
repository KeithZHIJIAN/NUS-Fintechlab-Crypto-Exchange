package orderbook

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/KeithZHIJIAN/nce-matchingengine/utils"

	"github.com/shopspring/decimal"
)

const unit = time.Minute

type OrderBook struct {
	symbol        string
	asks          *OrderTree
	bids          *OrderTree
	pendingOrders []*Order
	marketPrice   decimal.Decimal
	high          decimal.Decimal
	low           decimal.Decimal
	close         decimal.Decimal
	open          decimal.Decimal
	volume        decimal.Decimal
	vwap          decimal.Decimal
	trades        int
}

func NewOrderBook(symbol string) *OrderBook {
	ob := &OrderBook{
		symbol:        symbol,
		asks:          NewOrderTree(),
		bids:          NewOrderTree(),
		pendingOrders: make([]*Order, 0),
		marketPrice:   decimal.Zero,
		high:          decimal.NewFromFloat(100),
		low:           decimal.NewFromFloat(100),
		close:         decimal.NewFromFloat(100),
		open:          decimal.NewFromFloat(100),
		volume:        decimal.Zero,
		vwap:          decimal.Zero,
		trades:        0,
	}

	return ob
}

func (ob *OrderBook) Symbol() string {
	return ob.symbol
}

func (ob *OrderBook) Start() {
	go ob.startMarketHistoryAgent()
	ob.Listen()
}

// Message format:
// -   Add, Symbol, Type, Side, Quantity, Price, Owner ID, Wallet ID, Stop Price (Optional)
//     add ETHUSD limit ask 100 64000 user1 Alice1 (60000)
//     add ethusd market ask 100 0 user1 Alice1 (60000)

// -   Modify, Symbol, Side, Order ID, prev Quantity, prev Price, new Quantity, new Price
//     modify ETHUSD buy 0000000002 100 63000 100 64000 //change price to 64000 only

// -   Cancel, Symbol, Side, Price, Order ID
//     cancel ETHUSD buy 100 0000000001
func (ob *OrderBook) Apply(msg string) {
	msgList := strings.Fields(msg)
	if len(msgList) < 5 {
		log.Println("invalid message: ", msg)
		return
	}
	switch strings.ToUpper(msgList[0]) {
	case "ADD":
		ob.Add(msgList)
	case "MODIFY":
		ob.Modify(msgList)
	case "CANCEL":
		ob.Cancel(msgList)
	default:
		log.Println("unknown message: ", msg)
	}
	ob.UpdateAskOrder()
	ob.UpdateBidOrder()
	log.Println(ob)
}

func (ob *OrderBook) MarketPrice() decimal.Decimal {
	return ob.marketPrice
}

func (ob *OrderBook) SetMarketPrice(price decimal.Decimal) {
	ob.marketPrice = price
}

func (ob *OrderBook) SubmitPendingOrders() {
	for _, order := range ob.pendingOrders {
		ob.AddOrder(order)
	}
	ob.pendingOrders = make([]*Order, 0)
}

func (ob *OrderBook) Add(orderInfo []string) {
	order := ob.ParseOrder(orderInfo)
	if order == nil {
		log.Println("Failed parsing order")
		return
	}
	if order.Symbol() != ob.symbol {
		log.Println("Symbol mismatch: ", order.Symbol(), ob.symbol)
		return
	}
	utils.CreateOpenOrder(order.IsBuy(), order.ID(), order.WalletId(), order.OwnerId(), ob.symbol, order.Price(), order.quantity, time.Now())
	ob.AddOrder(order)
}

// Add, Symbol, Type, Side, Quantity, Price, Owner ID, Wallet ID, Stop Price
func (ob *OrderBook) ParseOrder(orderInfo []string) *Order {
	symbol := strings.ToUpper(orderInfo[1])
	isBuy := strings.ToUpper(orderInfo[3]) == "BUY" || strings.ToUpper(orderInfo[3]) == "BID"
	quantity, err := decimal.NewFromString(orderInfo[4])
	price, err := decimal.NewFromString(orderInfo[5])
	if err != nil {
		return nil
	}
	if strings.ToUpper(orderInfo[2]) == "MARKET" {
		price = decimal.Zero
	}
	ownerId := orderInfo[6]
	walletId := orderInfo[7]
	curr := time.Now()
	return NewOrder(symbol, ownerId, walletId, isBuy, quantity, price, curr, curr)
}

func (ob *OrderBook) AddOrder(inbound *Order) {
	inboundList, outboundList := ob.asks, ob.bids
	if inbound.IsBuy() {
		inboundList, outboundList = ob.bids, ob.asks
	}

	ob.MatchOrder(inbound, outboundList)

	p := &Price{price: inbound.Price(), isBuy: inbound.IsBuy()}

	if inbound.Filled() {
		utils.DeleteOpenOrder(inbound.IsBuy(), ob.symbol, inbound.ID())
		utils.CreateClosedOrder(inbound.IsBuy(), ob.symbol, inbound.ID(), inbound.WalletId(), inbound.OwnerId(), inbound.Quantity(), inbound.Price(), inbound.FillCost(), inbound.CreateTime(), time.Now())
	} else {
		inboundList.Add(p, inbound)
	}
}

func (ob *OrderBook) MatchOrder(inbound *Order, ot *OrderTree) {
	ob.MatchRegularOrder(inbound, ot)
}

func (ob *OrderBook) MatchRegularOrder(inbound *Order, ot *OrderTree) {
	inboudPrice := inbound.Price()
	filledList := make([]*Order, 0)
	otIter := ot.Iterator()
	for otIter.Next() {
		price := otIter.Key().(*Price)
		orderList := otIter.Value().(*OrderList)
		if price.Match(inboudPrice) {
			olIter := orderList.Iterator()
			for olIter.Next() {
				outbound := olIter.Value().(*Order)
				ob.CreateTrade(inbound, outbound, &filledList)
			}
		}
	}
	curr := time.Now()
	for _, order := range filledList {
		ot.Remove(&Price{price: order.Price(), isBuy: order.IsBuy()}, order.ID())
		utils.DeleteOpenOrder(order.IsBuy(), ob.symbol, order.ID())
		utils.CreateClosedOrder(order.IsBuy(), ob.symbol, order.ID(), order.WalletId(), order.OwnerId(), order.Quantity(), order.Price(), order.FillCost(), order.CreateTime(), curr)
	}
}

func (ob *OrderBook) CreateTrade(inbound, outbound *Order, filledList *[]*Order) {
	crossPrice := ob.ComputeCrossPrice(inbound, outbound)
	if crossPrice.Equal(decimal.Zero) {
		return
	}

	buyer, seller := outbound, inbound
	if inbound.IsBuy() {
		buyer, seller = inbound, outbound
	}

	fillQty := ob.ComputeFillQuantity(buyer, seller, crossPrice)
	if fillQty.Equal(decimal.Zero) {
		return
	}

	ob.AddTradeRecord(buyer, seller, outbound, crossPrice, fillQty)

	ob.SetMarketPrice(crossPrice)

	if fillQty.GreaterThan(decimal.Zero) {
		log.Println("order ", outbound.ID(), " & ", inbound.ID(), " filled ", fillQty, " @ $", crossPrice)
	}
	if outbound.Filled() {
		*filledList = append(*filledList, outbound)
	}
}

func (ob *OrderBook) AddTradeRecord(buyer, seller, outbound *Order, crossPrice, fillQty decimal.Decimal) {
	curr := time.Now()
	buyer.Fill(crossPrice, fillQty, curr)
	seller.Fill(crossPrice, fillQty, curr)
	ob.FillOrderList(outbound, fillQty)
	ob.UpdateMarket(crossPrice, fillQty)
	// UpdateOrderFilled(ob.OrderFilledString(curr, crossPrice, fillQty))
	utils.CreateTradeRecord(buyer.Symbol(), buyer.ID(), seller.ID(), crossPrice, fillQty, curr)
	utils.UpdateOrder(buyer.Symbol(), buyer.ID(), buyer.IsBuy(), buyer.Price(), buyer.Quantity(), buyer.OpenQuantity(), curr)
	utils.UpdateOrder(seller.Symbol(), seller.ID(), seller.IsBuy(), seller.Price(), seller.Quantity(), seller.OpenQuantity(), curr)
	utils.UpdateUserBalance(buyer.OwnerId(), seller.OwnerId(), crossPrice.Mul(fillQty))
	utils.UpdateWalletAsset(buyer.Symbol(), buyer.WalletId(), seller.WalletId(), fillQty)
}

func (ob *OrderBook) FillOrderList(outbound *Order, fillQty decimal.Decimal) {
	ot := ob.asks
	isBuy := false
	if outbound.IsBuy() {
		ot = ob.bids
		isBuy = true
	}
	if ol, ok := ot.Get(&Price{outbound.Price(), isBuy}); ok {
		ol.Fill(fillQty)
	}
}

func (ob *OrderBook) ComputeCrossPrice(inbound, outbound *Order) decimal.Decimal {
	crossPrice := outbound.Price()
	inboundPrice := inbound.Price()

	if crossPrice.Equal(decimal.Zero) {
		crossPrice = inboundPrice
	}

	if crossPrice.Equal(decimal.Zero) {
		crossPrice = ob.marketPrice
	}

	return crossPrice
}

func (ob *OrderBook) ComputeFillQuantity(buyer, seller *Order, crossPrice decimal.Decimal) decimal.Decimal {
	fillQty := buyer.OpenQuantity()
	if seller.OpenQuantity().LessThan(fillQty) {
		fillQty = seller.OpenQuantity()
	}
	if fillQty.GreaterThan(decimal.Zero) {
		// Check buyer's balance before trade, if not enough, buy as much as possible
		fillQty = ob.CheckBuyerBalance(buyer, crossPrice, fillQty)
		// Check seller's asset before trade, if not enough, sell as much as possible
		fillQty = ob.CheckSellerAsset(seller, fillQty)
		return fillQty
	}
	return decimal.Zero
}

func (ob *OrderBook) CheckBuyerBalance(buyer *Order, crossPrice, fillQty decimal.Decimal) decimal.Decimal {
	balance := utils.ReadUserBalance(buyer.OwnerId())
	if balance.LessThan(crossPrice.Mul(fillQty)) {
		fillQty = balance.Div(crossPrice)
		if fillQty.LessThanOrEqual(decimal.NewFromFloat(0.01)) {
			return decimal.Zero
		}
	}
	return fillQty
}

func (ob *OrderBook) CheckSellerAsset(seller *Order, fillQty decimal.Decimal) decimal.Decimal {
	quantity := utils.ReadWalletAsset(seller.WalletId(), seller.Symbol())
	if quantity.LessThan(fillQty) {
		fillQty = quantity
	}
	return fillQty
}

func (ob *OrderBook) String() string {
	return fmt.Sprintf("\nsymbol:%s\n\nasks:\n%v\nbids:\n%v\n", ob.symbol, ob.asks, ob.bids)
}

func (ob *OrderBook) Get(isBuy bool, price decimal.Decimal, orderId string) (*Order, bool) {
	ot := ob.asks
	if isBuy {
		ot = ob.bids
	}
	if ol, ok := ot.Get(&Price{price, isBuy}); ok {
		if o, ok := ol.Get(orderId); ok {
			return o, true
		}
	}
	return nil, false
}

// Modify, Symbol, Side, Order ID, prev Quantity, prev Price, new Quantity, new Price
func (ob *OrderBook) Modify(orderInfo []string) {
	symbol := strings.ToUpper(orderInfo[1])
	if symbol != ob.symbol {
		log.Println("Symbol mismatch:", symbol, ob.symbol)
		return
	}
	curr := time.Now()
	isBuy := strings.ToUpper(orderInfo[2]) == "BUY" || strings.ToUpper(orderInfo[2]) == "BID"
	orderId := orderInfo[3]
	prevPrice, err := decimal.NewFromString(orderInfo[5])
	newQuantity, err := decimal.NewFromString(orderInfo[6])
	newPrice, err := decimal.NewFromString(orderInfo[7])
	if err != nil {
		log.Println("Failed parsing order price/quantity:", err)
		return
	}
	if prevPrice.Equal(decimal.Zero) && !newPrice.Equal(decimal.Zero) {
		log.Println("Cannot change market order price")
		return
	}
	if order, ok := ob.Get(isBuy, prevPrice, orderId); ok {
		order.ModifyPrice(newPrice, curr)
		order.ModifyQuantity(newQuantity, curr)
		ot := ob.asks
		if isBuy {
			ot = ob.bids
		}
		ot.Remove(&Price{prevPrice, isBuy}, orderId)
		utils.UpdateOrder(symbol, orderId, isBuy, newPrice, newQuantity, order.OpenQuantity(), curr)
		ob.AddOrder(order)
		for len(ob.pendingOrders) > 0 {
			ob.SubmitPendingOrders()
		}
	} else {
		log.Printf("Orderbook: order not found")
	}
}

// Cancel, Symbol, Side, Price, Order ID
func (ob *OrderBook) Cancel(orderInfo []string) {
	symbol := strings.ToUpper(orderInfo[1])
	if symbol != ob.symbol {
		log.Println("Symbol mismatch:", symbol, ob.symbol)
		return
	}
	isBuy := strings.ToUpper(orderInfo[2]) == "BUY" || strings.ToUpper(orderInfo[2]) == "BID"
	price, err := decimal.NewFromString(orderInfo[4])
	if err != nil {
		log.Println("Failed parsing order price:", err)
		return
	}
	orderId := orderInfo[3]
	if order, ok := ob.Get(isBuy, price, orderId); ok {
		sizeDelta := order.OpenQuantity()
		order.ModifyQuantity(order.Quantity().Sub(sizeDelta), time.Now())
		ot := ob.asks
		if isBuy {
			ot = ob.bids
		}
		ot.Remove(&Price{price, isBuy}, orderId)
		ob.AddOrder(order)
	} else {
		log.Printf("Orderbook: order not found")
	}
}

func (ob *OrderBook) startMarketHistoryAgent() {
	currentTime := time.Now()
	startTime := currentTime.Truncate(unit).Add(unit)

	duration := startTime.Sub(currentTime)
	time.Sleep(duration)

	ticker := time.NewTicker(1 * unit)
	for {
		select {
		case curr := <-ticker.C:
			// do stuff
			ob.InitCandlestick()
			ob.UpdateMarketSubscription()
			utils.CreateMarketHistory(ob.symbol, curr, ob.high, ob.low, ob.open, ob.close, ob.volume, ob.vwap, ob.trades)
		}
	}
}

func (ob *OrderBook) UpdateMarket(price, qty decimal.Decimal) {
	//vwap calculation
	if qty.Equal(decimal.Zero) {
		ob.vwap = decimal.Zero
	} else {
		ob.vwap = (ob.vwap.Mul(ob.volume).Add(price.Mul(qty))).Div(ob.volume.Add(qty))
	}
	if qty.GreaterThan(decimal.Zero) {
		ob.trades++
	}
	ob.close = price
	ob.volume = ob.volume.Add(qty)
	if price.GreaterThan(ob.high) {
		ob.high = price
	} else if price.LessThan(ob.low) {
		ob.low = price
	}
	ob.UpdateMarketSubscription()

	utils.UpdateMarketHistory(ob.symbol, time.Now().Format("2006-01-02 15:04"), ob.high, ob.low, ob.open, ob.close, ob.volume, ob.vwap, ob.trades)
}

func (ob *OrderBook) InitCandlestick() {
	ob.open = ob.close
	ob.high = ob.close
	ob.low = ob.close
	ob.volume = decimal.Zero
	ob.vwap = decimal.Zero
	ob.trades = 0
}

func (ob *OrderBook) UpdateMarketString() string {
	return fmt.Sprintf("{\"updateMarketHistory\":{ \"time\":\"%v\", \"open\":%s, \"high\":%s, \"low\":%s, \"close\":%s, \"volume\":%s }}", time.Now().UnixMilli(), ob.open, ob.high, ob.low, ob.close, ob.volume)
}

func (ob *OrderBook) Listen() {
	msgs := utils.RabbitmqConsume(ob.symbol)
	var forever chan struct{}

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
			ob.Apply(string(d.Body[:]))
		}
	}()

	log.Printf(" [*] %s waiting for messages. To exit press CTRL+C", ob.symbol)
	<-forever
}

var prevAsks = ""

func (ob *OrderBook) UpdateAskOrder() {
	str := ob.asks.UpdateString()
	if prevAsks == str {
		return
	}
	exchange := fmt.Sprintf("UPDATE_ASK_ORDER_%s.DLQ.Exchange", ob.symbol)
	utils.RabbitmqSend(exchange, str)
	prevAsks = str
}

var prevBids = ""

func (ob *OrderBook) UpdateBidOrder() {
	str := ob.bids.UpdateString()
	if prevBids == str {
		return
	}
	exchange := fmt.Sprintf("UPDATE_BID_ORDER_%s.DLQ.Exchange", ob.symbol)
	utils.RabbitmqSend(exchange, str)
	prevBids = str
}

func (ob *OrderBook) UpdateMarketSubscription() {
	msg := ob.UpdateMarketString()
	exchange := fmt.Sprintf("MARKET_HISTORY_%s.DLQ.Exchange", ob.symbol)
	utils.RabbitmqSend(exchange, msg)
	fmt.Println("Message sent: ", msg)
}
