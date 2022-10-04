package main

import "github.com/KeithZHIJIAN/nce-matchingengine/utils"

var SYMBOLS = [...]string{"BTCUSD", "ETHUSD", "XRPUSD"}

func main() {
	for _, symbol := range SYMBOLS {
		utils.InitQueueForSymbol(symbol)
	}
}
