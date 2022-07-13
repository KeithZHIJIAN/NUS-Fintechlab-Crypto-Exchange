package main

import (
	"database/sql"
	"os"
	"strings"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/shopspring/decimal"
)

func ConnectDB() *sql.DB {
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
	db := ConnectDB()
	defer db.Close()

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
	db := ConnectDB()
	defer db.Close()

	query := "SELECT amount FROM wallet_assets WHERE walletid = $1 AND symbol = $2"
	row := db.QueryRow(query, walletId, strings.ToLower(symbol))
	var amount decimal.Decimal
	err := row.Scan(&amount)
	if err != nil {
		panic(err)
	}
	return amount
}
