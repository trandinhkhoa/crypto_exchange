package main

import (
	"reflect"
	"testing"
)

func assert(t *testing.T, a, b any) {
	if !reflect.DeepEqual(a, b) {
		t.Errorf("%+v != %+v", a, b)
	}
}

func TestLimit(t *testing.T) {
	l := NewLimit(10_000)
	buyOrderA := NewOrder(true, 5)
	buyOrderB := NewOrder(true, 8)
	buyOrderC := NewOrder(true, 10)

	l.AddOrder(buyOrderA)
	l.AddOrder(buyOrderB)
	l.AddOrder(buyOrderC)

	l.DeleteOrder(buyOrderB)

	a := 2
	b := 2
	assert(t, a, b)
	assert(t, l.price, float64(10_000))
	assert(t, l.totalVolume, float64(15))
	assert(t, len(l.Orders), 2)
	assert(t, l.Orders[0].size, float64(5))
	assert(t, l.Orders[1].size, float64(10))
}

func TestOrderBookPlaceLimitOrder(t *testing.T) {
	ob := NewOrderbook()

	sellOrder := NewOrder(false, 20)
	ob.placeLimitOrder(10_000, sellOrder)

	sellOrder = NewOrder(false, 30)
	ob.placeLimitOrder(20_000, sellOrder)

	sellOrder = NewOrder(false, 50)
	ob.placeLimitOrder(20_000, sellOrder)

	assert(t, ob.askLimits.Len(), 2)
	assert(t, len(ob.priceToAsksMap), 2)
	assert(t, ob.getTotalVolumeAllAsks(), float64(100))
}

func TestLimitsInterface(t *testing.T) {
	var orderbook OrderBook
	orderbook.askLimits = make(AskLimitsInterface, 0)
	limit1 := NewLimit(10000)
	orderbook.askLimits = append(orderbook.askLimits, limit1)
	limit2 := NewLimit(20000)
	orderbook.askLimits = append(orderbook.askLimits, limit2)

	assert(t, orderbook.askLimits.Len(), 2)
	assert(t, orderbook.askLimits[0].price, float64(10000))
	assert(t, orderbook.askLimits[1].price, float64(20000))
	orderbook.askLimits.Swap(0, 1)
	assert(t, orderbook.askLimits[0].price, float64(20000))
	assert(t, orderbook.askLimits[1].price, float64(10000))
	limit3 := NewLimit(30000)
	orderbook.askLimits = append(orderbook.askLimits, limit3)
	orderbook.sortAskLimits()
	assert(t, orderbook.askLimits[0].price, float64(10000))
	assert(t, orderbook.askLimits[1].price, float64(20000))
	assert(t, orderbook.askLimits[2].price, float64(30000))
}

func TestPlaceMarketOrderNotEnoughLiquidity(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	ob := NewOrderbook()

	// ask
	sellOrder := NewOrder(false, 20)
	ob.placeLimitOrder(10_000, sellOrder)

	// bid
	buyOrder := NewOrder(true, 400)
	ob.placeMarketOrder(buyOrder)
}

func TestPlaceMarketOrderBid(t *testing.T) {
	ob := NewOrderbook()

	// ask
	sellOrder := NewOrder(false, 20)
	ob.placeLimitOrder(10_000, sellOrder)

	// bid
	buyOrder := NewOrder(true, 4)
	matches := ob.placeMarketOrder(buyOrder)

	assert(t, len(matches), 1)

	// the smaller one is filled first
	assert(t, matches[0].sizeFilled, 4.0)
	// the bigger one is partially filled
	assert(t, ob.getTotalVolumeAllAsks(), 16.0)
	// 1 ask as before, it's still there with a reduced size since not enough bidder yet
	assert(t, len(ob.askLimits), 1)

	assert(t, matches[0].ask, sellOrder)
	assert(t, matches[0].bid, buyOrder)
	assert(t, matches[0].price, 10_000.0)
	assert(t, buyOrder.isFilled(), true)
}

func TestPlaceMarketOrderAskMultiFill(t *testing.T) {
	ob := NewOrderbook()

	// bid
	buyOrder0 := NewOrder(true, 25)
	ob.placeLimitOrder(5_000, buyOrder0)
	buyOrder1 := NewOrder(true, 10)
	ob.placeLimitOrder(15_000, buyOrder1)
	buyOrder2 := NewOrder(true, 15)
	ob.placeLimitOrder(10_000, buyOrder2)
	buyOrder3 := NewOrder(true, 20)
	ob.placeLimitOrder(10_000, buyOrder3)

	// ask
	sellOrder := NewOrder(false, 30)
	matches := ob.placeMarketOrder(sellOrder)

	assert(t, len(matches), 3)

	// the smaller one is filled first
	assert(t, matches[0].sizeFilled, 10.0)
	assert(t, matches[1].sizeFilled, 15.0)
	assert(t, matches[2].sizeFilled, 5.0)
	// the bigger one is partially filled
	assert(t, ob.getTotalVolumeAllBids(), 40.0)
	// 2 bid: 1 whose size reduced from 20 to 5, 1 untouched of size 25
	assert(t, len(ob.bidLimits), 2)

	assert(t, matches[0].ask, sellOrder)
	assert(t, matches[0].bid, buyOrder1)
	assert(t, matches[0].price, 15_000.0)
	assert(t, matches[0].sizeFilled, 10.0)

	assert(t, matches[1].ask, sellOrder)
	assert(t, matches[1].bid, buyOrder2)
	assert(t, matches[1].price, 10_000.0)
	assert(t, matches[1].sizeFilled, 15.0)

	assert(t, matches[2].ask, sellOrder)
	assert(t, matches[2].bid, buyOrder3)
	assert(t, matches[2].price, 10_000.0)
	assert(t, matches[2].sizeFilled, 5.0)

	assert(t, sellOrder.isFilled(), true)
}
