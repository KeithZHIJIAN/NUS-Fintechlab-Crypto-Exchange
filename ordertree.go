package main

import (
	"fmt"
	"strings"

	"github.com/emirpasic/gods/maps/treemap"
)

type OrderTree treemap.Map

var DEPTH = 4

func NewOrderTree() *OrderTree {
	t := treemap.NewWith(func(a, b interface{}) int {
		return a.(*Price).Cmp(*b.(*Price))
	})
	return (*OrderTree)(t)
}

func (ot *OrderTree) Add(p *Price, o *Order) {
	t := treemap.Map(*ot)
	ls, ok := t.Get(p)
	if !ok {
		ls = NewOrderList()
		t.Put(p, ls)
	}
	ls.(*OrderList).Add(o)
}

func (ot *OrderTree) Remove(p *Price, id string) {
	t := treemap.Map(*ot)
	ls, ok := t.Get(p)
	if !ok {
		fmt.Println("OrderTree: Price not found")
		return
	}
	ls.(*OrderList).Remove(id)
	if ls.(*OrderList).Size() == 0 {
		t.Remove(p)
	}
}

func (ot *OrderTree) Iterator() treemap.Iterator {
	t := treemap.Map(*ot)
	return t.Iterator()
}

func (ot *OrderTree) String() string {
	t := treemap.Map(*ot)
	str := ""
	OTiter := t.Iterator()
	cnt := 0
	for OTiter.Next() && cnt < DEPTH {
		OLiter := OTiter.Value().(*OrderList).Iterator()
		for OLiter.Next() && cnt < DEPTH {
			order := OLiter.Value().(*Order)
			if order.IsBuy() {
				str += fmt.Sprintf("%v\n", OLiter.Value())
			} else {
				str = fmt.Sprintf("%v\n", OLiter.Value()) + str
			}
			cnt += 1
		}
	}
	return str
}

func (ot *OrderTree) UpdateString() string {
	t := treemap.Map(*ot)
	str := ""
	isBuy := true
	OTiter := t.Iterator()
	cnt := 0
	for OTiter.Next() && cnt < DEPTH {
		price := OTiter.Key().(*Price)
		OLiter := OTiter.Value().(*OrderList).Iterator()
		if OLiter.Next() {
			order := OLiter.Value().(*Order)
			isBuy = order.IsBuy()
		}

		if isBuy {
			str += fmt.Sprintf("{\"price\": %s, \"openquantity\": %s}, ", price, OTiter.Value().(*OrderList).Quantity())
		} else {
			str = fmt.Sprintf("{\"price\": %s, \"openquantity\": %s}, ", price, OTiter.Value().(*OrderList).Quantity()) + str
		}
		cnt += 1
	}

	if isBuy {
		return "{\"getOpenBidOrdersForSymbol\": [" + strings.TrimSuffix(str, ", ") + "]}"
	} else {
		return "{\"getOpenAskOrdersForSymbol\": [" + strings.TrimSuffix(str, ", ") + "]}"
	}
}
