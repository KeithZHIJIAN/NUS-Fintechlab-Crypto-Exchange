package main

import (
	"github.com/KeithZHIJIAN/nce-realmarket/realmarket"
)

func main() {
	var forever chan struct{}
	go realmarket.OrderBookAgentStart()
	//Listen to live order book
	go realmarket.LoadOrderBook()
	//Listen to historical market
	go realmarket.HistoricalMarketAgentStart()
	//Serve orderbook and financial chart to front end
	go realmarket.WebsocketServerStart()
	<-forever
}
