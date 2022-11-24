package main

import (
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"os"
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
	"order_fillings_xrpusd",
	"topups",
	"pls",
}

var INDEXES = [...]string{
	"idx_topups_id",
	"idx_wallet_assets_id_symbol",
	"idx_users_id",
	"idx_open_bid_orders_btcusd_owner_orderid",
	"idx_open_bid_orders_ethusd_owner_orderid",
	"idx_open_ask_orders_btcusd_owner_orderid",
	"idx_open_ask_orders_ethusd_owner_orderid",
	"idx_closed_orders_btcusd_owner",
	"idx_closed_orders_ethusd_owner",
	"idx_market_history_btcusd_time",
	"idx_market_history_ethusd_time",
	"idx_topups_id_time",
	"idx_pls_id_time",
}

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
		query := fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table)
		_, err := utils.DB.Exec(query)
		if err != nil {
			panic(err)
		}
	}
}

func dropAllIndexes() {
	for _, index := range INDEXES {
		query := fmt.Sprintf("DROP INDEX IF EXISTS %s CASCADE", index)
		_, err := utils.DB.Exec(query)
		if err != nil {
			panic(err)
		}
	}
}

func hashedPassword(pwd string) string {
	h := sha256.New()
	_, err := h.Write([]byte(pwd))
	if err != nil {
		panic(err)
	}
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}

func createInitUser() {
	for i := 1; i < 5; i++ {
		query := os.Getenv(fmt.Sprintf("USER%dQUERY", i))
		_, err := utils.DB.Exec(query, hashedPassword(os.Getenv(fmt.Sprintf("USER%dPWD", i))))
		if err != nil {
			panic(err)
		}

	}
}

func main() {
	dropAllTables()
	executeSqlScript("./init/db/CREATE_GLOBAL_TABLES.txt")
	err := utils.AddSymbol("btcusd")
	if err != nil {
		panic(err)
	}
	err = utils.AddSymbol("ethusd")
	if err != nil {
		panic(err)
	}
	err = utils.AddSymbol("xrpusd")
	if err != nil {
		panic(err)
	}
	createInitUser()
	executeSqlScript("./init/db/INIT_ORDERBOOK.txt")
	fmt.Println("Database initialized!")
}
