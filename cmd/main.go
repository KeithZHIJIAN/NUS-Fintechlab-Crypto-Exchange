package main

import (
	"github.com/KeithZHIJIAN/nce-realmarket/realmarket"
)

func main() {
	var forever chan struct{}
	go realmarket.MatchingAgentStart()
	//Listen to live order book
	go realmarket.OrderBookAgentStart()
	//Listen to historical market
	go realmarket.HistoricalMarketAgentStart()
	//Serve orderbook and financial chart to front end
	go realmarket.WebsocketServerStart()
	//Record profit or loss trend
	go realmarket.PLAgentStart()
	<-forever
}
