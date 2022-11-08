package orderbook

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/emirpasic/gods/maps/treemap"
)

type OrderTree struct {
	sync.Mutex
	t *treemap.Map
}

var DEPTH = 4

func NewOrderTree() *OrderTree {
	return &OrderTree{t: treemap.NewWith(func(a, b interface{}) int {
		return a.(*Price).Cmp(*b.(*Price))
	})}
}

func (ot *OrderTree) Add(p *Price, o *Order) {
	ls, ok := ot.t.Get(p)
	if !ok {
		ls = NewOrderList()
		ot.t.Put(p, ls)
	}
	ls.(*OrderList).Add(o)
}

func (ot *OrderTree) Remove(p *Price, id string) error {
	ls, ok := ot.t.Get(p)
	if !ok {
		return fmt.Errorf("OrderTree: Price not found")
	}
	ls.(*OrderList).Remove(id)
	if ls.(*OrderList).Size() == 0 {
		ot.t.Remove(p)
	}
	return nil
}

func (ot *OrderTree) Pop(p *Price, id string) (*Order, error) {
	log.Println(p.isBuy)
	ls, ok := ot.t.Get(p)
	if !ok {
		return nil, fmt.Errorf("OrderTree: Price not found")
	}
	log.Println(id)
	order, ok := ls.(*OrderList).Get(id)
	if !ok {
		return nil, fmt.Errorf("OrderTree: ID not found")
	}
	ls.(*OrderList).Remove(id)
	if ls.(*OrderList).Size() == 0 {
		ot.t.Remove(p)
	}
	return order, nil
}

func (ot *OrderTree) Get(p *Price) (*OrderList, bool) {
	ls, ok := ot.t.Get(p)
	if !ok {
		return nil, false
	}
	return ls.(*OrderList), true
}

func (ot *OrderTree) Iterator() treemap.Iterator {
	return ot.t.Iterator()
}

func (ot *OrderTree) String() string {
	str := ""
	OTiter := ot.t.Iterator()
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
	str := ""
	isBuy := true
	OTiter := ot.t.Iterator()
	cnt := 0
	for OTiter.Next() && cnt < DEPTH {
		price := OTiter.Key().(*Price)
		OLiter := OTiter.Value().(*OrderList).Iterator()

		if OLiter.Next() {
			isBuy = OLiter.Value().(*Order).IsBuy()
		}

		if isBuy {
			str += fmt.Sprintf("{\"price\": %s, \"openquantity\": %s}, ", price, OTiter.Value().(*OrderList).Quantity())
		} else {
			str = fmt.Sprintf("{\"price\": %s, \"openquantity\": %s}, ", price, OTiter.Value().(*OrderList).Quantity()) + str
		}
		cnt++
	}

	if isBuy {
		return "{\"getOpenBidOrdersForSymbol\": [" + strings.TrimSuffix(str, ", ") + "]}"
	} else {
		return "{\"getOpenAskOrdersForSymbol\": [" + strings.TrimSuffix(str, ", ") + "]}"
	}
}
