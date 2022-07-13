package main

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func TestNewOrder(t *testing.T) {
	t.Log(NewOrder("A", "1", "1", true, decimal.NewFromFloat(1.0), decimal.NewFromFloat(1.0), time.Now(), time.Now()))
}
