package main

var symbols = [...]string{"BTCUSD", "ETHUSD", "XRPUSD"}

// Message format:
// -   Add, Symbol, Type, Side, Quantity, Price, Owner ID, Wallet ID, Stop Price (Optional)
//     add ETHUSD limit ask 100 64000 user1 Alice1 (60000)
//     add ethusd market ask 100 0 user1 Alice1 (60000)

// -   Modify, Symbol, Side, Order ID, prev Quantity, prev Price, new Quantity, new Price
//     modify ETHUSD buy 0000000002 100 63000 100 64000 //change price to 64000 only

// -   Cancel, Symbol, Side, Price, Order ID
//     cancel ETHUSD buy 100 0000000001

func main() {
	var forever chan struct{}
	for _, symbol := range symbols {
		ob := NewOrderBook(symbol)
		go ob.Start()
	}
	<-forever
}
