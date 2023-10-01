package entities

import "fmt"

// for each price level(limit) we need to know total volume and the corresponding orders
type Limit struct {
	limitPrice  float64
	TotalVolume float64
	Parent      *Limit
	LeftChild   *Limit
	RightChild  *Limit
	//TODO: how to make HeadOrder/TailOrder "readonly"
	HeadOrder *Order
	TailOrder *Order
}

func (l Limit) String() string {
	str := fmt.Sprintf("{\"limitPrice\": %.2f, \"totalVolume\": %.2f, \"orders\":", l.GetLimitPrice(), l.TotalVolume)
	str += "["
	iterator := l.HeadOrder
	for iterator != nil {
		str += iterator.String()
		iterator = iterator.NextOrder
		if iterator != nil {
			str += ","
		}
	}
	str += "]}"
	return str
}

func NewLimit(limitPrice float64) *Limit {
	return &Limit{
		limitPrice: limitPrice,
	}
}

func (l *Limit) GetLimitPrice() float64 {
	return l.limitPrice
}

func (l *Limit) AddOrder(newOrder *Order) {
	if l.TailOrder == nil {
		// if empty
		l.HeadOrder = newOrder
		l.TailOrder = newOrder
	} else {
		l.TailOrder.NextOrder = newOrder
		newOrder.PrevOrder = l.TailOrder
		l.TailOrder = l.TailOrder.NextOrder
	}
	newOrder.ParentLimit = l
	l.TotalVolume += float64(newOrder.Size)
	// TODO: throw error if sell limit but o is bid
}

func (l *Limit) DeleteOrder(order *Order) {
	l.TotalVolume -= order.Size
	if (order.PrevOrder == nil) && (order.NextOrder == nil) {
		// if the only one left
		l.HeadOrder = nil
		l.TailOrder = nil
	} else if order.PrevOrder == nil {
		// if deleting the head
		l.HeadOrder = order.NextOrder
		l.HeadOrder.PrevOrder = nil
	} else {
		prevOrder := order.PrevOrder
		nextOrder := order.NextOrder
		prevOrder.NextOrder = nextOrder
		if nextOrder != nil {
			// if not the last one
			nextOrder.PrevOrder = prevOrder
		}
	}
}
