package realmarket

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/KeithZHIJIAN/nce-realmarket/utils"
	"github.com/shopspring/decimal"
)

const Start = 1451606400 // January 1, 2016 00:00:00 AM
const MarketHistoryUpdateInterval = time.Minute
const FiveHours = 18000

func HistoricalMarketAgentStart() {
	// 18000 = 300min
	forever := make(chan interface{})
	go loadMarketHistory("BTCUSD")
	go loadMarketHistory("ETHUSD")
	log.Println("HistoricalMarketAgent: Market history websocket connection estbalished")
	<-forever
}

func loadMarketHistory(symbol string) {
	utils.DropTable(fmt.Sprintf("MARKET_HISTORY_%s", symbol))
	utils.CreateMarketHistoryTable(symbol)
	go func() {
		now := time.Now().Unix()
		for now-FiveHours > Start {
			requestMarketHistory(symbol, now, FiveHours)
			now -= FiveHours
		}
		requestMarketHistory(symbol, now, now-Start)
		fmt.Printf("%s historical data loaded till %v.\n", symbol[:3], time.Unix(Start, 0))
	}()
	time.Sleep(time.Minute)
	for {
		now := time.Now()
		tick := now.Truncate(MarketHistoryUpdateInterval).Add(MarketHistoryUpdateInterval)
		time.Sleep(tick.Sub(now) + time.Second)
		rows := requestMarketHistory(symbol, now.Unix(), 55)
		if len(rows) > 1 {
			panic("Update new market but get two rows")
		}
		// TIME, OPEN, CLOSE, HIGH, LOW, VOLUME
		LastCandlesticks.Store(symbol, &Candlestick{Time: time.Unix(rows[0][0].IntPart(), 0), Open: rows[0][1],
			Close: rows[0][2], High: rows[0][3], Low: rows[0][4], Vol: rows[0][5]})
		CurrCandlesticks.Store(symbol, &Candlestick{Open: rows[0][2], Close: rows[0][2], High: rows[0][2], Low: rows[0][2], Vol: decimal.Zero})
		fmt.Printf("[Time %v] %s new data added.\n", now, symbol[:3])
	}
}

func requestMarketHistory(symbol string, end, timerange int64) [][6]decimal.Decimal {
	url := fmt.Sprintf("https://api.exchange.coinbase.com/products/%s-%s/candles?granularity=60&start=%d&end=%d", symbol[:3], symbol[3:], end-timerange, end)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("accept", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(fmt.Errorf(err.Error()))
	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	ls := make([][6]decimal.Decimal, 0)
	err = json.Unmarshal(body, &ls)
	if len(ls) == 0 {
		fmt.Printf("request failed, start: %v, end: %v", time.Unix(end-timerange, 0), time.Unix(end, 0))
		return ls
	}
	if err != nil {
		panic(fmt.Errorf(err.Error()))
	}
	utils.InsertMarketHistory(symbol, ls...)
	return ls
}
