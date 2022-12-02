package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/shopspring/decimal"
)

type RequestBody struct {
	Operation    string          `json:"operation"`
	Symbol       string          `json:"symbol"`
	Type         string          `json:"type"`
	Side         string          `json:"side"`
	Quantity     decimal.Decimal `json:"quantity"`
	Price        decimal.Decimal `json:"price"`
	OwnerID      string          `json:"owner_id"`
	WalletID     string          `json:"wallet_id"`
	OrderID      string          `json:"order_id"`
	PrevQuantity decimal.Decimal `json:"prev_quantity"`
	PrevPrice    decimal.Decimal `json:"prev_Price"`
	NewQuantity  decimal.Decimal `json:"new_quantity"`
	NewPrice     decimal.Decimal `json:"new_price"`
}

//	func parseOrder(req *RequestBody) *orderbook.Order {
//		symbol := strings.ToUpper(req.Symbol)
//		isBuy := strings.ToUpper(req.Side) == "BUY"
//		quantity := req.Quantity
//		price := req.Price
//		if strings.ToUpper(req.Type) == "MARKET" {
//			price = decimal.Zero
//		}
//		ownerId := req.OwnerID
//		walletId := req.WalletID
//		curr := time.Now()
//		return orderbook.NewOrder(symbol, ownerId, walletId, isBuy, quantity, price, curr, curr)
//	}
func main() {
	buy := RequestBody{Symbol: "BTCUSD", Side: "BUY", Quantity: decimal.NewFromFloat(0.05), Type: "MARKET", OwnerID: "1", WalletID: "1"}
	json_buy, err := json.Marshal(buy)
	if err != nil {
		log.Fatal(err)
	}
	sell := RequestBody{Symbol: "BTCUSD", Side: "SELL", Quantity: decimal.NewFromFloat(0.05), Type: "MARKET", OwnerID: "1", WalletID: "1"}
	json_sell, err := json.Marshal(sell)
	if err != nil {
		log.Fatal(err)
	}
	ticker := time.NewTicker(10 * time.Second)
	i := 0
	for {
		select {
		case <-ticker.C:
			log.Println("number of instructions executed: ", 2*i)
			i = 0
		default:
			http.Post("http://localhost:8000/order", "application/json", bytes.NewBuffer(json_buy))
			http.Post("http://localhost:8000/order", "application/json", bytes.NewBuffer(json_sell))
			i += 2
		}
	}
}
