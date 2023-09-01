package main

import (
	"fmt"
	"time"
)

type Order struct {
	isBid     bool
	size      float64
	limit     *Limit
	timestamp int64
	// false = ask
}

func NewOrder(isBid bool, size float64) *Order {
	return &Order{
		isBid:     isBid,
		size:      size,
		timestamp: time.Now().UnixNano(),
	}
}

// implement Stringer interface
func (o *Order) String() string {
	return fmt.Sprintf("size: %.2f", o.size)
}

type Limit struct {
	price       float64
	totalVolume float64
	Orders      []*Order
}

func NewLimit(price int) *Limit {
	return &Limit{
		price:  float64(price),
		Orders: []*Order{},
	}
}

func (l *Limit) AddOrder(o *Order) {
	o.limit = l
	l.Orders = append(l.Orders, o)
	l.totalVolume += float64(o.size)
}

func (l *Limit) DeleteOrder(o *Order) {
	// delete by swapping the deleted with the last entry and re-slicing
	for index, value := range l.Orders {
		if value == o {
			l.Orders[index] = l.Orders[len(l.Orders)-1]
			l.Orders = l.Orders[:len(l.Orders)-1]
		}
	}
	// remove the size of the deleted order from the total volume
	l.totalVolume -= float64(o.size)
}

type OrderBook struct {
	asks []Order
	bids []Order
}

func (ob *OrderBook) getTotalVolumeAllBids() float64 {
	return 0
}

func (ob *OrderBook) getTotalVolumeAllAsks() float64 {
	return 0
}

func (ob *OrderBook) placeMarketOrder() {
}
