package main

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Order strores information about request
type Order struct {
	id           string
	symbol       string
	isBuy        bool
	quantity     decimal.Decimal
	openQuantity decimal.Decimal
	price        decimal.Decimal
	fillCost     decimal.Decimal
	ownerId      string
	walletId     string
	createTime   time.Time
	updateTime   time.Time
}

// NewOrder creates new constant object Order
func NewOrder(symbol, ownerId, walletId string, isBuy bool, quantity, price decimal.Decimal, createTime, updateTime time.Time) *Order {
	return &Order{
		id:           uuid.New().String(),
		symbol:       symbol,
		isBuy:        isBuy,
		quantity:     quantity,
		openQuantity: quantity,
		fillCost:     decimal.Zero,
		price:        price,
		ownerId:      ownerId,
		walletId:     walletId,
		createTime:   createTime,
		updateTime:   updateTime,
	}
}

// ID returns orderID field copy
func (o *Order) ID() string {
	return o.id
}

func (o *Order) Symbol() string {
	return o.symbol
}

// Side returns side of the order
func (o *Order) IsBuy() bool {
	return o.isBuy
}

// Quantity returns quantity field copy
func (o *Order) Quantity() decimal.Decimal {
	return o.quantity
}

// Price returns price field copy
func (o *Order) Price() decimal.Decimal {
	return o.price
}

func (o *Order) OpenQuantity() decimal.Decimal {
	return o.openQuantity
}

func (o *Order) FillCost() decimal.Decimal {
	return o.fillCost
}

func (o *Order) OwnerId() string {
	return o.ownerId
}

func (o *Order) WalletId() string {
	return o.walletId
}

func (o *Order) CreateTime() time.Time {
	return o.createTime
}

func (o *Order) ModifyPrice(newPrice decimal.Decimal, currTime time.Time) {
	if o.price == decimal.Zero {
		print("Order: Cannot modify market order price")
		return
	}
	o.price = newPrice
	o.updateTime = currTime
}

func (o *Order) ModifyQuantity(newQuantity decimal.Decimal, currTime time.Time) {
	delta := newQuantity.Sub(o.quantity)
	// if the reduction in quantity is greater than the open quantity, need to cancel the order
	if o.openQuantity.Add(delta).LessThan(decimal.Zero) {
		delta = o.openQuantity.Neg()
	}
	o.openQuantity = o.openQuantity.Add(delta)
	o.quantity = newQuantity
	o.updateTime = currTime
}

func (o *Order) Fill(price, quantity decimal.Decimal, currTime time.Time) {
	if o.openQuantity.LessThan(quantity) {
		print("Order: Fill quantity exceeds open quantity")
		return
	}
	o.fillCost = o.fillCost.Add(price.Mul(quantity))
	o.openQuantity = o.openQuantity.Sub(quantity)
	o.updateTime = currTime
}

func (o *Order) Filled() bool {
	return o.openQuantity.Equal(decimal.Zero)
}

func (o *Order) String() string {
	side := "Sell"
	if o.isBuy {
		side = "Buy"
	}
	return fmt.Sprintf("%s\t%s %s @ $%s", o.id, side, o.openQuantity, o.price)
}

func (o *Order) UpdateString() string {
	return fmt.Sprintf("{\"price\": %s, \"openquantity\": %s}", o.price, o.openQuantity)
}
