package entities

import "time"

type Trade struct {
	buyer        *Order
	seller       *Order
	Price        float64
	Size         float64
	isBuyerMaker bool
	Timestamp    int64
}

// TODO: dont return order as it contains pointer,
// as w/ the pointer, the content of h field can be modified
// defeat the point of encapsulation
func (t Trade) GetBuyer() Order {
	return *t.buyer
}

func (t Trade) GetSeller() Order {
	return *t.seller
}

func (t Trade) GetPrice() float64 {
	return t.Price
}

func (t Trade) GetSize() float64 {
	return t.Size
}

func (t Trade) GetIsBuyerMaker() bool {
	return t.isBuyerMaker
}

func (t Trade) GetTimeStamp() int64 {
	return t.Timestamp
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
		Price:        price,
		Size:         size,
		isBuyerMaker: isBuyerMaker,
		Timestamp:    time.Now().UnixNano(),
	}
}
