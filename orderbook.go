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

func (o *Order) isFilled() bool {
	return o.size == float64(0)
}

// for each price level(limit) we need to know total volume and the corresponding orders
type Limit struct {
	price       float64
	totalVolume float64
	Orders      []*Order
}

func NewLimit(price float64) *Limit {
	return &Limit{
		price:  price,
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
	asks           []*Limit
	bids           []*Limit
	priceToAsksMap map[float64]*Limit
	priceToBidsMap map[float64]*Limit
}

func NewOrderbook() *OrderBook {
	return &OrderBook{
		priceToAsksMap: make(map[float64]*Limit),
		priceToBidsMap: make(map[float64]*Limit),
	}
}

// fill at `price`
func (ob *OrderBook) placeLimitOrder(price float64, o *Order) {
	var limit *Limit

	// find the limit object with the corresponding price
	if o.isBid {
		limit = ob.priceToAsksMap[price]
	} else {
		limit = ob.priceToBidsMap[price]
	}
	if limit == nil {
		limit = NewLimit(price)
		if o.isBid {
			ob.bids = append(ob.bids, limit)
			ob.priceToBidsMap[price] = limit
		} else {
			ob.asks = append(ob.asks, limit)
			ob.priceToAsksMap[price] = limit
		}
	}
	limit.AddOrder(o)
}

// fill at best price
func (ob *OrderBook) placeMarketOrder(o *Order) []Match {
	// check if there is enough liquidity
	if o.isBid && o.size > ob.getTotalVolumeAllAsks() {
		panic("Not enough ask liquidity")
	} else if !o.isBid && o.size > ob.getTotalVolumeAllBids() {
		panic("Not enough ask liquidity")
	}
	return nil
}

func (ob *OrderBook) getTotalVolumeAllBids() float64 {
	total := float64(0)
	for _, limit := range ob.bids {
		total += limit.totalVolume
	}
	return total
}

func (ob *OrderBook) getTotalVolumeAllAsks() float64 {
	total := float64(0)
	for _, limit := range ob.asks {
		total += limit.totalVolume
	}
	return total
}

type Match struct {
	ask        Order
	bid        Order
	sizeFilled float64
	price      float64
}
