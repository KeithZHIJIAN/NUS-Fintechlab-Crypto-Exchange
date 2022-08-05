package orderbook

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/shopspring/decimal"
)

var db = NewDB()

func NewDB() *sql.DB {
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}
	psqlInfo := os.Getenv("DATABASE_URL")
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	return db
}

func ReadUserBalance(userId string) decimal.Decimal {
	query := "SELECT balance FROM users WHERE userid = $1"
	row := db.QueryRow(query, userId)
	var balance decimal.Decimal
	err := row.Scan(&balance)
	if err != nil {
		panic(err)
	}
	return balance
}

func ReadWalletAsset(walletId, symbol string) decimal.Decimal {
	query := "SELECT amount FROM wallet_assets WHERE walletid = $1 AND symbol = $2"
	row := db.QueryRow(query, walletId, strings.ToLower(symbol))
	var amount decimal.Decimal
	err := row.Scan(&amount)
	if err != nil {
		panic(err)
	}
	return amount
}

func CreateTradeRecord(symbol, buyerId, sellerId string, crossPrice, fillQty decimal.Decimal, curr time.Time) {
	query := fmt.Sprintf("INSERT INTO ORDER_FILLINGS_%s (BUYORDERID, SELLORDERID, SYMBOL, PRICE, QUANTITY, TIME) VALUES ($1, $2, $3, $4, $5, $6)", symbol)
	_, err := db.Exec(query, buyerId, sellerId, symbol, crossPrice, fillQty, curr)
	if err != nil {
		panic(err)
	}
}

func CreateOpenOrder(isBuy bool, orderId, wallerId, userId, symbol string, price, quantity decimal.Decimal, curr time.Time) {
	side := "ASK"
	if isBuy {
		side = "BID"
	}
	query := fmt.Sprintf(`INSERT INTO OPEN_%s_ORDERS_%s 
	(ORDERID, WALLETID, OWNER, QUANTITY, SYMBOL, PRICE, OPENQUANTITY, FILLCOST, CREATEDAT, UPDATEDAT) 
	VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10);`, side, symbol)
	_, err := db.Exec(query, orderId, wallerId, userId, quantity, symbol, price, quantity, decimal.Zero, curr, curr)
	if err != nil {
		panic(err)
	}
}

func CreateMarketHistory(symbol string, curr time.Time, high, low, open, close, volume, vwap decimal.Decimal, trades int) {
	time := curr.Format("2006-01-02 15:04:05")
	query := fmt.Sprintf(`INSERT INTO MARKET_HISTORY_%s (TIME, HIGH, LOW, OPEN, CLOSE, VOLUME, VWAP, NUM_TRADES)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8);`, symbol)
	_, err := db.Exec(query, time, high, low, open, close, volume, vwap, trades)
	if err != nil {
		panic(err)
	}
}

func UpdateMarketHistory(symbol, curr string, high, low, open, close, volume, vwap decimal.Decimal, trades int) {
	query := fmt.Sprintf(`INSERT INTO MARKET_HISTORY_%s (TIME, HIGH, LOW, OPEN, CLOSE, VOLUME, VWAP, NUM_TRADES) 
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8) 
	ON CONFLICT (TIME) 
	DO UPDATE SET HIGH = $2, LOW = $3, OPEN = $4, CLOSE = $5, VOLUME = $6, VWAP = $7, NUM_TRADES = $8;`, symbol)

	_, err := db.Exec(query, curr, high, low, open, close, volume, vwap, trades)
	if err != nil {
		panic(err)
	}
}

func UpdateOrder(symbol, orderId string, isBuy bool, price, quantity, openQuantity decimal.Decimal, curr time.Time) {
	side := "ASK"
	if isBuy {
		side = "BID"
	}
	query := fmt.Sprintf(`UPDATE OPEN_%s_ORDERS_%s 
	SET PRICE = $1, QUANTITY = $2, OPENQUANTITY = $3, UPDATEDAT = $4 
	WHERE ORDERID = $5;`, side, symbol)
	_, err := db.Exec(query, price, quantity, openQuantity, curr, orderId)
	if err != nil {
		panic(err)
	}
}

func UpdateUserBalance(buyerId, sellerId string, fillCost decimal.Decimal) {
	buyerBalance := ReadUserBalance(buyerId)
	query := "UPDATE users SET balance = $1 WHERE userid = $2"
	_, err := db.Exec(query, buyerBalance.Sub(fillCost), buyerId)
	if err != nil {
		panic(err)
	}
	sellerBalance := ReadUserBalance(sellerId)
	_, err = db.Exec(query, sellerBalance.Add(fillCost), sellerId)
	if err != nil {
		panic(err)
	}
}

func UpdateWalletAsset(symbol, buyerId, sellerId string, amount decimal.Decimal) {
	buyerAsset := ReadWalletAsset(buyerId, symbol)

	query := "UPDATE wallet_assets SET amount = $1 WHERE walletid = $2 AND symbol = $3"
	_, err := db.Exec(query, buyerAsset.Add(amount), buyerId, strings.ToLower(symbol))
	if err != nil {
		panic(err)
	}
	sellerAsset := ReadWalletAsset(sellerId, symbol)
	_, err = db.Exec(query, sellerAsset.Sub(amount), sellerId, strings.ToLower(symbol))
	if err != nil {
		panic(err)
	}
}

func DeleteOpenOrder(isBuy bool, symbol, orderId string) {
	side := "ASK"
	if isBuy {
		side = "BID"
	}
	query := fmt.Sprintf("DELETE FROM OPEN_%s_ORDERS_%s WHERE ORDERID = $1;", side, symbol)
	_, err := db.Exec(query, orderId)
	if err != nil {
		panic(err)
	}
}

func CreateClosedOrder(isBuy bool, symbol, orderId, walletId, userId string, quantity, price, fillCost decimal.Decimal, createdAt, filledAt time.Time) {
	side := "SELL"
	if isBuy {
		side = "BUY"
	}
	fillPrice := decimal.Zero
	if quantity.GreaterThan(decimal.Zero) {
		fillPrice = fillCost.Div(quantity)
	}
	query := fmt.Sprintf(`INSERT INTO CLOSED_ORDERS_%s 
	(ORDERID, WALLETID, OWNER, BUYSIDE, QUANTITY, SYMBOL, PRICE, FILLCOST, FILLPRICE, CREATEDAT, FILLEDAT) 
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);`, symbol)
	db.Exec(query, orderId, walletId, userId, side, quantity, symbol, price, fillCost, fillPrice, createdAt, filledAt)
}
