package main

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/KeithZHIJIAN/nce-realmarket/utils"
	_ "github.com/lib/pq"
)

var TABLES = [...]string{
	"symbol",
	"users",
	"wallet",
	"wallet_assets",
	"contests",
	"user_contests",
	"closed_orders_btcusd",
	"closed_orders_ethusd",
	"closed_orders_xrpusd",
	"market_history_btcusd",
	"market_history_ethusd",
	"market_history_xrpusd",
	"open_ask_orders_btcusd",
	"open_ask_orders_ethusd",
	"open_ask_orders_xrpusd",
	"open_bid_orders_btcusd",
	"open_bid_orders_ethusd",
	"open_bid_orders_xrpusd",
	"order_fillings_btcusd",
	"order_fillings_ethusd",
	"order_fillings_xrpusd"}

func executeSqlScript(filepath string) {
	sqlFile := filepath
	query, err := ioutil.ReadFile(sqlFile)
	if err != nil {
		panic(err)
	}

	queries := strings.Split(string(query), "\n")
	sqlCommand := ""
	for _, query := range queries {
		query = strings.TrimSpace(query)
		if strings.HasPrefix(query, "--") || len(query) == 0 {
			continue
		}
		sqlCommand += query + "\n"
		if strings.HasSuffix(query, ";") {
			_, err := utils.DB.Exec(sqlCommand)
			if err != nil {
				panic(err)
			}
			sqlCommand = ""
		}
	}
}

func dropAllTables() {
	for _, table := range TABLES {
		query := "DROP TABLE IF EXISTS " + table + " CASCADE"
		_, err := utils.DB.Exec(query)
		if err != nil {
			panic(err)
		}
	}
}

func main() {
	dropAllTables()
	executeSqlScript("./init/db/CREATE_GLOBAL_TABLES.txt")
	utils.AddSymbol("btcusd")
	utils.AddSymbol("ethusd")
	utils.AddSymbol("xrpusd")
	executeSqlScript("./init/db/INIT_ORDERBOOK.txt")
	fmt.Println("Database initialized!")
}
