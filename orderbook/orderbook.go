package orderbook

import (
	"fmt"
	"math/rand"
	"sort"
	"time"
)

type Order struct {
	// false = ask
	ID        int
	IsBid     bool
	Size      float64
	limit     *Limit
	Timestamp int64
}

// TODO: make sure ID is unique among orderbooks of all markets
func NewOrder(isBid bool, size float64) *Order {
	return &Order{
		ID:        int(rand.Int31()),
		IsBid:     isBid,
		Size:      size,
		Timestamp: time.Now().UnixNano(),
	}
}

// implement Stringer interface
func (o *Order) String() string {
	return fmt.Sprintf("size: %.2f", o.Size)
}

func (o *Order) isFilled() bool {
	return o.Size == float64(0)
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
	l.totalVolume += float64(o.Size)
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
	l.totalVolume -= float64(o.Size)
}

func (l *Limit) fill(incomingOrder *Order) []Match {
	matchArray := make([]Match, 0)
	for _, existingOrder := range l.Orders {

		if incomingOrder.isFilled() {
			break
		}
		if existingOrder.Size >= incomingOrder.Size {
			if incomingOrder.IsBid {
				matchArray = append(matchArray, Match{
					bid:        incomingOrder,
					ask:        existingOrder,
					price:      l.price,
					sizeFilled: incomingOrder.Size,
				})
			} else {
				matchArray = append(matchArray, Match{
					bid:        existingOrder,
					ask:        incomingOrder,
					price:      l.price,
					sizeFilled: incomingOrder.Size,
				})
			}
			l.totalVolume = l.totalVolume - incomingOrder.Size
			existingOrder.Size = existingOrder.Size - incomingOrder.Size
			incomingOrder.Size = 0
		} else {
			if incomingOrder.IsBid {
				matchArray = append(matchArray, Match{
					bid:        incomingOrder,
					ask:        existingOrder,
					price:      l.price,
					sizeFilled: existingOrder.Size,
				})
			} else {
				matchArray = append(matchArray, Match{
					bid:        existingOrder,
					ask:        incomingOrder,
					price:      l.price,
					sizeFilled: existingOrder.Size,
				})
			}
			l.totalVolume = l.totalVolume - existingOrder.Size
			incomingOrder.Size = incomingOrder.Size - existingOrder.Size
			existingOrder.Size = 0
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
	AskLimits      AskLimitsInterface
	BidLimits      BidLimitsInterface
	PriceToAsksMap map[float64]*Limit
	PriceToBidsMap map[float64]*Limit
	IDToOrderMap   map[int]*Order
}

func NewOrderbook() *OrderBook {
	// w/o make :  assignment to entry in nil map
	return &OrderBook{
		PriceToAsksMap: make(map[float64]*Limit),
		PriceToBidsMap: make(map[float64]*Limit),
		IDToOrderMap:   make(map[int]*Order),
	}
}

// fill at `price`
func (ob *OrderBook) PlaceLimitOrder(price float64, o *Order) {
	var limit *Limit

	// find the limit object with the corresponding price
	if o.IsBid {
		limit = ob.PriceToBidsMap[price]
	} else {
		limit = ob.PriceToAsksMap[price]
	}
	if limit == nil {
		limit = NewLimit(price)
		if o.IsBid {
			ob.BidLimits = append(ob.BidLimits, limit)
			ob.PriceToBidsMap[price] = limit
		} else {
			ob.AskLimits = append(ob.AskLimits, limit)
			ob.PriceToAsksMap[price] = limit
		}
	}

	ob.IDToOrderMap[o.ID] = o
	limit.AddOrder(o)
}

func (ob *OrderBook) sortAskLimits() {
	// sort.Sort takes in as argument that implement a certain interface.
	sort.Sort(ob.AskLimits)
}

func (ob *OrderBook) sortBidLimits() {
	// sort.Sort takes in as argument that implement a certain interface.
	sort.Sort(ob.BidLimits)
}

// fill at best price
func (ob *OrderBook) PlaceMarketOrder(incomingOrder *Order) []Match {
	// check if there is enough liquidity
	if incomingOrder.IsBid && incomingOrder.Size > ob.GetTotalVolumeAllAsks() {
		panic("Not enough ask liquidity")
	} else if !incomingOrder.IsBid && incomingOrder.Size > ob.GetTotalVolumeAllBids() {
		panic("Not enough ask liquidity")
	}

	matchArray := make([]Match, 0)
	// if bid, search for asks, starting from best price
	if incomingOrder.IsBid {
		// ob.ask should be sorted according to limit price
		ob.sortAskLimits()
		for _, limit := range ob.AskLimits {
			// inside a limit, orders should be sorted according to timestamp
			matchArray = append(matchArray, limit.fill(incomingOrder)...)
			if len(limit.Orders) == 0 {
				delete(ob.PriceToAsksMap, limit.price)
				for index, toBeDeletedLimit := range ob.AskLimits {
					if toBeDeletedLimit.price == limit.price {
						ob.AskLimits[index] = ob.AskLimits[len(ob.AskLimits)-1]
						ob.AskLimits = ob.AskLimits[:len(ob.AskLimits)-1]
						break
					}
				}
			}
		}
	} else {
		ob.sortBidLimits()
		for _, limit := range ob.BidLimits {
			// inside a limit, orders should be sorted according to timestamp
			matchArray = append(matchArray, limit.fill(incomingOrder)...)
			if len(limit.Orders) == 0 {
				delete(ob.PriceToBidsMap, limit.price)
				for index, toBeDeletedLimit := range ob.BidLimits {
					if toBeDeletedLimit.price == limit.price {
						ob.BidLimits[index] = ob.BidLimits[len(ob.BidLimits)-1]
						ob.BidLimits = ob.BidLimits[:len(ob.BidLimits)-1]
						break
					}
				}
			}
		}
	}
	return matchArray
}

func (ob *OrderBook) GetTotalVolumeAllBids() float64 {
	total := float64(0)
	for _, limit := range ob.BidLimits {
		total += limit.totalVolume
	}
	return total
}

func (ob *OrderBook) GetTotalVolumeAllAsks() float64 {
	total := float64(0)
	for _, limit := range ob.AskLimits {
		total += limit.totalVolume
	}
	return total
}

func (ob *OrderBook) CancelOrder(o *Order) {
	o.limit.DeleteOrder(o)
	o.limit = nil
}

type Match struct {
	ask        *Order
	bid        *Order
	sizeFilled float64
	price      float64
}
