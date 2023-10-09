package entities

import (
	"fmt"
	"time"
)

type Trade struct {
	buyer        *Order
	seller       *Order
	price        float64
	size         float64
	isBuyerMaker bool
	timestamp    int64
}

// TODO: the comment below is not really an issue since the return value is a dereference pointer, essentially a copy
// dont return order as it contains pointer, as w/ the pointer, the content of h field can be modified
// defeat the point of encapsulation
func (t Trade) GetBuyer() Order {
	return *t.buyer
}

func (t Trade) GetSeller() Order {
	return *t.seller
}

func (t Trade) GetPrice() float64 {
	return t.price
}

func (t Trade) GetSize() float64 {
	return t.size
}

func (t Trade) GetIsBuyerMaker() bool {
	return t.isBuyerMaker
}

func (t Trade) GetTimeStamp() int64 {
	return t.timestamp
}

func NewTrade(
	buyer *Order,
	seller *Order,
	price float64,
	size float64,
	isBuyerMaker bool,
) *Trade {
	return &Trade{
		buyer:        buyer,
		seller:       seller,
		price:        price,
		size:         size,
		isBuyerMaker: isBuyerMaker,
		timestamp:    time.Now().UnixNano(),
	}
}

func NewTradeWithTimeStamp(
	buyer *Order,
	seller *Order,
	price float64,
	size float64,
	isBuyerMaker bool,
	timestamp int64,
) *Trade {
	return &Trade{
		buyer:        buyer,
		seller:       seller,
		price:        price,
		size:         size,
		isBuyerMaker: isBuyerMaker,
		timestamp:    timestamp,
	}
}

func (t Trade) String() string {
	str := fmt.Sprintf("{\"buyerOrderId\": %d, \"buyerUserId\": \"%s\", \"sellerOrderId\": %d, \"sellerUserId\": \"%s\", \"price\": %.2f, \"size\": %.2f}",
		t.buyer.id,
		t.buyer.userId,
		t.seller.id,
		t.seller.userId,
		t.price,
		t.size)
	return str

}
