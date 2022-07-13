package main

import (
	"fmt"

	"github.com/emirpasic/gods/maps/linkedhashmap"
)

type OrderList struct {
	list *linkedhashmap.Map
}

func NewOrderList() *OrderList {
	return &OrderList{
		linkedhashmap.New(),
	}
}

func (ol *OrderList) Add(o *Order) {
	ol.list.Put(o.ID(), o)
}

// func (ol *OrderList) Remove(o *Order) {
// 	ol.list.Remove(o.ID())
// }

func (ol *OrderList) Remove(id string) {
	ol.list.Remove(id)
}

func (ol *OrderList) String() string {
	str := ""
	it := ol.list.Iterator()
	for it.Next() {
		str += fmt.Sprintf("%v", it.Value())
	}
	return str
}
