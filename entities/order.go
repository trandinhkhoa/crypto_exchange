package entities

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/sirupsen/logrus"
)

// TODO: user need crypto wallet
type User struct {
	userId  string
	Balance map[string]float64
}

func (u User) GetUserId() string {
	return u.userId
}

func NewUser(userId string, balance map[string]float64) *User {
	return &User{
		userId:  userId,
		Balance: balance,
	}
}

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
	nextOrder   *Order
	prevOrder   *Order
	parentLimit *Limit
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
func (o Order) String() string {
	return fmt.Sprintf("{\"id\": %d, \"userId\": \"%s\", \"isBid\": %t, \"orderType\": \"%s\", \"size\": %.2f, \"limitPrice\": %.2f, \"timestamp\": %d }",
		o.id,
		o.userId,
		o.isBid,
		o.orderType,
		o.Size,
		o.limitPrice,
		o.timestamp)
}

func (o1 Order) IsBetter(o2 Order) bool {
	if o1.isBid && o2.isBid {
		return o1.limitPrice > o2.limitPrice
	} else if !o1.isBid && !o2.isBid {
		return o1.limitPrice < o2.limitPrice
	} else {
		logrus.Warn("Cant compare if not bid-bid/ask-ask")
		return false
	}
}

func (o Order) IsFilled() bool {
	return o.Size == float64(0)
}

func (o Order) GetId() int64 {
	return o.id
}

func (o Order) GetUserId() string {
	return o.userId
}
func (o Order) GetTicker() string {
	return o.ticker
}
func (o Order) GetIsBid() bool {
	return o.isBid
}
func (o Order) GetOrderType() OrderType {
	return o.orderType
}
func (o Order) GetLimitPrice() float64 {
	return o.limitPrice
}
func (o Order) GetTimeStamp() int64 {
	return o.timestamp
}
func (o Order) GetSize() float64 {
	return o.Size
}
