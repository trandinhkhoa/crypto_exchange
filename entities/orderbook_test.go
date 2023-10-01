package entities_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/trandinhkhoa/crypto-exchange/entities"
)

func limitArrToJson(ll []*entities.Limit) string {
	str := "["
	for index, limit := range ll {
		str += limit.String()
		if index < len(ll)-1 {
			str += ","
		}
	}
	str += "]"
	return str
}

func TestPlaceLimitOrder(t *testing.T) {
	ob := entities.NewOrderbook()

	// Root: 1000
	// L--- 1005
	//     L--- 1100
	// R--- 900
	incomingOrder := entities.NewOrder("john", "ticker", true, entities.LimitOrderType, 1, 1000)
	ob.PlaceLimitOrder(*incomingOrder)

	incomingOrder = entities.NewOrder("jim", "ticker", true, entities.LimitOrderType, 1, 900)
	ob.PlaceLimitOrder(*incomingOrder)

	incomingOrder = entities.NewOrder("jane", "ticker", true, entities.LimitOrderType, 4, 1100)
	ob.PlaceLimitOrder(*incomingOrder)

	incomingOrder = entities.NewOrder("jun", "ticker", true, entities.LimitOrderType, 9, 1005)
	ob.PlaceLimitOrder(*incomingOrder)

	incomingOrder = entities.NewOrder("jack", "ticker", true, entities.LimitOrderType, 9, 1005)
	ob.PlaceLimitOrder(*incomingOrder)

	assert.Equal(t, ob.HighestBuy.GetTotalVolume(), 4.0)
	assert.Equal(t, ob.HighestBuy.GetLimitPrice(), 1100.0)

	arr := entities.TreeToArray(ob.BuyTree)
	orderList := make([]entities.Order, 0)
	for index := 0; index < len(arr); index++ {
		if (index < len(arr)-1) && (arr[index].GetLimitPrice() <= arr[index+1].GetLimitPrice()) {
			t.Errorf("Buy Limits not sorted in descending order ")
			break
		}
		orderList = append(orderList, arr[index].GetAllOrders()...)
	}

	assert.Equal(t, "jane", orderList[0].GetUserId())
	assert.Equal(t, "jun", orderList[1].GetUserId())
	assert.Equal(t, "jack", orderList[2].GetUserId())
	assert.Equal(t, "john", orderList[3].GetUserId())
	assert.Equal(t, "jim", orderList[4].GetUserId())
}

func TestPlaceLimitOrderSell(t *testing.T) {
	ob := entities.NewOrderbook()

	// Root: 1000
	// L--- 900
	// R--- 1005
	//     R--- 1100

	incomingOrder := entities.NewOrder("john", "ticker", false, entities.LimitOrderType, 1, 1000)
	ob.PlaceLimitOrder(*incomingOrder)

	incomingOrder = entities.NewOrder("jim", "ticker", false, entities.LimitOrderType, 1, 900)
	ob.PlaceLimitOrder(*incomingOrder)

	incomingOrder = entities.NewOrder("jane", "ticker", false, entities.LimitOrderType, 4, 1100)
	ob.PlaceLimitOrder(*incomingOrder)

	incomingOrder = entities.NewOrder("jun", "ticker", false, entities.LimitOrderType, 9, 1005)
	ob.PlaceLimitOrder(*incomingOrder)

	incomingOrder = entities.NewOrder("jack", "ticker", false, entities.LimitOrderType, 9, 1005)
	ob.PlaceLimitOrder(*incomingOrder)

	arr := entities.TreeToArray(ob.SellTree)
	orderList := make([]entities.Order, 0)
	for index := 0; index < len(arr); index++ {
		if (index < len(arr)-1) && (arr[index].GetLimitPrice() >= arr[index+1].GetLimitPrice()) {
			t.Errorf("Sell Limits not sorted in ascending order ")
			break
		}
		orderList = append(orderList, arr[index].GetAllOrders()...)
	}

	assert.Equal(t, "jim", orderList[0].GetUserId())
	assert.Equal(t, "john", orderList[1].GetUserId())
	assert.Equal(t, "jun", orderList[2].GetUserId())
	assert.Equal(t, "jack", orderList[3].GetUserId())
	assert.Equal(t, "jane", orderList[4].GetUserId())
}
func TestPlaceMarketOrderBuyOneFill(t *testing.T) {
	// Root: 1000
	// L--- 900
	// R--- 1005
	//     R--- 1100
	ob := entities.NewOrderbook()

	incomingOrder := entities.NewOrder("john", "ticker", false, entities.LimitOrderType, 1, 1000)
	ob.PlaceLimitOrder(*incomingOrder)

	incomingOrder = entities.NewOrder("jim", "ticker", false, entities.LimitOrderType, 1, 900)
	ob.PlaceLimitOrder(*incomingOrder)

	incomingOrder = entities.NewOrder("jane", "ticker", false, entities.LimitOrderType, 4, 1100)
	ob.PlaceLimitOrder(*incomingOrder)

	incomingOrder = entities.NewOrder("jun", "ticker", false, entities.LimitOrderType, 9, 1005)
	ob.PlaceLimitOrder(*incomingOrder)

	incomingOrder = entities.NewOrder("jack", "ticker", false, entities.LimitOrderType, 9, 1005)
	ob.PlaceLimitOrder(*incomingOrder)
	assert.Equal(t, ob.GetTotalVolumeAllSells(), 24.0)

	incomingOrder = entities.NewOrder("lily", "ticker", true, entities.MarketOrderType, 1, 0)
	tradesArray := ob.PlaceMarketOrder(*incomingOrder)
	assert.Equal(t, ob.GetTotalVolumeAllSells(), 23.0)
	assert.Equal(t, ob.LowestSell.GetLimitPrice(), 1000.0)

	assert.Equal(t, tradesArray[0].GetBuyer().GetUserId(), "lily")
	assert.Equal(t, tradesArray[0].GetSeller().GetUserId(), "jim")
	assert.Equal(t, tradesArray[0].GetPrice(), 900.0)
	assert.Equal(t, tradesArray[0].GetSize(), 1.0)

	arr := entities.TreeToArray(ob.SellTree)
	orderList := make([]entities.Order, 0)
	for index := 0; index < len(arr); index++ {
		if (index < len(arr)-1) && (arr[index].GetLimitPrice() >= arr[index+1].GetLimitPrice()) {
			t.Errorf("Sell Limits not sorted in ascending order ")
			break
		}
		orderList = append(orderList, arr[index].GetAllOrders()...)
	}

	assert.Equal(t, "john", orderList[0].GetUserId())
	assert.Equal(t, "jun", orderList[1].GetUserId())
	assert.Equal(t, "jack", orderList[2].GetUserId())
	assert.Equal(t, "jane", orderList[3].GetUserId())
}

func TestPlaceMarketOrderBuyOnePartialFill(t *testing.T) {
	ob := entities.NewOrderbook()

	// Root: 1000
	// L--- 900
	// R--- 1005
	//     R--- 1100
	incomingOrder := entities.NewOrder("john", "ticker", false, entities.LimitOrderType, 1, 1000)
	ob.PlaceLimitOrder(*incomingOrder)

	incomingOrder = entities.NewOrder("jim", "ticker", false, entities.LimitOrderType, 1, 900)
	ob.PlaceLimitOrder(*incomingOrder)

	incomingOrder = entities.NewOrder("jane", "ticker", false, entities.LimitOrderType, 4, 1100)
	ob.PlaceLimitOrder(*incomingOrder)

	incomingOrder = entities.NewOrder("jun", "ticker", false, entities.LimitOrderType, 9, 1005)
	ob.PlaceLimitOrder(*incomingOrder)

	incomingOrder = entities.NewOrder("jack", "ticker", false, entities.LimitOrderType, 9, 1005)
	ob.PlaceLimitOrder(*incomingOrder)
	assert.Equal(t, ob.GetTotalVolumeAllSells(), 24.0)

	incomingOrder = entities.NewOrder("lily", "ticker", true, entities.MarketOrderType, 0.5, 0)
	tradesArray := ob.PlaceMarketOrder(*incomingOrder)
	assert.Equal(t, ob.GetTotalVolumeAllSells(), 23.5)
	assert.Equal(t, ob.LowestSell.GetLimitPrice(), 900.0)

	assert.Equal(t, tradesArray[0].GetBuyer().GetUserId(), "lily")
	assert.Equal(t, tradesArray[0].GetSeller().GetUserId(), "jim")
	assert.Equal(t, tradesArray[0].GetPrice(), 900.0)
	assert.Equal(t, tradesArray[0].GetSize(), 0.5)
}

func TestPlaceMarketOrderBuyMultiFill(t *testing.T) {
	ob := entities.NewOrderbook()

	// Root: 1000
	// L--- 900
	// R--- 1005
	//     R--- 1100

	incomingOrder := entities.NewOrder("john", "ticker", false, entities.LimitOrderType, 1, 1000)
	ob.PlaceLimitOrder(*incomingOrder)

	incomingOrder = entities.NewOrder("jim", "ticker", false, entities.LimitOrderType, 1, 900)
	ob.PlaceLimitOrder(*incomingOrder)

	incomingOrder = entities.NewOrder("jane", "ticker", false, entities.LimitOrderType, 4, 1100)
	ob.PlaceLimitOrder(*incomingOrder)

	incomingOrder = entities.NewOrder("jun", "ticker", false, entities.LimitOrderType, 9, 1005)
	ob.PlaceLimitOrder(*incomingOrder)

	incomingOrder = entities.NewOrder("jack", "ticker", false, entities.LimitOrderType, 9, 1005)
	ob.PlaceLimitOrder(*incomingOrder)
	assert.Equal(t, ob.GetTotalVolumeAllSells(), 24.0)

	incomingOrder = entities.NewOrder("lily", "ticker", true, entities.MarketOrderType, 23.5, 0)
	tradesArray := ob.PlaceMarketOrder(*incomingOrder)
	assert.Equal(t, ob.GetTotalVolumeAllSells(), 0.5)
	assert.Equal(t, ob.LowestSell.GetLimitPrice(), 1100.0)

	// check price priority
	assert.Equal(t, tradesArray[0].GetSeller().GetUserId(), "jim")
	assert.Equal(t, tradesArray[1].GetSeller().GetUserId(), "john")
	// check time priority jun > jack
	assert.Equal(t, tradesArray[2].GetSeller().GetUserId(), "jun")
	assert.Equal(t, tradesArray[3].GetSeller().GetUserId(), "jack")
	assert.Equal(t, tradesArray[4].GetSeller().GetUserId(), "jane")

	incomingOrder = entities.NewOrder("lily", "ticker", true, entities.MarketOrderType, 0.5, 0)
	tradesArray = ob.PlaceMarketOrder(*incomingOrder)
	assert.Equal(t, ob.GetTotalVolumeAllSells(), 0.0)
	// check returned trades array
	assert.Equal(t, tradesArray[0].GetSeller().GetUserId(), "jane")
	assert.Equal(t, tradesArray[0].GetSize(), 0.5)
	assert.Equal(t, tradesArray[0].GetPrice(), 1100.0)

	// check if orderbook recorded all trades across all placed market order
	assert.Equal(t, ob.GetLastTrades()[0].GetSeller().GetUserId(), "jim")
	assert.Equal(t, ob.GetLastTrades()[1].GetSeller().GetUserId(), "john")
	assert.Equal(t, ob.GetLastTrades()[2].GetSeller().GetUserId(), "jun")
	assert.Equal(t, ob.GetLastTrades()[3].GetSeller().GetUserId(), "jack")
	assert.Equal(t, ob.GetLastTrades()[4].GetSeller().GetUserId(), "jane")
	assert.Equal(t, ob.GetLastTrades()[5].GetSeller().GetUserId(), "jane")

	arr := entities.TreeToArray(ob.SellTree)
	assert.Equal(t, len(arr) == 0, true)

	assert.Equal(t, ob.LowestSell == nil, true)
}

func TestPlaceMarketOrderSellMultiFill(t *testing.T) {
	// Root: 1000
	// L--- 1005
	//     L--- 1100
	// R--- 900
	ob := entities.NewOrderbook()

	incomingOrder := entities.NewOrder("john", "ticker", true, entities.LimitOrderType, 1, 1000)
	ob.PlaceLimitOrder(*incomingOrder)

	incomingOrder = entities.NewOrder("jim", "ticker", true, entities.LimitOrderType, 1, 900)
	ob.PlaceLimitOrder(*incomingOrder)

	incomingOrder = entities.NewOrder("jane", "ticker", true, entities.LimitOrderType, 4, 1100)
	ob.PlaceLimitOrder(*incomingOrder)

	incomingOrder = entities.NewOrder("jun", "ticker", true, entities.LimitOrderType, 9, 1005)
	ob.PlaceLimitOrder(*incomingOrder)

	incomingOrder = entities.NewOrder("jack", "ticker", true, entities.LimitOrderType, 9, 1005)
	ob.PlaceLimitOrder(*incomingOrder)
	assert.Equal(t, ob.GetTotalVolumeAllBuys(), 24.0)

	incomingOrder = entities.NewOrder("lily", "ticker", false, entities.MarketOrderType, 23.5, 0)
	tradesArray := ob.PlaceMarketOrder(*incomingOrder)
	assert.Equal(t, ob.GetTotalVolumeAllBuys(), 0.5)
	assert.Equal(t, ob.HighestBuy.GetLimitPrice(), 900.0)

	// check price priority
	assert.Equal(t, tradesArray[0].GetBuyer().GetUserId(), "jane")
	assert.Equal(t, tradesArray[1].GetBuyer().GetUserId(), "jun")
	// 	// check time priority jun > jack
	assert.Equal(t, tradesArray[2].GetBuyer().GetUserId(), "jack")
	assert.Equal(t, tradesArray[3].GetBuyer().GetUserId(), "john")
	assert.Equal(t, tradesArray[4].GetBuyer().GetUserId(), "jim")

	incomingOrder = entities.NewOrder("lily", "ticker", false, entities.MarketOrderType, 0.5, 0)
	tradesArray = ob.PlaceMarketOrder(*incomingOrder)
	assert.Equal(t, ob.GetTotalVolumeAllBuys(), 0.0)

	// check returned trades array
	assert.Equal(t, tradesArray[0].GetBuyer().GetUserId(), "jim")
	assert.Equal(t, tradesArray[0].GetSize(), 0.5)
	assert.Equal(t, tradesArray[0].GetPrice(), 900.0)

	// check if orderbook recorded all trades across all placed market order
	assert.Equal(t, ob.GetLastTrades()[0].GetBuyer().GetUserId(), "jane")
	assert.Equal(t, ob.GetLastTrades()[1].GetBuyer().GetUserId(), "jun")
	assert.Equal(t, ob.GetLastTrades()[2].GetBuyer().GetUserId(), "jack")
	assert.Equal(t, ob.GetLastTrades()[3].GetBuyer().GetUserId(), "john")
	assert.Equal(t, ob.GetLastTrades()[4].GetBuyer().GetUserId(), "jim")
	assert.Equal(t, ob.GetLastTrades()[5].GetBuyer().GetUserId(), "jim")

	arr := entities.TreeToArray(ob.BuyTree)
	assert.Equal(t, len(arr) == 0, true)

	assert.Equal(t, ob.HighestBuy == nil, true)
}

func TestCancelOrderSimple(t *testing.T) {
	ob := entities.NewOrderbook()

	// Root: 1000

	incomingOrder := entities.NewOrder("john", "ticker", true, entities.LimitOrderType, 1, 1000)
	ob.PlaceLimitOrder(*incomingOrder)

	assert.Equal(t, 1.0, ob.GetTotalVolumeAllBuys(), 1.0)
	assert.Equal(t, 1000.0, ob.HighestBuy.GetLimitPrice())

	ob.CancelOrder(incomingOrder.GetId())

	assert.Equal(t, 0.0, ob.GetTotalVolumeAllBuys())
	assert.True(t, ob.HighestBuy == nil)
}

func TestCancelOrderBigTree(t *testing.T) {
	ob := entities.NewOrderbook()

	// Root: 7
	// L--- 3 (x2)
	//     L--- 1
	//         R--- 2
	//     R--- 5
	//         L--- 4
	//         R--- 6
	// R--- 11
	//     L--- 9
	//         L--- 8
	//         R--- 10
	//     R--- 13
	//         L--- 12
	//         R--- 14

	// Adding orders with limit prices from 1 to 14 (from the BST)
	johnOrder := entities.NewOrder("john", "ticker", false, entities.LimitOrderType, 1, 7)
	ob.PlaceLimitOrder(*johnOrder)

	jimOrder := entities.NewOrder("jim", "ticker", false, entities.LimitOrderType, 1, 3)
	ob.PlaceLimitOrder(*jimOrder)

	janeOrder := entities.NewOrder("jane", "ticker", false, entities.LimitOrderType, 4, 11)
	ob.PlaceLimitOrder(*janeOrder)

	junOrder := entities.NewOrder("jun", "ticker", false, entities.LimitOrderType, 9, 9)
	ob.PlaceLimitOrder(*junOrder)

	jackOrder := entities.NewOrder("jack", "ticker", false, entities.LimitOrderType, 9, 9)
	ob.PlaceLimitOrder(*jackOrder)

	jerryOrder := entities.NewOrder("jerry", "ticker", false, entities.LimitOrderType, 2, 1)
	ob.PlaceLimitOrder(*jerryOrder)

	jessicaOrder := entities.NewOrder("jessica", "ticker", false, entities.LimitOrderType, 2, 2)
	ob.PlaceLimitOrder(*jessicaOrder)

	jessicaTwinOrder := entities.NewOrder("jessicaTwin", "ticker", false, entities.LimitOrderType, 3, 2)
	ob.PlaceLimitOrder(*jessicaTwinOrder)

	jillOrder := entities.NewOrder("jill", "ticker", false, entities.LimitOrderType, 3, 4)
	ob.PlaceLimitOrder(*jillOrder)

	jeffOrder := entities.NewOrder("jeff", "ticker", false, entities.LimitOrderType, 3, 5)
	ob.PlaceLimitOrder(*jeffOrder)

	jacobOrder := entities.NewOrder("jacob", "ticker", false, entities.LimitOrderType, 4, 6)
	ob.PlaceLimitOrder(*jacobOrder)

	julieOrder := entities.NewOrder("julie", "ticker", false, entities.LimitOrderType, 4, 8)
	ob.PlaceLimitOrder(*julieOrder)

	jamesOrder := entities.NewOrder("james", "ticker", false, entities.LimitOrderType, 5, 10)
	ob.PlaceLimitOrder(*jamesOrder)

	joanOrder := entities.NewOrder("joan", "ticker", false, entities.LimitOrderType, 5, 12)
	ob.PlaceLimitOrder(*joanOrder)

	jamieOrder := entities.NewOrder("jamie", "ticker", false, entities.LimitOrderType, 6, 13)
	ob.PlaceLimitOrder(*jamieOrder)

	jodieOrder := entities.NewOrder("jodie", "ticker", false, entities.LimitOrderType, 6, 14)
	ob.PlaceLimitOrder(*jodieOrder)

	assert.Equal(t, 67.0, ob.GetTotalVolumeAllSells())

	assert.Equal(t, 1.0, ob.LowestSell.GetLimitPrice())

	ob.CancelOrder(jerryOrder.GetId())
	assert.Equal(t, 2.0, ob.LowestSell.GetLimitPrice())
	assert.Equal(t, 65.0, ob.GetTotalVolumeAllSells())

	ob.CancelOrder(jessicaOrder.GetId())
	assert.Equal(t, 2.0, ob.LowestSell.GetLimitPrice())
	assert.Equal(t, 63.0, ob.GetTotalVolumeAllSells())

	ob.CancelOrder(jessicaTwinOrder.GetId())
	assert.Equal(t, 3.0, ob.LowestSell.GetLimitPrice())
	assert.Equal(t, 60.0, ob.GetTotalVolumeAllSells())

	ob.CancelOrder(jodieOrder.GetId())
	assert.Equal(t, 3.0, ob.LowestSell.GetLimitPrice())
	assert.Equal(t, 54.0, ob.GetTotalVolumeAllSells())

	ob.CancelOrder(julieOrder.GetId())
	assert.Equal(t, 3.0, ob.LowestSell.GetLimitPrice())
	assert.Equal(t, 50.0, ob.GetTotalVolumeAllSells())

	ob.CancelOrder(jamesOrder.GetId())
	assert.Equal(t, 3.0, ob.LowestSell.GetLimitPrice())
	assert.Equal(t, 45.0, ob.GetTotalVolumeAllSells())

	arr := entities.TreeToArray(ob.SellTree)
	for index := 0; index < len(arr)-1; index++ {
		if arr[index].GetLimitPrice() >= arr[index+1].GetLimitPrice() {
			t.Errorf("Sell Limit not sorted in ascending order ")
			break
		}
	}
}

func TestGetKBestBuys(t *testing.T) {
	ob := entities.NewOrderbook()

	// Root: 7
	// L--- 3 (x2)
	//     L--- 1
	//         R--- 2
	//     R--- 5
	//         L--- 4
	//         R--- 6
	// R--- 11
	//     L--- 9
	//         L--- 8
	//         R--- 10
	//     R--- 13
	//         L--- 12
	//         R--- 14

	// Adding orders with limit prices from 1 to 14 (from the BST)
	johnOrder := entities.NewOrder("john", "ticker", false, entities.LimitOrderType, 1, 7)
	ob.PlaceLimitOrder(*johnOrder)

	jimOrder := entities.NewOrder("jim", "ticker", false, entities.LimitOrderType, 1, 3)
	ob.PlaceLimitOrder(*jimOrder)

	janeOrder := entities.NewOrder("jane", "ticker", false, entities.LimitOrderType, 4, 11)
	ob.PlaceLimitOrder(*janeOrder)

	junOrder := entities.NewOrder("jun", "ticker", false, entities.LimitOrderType, 9, 9)
	ob.PlaceLimitOrder(*junOrder)

	jackOrder := entities.NewOrder("jack", "ticker", false, entities.LimitOrderType, 9, 9)
	ob.PlaceLimitOrder(*jackOrder)

	jerryOrder := entities.NewOrder("jerry", "ticker", false, entities.LimitOrderType, 2, 1)
	ob.PlaceLimitOrder(*jerryOrder)

	jessicaOrder := entities.NewOrder("jessica", "ticker", false, entities.LimitOrderType, 2, 2)
	ob.PlaceLimitOrder(*jessicaOrder)

	jillOrder := entities.NewOrder("jill", "ticker", false, entities.LimitOrderType, 3, 4)
	ob.PlaceLimitOrder(*jillOrder)

	jeffOrder := entities.NewOrder("jeff", "ticker", false, entities.LimitOrderType, 3, 5)
	ob.PlaceLimitOrder(*jeffOrder)

	jacobOrder := entities.NewOrder("jacob", "ticker", false, entities.LimitOrderType, 4, 6)
	ob.PlaceLimitOrder(*jacobOrder)

	julieOrder := entities.NewOrder("julie", "ticker", false, entities.LimitOrderType, 4, 8)
	ob.PlaceLimitOrder(*julieOrder)

	jamesOrder := entities.NewOrder("james", "ticker", false, entities.LimitOrderType, 5, 10)
	ob.PlaceLimitOrder(*jamesOrder)

	joanOrder := entities.NewOrder("joan", "ticker", false, entities.LimitOrderType, 5, 12)
	ob.PlaceLimitOrder(*joanOrder)

	jamieOrder := entities.NewOrder("jamie", "ticker", false, entities.LimitOrderType, 6, 13)
	ob.PlaceLimitOrder(*jamieOrder)

	jodieOrder := entities.NewOrder("jodie", "ticker", false, entities.LimitOrderType, 6, 14)
	ob.PlaceLimitOrder(*jodieOrder)

	assert.Equal(t, 64.0, ob.GetTotalVolumeAllSells())

	assert.Equal(t, 1.0, ob.LowestSell.GetLimitPrice())

	arr := ob.GetBestLimits(ob.SellTree, 3)
	assert.Equal(t, 3, len(arr))
	assert.Equal(t, 1.0, arr[0].GetLimitPrice())
	assert.Equal(t, 2.0, arr[1].GetLimitPrice())
	assert.Equal(t, 3.0, arr[2].GetLimitPrice())
}
