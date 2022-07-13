package main

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func TestAddOrderTree(t *testing.T) {

	o1 := NewOrder("A", "1", "1", true, decimal.NewFromFloat(0), decimal.NewFromFloat(0), time.Now(), time.Now())
	o2 := NewOrder("A", "1", "1", true, decimal.NewFromFloat(1.5), decimal.NewFromFloat(1.5), time.Now(), time.Now())
	o3 := NewOrder("A", "1", "1", true, decimal.NewFromFloat(1.3), decimal.NewFromFloat(1.3), time.Now(), time.Now())
	ot := NewOrderTree()
	ot.Add(&Price{decimal.NewFromFloat(0), false}, o1)
	ot.Add(&Price{decimal.NewFromFloat(1.5), false}, o2)
	ot.Add(&Price{decimal.NewFromFloat(1.3), false}, o3)

	t.Log(ot)
}
