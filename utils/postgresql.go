package utils

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

var DB = NewDB()

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

func InsertMarketHistory(symbol string, rows ...[6]decimal.Decimal) {
	query := make([]string, len(rows)+1)
	query[0] = fmt.Sprintf("INSERT INTO MARKET_HISTORY_%s (TIME, OPEN, CLOSE, HIGH, LOW, VOLUME) VALUES", symbol)
	for i := 0; i < len(rows)-1; i++ {
		query[i+1] = fmt.Sprintf("(to_timestamp(%d), %s, %s, %s, %s, %s), ", rows[i][0].IntPart(), rows[i][1].String(), rows[i][2].String(), rows[i][3].String(), rows[i][4].String(), rows[i][5].String())
	}
	query[len(rows)] = fmt.Sprintf("(to_timestamp(%d), %s, %s, %s, %s, %s);", rows[len(rows)-1][0].IntPart(), rows[len(rows)-1][1].String(), rows[len(rows)-1][2].String(), rows[len(rows)-1][3].String(), rows[len(rows)-1][4].String(), rows[len(rows)-1][5].String())

	_, err := DB.Exec(strings.Join(query, ""))
	if err != nil {
		panic(fmt.Errorf(err.Error()))
	}
}

func CreateMarketHistoryTable(symbol string) {
	query := fmt.Sprintf(`CREATE TABLE MARKET_HISTORY_%s 
	(
		TIME  		timestamp without time zone 	NOT NULL,
		OPEN 		DECIMAL(15,5) 		NOT NULL,
		CLOSE 		DECIMAL(15,5) 		NOT NULL,
		HIGH 		DECIMAL(15,5) 		NOT NULL,
		LOW 		DECIMAL(15,5) 		NOT NULL,
		VOLUME 		DECIMAL(15,5) 		NOT NULL,
		PRIMARY KEY(TIME)
	)`, symbol)
	_, err := DB.Exec(query)
	if err != nil {
		panic(fmt.Errorf(err.Error()))
	}
	query = fmt.Sprintf("CREATE INDEX idx_market_history_%s_time ON MARKET_HISTORY_%s(TIME);", symbol, symbol)
}

func createSymbol(txn *sql.Tx, symbol string) error {
	query := "INSERT INTO symbol (symbol) VALUES ($1)"
	_, err := txn.Exec(query, symbol)
	if err != nil {
		return err
	}
	return nil
}

func createOpenAskOrderTable(txn *sql.Tx, symbol string) error {
	query := fmt.Sprintf(`CREATE TABLE OPEN_ASK_ORDERS_%s (
		ORDERID  	VARCHAR(64) 		NOT NULL,
		WALLETID  	INTEGER 		    NOT NULL,
		OWNER  		INTEGER 			NOT NULL,
		QUANTITY 	DECIMAL(15,5) 		NOT NULL,
		SYMBOL  	VARCHAR(64) 		NOT NULL,
		PRICE 		DECIMAL(15,5) 		NOT NULL,
		OPENQUANTITY DECIMAL(15,5) 		NOT NULL,
		FILLCOST 	DECIMAL(15,5) 		NOT NULL,
		CREATEDAT  	timestamp without time zone 	NOT NULL,
		UPDATEDAT  	timestamp without time zone 	NOT NULL,
		PRIMARY KEY(ORDERID),
		FOREIGN KEY (WALLETID) REFERENCES WALLET(WALLETID),
		FOREIGN KEY (OWNER) REFERENCES USERS(USERID)
		);`, symbol)
	_, err := txn.Exec(query)
	if err != nil {
		return err
	}
	query = fmt.Sprintf("CREATE INDEX idx_OPEN_ASK_ORDERS_%s_owner_orderid ON OPEN_ASK_ORDERS_%s(OWNER, ORDERID);", symbol, symbol)
	_, err = txn.Exec(query)
	if err != nil {
		return err
	}
	return nil
}

func createOpenBidOrderTable(txn *sql.Tx, symbol string) error {
	query := fmt.Sprintf(`CREATE TABLE OPEN_BID_ORDERS_%s (
		ORDERID  	VARCHAR(64) 		NOT NULL,
		WALLETID  	INTEGER 		    NOT NULL,
		OWNER  		INTEGER 			NOT NULL,
		QUANTITY 	DECIMAL(15,5) 		NOT NULL,
		SYMBOL  	VARCHAR(64) 		NOT NULL,
		PRICE 		DECIMAL(15,5) 		NOT NULL,
		OPENQUANTITY DECIMAL(15,5) 		NOT NULL,
		FILLCOST 	DECIMAL(15,5) 		NOT NULL,
		CREATEDAT  	timestamp without time zone 	NOT NULL,
		UPDATEDAT  	timestamp without time zone 	NOT NULL,
		PRIMARY KEY(ORDERID),
		FOREIGN KEY (WALLETID) REFERENCES WALLET(WALLETID),
		FOREIGN KEY (OWNER) REFERENCES USERS(USERID)
		);`, symbol)
	_, err := txn.Exec(query)
	if err != nil {
		return err
	}
	query = fmt.Sprintf("CREATE INDEX idx_OPEN_BID_ORDERS_%s_owner_orderid ON OPEN_BID_ORDERS_%s(OWNER, ORDERID);", symbol, symbol)
	_, err = txn.Exec(query)
	if err != nil {
		return err
	}
	return nil
}

func createClosedOrderTable(txn *sql.Tx, symbol string) error {
	query := fmt.Sprintf(`CREATE TABLE CLOSED_ORDERS_%s (
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
		);`, symbol)
	_, err := txn.Exec(query)
	if err != nil {
		return err
	}
	query = fmt.Sprintf("CREATE INDEX idx_CLOSED_ORDERS_%s_owner ON CLOSED_ORDERS_%s(OWNER);", symbol, symbol)
	_, err = txn.Exec(query)
	if err != nil {
		return err
	}
	return nil
}

func AddSymbol(symbol string) error {
	txn, err := DB.Begin()
	if err != nil {
		return err
	}
	err = createSymbol(txn, symbol)
	if err != nil {
		txn.Rollback()
		return err
	}
	err = createOpenAskOrderTable(txn, symbol)
	if err != nil {
		txn.Rollback()
		return err
	}
	err = createOpenBidOrderTable(txn, symbol)
	if err != nil {
		txn.Rollback()
		return err
	}
	err = createClosedOrderTable(txn, symbol)
	if err != nil {
		txn.Rollback()
		return err
	}
	return txn.Commit()
}

func DropTable(table string) {
	query := fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table)
	_, err := DB.Exec(query)
	if err != nil {
		panic(fmt.Errorf(err.Error()))
	}
}

func deleteOpenOrder(txn *sql.Tx, isBuy bool, symbol, orderID string) {
	query := fmt.Sprintf("DELETE FROM OPEN_ASK_ORDERS_%s WHERE ORDERID = $1;", symbol)
	if isBuy {
		query = fmt.Sprintf("DELETE FROM OPEN_BID_ORDERS_%s WHERE ORDERID = $1;", symbol)
	}
	_, err := txn.Exec(query, orderID)
	if err != nil {
		panic(err)
	}
}

func createClosedOrder(txn *sql.Tx, isBuy bool, symbol, orderID, walletID, userID string, quantity, price, fillCost decimal.Decimal, createdAt, filledAt time.Time) {
	side := "SELL"
	if isBuy {
		side = "BUY"
	}
	fillPrice := decimal.Zero
	if quantity.GreaterThan(decimal.Zero) {
		fillPrice = fillCost.Div(quantity)
	}
	query := fmt.Sprintf("INSERT INTO CLOSED_ORDERS_%s (ORDERID, WALLETID, OWNER, BUYSIDE, QUANTITY, SYMBOL, PRICE, FILLCOST, FILLPRICE, CREATEDAT, FILLEDAT) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);", symbol)
	_, err := txn.Exec(query, orderID, walletID, userID, side, quantity, symbol, price, fillCost, fillPrice, createdAt, filledAt)
	if err != nil {
		panic(err)
	}
}

func lockBalanceAndAssetRows(txn *sql.Tx, userID, walletID, symbol string) {
	query := "SELECT * FROM users WHERE userid = $1 FOR UPDATE;"
	_, err := txn.Exec(query, userID)
	if err != nil {
		panic(err)
	}

	query = "SELECT * FROM WALLET_ASSETS WHERE walletid = $1 AND symbol = $2 FOR UPDATE;"
	_, err = txn.Exec(query, walletID, strings.ToLower(symbol))
	if err != nil {
		panic(err)
	}
}

func updateBalanceAndAssets(txn *sql.Tx, isBuy bool, userID, walletID, symbol string, amount, fillCost decimal.Decimal) {
	if isBuy {
		query := "UPDATE USERS SET locked = locked - $2 WHERE userid = $1;"
		_, err := txn.Exec(query, userID, fillCost)
		if err != nil {
			panic(err)
		}
		query = "UPDATE WALLET_ASSETS SET amount = amount + $3 WHERE walletid = $1 AND symbol = $2;"
		_, err = txn.Exec(query, walletID, strings.ToLower(symbol), amount)
		if err != nil {
			panic(err)
		}
	} else {
		query := "UPDATE WALLET_ASSETS SET locked = locked - $3 WHERE walletid = $1 AND symbol = $2;"
		_, err := txn.Exec(query, walletID, strings.ToLower(symbol), amount)
		if err != nil {
			panic(err)
		}
		query = "UPDATE USERS SET balance = balance + $2 WHERE userid = $1;"
		_, err = txn.Exec(query, userID, fillCost)
		if err != nil {
			panic(err)
		}
	}
}

func rollbackBalance(txn *sql.Tx, userID, symbol string, lockedBalance decimal.Decimal) {
	query := "UPDATE USERS SET locked = locked - $2 WHERE userid = $1;"
	_, err := txn.Exec(query, userID, lockedBalance)
	if err != nil {
		panic(err)
	}
	query = "UPDATE USERS SET balance = balance + $2 WHERE userid = $1;"
	_, err = txn.Exec(query, userID, lockedBalance)
	if err != nil {
		panic(err)
	}
}

func rollbackAsset(txn *sql.Tx, walletID, symbol string, amount decimal.Decimal) {
	query := "UPDATE WALLET_ASSETS SET locked = locked - $3 WHERE walletid = $1 AND symbol = $2;"
	_, err := txn.Exec(query, walletID, strings.ToLower(symbol), amount)
	if err != nil {
		panic(err)
	}
	query = "UPDATE WALLET_ASSETS SET amount = amount + $3 WHERE walletid = $1 AND symbol = $2;"
	_, err = txn.Exec(query, walletID, strings.ToLower(symbol), amount)
	if err != nil {
		panic(err)
	}
}

func CancelOrder(isBuy bool, symbol, orderID, walletID, userID string, quantity, price decimal.Decimal, createdAt, filledAt time.Time) error {
	txn, err := DB.Begin()
	if err != nil {
		return err
	}
	lockBalanceAndAssetRows(txn, userID, walletID, symbol)
	if isBuy {
		rollbackBalance(txn, userID, symbol, quantity.Mul(price))
	} else {
		rollbackAsset(txn, walletID, symbol, quantity)
	}
	deleteOpenOrder(txn, isBuy, symbol, orderID)
	createClosedOrder(txn, isBuy, symbol, orderID, walletID, userID, decimal.Zero, price, decimal.Zero, createdAt, filledAt)
	return txn.Commit()
}

// Remove record from open order, add it to db closed_order, consume from locked balance, and add to wallet asset
func SettleTrade(isBuy bool, symbol, orderID, walletID, userID string, quantity, price, fillCost decimal.Decimal, createdAt, filledAt time.Time) error {
	txn, err := DB.Begin()
	if err != nil {
		return err
	}
	lockBalanceAndAssetRows(txn, userID, walletID, symbol)
	updateBalanceAndAssets(txn, isBuy, userID, walletID, symbol, quantity, fillCost)
	deleteOpenOrder(txn, isBuy, symbol, orderID)
	createClosedOrder(txn, isBuy, symbol, orderID, walletID, userID, quantity, price, fillCost, createdAt, filledAt)
	return txn.Commit()
}

func lockBalanceRowAndGet(txn *sql.Tx, userID string) decimal.Decimal {
	query := "SELECT balance FROM users WHERE userid=$1 FOR UPDATE"
	row := txn.QueryRow(query, userID)
	var balance decimal.Decimal
	err := row.Scan(&balance)
	if err != nil {
		panic(err)
	}
	return balance
}

func lockAssetRowAndGet(txn *sql.Tx, walletID, symbol string) decimal.Decimal {
	query := "SELECT amount FROM wallet_assets WHERE walletid = $1 AND symbol = $2 FOR UPDATE;"
	row := txn.QueryRow(query, walletID, strings.ToLower(symbol))
	var asset decimal.Decimal
	err := row.Scan(&asset)
	if err != nil {
		panic(err)
	}
	return asset
}

func createOpenBidOrder(txn *sql.Tx, orderID, walletID, userID, symbol string, price, quantity decimal.Decimal, currTime time.Time) {
	query := fmt.Sprintf("INSERT INTO OPEN_BID_ORDERS_%s (ORDERID, WALLETID, OWNER, QUANTITY, SYMBOL, PRICE, OPENQUANTITY, FILLCOST, CREATEDAT, UPDATEDAT) VALUES ($1,$2,$3,$4,$5,$6,$4,0,$7,$7);", symbol)
	_, err := txn.Exec(query, orderID, walletID, userID, quantity, symbol, price, currTime)
	if err != nil {
		panic(err)
	}
}

func createOpenAskOrder(txn *sql.Tx, orderID, walletID, userID, symbol string, price, quantity decimal.Decimal, currTime time.Time) {
	query := fmt.Sprintf("INSERT INTO OPEN_ASK_ORDERS_%s (ORDERID, WALLETID, OWNER, QUANTITY, SYMBOL, PRICE, OPENQUANTITY, FILLCOST, CREATEDAT, UPDATEDAT) VALUES ($1,$2,$3,$4,$5,$6,$4,0,$7,$7);", symbol)
	_, err := txn.Exec(query, orderID, walletID, userID, quantity, symbol, price, currTime)
	if err != nil {
		panic(err)
	}
}

func lockBalance(txn *sql.Tx, userID string, quantity, price decimal.Decimal) {
	query := "UPDATE users SET balance = balance - $2, LOCKED = locked + $2 WHERE userid = $1;"
	_, err := txn.Exec(query, userID, quantity.Mul(price))
	if err != nil {
		panic(err)
	}
}

func lockAsset(txn *sql.Tx, walletID, symbol string, quantity decimal.Decimal) {
	query := "UPDATE wallet_assets SET amount = amount - $3, locked = locked + $3 WHERE walletid = $1 AND symbol = $2;"
	_, err := txn.Exec(query, walletID, strings.ToLower(symbol), quantity)
	if err != nil {
		panic(err)
	}
}

// Create open order and lock corresponding asset
func CreateOpenOrder(isBuy bool, orderID, walletID, userID, symbol string, price, quantity decimal.Decimal, currTime time.Time) error {
	txn, err := DB.Begin()
	if err != nil {
		return err
	}
	if isBuy {
		balance := lockBalanceRowAndGet(txn, userID)
		if balance.GreaterThan(price.Mul(quantity)) {
			createOpenBidOrder(txn, orderID, walletID, userID, symbol, price, quantity, currTime)
			lockBalance(txn, userID, quantity, price)
		} else {
			txn.Rollback()
			return fmt.Errorf("Balance not enough")
		}
	} else {
		asset := lockAssetRowAndGet(txn, walletID, symbol)
		if asset.GreaterThan(quantity) {
			createOpenAskOrder(txn, orderID, walletID, userID, symbol, price, quantity, currTime)
			lockAsset(txn, walletID, symbol, quantity)
		} else {
			txn.Rollback()
			return fmt.Errorf("Asset %s not enough", symbol)
		}
	}
	err = txn.Commit()
	return err
}

func ModifyOpenBidOrder(symbol, orderID, userID string, prevFillCost, price, quantity decimal.Decimal, currTime time.Time) error {
	txn, err := DB.Begin()
	if err != nil {
		return err
	}
	updateBidOrder(txn, symbol, orderID, price, quantity, currTime)
	rollbackBalance(txn, userID, symbol, prevFillCost.Sub(price.Mul(quantity)))
	return txn.Commit()
}

func ModifyOpenAskOrder(symbol, orderID, walletID string, prevQuantity, price, quantity decimal.Decimal, currTime time.Time) error {
	txn, err := DB.Begin()
	if err != nil {
		return err
	}
	updateAskOrder(txn, symbol, orderID, price, quantity, currTime)
	rollbackAsset(txn, walletID, symbol, prevQuantity.Sub(quantity))
	return txn.Commit()
}

func updateBidOrder(txn *sql.Tx, symbol, orderID string, price, quantity decimal.Decimal, currTime time.Time) {
	query := fmt.Sprintf("UPDATE OPEN_BID_ORDERS_%s SET PRICE = $1, QUANTITY = $2, OPENQUANTITY = $2, UPDATEDAT = $3 WHERE ORDERID = $4;", symbol)
	_, err := txn.Exec(query, price, quantity, currTime, orderID)
	if err != nil {
		panic(err)
	}
}

func updateAskOrder(txn *sql.Tx, symbol, orderID string, price, quantity decimal.Decimal, currTime time.Time) {
	query := fmt.Sprintf("UPDATE OPEN_ASK_ORDERS_%s SET PRICE = $1, QUANTITY = $2, OPENQUANTITY = $2, UPDATEDAT = $3 WHERE ORDERID = $4;", symbol)
	_, err := txn.Exec(query, price, quantity, currTime, orderID)
	if err != nil {
		panic(err)
	}
}

func ReadHistoricalMarket(symbol string, limit int) *sql.Rows {
	var query string
	if limit != 0 {
		query = fmt.Sprintf("SELECT * FROM MARKET_HISTORY_%s ORDER BY TIME DESC LIMIT %d", symbol, limit)
	} else {
		query = fmt.Sprintf("SELECT * FROM MARKET_HISTORY_%s ORDER BY TIME DESC LIMIT 3600", symbol)
	}
	rows, err := DB.Query(query)
	if err != nil {
		panic(fmt.Errorf(err.Error()))
	}
	return rows
}

func ReadOpenAskOrderBySymbolAndOwnerID(symbol, ownerID string) *sql.Rows {
	query := fmt.Sprintf("SELECT orderid, quantity, price, openquantity, fillcost, createdat, updatedat FROM OPEN_ASK_ORDERS_%s WHERE owner = $1", symbol)
	rows, err := DB.Query(query, ownerID)
	if err != nil {
		panic(err)
	}
	return rows
}

func ReadOpenBidOrderBySymbolAndOwnerID(symbol, ownerID string) *sql.Rows {
	query := fmt.Sprintf("SELECT orderid, quantity, price, openquantity, fillcost, createdat, updatedat FROM OPEN_BID_ORDERS_%s WHERE owner = $1", symbol)
	rows, err := DB.Query(query, ownerID)
	if err != nil {
		panic(err)
	}
	return rows
}

func ReadClosedOrderBySymbolAndOwnerID(symbol, ownerID string) *sql.Rows {
	query := fmt.Sprintf("SELECT orderid, buyside, quantity, price, fillprice, createdat, filledat FROM CLOSED_ORDERS_%s WHERE owner = $1", symbol)
	rows, err := DB.Query(query, ownerID)
	if err != nil {
		panic(err)
	}
	return rows
}

func SettleMarketOrder(isBuy bool, orderID, walletID, userID, symbol string, price, quantity decimal.Decimal, currTime time.Time) error {
	txn, err := DB.Begin()
	if err != nil {
		return err
	}
	if isBuy {
		balance := lockBalanceRowAndGet(txn, userID)
		if balance.GreaterThan(price.Mul(quantity)) {
			createClosedOrder(txn, isBuy, symbol, orderID, walletID, userID, quantity, price, price.Mul(quantity), currTime, currTime)
			settleMarketBuy(txn, userID, walletID, symbol, quantity, price)
		} else {
			txn.Rollback()
			return fmt.Errorf("Balance not enough")
		}
	} else {
		asset := lockAssetRowAndGet(txn, walletID, symbol)
		if asset.GreaterThan(quantity) {
			createClosedOrder(txn, isBuy, symbol, orderID, walletID, userID, quantity, price, price.Mul(quantity), currTime, currTime)
			settleMarketSell(txn, userID, walletID, symbol, quantity, price)
		} else {
			txn.Rollback()
			return fmt.Errorf("Asset %s not enough", symbol)
		}
	}
	err = txn.Commit()
	return err
}

func settleMarketBuy(txn *sql.Tx, userID, walletID, symbol string, quantity, price decimal.Decimal) {
	query := "UPDATE users SET balance = balance - $2 WHERE userid = $1;"
	_, err := txn.Exec(query, userID, quantity.Mul(price))
	if err != nil {
		panic(err)
	}
	query = "UPDATE wallet_assets SET amount = amount + $3 WHERE walletid = $1 AND symbol = $2;"
	_, err = txn.Exec(query, walletID, strings.ToLower(symbol), quantity)
	if err != nil {
		panic(err)
	}
}

func settleMarketSell(txn *sql.Tx, userID, walletID, symbol string, quantity, price decimal.Decimal) {
	query := "UPDATE users SET balance = balance + $2 WHERE userid = $1;"
	_, err := txn.Exec(query, userID, quantity.Mul(price))
	if err != nil {
		panic(err)
	}
	query = "UPDATE wallet_assets SET amount = amount - $3 WHERE walletid = $1 AND symbol = $2;"
	_, err = txn.Exec(query, walletID, strings.ToLower(symbol), quantity)
	if err != nil {
		panic(err)
	}
}

// USERID		INTEGER	NOT NULL,
// AMOUNT      DECIMAL(15,5)	NOT NULL,
// TIME        timestamp without time zone 	NOT NULL,

func CreatePL(userID string, amount decimal.Decimal, currTime time.Time) {
	query := "INSERT INTO PLS (USERID, AMOUNT, TIME) VALUES ($1,$2,$3);"
	_, err := DB.Exec(query, userID, amount, currTime)
	if err != nil {
		panic(err)
	}
}

func ReadUsersBalance() *sql.Rows {
	query := "SELECT userid, balance, locked FROM users;"
	rows, err := DB.Query(query)
	if err != nil {
		panic(fmt.Errorf(err.Error()))
	}
	return rows
}

func ReadUserAsset(userID string) *sql.Rows {
	query := "SELECT symbol, amount, locked FROM wallet_assets where walletid = $1;"
	rows, err := DB.Query(query, userID)
	if err != nil {
		panic(fmt.Errorf(err.Error()))
	}
	return rows
}

func ReadUserPL(userID string) *sql.Rows {
	query := "SELECT time, amount FROM pls WHERE userid = $1 ORDER BY time LIMIT 288;"
	rows, err := DB.Query(query, userID)
	if err != nil {
		panic(fmt.Errorf(err.Error()))
	}
	return rows
}

func CreateTopUps(userID string, amount decimal.Decimal, currTime time.Time) error {
	txn, err := DB.Begin()
	if err != nil {
		return err
	}
	query := "INSERT INTO TOPUPS (USERID, AMOUNT, TIME) VALUES ($1,$2,$3);"
	_, err = txn.Exec(query, userID, amount, currTime)
	if err != nil {
		return err
	}
	query = "UPDATE USERS SET BALANCE = BALANCE + $2 WHERE userid= $1"
	_, err = txn.Exec(query, userID, amount)
	if err != nil {
		return err
	}
	return txn.Commit()
}

func ReadLastPL(userID string) *sql.Rows {
	query := "SELECT amount FROM PLS WHERE USERID = $1 ORDER BY TIME DESC LIMIT 1;"
	rows, err := DB.Query(query, userID)
	if err != nil {
		panic(fmt.Errorf(err.Error()))
	}
	return rows
}

func ReadFirstPL(userID string) *sql.Rows {
	query := "SELECT amount FROM PLS WHERE USERID = $1 ORDER BY TIME ASC LIMIT 1;"
	rows, err := DB.Query(query, userID)
	if err != nil {
		panic(fmt.Errorf(err.Error()))
	}
	return rows
}

func ReadCumulativeTopups(userID string) *sql.Rows {
	query := "SELECT SUM(amount) FROM topups WHERE USERID = $1;"
	rows, err := DB.Query(query, userID)
	if err != nil {
		panic(fmt.Errorf(err.Error()))
	}
	return rows
}

func ReadUserLearnStageByID(userID string) *sql.Rows {
	query := "SELECT learnstage FROM users WHERE userid = $1"
	rows, err := DB.Query(query, userID)
	if err != nil {
		panic(fmt.Errorf(err.Error()))
	}
	return rows
}

func ReadUserBalanceByID(userID string) *sql.Rows {
	query := "SELECT balance FROM users WHERE userid = $1"
	rows, err := DB.Query(query, userID)
	if err != nil {
		panic(fmt.Errorf(err.Error()))
	}
	return rows
}

func ReadWalletAssetByIDAndSymbol(walletID, symbol string) *sql.Rows {
	query := "SELECT amount FROM wallet_assets WHERE walletid = $1 and symbol = $2"
	rows, err := DB.Query(query, walletID, symbol)
	if err != nil {
		panic(fmt.Errorf(err.Error()))
	}
	return rows
}
