package realmarket

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/huandu/skiplist"
	"github.com/shopspring/decimal"

	"github.com/gorilla/websocket"
)

const (
	OrderBookDepth          = 4
	OrderBookUpdateInterval = time.Second
)

var SymbolMap = map[string]string{"BTC-USD": "BTCUSD", "ETH-USD": "ETHUSD"}

type WebsocketResponse struct {
	Type      string      `json:"type"`
	ProductID string      `json:"product_id"`
	Asks      [][2]string `json:"asks"`
	Bids      [][2]string `json:"bids"`
	Changes   [][3]string `json:"changes"`
	Time      time.Time   `json:"time"`
	Price     string      `json:"price"`
}

type Candlestick struct {
	Time  time.Time       `json:"time"`
	Open  decimal.Decimal `json:"open"`
	Close decimal.Decimal `json:"close"`
	High  decimal.Decimal `json:"high"`
	Low   decimal.Decimal `json:"low"`
	Vol   decimal.Decimal `json:"volume"`
}

// every second update_candlestick
type CandlestickMsg struct {
	Type        string      `json:"type"`
	Symbol      string      `json:"symbol"`
	Candlestick Candlestick `json:"candlestick"`
}

type OrderBook struct {
	Symbol string      `json:"symbol"`
	Asks   [][2]string `json:"asks"`
	Bids   [][2]string `json:"bids"`
}

var MarketOrderBooks map[string]map[string]*skiplist.SkipList
var LastCandlesticks sync.Map
var CurrCandlesticks sync.Map
var MarketPrices sync.Map

type LiveMarket struct {
	Symbol          string
	Bids            *skiplist.SkipList
	Asks            *skiplist.SkipList
	LastCandlestick Candlestick
	CurrCandlestick Candlestick
}

func OrderBookAgentStart() {
	// "BTC-USD": asks: [[100,0.001],[100.1,0.002]], bids: []
	MarketOrderBooks = make(map[string]map[string]*skiplist.SkipList)
	c, _, err := websocket.DefaultDialer.Dial("wss://ws-feed.exchange.coinbase.com", nil)
	if err != nil {
		panic(fmt.Errorf(err.Error()))
	}
	SubRequest := []byte(`{"type": "subscribe", "product_ids": ["ETH-USD", "BTC-USD"], "channels": ["level2", "heartbeat", { "name": "ticker", "product_ids": ["ETH-USD", "BTC-USD"]}]}`)
	err = c.WriteMessage(websocket.TextMessage, SubRequest)
	if err != nil {
		panic(fmt.Errorf(err.Error()))
	}
	log.Println("OrderBookAgentStart: Order Book websocket connection estbalished")
	defer c.Close()

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println(err.Error())
			continue
		}
		var res WebsocketResponse
		json.Unmarshal(message, &res)
		switch res.Type {
		case "snapshot":
			MarketOrderBooks[SymbolMap[res.ProductID]] = make(map[string]*skiplist.SkipList)
			MarketOrderBooks[SymbolMap[res.ProductID]]["asks"] = parseAsks(res.Asks)
			MarketOrderBooks[SymbolMap[res.ProductID]]["bids"] = parseBids(res.Bids)
		case "l2update":
			for _, change := range res.Changes {
				price := change[1]
				quantity, _ := strconv.ParseFloat(change[2], 64)
				side := "asks"
				if change[0] == "buy" {
					side = "bids"
				}
				if quantity == 0 {
					MarketOrderBooks[SymbolMap[res.ProductID]][side].Remove(price)
				} else {
					MarketOrderBooks[SymbolMap[res.ProductID]][side].Set(price, quantity)
				}
			}
		case "ticker":
			// fmt.Println(res)
			price, _ := decimal.NewFromString(res.Price)
			MarketPrices.Store(SymbolMap[res.ProductID], price)
			if value, ok := CurrCandlesticks.Load(SymbolMap[res.ProductID]); ok {
				candlestick := value.(*Candlestick)
				if candlestick.High.LessThan(price) {
					candlestick.High = price
				} else if price.LessThan(candlestick.Low) {
					candlestick.Low = price
				}
				candlestick.Close = price
				CurrCandlesticks.Store(SymbolMap[res.ProductID], candlestick)
			}
		}
	}
}

func getOrderBook(symbol string) (*OrderBook, bool) {
	//{"symbol":"BTC-USD", "asks": [[price, quantity], ...], "bids": [[price, quantity], ...], "market_price": price}
	ob := &OrderBook{}
	ob.Symbol = symbol
	ob.Asks = make([][2]string, OrderBookDepth)
	askOrder := MarketOrderBooks[symbol]["asks"].Front()
	for i := 0; i < OrderBookDepth; i++ {
		if askOrder == nil {
			return ob, false
		}
		ob.Asks[i] = [2]string{askOrder.Key().(string), fmt.Sprintf("%.6f", askOrder.Value.(float64))}
		askOrder = askOrder.Next()
	}
	ob.Bids = make([][2]string, OrderBookDepth)
	bidOrder := MarketOrderBooks[symbol]["bids"].Front()
	for i := 0; i < OrderBookDepth; i++ {
		if bidOrder == nil {
			return ob, false
		}
		ob.Bids[i] = [2]string{bidOrder.Key().(string), fmt.Sprintf("%.6f", bidOrder.Value.(float64))}
		bidOrder = bidOrder.Next()
	}
	return ob, true
}

func parseAsks(orderbook [][2]string) *skiplist.SkipList {
	res := skiplist.New(skiplist.GreaterThanFunc(func(s1, s2 interface{}) int {
		if s1.(string) == s2.(string) {
			return 0
		}
		f1, err := strconv.ParseFloat(s1.(string), 64)
		f2, err := strconv.ParseFloat(s2.(string), 64)
		if err != nil {
			panic("Error parse price")
		}
		if f1 > f2 {
			return 1
		}
		return -1
	}))

	for _, v := range orderbook {
		price := v[0]
		quantity, _ := strconv.ParseFloat(v[1], 64)
		res.Set(price, quantity)
	}
	return res
}

func parseBids(orderbook [][2]string) *skiplist.SkipList {
	res := skiplist.New(skiplist.GreaterThanFunc(func(s1, s2 interface{}) int {
		if s1.(string) == s2.(string) {
			return 0
		}
		f1, err := strconv.ParseFloat(s1.(string), 64)
		f2, err := strconv.ParseFloat(s2.(string), 64)
		if err != nil {
			panic("Error parse price")
		}
		if f1 < f2 {
			return 1
		}
		return -1
	}))
	for _, v := range orderbook {
		price := v[0]
		quantity, _ := strconv.ParseFloat(v[1], 64)
		res.Set(price, quantity)
	}
	return res
}
