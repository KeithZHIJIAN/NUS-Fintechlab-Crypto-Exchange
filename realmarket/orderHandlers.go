package realmarket

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/KeithZHIJIAN/nce-realmarket/orderbook"
	"github.com/KeithZHIJIAN/nce-realmarket/utils"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

var OrderBooks map[string]*orderbook.OrderBook

func OrderBookAgentStart() {
	OrderBooks = make(map[string]*orderbook.OrderBook)
	for _, symbol := range SymbolMap {
		OrderBooks[symbol] = orderbook.NewOrderBook(symbol)
	}
	// log.Println(OrderBooks)
	ticker := time.NewTicker(500 * time.Millisecond)
	for {
		select {
		case <-ticker.C:
			for _, symbol := range SymbolMap {
				// price = MarketPrices[symbol]
				// check price(buy) with current ask
				value, ok := MarketPrices.Load(symbol)
				if !ok {
					continue
				}
				go matchOrders(value.(decimal.Decimal), OrderBooks[symbol].GetAsks())
				go matchOrders(value.(decimal.Decimal), OrderBooks[symbol].GetBids())
			}
		}
	}
}

func matchOrders(marketPrice decimal.Decimal, ot *orderbook.OrderTree) {
	filledList := make([]*orderbook.Order, 0)
	otIter := ot.Iterator()
	for otIter.Next() {
		price := otIter.Key().(*orderbook.Price)
		orderList := otIter.Value().(*orderbook.OrderList)
		if price.Match(marketPrice) {
			olIter := orderList.Iterator()
			for olIter.Next() {
				outbound := olIter.Value().(*orderbook.Order)
				filledList = append(filledList, outbound)
			}
		} else {
			break
		}
	}
	currTime := time.Now()
	for _, order := range filledList {
		fillOrder(order, currTime)
	}
}

// Remove locked asset and close order
func fillOrder(order *orderbook.Order, currTime time.Time) error {
	order.Fill(order.Price(), order.Quantity(), currTime)
	ot := OrderBooks[order.Symbol()].GetAsks()
	if order.IsBuy() {
		ot = OrderBooks[order.Symbol()].GetBids()
	}
	ot.Lock()
	defer ot.Unlock()
	ot.Remove(orderbook.NewPrice(order.Price(), order.IsBuy()), order.ID())
	return utils.SettleTrade(order.IsBuy(), order.Symbol(), order.ID(), order.WalletId(), order.OwnerId(), order.Quantity(), order.Price(), order.FillCost(), order.CreateTime(), currTime)
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

// Message format:
// -   Add, Symbol, Type, Side, Quantity, Price, Owner ID, Wallet ID, Stop Price (Optional)
//     add ETHUSD limit ask 100 64000 user1 Alice1 (60000)
//     add ethusd market ask 100 0 user1 Alice1 (60000)

// -   Modify, Symbol, Side, Order ID, prev Quantity, prev Price, new Quantity, new Price
//     modify ETHUSD buy 0000000002 100 63000 100 64000 //change price to 64000 only

//   - Cancel, Symbol, Side, Price, Order ID
//     cancel ETHUSD buy 100 0000000001
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

func addOrder(order *orderbook.Order) error {
	ot := OrderBooks[order.Symbol()].GetAsks()
	if order.IsBuy() {
		ot = OrderBooks[order.Symbol()].GetBids()
	}
	ot.Lock()
	defer ot.Unlock()
	ot.Add(orderbook.NewPrice(order.Price(), order.IsBuy()), order)
	return utils.CreateOpenOrder(order.IsBuy(), order.ID(), order.WalletId(), order.OwnerId(), order.Symbol(), order.Price(), order.Quantity(), time.Now())
}

func modifyOrder(symbol, id string, isBuy bool, newQuantity, newPrice, prevQuantity, prevPrice decimal.Decimal) error {
	ot := OrderBooks[symbol].GetAsks()
	if isBuy {
		ot = OrderBooks[symbol].GetBids()
	}
	ot.Lock()
	defer ot.Unlock()
	order, err := ot.Pop(orderbook.NewPrice(prevPrice, isBuy), id)
	if err != nil {
		return err
	}
	currTime := time.Now()
	order.ModifyPrice(newPrice, currTime)
	order.ModifyQuantity(newQuantity, currTime)
	ot.Add(orderbook.NewPrice(newPrice, isBuy), order)
	if isBuy {
		return utils.ModifyOpenBidOrder(symbol, id, order.OwnerId(), prevPrice.Mul(prevQuantity), newPrice, newQuantity, currTime)
	} else {
		return utils.ModifyOpenAskOrder(symbol, id, order.WalletId(), prevQuantity, newPrice, newQuantity, currTime)
	}
}

func cancelOrder(symbol, id string, isBuy bool, price decimal.Decimal) error {
	ot := OrderBooks[symbol].GetAsks()
	if isBuy {
		ot = OrderBooks[symbol].GetBids()
	}
	ot.Lock()
	defer ot.Unlock()
	order, err := ot.Pop(orderbook.NewPrice(price, isBuy), id)
	if err != nil {
		return err
	}
	return utils.CancelOrder(order.IsBuy(), order.Symbol(), order.ID(), order.WalletId(), order.OwnerId(), order.Quantity(), order.Price(), order.CreateTime(), time.Now())
}

func cancelOrderHandler(c *gin.Context) {
	log.Println("cancel order handler called")
	var requestBody RequestBody
	if err := c.BindJSON(&requestBody); err != nil {
		log.Println("[websocket server]: Cancel order bind JSON failed, error: ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := cancelOrder(strings.ToUpper(requestBody.Symbol), requestBody.OrderID, strings.ToUpper(requestBody.Side) == "BUY", requestBody.Price); err != nil {
		log.Println("[websocket server]: Cancel order failed, error: ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"result": "Modify order successful"})
}

func modifyOrderHandler(c *gin.Context) {
	var requestBody RequestBody
	if err := c.BindJSON(&requestBody); err != nil {
		log.Println("[websocket server]: Modify order bind JSON failed, error: ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if requestBody.NewQuantity.IsZero() {
		if err := cancelOrder(strings.ToUpper(requestBody.Symbol), requestBody.OrderID, strings.ToUpper(requestBody.Side) == "BUY", requestBody.NewPrice); err != nil {
			log.Println("[websocket server]: Cancel order failed, error: ", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}
	if err := modifyOrder(strings.ToUpper(requestBody.Symbol), requestBody.OrderID, strings.ToUpper(requestBody.Side) == "BUY", requestBody.NewQuantity, requestBody.NewPrice, requestBody.PrevQuantity, requestBody.PrevPrice); err != nil {
		log.Println("[websocket server]: Modify order failed, error: ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"result": "Modify order successful"})
}

func addOrderHandler(c *gin.Context) {
	var requestBody RequestBody
	if err := c.BindJSON(&requestBody); err != nil {
		log.Println("[websocket server]: Add order bind JSON failed, error: ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	order := parseOrder(&requestBody)
	if order.Price().IsZero() {
		marketPrice, ok := MarketPrices.Load(order.Symbol())
		for !ok {
			marketPrice, ok = MarketPrices.Load(order.Symbol())
		}
		utils.SettleMarketOrder(order.IsBuy(), order.ID(), order.WalletId(), order.OwnerId(), order.Symbol(), marketPrice.(decimal.Decimal), order.Quantity(), time.Now())
		log.Println("market order settled aha")
	} else if err := addOrder(order); err != nil {
		log.Println("[websocket server]: Add order failed, error: ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"result": "Add order successful"})
}

func parseOrder(req *RequestBody) *orderbook.Order {
	symbol := strings.ToUpper(req.Symbol)
	isBuy := strings.ToUpper(req.Side) == "BUY"
	quantity := req.Quantity
	price := req.Price
	if strings.ToUpper(req.Type) == "MARKET" {
		price = decimal.Zero
	}
	ownerId := req.OwnerID
	walletId := req.WalletID
	curr := time.Now()
	return orderbook.NewOrder(symbol, ownerId, walletId, isBuy, quantity, price, curr, curr)
}

type ClosedOrder struct {
	ID        string          `json:"orderid"`
	Action    string          `json:"action"`
	Quantity  decimal.Decimal `json:"quantity"`
	Price     decimal.Decimal `json:"price"`
	FillPrice decimal.Decimal `json:"fill_price"`
	CreatedAt time.Time       `json:"created_at"`
	FilledAt  time.Time       `json:"filled_at"`
}

type Orders struct {
	Asks   []OpenOrder   `json:"asks"`
	Bids   []OpenOrder   `json:"bids"`
	Closed []ClosedOrder `json:"closed"`
}

type OpenOrder struct {
	ID           string          `json:"orderid"`
	Quantity     decimal.Decimal `json:"quantity"`
	Price        decimal.Decimal `json:"price"`
	OpenQuantity decimal.Decimal `json:"open_quantity"`
	FillCost     decimal.Decimal `json:"fill_cost"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// symbol ownerid
func getOrderHandler(c *gin.Context) {
	symbol := c.Query("symbol")
	ownerID := c.Query("owner_id")
	orders := &Orders{}
	rows := utils.ReadOpenAskOrderBySymbolAndOwnerID(symbol, ownerID)
	orders.Asks = make([]OpenOrder, 0)
	for rows.Next() {
		openOrder := OpenOrder{}
		//orderid, quantity, price, openquantity, fillcost, createdat, updatedat
		if err := rows.Scan(&openOrder.ID, &openOrder.Quantity, &openOrder.Price, &openOrder.OpenQuantity,
			&openOrder.FillCost, &openOrder.CreatedAt, &openOrder.UpdatedAt); err != nil {
			panic(fmt.Errorf(err.Error()))
		}
		orders.Asks = append(orders.Asks, openOrder)
	}

	rows = utils.ReadOpenBidOrderBySymbolAndOwnerID(symbol, ownerID)
	orders.Bids = make([]OpenOrder, 0)
	for rows.Next() {
		openOrder := OpenOrder{}
		//orderid, quantity, price, openquantity, fillcost, createdat, updatedat
		if err := rows.Scan(&openOrder.ID, &openOrder.Quantity, &openOrder.Price, &openOrder.OpenQuantity,
			&openOrder.FillCost, &openOrder.CreatedAt, &openOrder.UpdatedAt); err != nil {
			panic(fmt.Errorf(err.Error()))
		}
		orders.Bids = append(orders.Bids, openOrder)
	}

	rows = utils.ReadClosedOrderBySymbolAndOwnerID(symbol, ownerID)
	for rows.Next() {
		closedOrder := ClosedOrder{}
		//orderid, quantity, price, openquantity, fillcost, createdat, updatedat
		if err := rows.Scan(&closedOrder.ID, &closedOrder.Action,
			&closedOrder.Quantity, &closedOrder.Price, &closedOrder.FillPrice,
			&closedOrder.CreatedAt, &closedOrder.FilledAt); err != nil {
			panic(fmt.Errorf(err.Error()))
		}
		orders.Closed = append(orders.Closed, closedOrder)
	}
	c.JSON(http.StatusOK, *orders)
}
