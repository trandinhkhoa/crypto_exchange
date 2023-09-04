package main

import (
	"fmt"
	"sort"
	"time"
)

type Order struct {
	// false = ask
	isBid     bool
	size      float64
	limit     *Limit
	timestamp int64
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
	// uppercase O for Order Springer
	Orders []*Order
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

func (l *Limit) fill(incomingOrder *Order) []Match {
	matchArray := make([]Match, 0)
	for _, existingOrder := range l.Orders {

		if incomingOrder.isFilled() {
			break
		}
		if existingOrder.size >= incomingOrder.size {
			if incomingOrder.isBid {
				matchArray = append(matchArray, Match{
					bid:        incomingOrder,
					ask:        existingOrder,
					price:      l.price,
					sizeFilled: incomingOrder.size,
				})
			} else {
				matchArray = append(matchArray, Match{
					bid:        existingOrder,
					ask:        incomingOrder,
					price:      l.price,
					sizeFilled: incomingOrder.size,
				})
			}
			l.totalVolume = l.totalVolume - incomingOrder.size
			existingOrder.size = existingOrder.size - incomingOrder.size
			incomingOrder.size = 0
		} else {
			if incomingOrder.isBid {
				matchArray = append(matchArray, Match{
					bid:        incomingOrder,
					ask:        existingOrder,
					price:      l.price,
					sizeFilled: existingOrder.size,
				})
			} else {
				matchArray = append(matchArray, Match{
					bid:        existingOrder,
					ask:        incomingOrder,
					price:      l.price,
					sizeFilled: existingOrder.size,
				})
			}
			l.totalVolume = l.totalVolume - existingOrder.size
			incomingOrder.size = incomingOrder.size - existingOrder.size
			existingOrder.size = 0
			l.DeleteOrder(existingOrder)
		}
	}
	return matchArray
}

// ob.askLimits should be sorted according to limit price
type AskLimitsInterface []*Limit

func (ls AskLimitsInterface) Less(a int, b int) bool {
	if ls[a].price < ls[b].price {
		return true
	} else {
		return false
	}
}

func (ls AskLimitsInterface) Swap(a int, b int) {
	ls[a], ls[b] = ls[b], ls[a]
}

func (ls AskLimitsInterface) Len() int {
	return len(ls)
}

type BidLimitsInterface []*Limit

func (ls BidLimitsInterface) Less(a int, b int) bool {
	if ls[a].price > ls[b].price {
		return true
	} else {
		return false
	}
}

func (ls BidLimitsInterface) Swap(a int, b int) {
	ls[a], ls[b] = ls[b], ls[a]
}

func (ls BidLimitsInterface) Len() int {
	return len(ls)
}

type OrderBook struct {
	askLimits      AskLimitsInterface
	bidLimits      BidLimitsInterface
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
		limit = ob.priceToBidsMap[price]
	} else {
		limit = ob.priceToAsksMap[price]
	}
	if limit == nil {
		limit = NewLimit(price)
		if o.isBid {
			ob.bidLimits = append(ob.bidLimits, limit)
			ob.priceToBidsMap[price] = limit
		} else {
			ob.askLimits = append(ob.askLimits, limit)
			ob.priceToAsksMap[price] = limit
		}
	}
	limit.AddOrder(o)
}

func (ob *OrderBook) sortAskLimits() {
	// sort.Sort takes in as argument that implement a certain interface.
	sort.Sort(ob.askLimits)
}

func (ob *OrderBook) sortBidLimits() {
	// sort.Sort takes in as argument that implement a certain interface.
	sort.Sort(ob.bidLimits)
}

// fill at best price
func (ob *OrderBook) placeMarketOrder(incomingOrder *Order) []Match {
	// check if there is enough liquidity
	if incomingOrder.isBid && incomingOrder.size > ob.getTotalVolumeAllAsks() {
		panic("Not enough ask liquidity")
	} else if !incomingOrder.isBid && incomingOrder.size > ob.getTotalVolumeAllBids() {
		panic("Not enough ask liquidity")
	}

	matchArray := make([]Match, 0)
	// if bid, search for asks, starting from best price
	if incomingOrder.isBid {
		// ob.ask should be sorted according to limit price
		ob.sortAskLimits()
		for _, limit := range ob.askLimits {
			// inside a limit, orders should be sorted according to timestamp
			matchArray = append(matchArray, limit.fill(incomingOrder)...)
			if len(limit.Orders) == 0 {
				delete(ob.priceToAsksMap, limit.price)
				for index, toBeDeletedLimit := range ob.askLimits {
					if toBeDeletedLimit.price == limit.price {
						ob.askLimits[index] = ob.askLimits[len(ob.askLimits)-1]
						ob.askLimits = ob.askLimits[:len(ob.askLimits)-1]
						break
					}
				}
			}
		}
	} else {
		ob.sortBidLimits()
		for _, limit := range ob.bidLimits {
			// inside a limit, orders should be sorted according to timestamp
			matchArray = append(matchArray, limit.fill(incomingOrder)...)
			if len(limit.Orders) == 0 {
				delete(ob.priceToBidsMap, limit.price)
				for index, toBeDeletedLimit := range ob.bidLimits {
					if toBeDeletedLimit.price == limit.price {
						ob.bidLimits[index] = ob.bidLimits[len(ob.bidLimits)-1]
						ob.bidLimits = ob.bidLimits[:len(ob.bidLimits)-1]
						break
					}
				}
			}
		}
	}
	return matchArray
}

func (ob *OrderBook) getTotalVolumeAllBids() float64 {
	total := float64(0)
	for _, limit := range ob.bidLimits {
		total += limit.totalVolume
	}
	return total
}

func (ob *OrderBook) getTotalVolumeAllAsks() float64 {
	total := float64(0)
	for _, limit := range ob.askLimits {
		total += limit.totalVolume
	}
	return total
}

type Match struct {
	ask        *Order
	bid        *Order
	sizeFilled float64
	price      float64
}
