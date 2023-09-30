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
	id          int64
	userId      string
	ticker      string
	isBid       bool
	orderType   OrderType
	Size        float64
	limitPrice  float64
	timestamp   int64
	NextOrder   *Order
	PrevOrder   *Order
	ParentLimit *Limit
}

func NewOrder(
	userId string,
	ticker string,
	isBid bool,
	orderType OrderType,
	size float64,
	limitPrice float64) *Order {
	return &Order{
		// TODO: incremental unique id
		id:         int64(rand.Int31()),
		ticker:     ticker,
		userId:     userId,
		isBid:      isBid,
		orderType:  orderType,
		Size:       size,
		limitPrice: limitPrice,
		timestamp:  time.Now().UnixNano(),
	}
}

// implement Stringer interface
func (o *Order) String() string {
	return fmt.Sprintf("{\"id\": %d, \"userId\": \"%s\", \"isBid\": %t, \"orderType\": \"%s\", \"size\": %.2f, \"limitPrice\": %.2f, \"timestamp\": %d }",
		o.id,
		o.userId,
		o.isBid,
		o.orderType,
		o.Size,
		o.limitPrice,
		o.timestamp)
}

func (o1 *Order) IsBetter(o2 *Order) bool {
	if o1.isBid && o2.isBid {
		return o1.limitPrice > o2.limitPrice
	} else if !o1.isBid && !o2.isBid {
		return o1.limitPrice < o2.limitPrice
	} else {
		// TODO: throw error if not bid-bid/ask-ask
		panic("Cant compare if not bid-bid/ask-ask")
	}
}
func (o *Order) IsFilled() bool {
	return o.Size == float64(0)
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
func (o Order) GetUserId() string {
	return o.userId
}
func (o Order) GetTicker() string {
	return o.ticker
}
func (o *Order) GetIsBid() bool {
	return o.isBid
}
func (o Order) GetOrderType() OrderType {
	return o.orderType
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

// TODO: user need crypto wallet
type User struct {
	UserId  string
	Balance map[string]float64
}
