package realmarket

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"syscall"
	"time"

	"github.com/KeithZHIJIAN/nce-realmarket/utils"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func realOrderBookHandler(c *gin.Context) {
	//upgrade get request to websocket protocol
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		panic(fmt.Errorf(err.Error()))
	}
	defer ws.Close()
	_, message, err := ws.ReadMessage()
	if err != nil {
		panic(fmt.Errorf(err.Error()))
	}
	var res WebsocketResponse
	//If client message is ping will return pong
	err = json.Unmarshal(message, &res)
	if _, ok := OrderBooks[res.ProductID]; !ok {
		fmt.Println(res.ProductID)
		fmt.Println(OrderBooks)
		panic("ProductID not exists")
	}
	for {
		now := time.Now()
		tick := now.Truncate(OrderBookUpdateInterval).Add(OrderBookUpdateInterval)
		time.Sleep(tick.Sub(now))
		//Response message to client
		ob, ok := getOrderBook(res.ProductID)
		if !ok {
			log.Println("realOrderBookHandler: Order book not loaded yet")
			continue
		}
		err = ws.WriteJSON(*ob)
		if err != nil {
			fmt.Println(err)
			break
		}
	}
}

var marketHistoryLock sync.Mutex

func realMarketHistoryHandler(c *gin.Context) {
	fmt.Println("realMarketHistoryHandler called")
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		panic(fmt.Errorf(err.Error()))
	}
	defer ws.Close()
	_, message, err := ws.ReadMessage()
	if err != nil {
		panic(fmt.Errorf(err.Error()))
	}
	var res WebsocketResponse
	//If client message is ping will return pong
	err = json.Unmarshal(message, &res)
	if _, ok := OrderBooks[res.ProductID]; !ok {
		fmt.Println(res.ProductID)
		fmt.Println(OrderBooks)
		panic("ProductID not exists")
	}

	// first reply returns all historical data, type == "snapshot"
	var marketHistoryMsg struct {
		Symbol        string        `json:"symbol"`
		Type          string        `json:"type"`
		MarketHistory []Candlestick `json:"market_history"`
	}
	marketHistoryMsg.Symbol = res.ProductID
	marketHistoryMsg.Type = "snapshot"
	MarketHistory := make([]Candlestick, 0)
	rows := utils.ReadHistoricalMarket(res.ProductID)
	for rows.Next() {
		mhr := Candlestick{}
		if err := rows.Scan(&mhr.Time, &mhr.Open, &mhr.Close,
			&mhr.High, &mhr.Low, &mhr.Vol); err != nil {
			panic(fmt.Errorf(err.Error()))
		}
		MarketHistory = append(MarketHistory, mhr)
	}
	if err = rows.Err(); err != nil {
		panic(fmt.Errorf(err.Error()))
	}
	marketHistoryMsg.MarketHistory = MarketHistory
	marketHistoryLock.Lock()
	err = ws.WriteJSON(marketHistoryMsg)
	marketHistoryLock.Unlock()
	if errors.Is(err, syscall.EPIPE) || errors.Is(err, syscall.ECONNRESET) {
		log.Printf("websocket server: %v\n", err)
		ws.Close()
		return
	} else if err != nil {
		panic(fmt.Errorf(err.Error()))
	}

	//every minute add_candlestick
	go addCandlestick(res.ProductID, ws)
	updateCandlestick(res.ProductID, ws)
}

func addCandlestick(symbol string, ws *websocket.Conn) {
	candlestickMsg := &CandlestickMsg{}
	for {
		now := time.Now()
		tick := now.Truncate(MarketHistoryUpdateInterval).Add(MarketHistoryUpdateInterval)
		time.Sleep(tick.Sub(now) + 1*time.Second)
		candlestickMsg.Type = "add_candlestick"
		candlestickMsg.Symbol = symbol
		value, ok := LastCandlesticks.Load(symbol)
		if !ok {
			continue
		}
		candlestickMsg.Candlestick = *value.(*Candlestick)
		marketHistoryLock.Lock()
		err := ws.WriteJSON(*candlestickMsg)
		marketHistoryLock.Unlock()
		if errors.Is(err, syscall.EPIPE) || errors.Is(err, syscall.ECONNRESET) {
			log.Printf("websocket server: %v\n", err)
			ws.Close()
			return
		} else if err != nil {
			panic(fmt.Errorf(err.Error()))
		}
	}
}

func updateCandlestick(symbol string, ws *websocket.Conn) {
	candlestickMsg := &CandlestickMsg{}
	for {
		now := time.Now()
		tick := now.Truncate(OrderBookUpdateInterval).Add(OrderBookUpdateInterval)
		time.Sleep(tick.Sub(now))
		candlestickMsg.Type = "update_candlestick"
		candlestickMsg.Symbol = symbol
		value, ok := CurrCandlesticks.Load(symbol)
		if !ok {
			continue
		}
		candlestickMsg.Candlestick = *value.(*Candlestick)
		marketHistoryLock.Lock()
		err := ws.WriteJSON(*candlestickMsg)
		marketHistoryLock.Unlock()
		if errors.Is(err, syscall.EPIPE) || errors.Is(err, syscall.ECONNRESET) {
			log.Printf("websocket server: %v\n", err)
			ws.Close()
			return
		} else if err != nil {
			panic(fmt.Errorf(err.Error()))
		}
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		return origin == "http://localhost:3000"
	},
}
