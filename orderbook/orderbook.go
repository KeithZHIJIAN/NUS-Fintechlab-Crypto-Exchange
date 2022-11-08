package orderbook

import (
	"time"

	"github.com/shopspring/decimal"
)

const unit = time.Minute

var UpdateMarketChan = make(chan *MarketInfo)

type MarketInfo struct {
	MarketPrice decimal.Decimal `json:"MarketPrice"`
	High        decimal.Decimal `json:"High"`
	Low         decimal.Decimal `json:"Low"`
	Close       decimal.Decimal `json:"Close"`
	Open        decimal.Decimal `json:"Open"`
	Volume      decimal.Decimal `json:"Volume"`
	Vwap        decimal.Decimal `json:"Vwap"`
	Trades      int             `json:"Trades"`
}

type OrderBook struct {
	symbol        string
	asks          *OrderTree
	bids          *OrderTree
	pendingOrders []*Order
	marketInfo    *MarketInfo
}

func (ob *OrderBook) GetBids() *OrderTree {
	return ob.bids
}

func (ob *OrderBook) GetAsks() *OrderTree {
	return ob.asks
}

func NewMarketInfo() *MarketInfo {
	return &MarketInfo{
		MarketPrice: decimal.Zero,
		High:        decimal.NewFromFloat(100),
		Low:         decimal.NewFromFloat(100),
		Close:       decimal.NewFromFloat(100),
		Open:        decimal.NewFromFloat(100),
		Volume:      decimal.Zero,
		Vwap:        decimal.Zero,
		Trades:      0,
	}
}

func NewOrderBook(symbol string) *OrderBook {
	ob := &OrderBook{
		symbol:        symbol,
		asks:          NewOrderTree(),
		bids:          NewOrderTree(),
		pendingOrders: make([]*Order, 0),
		marketInfo:    NewMarketInfo(),
	}

	return ob
}

func (ob *OrderBook) Symbol() string {
	return ob.symbol
}

func (ob *OrderBook) MarketPrice() decimal.Decimal {
	return ob.marketInfo.MarketPrice
}

func (ob *OrderBook) SetMarketPrice(price decimal.Decimal) {
	ob.marketInfo.MarketPrice = price
}
