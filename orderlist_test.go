package main

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func TestAddOrderList(t *testing.T) {
	o1 := NewOrder("A", "1", "1", true, decimal.NewFromFloat(1.0), decimal.NewFromFloat(1.0), time.Now(), time.Now())
	o2 := NewOrder("A", "1", "1", true, decimal.NewFromFloat(1.2), decimal.NewFromFloat(1.2), time.Now(), time.Now())
	ol := NewOrderList()
	ol.Add(o1)
	ol.Add(o2)
	t.Log(ol)
	ol.Remove("3")
	t.Log(ol)
}
