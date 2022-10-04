package main

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/KeithZHIJIAN/nce-matchingengine/utils"
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

func addSymbol(symbol string) {
	query := "INSERT INTO symbol (symbol) VALUES ($1)"
	_, err := utils.DB.Exec(query, symbol)
	if err != nil {
		panic(err)
	}

	query = "CREATE TABLE OPEN_ASK_ORDERS_" + symbol + `(
		ORDERID  	VARCHAR(64) 		NOT NULL,
		WALLETID  	INTEGER 		    NOT NULL,
		OWNER  		INTEGER 			NOT NULL,
		QUANTITY 	DECIMAL(15,5) 		NOT NULL,
		SYMBOL  	    VARCHAR(64) 		NOT NULL,
		PRICE 		DECIMAL(15,5) 		NOT NULL,
		OPENQUANTITY DECIMAL(15,5) 		NOT NULL,
		FILLCOST 	DECIMAL(15,5) 		NOT NULL,
		CREATEDAT  	timestamp without time zone 	NOT NULL,
		UPDATEDAT  	timestamp without time zone 	NOT NULL,
		PRIMARY KEY(ORDERID),
		FOREIGN KEY (WALLETID) REFERENCES WALLET(WALLETID),
		FOREIGN KEY (OWNER) REFERENCES USERS(USERID)
		);`
	_, err = utils.DB.Exec(query)
	if err != nil {
		panic(err)
	}

	query = "CREATE TABLE OPEN_BID_ORDERS_" + symbol + `(
		ORDERID  	VARCHAR(64) 		NOT NULL,
		WALLETID  	INTEGER 		    NOT NULL,
		OWNER  		INTEGER 			NOT NULL,
		QUANTITY 	DECIMAL(15,5) 		NOT NULL,
		SYMBOL  	    VARCHAR(64) 		NOT NULL,
		PRICE 		DECIMAL(15,5) 		NOT NULL,
		OPENQUANTITY DECIMAL(15,5) 		NOT NULL,
		FILLCOST 	DECIMAL(15,5) 		NOT NULL,
		CREATEDAT  	timestamp without time zone 	NOT NULL,
		UPDATEDAT  	timestamp without time zone 	NOT NULL,
		PRIMARY KEY(ORDERID),
		FOREIGN KEY (WALLETID) REFERENCES WALLET(WALLETID),
		FOREIGN KEY (OWNER) REFERENCES USERS(USERID)
		);`
	_, err = utils.DB.Exec(query)
	if err != nil {
		panic(err)
	}

	query = "CREATE TABLE CLOSED_ORDERS_" + symbol + `(
		ORDERID  	VARCHAR(64) 		NOT NULL,
		WALLETID  	INTEGER 		    NOT NULL,
		OWNER  		INTEGER 			NOT NULL,
		BUYSIDE  	VARCHAR(64) 		NOT NULL,
		QUANTITY 	DECIMAL(15,5) 		NOT NULL,
		SYMBOL  	    VARCHAR(64) 		NOT NULL,
		PRICE 		DECIMAL(15,5) 		NOT NULL,
		FILLCOST 	DECIMAL(15,5) 		NOT NULL,
		FILLPRICE 	DECIMAL(15,5) 		NOT NULL,
		CREATEDAT  	timestamp without time zone 	NOT NULL,
		FILLEDAT  	timestamp without time zone 	NOT NULL,
		PRIMARY KEY(ORDERID),
		FOREIGN KEY (WALLETID) REFERENCES WALLET(WALLETID),
		FOREIGN KEY (OWNER) REFERENCES USERS(USERID)
		);`
	_, err = utils.DB.Exec(query)
	if err != nil {
		panic(err)
	}

	query = "CREATE TABLE ORDER_FILLINGS_" + symbol + `(
		MATCHID 	    SERIAL 			NOT NULL,
		BUYORDERID  	VARCHAR(64),
		SELLORDERID  VARCHAR(64),
		SYMBOL  	    VARCHAR(64) 	NOT NULL,
		PRICE 		DECIMAL(15,5) 	NOT NULL,
		QUANTITY 	DECIMAL(15,5) 	NOT NULL,
		TIME  		timestamp without time zone 	NOT NULL,
		PRIMARY KEY(MATCHID)
		);`
	_, err = utils.DB.Exec(query)
	if err != nil {
		panic(err)
	}

	query = "CREATE TABLE MARKET_HISTORY_" + symbol + `(
		TIME  		timestamp without time zone 	NOT NULL,
		OPEN 		DECIMAL(15,5) 		NOT NULL,
		CLOSE 		DECIMAL(15,5) 		NOT NULL,
		HIGH 		DECIMAL(15,5) 		NOT NULL,
		LOW 		    DECIMAL(15,5) 		NOT NULL,
		VOLUME 		DECIMAL(15,5) 		NOT NULL,
		VWAP 		DECIMAL(15,5) 		NOT NULL,
		NUM_TRADES 	INTEGER 			NOT NULL	DEFAULT 0,
		PRIMARY KEY(TIME)
		);`
	_, err = utils.DB.Exec(query)
	if err != nil {
		panic(err)
	}

}

func main() {
	dropAllTables()
	executeSqlScript("./init/db/CREATE_GLOBAL_TABLES.txt")
	addSymbol("btcusd")
	addSymbol("ethusd")
	addSymbol("xrpusd")
	executeSqlScript("./init/db/INIT_ORDERBOOK.txt")
	fmt.Println("Database initialized!")
}
