package main

import (
	"fmt"

	"github.com/emirpasic/gods/maps/treemap"
)

type OrderTree treemap.Map

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
	str := "\n"
	it := t.Iterator()
	for it.Next() {
		str += fmt.Sprintf("%v\n", it.Value())
	}
	return str
}
