package main

import (
	"github.com/emirpasic/gods/maps/linkedhashmap"
)

type OrderList linkedhashmap.Map

func NewOrderList() *OrderList {
	m := linkedhashmap.New()
	return (*OrderList)(m)
}

func (ol *OrderList) Add(o *Order) {
	m := linkedhashmap.Map(*ol)
	m.Put(o.ID(), o)
}

func (ol *OrderList) Remove(id string) {
	m := linkedhashmap.Map(*ol)
	m.Remove(id)
}

func (ol *OrderList) Iterator() linkedhashmap.Iterator {
	m := linkedhashmap.Map(*ol)
	return m.Iterator()
}

func (ol *OrderList) Size() int {
	m := linkedhashmap.Map(*ol)
	return m.Size()
}
