package main

import (
	"fmt"
	"strings"
	"sync"
)

var once sync.Once

// type global
type MatchingEngine struct {
	orderbooks map[string]*OrderBook
}

var (
	instance MatchingEngine
)

func NewMatchingEngine() MatchingEngine {

	once.Do(func() { // <-- atomic, does not allow repeating

		instance = MatchingEngine{
			orderbooks: make(map[string]*OrderBook),
		} // <-- thread safe

	})

	return instance
}

func (me *MatchingEngine) Start() {
	Listen(me)
}

func (me *MatchingEngine) Apply(msg string) {
	msgList := strings.Fields(msg)
	if strings.ToUpper(msgList[0]) == "ADD" {
		me.DoAdd(msgList)
	}
}

// Message format:
// -   Add, Symbol, Type, Side, Quantity, Price, Owner ID, Wallet ID, Stop Price (Optional)
//     add ETHUSD limit ask 100 64000 user1 Alice1 (60000)
//     add ethusd market ask 100 0 user1 Alice1 (60000)

// -   Modify, Symbol, Side, Order ID, prev Quantity, prev Price, new Quantity, new Price
//     modify ETHUSD buy 0000000002 100 63000 100 64000 //change price to 64000 only

// -   Cancel, Symbol, Side, Price, Order ID
//     cancel ETHUSD buy 100 0000000001

func (me *MatchingEngine) DoAdd(msgList []string) {
	symbol := strings.ToUpper(msgList[1])
	orderbook, ok := me.orderbooks[symbol]
	if !ok {
		orderbook = NewOrderBook(symbol)
		me.orderbooks[symbol] = orderbook
	}
	orderbook.Add(msgList)
	fmt.Println(orderbook)
}
