package main

import (
	"testing"
)

func TestNewOrderBook(t *testing.T) {
	ob := NewOrderBook("A")
	ob.Add([]string{"A", "A", "BUY", "1", "1", "1", "1", "1"})

	t.Log(ob)
}
