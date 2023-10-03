package entities

import "fmt"

// for each price level(limit) we need to know total volume and the corresponding orders
type Limit struct {
	limitPrice  float64
	totalVolume float64
	parent      *Limit
	leftChild   *Limit
	rightChild  *Limit
	headOrder   *Order
	tailOrder   *Order
}

func (l Limit) GetTotalVolume() float64 {
	return l.totalVolume
}

func (l Limit) String() string {
	str := fmt.Sprintf("{\"limitPrice\": %.2f, \"totalVolume\": %.2f, \"orders\":", l.GetLimitPrice(), l.totalVolume)
	str += "["
	iterator := l.headOrder
	for iterator != nil {
		str += iterator.String()
		iterator = iterator.nextOrder
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

func (l Limit) GetLimitPrice() float64 {
	return l.limitPrice
}

func (l *Limit) AddOrder(newOrder *Order) {
	if l.tailOrder == nil {
		// if empty
		l.headOrder = newOrder
		l.tailOrder = newOrder
	} else {
		l.tailOrder.nextOrder = newOrder
		newOrder.prevOrder = l.tailOrder
		l.tailOrder = l.tailOrder.nextOrder
	}
	newOrder.parentLimit = l
	l.totalVolume += float64(newOrder.Size)
	// TODO: throw error if sell limit but o is bid
}

func (l *Limit) DeleteOrderById(id int64) {
	iterator := l.headOrder
	for iterator != nil {
		if iterator.id == id {
			break
		}
		iterator = iterator.nextOrder
	}
	l.deleteOrder(iterator)
}

func (l *Limit) deleteOrder(order *Order) {
	l.totalVolume -= order.Size
	if (order.prevOrder == nil) && (order.nextOrder == nil) {
		// if the only one left
		l.headOrder = nil
		l.tailOrder = nil
	} else if order.prevOrder == nil {
		// if deleting the head
		l.headOrder = order.nextOrder
		l.headOrder.prevOrder = nil
	} else {
		prevOrder := order.prevOrder
		nextOrder := order.nextOrder
		prevOrder.nextOrder = nextOrder
		if nextOrder != nil {
			// if not the last one
			nextOrder.prevOrder = prevOrder
		}
	}
}

func (l Limit) GetAllOrders() []Order {
	ordersList := make([]Order, 0)
	iterator := l.headOrder
	for iterator != nil {
		ordersList = append(ordersList, *iterator)
		iterator = iterator.nextOrder
	}

	return ordersList
}
