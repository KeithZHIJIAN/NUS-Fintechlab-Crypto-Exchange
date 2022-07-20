package main

import (
	"github.com/emirpasic/gods/maps/linkedhashmap"
	"github.com/shopspring/decimal"
)

type OrderList struct {
	m        *linkedhashmap.Map
	quantity decimal.Decimal
}

func NewOrderList() *OrderList {
	m := linkedhashmap.New()
	return &OrderList{m, decimal.Zero}
}

func (ol *OrderList) Add(o *Order) {
	ol.m.Put(o.ID(), o)
	ol.quantity = ol.quantity.Add(o.quantity)
}

func (ol *OrderList) Remove(id string) {
	if order, ok := ol.m.Get(id); ok {
		ol.m.Remove(id)
		ol.quantity = ol.quantity.Sub(order.(*Order).quantity)
	}
}

func (ol *OrderList) Iterator() linkedhashmap.Iterator {
	return ol.m.Iterator()
}

func (ol *OrderList) Size() int {
	return ol.m.Size()
}

func (ol *OrderList) Quantity() decimal.Decimal {
	return ol.quantity
}
