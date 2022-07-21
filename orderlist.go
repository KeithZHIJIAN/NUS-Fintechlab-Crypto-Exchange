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

func (ol *OrderList) Get(id string) (*Order, bool) {
	if o, ok := ol.m.Get(id); ok {
		return o.(*Order), true
	}
	return nil, false
}

func (ol *OrderList) Remove(id string) {
	ol.m.Remove(id)
}

func (ol *OrderList) Fill(quantity decimal.Decimal) {
	ol.quantity = ol.quantity.Sub(quantity)
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
