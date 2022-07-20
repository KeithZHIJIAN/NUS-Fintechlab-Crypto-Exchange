package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

type OrderBook struct {
	symbol        string
	asks          *OrderTree
	bids          *OrderTree
	pendingOrders []*Order
	marketPrice   decimal.Decimal
}

func NewOrderBook(symbol string) *OrderBook {
	return &OrderBook{
		symbol:        symbol,
		asks:          NewOrderTree(),
		bids:          NewOrderTree(),
		pendingOrders: make([]*Order, 0),
		marketPrice:   decimal.Zero,
	}
}

func (ob *OrderBook) Start() {
	Listen(ob)
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
		fmt.Println("invalid message: ", msg)
		return
	}
	switch strings.ToUpper(msgList[0]) {
	case "ADD":
		ob.Add(msgList)
	default:
		fmt.Println("unknown message: ", msg)
	}
	UpdateAskOrder(ob)
	UpdateBidOrder(ob)
	fmt.Println(ob.asks.UpdateString())
	fmt.Println(ob.bids.UpdateString())
	fmt.Println(ob)
}

func (ob *OrderBook) MarketPrice() decimal.Decimal {
	return ob.marketPrice
}

func (ob *OrderBook) SetMarketPrice(price decimal.Decimal) {
	ob.marketPrice = price
}

func (ob *OrderBook) Add(orderInfo []string) {
	order := ob.ParseOrder(orderInfo)
	if order.Symbol() != ob.symbol {
		fmt.Println("symbol mismatch: ", order.Symbol(), ob.symbol)
		return
	}
	// TODO: add order to db
	ob.AddOrder(order)
}

// Add, Symbol, Type, Side, Quantity, Price, Owner ID, Wallet ID, Stop Price
func (ob *OrderBook) ParseOrder(orderInfo []string) *Order {
	symbol := strings.ToUpper(orderInfo[1])
	isBuy := strings.ToUpper(orderInfo[3]) == "BUY" || strings.ToUpper(orderInfo[3]) == "BID"
	quantity, _ := decimal.NewFromString(orderInfo[4])
	price, _ := decimal.NewFromString(orderInfo[5])
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
		// TODO: close order in db
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
	for _, order := range filledList {
		ot.Remove(&Price{price: order.Price(), isBuy: order.IsBuy()}, order.ID())
		// TODO: close order in db
	}
}

func (ob *OrderBook) CreateTrade(inbound, outbound *Order, filledList *[]*Order) {
	crossPrice := ob.ComputeCrossPrice(inbound, outbound)
	if crossPrice.Equal(decimal.Zero) {
		return
	}

	fillQty := ob.ComputeFillQuantity(inbound, outbound, crossPrice)
	if fillQty.Equal(decimal.Zero) {
		return
	}

	ob.AddTradeRecord(inbound, outbound, crossPrice, fillQty)

	ob.SetMarketPrice(crossPrice)

	if fillQty.Cmp(decimal.Zero) > 0 {
		fmt.Println("order ", outbound.ID(), " & ", inbound.ID(), " filled ", fillQty, " @ $", crossPrice)
	}
	if outbound.Filled() {
		*filledList = append(*filledList, outbound)
	}
}

func (ob *OrderBook) AddTradeRecord(inbound, outbound *Order, crossPrice, fillQty decimal.Decimal) {
	curr := time.Now()
	inbound.Fill(crossPrice, fillQty, curr)
	outbound.Fill(crossPrice, fillQty, curr)
	ob.FillOrderList(outbound, fillQty)
	// TODO: add db trade record, update db order table, update db balance table, update db asset table
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

func (ob *OrderBook) ComputeFillQuantity(inbound, outbound *Order, crossPrice decimal.Decimal) decimal.Decimal {
	fillQty := inbound.OpenQuantity()
	if outbound.OpenQuantity().Cmp(fillQty) < 0 {
		fillQty = outbound.OpenQuantity()
	}
	if fillQty.Cmp(decimal.Zero) > 0 {
		buyer, seller := outbound, inbound
		if inbound.IsBuy() {
			buyer, seller = inbound, outbound
		}

		// Check buyer's balance before trade, if not enough, buy as much as possible
		fillQty = ob.CheckBuyerBalance(buyer, crossPrice, fillQty)

		// Check seller's asset before trade, if not enough, sell as much as possible
		fillQty = ob.CheckSellerAsset(seller, fillQty)

		return fillQty
	}
	return decimal.Zero
}

func (ob *OrderBook) CheckBuyerBalance(buyer *Order, crossPrice, fillQty decimal.Decimal) decimal.Decimal {
	balance := ReadUserBalance(buyer.OwnerId())
	if balance.Cmp(crossPrice.Mul(fillQty)) < 0 {
		fillQty = balance.Div(crossPrice)
		if fillQty.Cmp(decimal.NewFromFloat(0.01)) <= 0 {
			return decimal.Zero
		}
	}
	return fillQty
}

func (ob *OrderBook) CheckSellerAsset(seller *Order, fillQty decimal.Decimal) decimal.Decimal {
	quantity := ReadWalletAsset(seller.WalletId(), seller.Symbol())
	if quantity.Cmp(fillQty) < 0 {
		fillQty = quantity
	}
	return fillQty
}

func (ob *OrderBook) String() string {
	return fmt.Sprintf("\nsymbol:%s\n\nasks:\n%v\nbids:\n%v\n", ob.symbol, ob.asks, ob.bids)
}
