package domain

import (
	"fmt"
	"math/rand"
	"time"
)

type OrderType string

const (
	MarketOrderType OrderType = "MARKET"
	LimitOrderType  OrderType = "LIMIT"
)

// make sure outer packages using &Order{} cant use it
// hiding the most essential info w/ lowercase
type Order struct {
	id         int64
	userId     string
	isBid      bool
	orderType  string
	size       float64
	limitPrice float64
	timestamp  int64
	NextOrder  *Order
	PrevOrder  *Order
}

func NewOrder(
	userId string,
	isBid bool,
	orderType string,
	size float64,
	limitPrice float64) *Order {
	return &Order{
		// TODO: incremental unique id
		id:         int64(rand.Int31()),
		userId:     userId,
		isBid:      isBid,
		orderType:  orderType,
		size:       size,
		limitPrice: limitPrice,
		timestamp:  time.Now().UnixNano(),
	}
}

// implement Stringer interface
func (o *Order) String() string {
	return fmt.Sprintf("{id: %d userId: %s isBid: %t orderType: %s size: %.2f limitPrice: %.2f timestamp: %d }",
		o.id,
		o.userId,
		o.isBid,
		o.orderType,
		o.size,
		o.limitPrice,
		o.timestamp)
}

func (o *Order) IsFilled() bool {
	return o.size == float64(0)
}

func (o *Order) GetId() int64 {
	return o.id
}

// id         int64
// userId     string
// isBid      bool
// orderType  string
// size       float64
// limitPrice float64
// timestamp  int64
func (o *Order) GetUserId() string {
	return o.userId
}
func (o *Order) GetIsBid() bool {
	return o.isBid
}
func (o *Order) GetOrderType() string {
	return o.orderType
}
func (o *Order) GetSize() float64 {
	return o.size
}
func (o *Order) GetLimitPrice() float64 {
	return o.limitPrice
}
func (o *Order) GetTimeStamp() int64 {
	return o.timestamp
}

// for each price level(limit) we need to know total volume and the corresponding orders
type Limit struct {
	limitPrice  float64
	TotalVolume float64
	Parent      *Limit
	LeftChild   *Limit
	RightChild  *Limit
	HeadOrder   *Order
	TailOrder   *Order
}

func NewLimit(limitPrice float64) *Limit {
	return &Limit{
		limitPrice: limitPrice,
	}
}

func (l *Limit) AddOrder(o *Order) {
	// if o is a pointer and is assigned directly to l.HeadOrder,
	// caller of this function may modify o after calling AddOrder and mess thing up
	newOrder := NewOrder(o.userId, o.isBid, o.orderType, o.size, o.limitPrice)
	if l.TailOrder == nil {
		// if empty
		l.HeadOrder = newOrder
		l.TailOrder = newOrder
	} else {
		l.TailOrder.NextOrder = newOrder
		newOrder.PrevOrder = l.TailOrder
		l.TailOrder = l.TailOrder.NextOrder
	}
	l.TotalVolume += float64(o.size)
	// TODO: throw error if sell limit but o is bid
}

func (l *Limit) DeleteOrder(id int64) {
	iterator := l.HeadOrder
	for iterator != nil && iterator.id != id {
		iterator = iterator.NextOrder
	}
	if iterator == nil {
		// TODO: throw error deleted order not in list
		return
	}

	l.TotalVolume -= iterator.size
	if (iterator.PrevOrder == nil) && (iterator.NextOrder == nil) {
		// if the only one left
		l.HeadOrder = nil
		l.TailOrder = nil
	} else if iterator.PrevOrder == nil {
		// if deleting the head
		l.HeadOrder = iterator.NextOrder
		l.HeadOrder.PrevOrder = nil
	} else {
		prevOrder := iterator.PrevOrder
		nextOrder := iterator.NextOrder
		prevOrder.NextOrder = nextOrder
		if nextOrder != nil {
			// if not the last one
			nextOrder.PrevOrder = prevOrder
		}
	}
}
