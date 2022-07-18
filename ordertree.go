package main

import (
	"fmt"

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
			str += fmt.Sprintf("%v\n", OLiter.Value())
			cnt += 1
		}
	}
	return str
}

func (ot *OrderTree) UpdateString() string {
	t := treemap.Map(*ot)
	str := "[\n"
	OTiter := t.Iterator()
	cnt := 0
	for OTiter.Next() && cnt < DEPTH {
		OLiter := OTiter.Value().(*OrderList).Iterator()
		for OLiter.Next() && cnt < DEPTH {
			str += fmt.Sprintf("%v\n", OLiter.Value().(*Order).UpdateString())
			cnt += 1
		}
	}
	str += "]"
	return str
}
