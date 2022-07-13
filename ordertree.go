package main

import (
	"fmt"

	"github.com/emirpasic/gods/maps/treemap"
)

type OrderTree struct {
	tree *treemap.Map
}

func NewOrderTree() *OrderTree {
	return &OrderTree{
		tree: treemap.NewWith(func(a, b interface{}) int {
			return a.(*Price).Cmp(*b.(*Price))
		}),
	}
}

func (ot *OrderTree) Add(p *Price, o *Order) {
	ls, ok := ot.tree.Get(p)
	if !ok {
		ls = NewOrderList()
		ot.tree.Put(p, ls)
	}
	ls.(*OrderList).Add(o)
}

func (ot *OrderTree) Remove(p *Price, id string) {
	ls, ok := ot.tree.Get(p)
	if !ok {
		fmt.Println("OrderTree: Price not found")
		return
	}
	ls.(*OrderList).Remove(id)
	if ls.(*OrderList).list.Size() == 0 {
		ot.tree.Remove(p)
	}
}

func (ot *OrderTree) String() string {
	str := "\n"
	it := ot.tree.Iterator()
	for it.Next() {
		str += fmt.Sprintf("%v\n", it.Value())
	}
	return str
}
