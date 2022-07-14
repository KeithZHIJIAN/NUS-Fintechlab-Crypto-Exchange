package main

var symbols = [...]string{"BTCUSD", "ETHUSD", "XRPUSD"}

func main() {
	var forever chan struct{}
	for _, symbol := range symbols {
		ob := NewOrderBook(symbol)
		go ob.Start()
	}
	<-forever
}
