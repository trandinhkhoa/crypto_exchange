package usecases_test

import (
	"reflect"
	"testing"

	"github.com/trandinhkhoa/crypto-exchange/domain"
	"github.com/trandinhkhoa/crypto-exchange/usecases"
)

func assert(t *testing.T, a, b any) {
	if !reflect.DeepEqual(a, b) {
		t.Errorf("%+v != %+v", a, b)
	}
}

func limitArrToJson(ll []*domain.Limit) string {
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
	ob := &usecases.Orderbook{}

	// 1000(*)
	incomingOrder := domain.NewOrder("john", "ticker", true, domain.LimitOrderType, 1, 1000)
	ob.PlaceLimitOrder(*incomingOrder)

	assert(t, ob.HighestBuy.TotalVolume, 1.0)
	assert(t, ob.HighestBuy.GetLimitPrice(), 1000.0)
	assert(t, ob.HighestBuy.HeadOrder.GetUserId(), "john")

	// 1000(*) > 900
	incomingOrder = domain.NewOrder("jim", "ticker", true, domain.LimitOrderType, 1, 900)
	ob.PlaceLimitOrder(*incomingOrder)
	assert(t, ob.HighestBuy.TotalVolume, 1.0)
	assert(t, ob.HighestBuy.GetLimitPrice(), 1000.0)
	assert(t, ob.HighestBuy.HeadOrder.GetUserId(), "john")

	assert(t, ob.BuyTree.RightChild.TotalVolume, 1.0)
	assert(t, ob.BuyTree.RightChild.GetLimitPrice(), 900.0)
	assert(t, ob.BuyTree.RightChild.HeadOrder.GetUserId(), "jim")

	// 1100 > 1000(*) > 900
	incomingOrder = domain.NewOrder("jane", "ticker", true, domain.LimitOrderType, 4, 1100)
	ob.PlaceLimitOrder(*incomingOrder)
	assert(t, ob.HighestBuy.TotalVolume, 4.0)
	assert(t, ob.HighestBuy.GetLimitPrice(), 1100.0)
	assert(t, ob.HighestBuy.HeadOrder.GetUserId(), "jane")

	assert(t, ob.BuyTree.RightChild.TotalVolume, 1.0)
	assert(t, ob.BuyTree.RightChild.GetLimitPrice(), 900.0)
	assert(t, ob.BuyTree.RightChild.HeadOrder.GetUserId(), "jim")

	assert(t, ob.BuyTree.LeftChild.TotalVolume, 4.0)
	assert(t, ob.BuyTree.LeftChild.GetLimitPrice(), 1100.0)
	assert(t, ob.BuyTree.LeftChild.HeadOrder.GetUserId(), "jane")

	// 1100 > 1005 > 1000(*) > 900
	incomingOrder = domain.NewOrder("jun", "ticker", true, domain.LimitOrderType, 9, 1005)
	ob.PlaceLimitOrder(*incomingOrder)
	assert(t, ob.HighestBuy.TotalVolume, 4.0)
	assert(t, ob.HighestBuy.GetLimitPrice(), 1100.0)
	assert(t, ob.HighestBuy.HeadOrder.GetUserId(), "jane")

	assert(t, ob.BuyTree.RightChild.TotalVolume, 1.0)
	assert(t, ob.BuyTree.RightChild.GetLimitPrice(), 900.0)
	assert(t, ob.BuyTree.RightChild.HeadOrder.GetUserId(), "jim")

	assert(t, ob.BuyTree.LeftChild.TotalVolume, 4.0)
	assert(t, ob.BuyTree.LeftChild.GetLimitPrice(), 1100.0)
	assert(t, ob.BuyTree.LeftChild.HeadOrder.GetUserId(), "jane")

	assert(t, ob.BuyTree.LeftChild.RightChild.TotalVolume, 9.0)
	assert(t, ob.BuyTree.LeftChild.RightChild.GetLimitPrice(), 1005.0)
	assert(t, ob.BuyTree.LeftChild.RightChild.HeadOrder.GetUserId(), "jun")

	// 1100 > 1005[2] > 1000(*) > 900
	incomingOrder = domain.NewOrder("jack", "ticker", true, domain.LimitOrderType, 9, 1005)
	ob.PlaceLimitOrder(*incomingOrder)
	assert(t, ob.HighestBuy.TotalVolume, 4.0)
	assert(t, ob.HighestBuy.GetLimitPrice(), 1100.0)
	assert(t, ob.HighestBuy.HeadOrder.GetUserId(), "jane")

	assert(t, ob.BuyTree.RightChild.TotalVolume, 1.0)
	assert(t, ob.BuyTree.RightChild.GetLimitPrice(), 900.0)
	assert(t, ob.BuyTree.RightChild.HeadOrder.GetUserId(), "jim")

	assert(t, ob.BuyTree.LeftChild.TotalVolume, 4.0)
	assert(t, ob.BuyTree.LeftChild.GetLimitPrice(), 1100.0)
	assert(t, ob.BuyTree.LeftChild.HeadOrder.GetUserId(), "jane")

	assert(t, ob.BuyTree.LeftChild.RightChild.TotalVolume, 18.0)
	assert(t, ob.BuyTree.LeftChild.RightChild.GetLimitPrice(), 1005.0)
	assert(t, ob.BuyTree.LeftChild.RightChild.HeadOrder.GetUserId(), "jun")
	assert(t, ob.BuyTree.LeftChild.RightChild.TailOrder.GetUserId(), "jack")

	arr := usecases.TreeToArray(ob.BuyTree)
	for index := 0; index < len(arr)-1; index++ {
		if arr[index].GetLimitPrice() <= arr[index+1].GetLimitPrice() {
			t.Errorf("Buy Limit not sorted in descending order ")
			break
		}
	}
	// fmt.Println(limitArrToJson(arr))
}

func TestPlaceLimitOrderSell(t *testing.T) {
	ob := &usecases.Orderbook{}

	// 1000(*)
	incomingOrder := domain.NewOrder("john", "ticker", false, domain.LimitOrderType, 1, 1000)
	ob.PlaceLimitOrder(*incomingOrder)

	assert(t, ob.LowestSell.TotalVolume, 1.0)
	assert(t, ob.LowestSell.GetLimitPrice(), 1000.0)
	assert(t, ob.LowestSell.HeadOrder.GetUserId(), "john")

	// 900 < 1000(*)
	incomingOrder = domain.NewOrder("jim", "ticker", false, domain.LimitOrderType, 1, 900)
	ob.PlaceLimitOrder(*incomingOrder)
	assert(t, ob.LowestSell.TotalVolume, 1.0)
	assert(t, ob.LowestSell.GetLimitPrice(), 900.0)
	assert(t, ob.LowestSell.HeadOrder.GetUserId(), "jim")

	assert(t, ob.SellTree.LeftChild.TotalVolume, 1.0)
	assert(t, ob.SellTree.LeftChild.GetLimitPrice(), 900.0)
	assert(t, ob.SellTree.LeftChild.HeadOrder.GetUserId(), "jim")

	// 900 < 1000(*) < 1100
	incomingOrder = domain.NewOrder("jane", "ticker", false, domain.LimitOrderType, 4, 1100)
	ob.PlaceLimitOrder(*incomingOrder)
	assert(t, ob.LowestSell.TotalVolume, 1.0)
	assert(t, ob.LowestSell.GetLimitPrice(), 900.0)
	assert(t, ob.LowestSell.HeadOrder.GetUserId(), "jim")

	assert(t, ob.SellTree.LeftChild.TotalVolume, 1.0)
	assert(t, ob.SellTree.LeftChild.GetLimitPrice(), 900.0)
	assert(t, ob.SellTree.LeftChild.HeadOrder.GetUserId(), "jim")

	assert(t, ob.SellTree.RightChild.TotalVolume, 4.0)
	assert(t, ob.SellTree.RightChild.GetLimitPrice(), 1100.0)
	assert(t, ob.SellTree.RightChild.HeadOrder.GetUserId(), "jane")

	// 900 < 1000(*) < 1005 < 1100
	incomingOrder = domain.NewOrder("jun", "ticker", false, domain.LimitOrderType, 9, 1005)
	ob.PlaceLimitOrder(*incomingOrder)
	assert(t, ob.LowestSell.TotalVolume, 1.0)
	assert(t, ob.LowestSell.GetLimitPrice(), 900.0)
	assert(t, ob.LowestSell.HeadOrder.GetUserId(), "jim")

	assert(t, ob.SellTree.LeftChild.TotalVolume, 1.0)
	assert(t, ob.SellTree.LeftChild.GetLimitPrice(), 900.0)
	assert(t, ob.SellTree.LeftChild.HeadOrder.GetUserId(), "jim")

	assert(t, ob.SellTree.RightChild.TotalVolume, 4.0)
	assert(t, ob.SellTree.RightChild.GetLimitPrice(), 1100.0)
	assert(t, ob.SellTree.RightChild.HeadOrder.GetUserId(), "jane")

	assert(t, ob.SellTree.RightChild.LeftChild.TotalVolume, 9.0)
	assert(t, ob.SellTree.RightChild.LeftChild.GetLimitPrice(), 1005.0)
	assert(t, ob.SellTree.RightChild.LeftChild.HeadOrder.GetUserId(), "jun")

	// 900 < 1000(*) < 1005[2] < 1100
	incomingOrder = domain.NewOrder("jack", "ticker", false, domain.LimitOrderType, 9, 1005)
	ob.PlaceLimitOrder(*incomingOrder)
	assert(t, ob.LowestSell.TotalVolume, 1.0)
	assert(t, ob.LowestSell.GetLimitPrice(), 900.0)
	assert(t, ob.LowestSell.HeadOrder.GetUserId(), "jim")

	assert(t, ob.SellTree.LeftChild.TotalVolume, 1.0)
	assert(t, ob.SellTree.LeftChild.GetLimitPrice(), 900.0)
	assert(t, ob.SellTree.LeftChild.HeadOrder.GetUserId(), "jim")

	assert(t, ob.SellTree.RightChild.TotalVolume, 4.0)
	assert(t, ob.SellTree.RightChild.GetLimitPrice(), 1100.0)
	assert(t, ob.SellTree.RightChild.HeadOrder.GetUserId(), "jane")

	assert(t, ob.SellTree.RightChild.LeftChild.TotalVolume, 18.0)
	assert(t, ob.SellTree.RightChild.LeftChild.GetLimitPrice(), 1005.0)
	assert(t, ob.SellTree.RightChild.LeftChild.HeadOrder.GetUserId(), "jun")
	assert(t, ob.SellTree.RightChild.LeftChild.TailOrder.GetUserId(), "jack")

	arr := usecases.TreeToArray(ob.SellTree)
	for index := 0; index < len(arr)-1; index++ {
		if arr[index].GetLimitPrice() >= arr[index+1].GetLimitPrice() {
			t.Errorf("Sell Limit not sorted in ascending order ")
			break
		}
	}
	// fmt.Println(limitArrToJson(arr))
}
func TestPlaceMarketOrderBuyOneFill(t *testing.T) {
	ob := &usecases.Orderbook{}

	// 1000(*)
	incomingOrder := domain.NewOrder("john", "ticker", false, domain.LimitOrderType, 1, 1000)
	ob.PlaceLimitOrder(*incomingOrder)

	// 900 < 1000(*)
	incomingOrder = domain.NewOrder("jim", "ticker", false, domain.LimitOrderType, 1, 900)
	ob.PlaceLimitOrder(*incomingOrder)

	// 900 < 1000(*) < 1100
	incomingOrder = domain.NewOrder("jane", "ticker", false, domain.LimitOrderType, 4, 1100)
	ob.PlaceLimitOrder(*incomingOrder)

	// 900 < 1000(*) < 1005 < 1100
	incomingOrder = domain.NewOrder("jun", "ticker", false, domain.LimitOrderType, 9, 1005)
	ob.PlaceLimitOrder(*incomingOrder)

	// 900 < 1000(*) < 1005[2] < 1100
	incomingOrder = domain.NewOrder("jack", "ticker", false, domain.LimitOrderType, 9, 1005)
	ob.PlaceLimitOrder(*incomingOrder)
	assert(t, ob.GetTotalVolumeAllSells(), 24.0)

	incomingOrder = domain.NewOrder("lily", "ticker", true, domain.MarketOrderType, 1, 0)
	tradesArray := ob.PlaceMarketOrder(*incomingOrder)
	assert(t, ob.GetTotalVolumeAllSells(), 23.0)
	assert(t, ob.LowestSell.GetLimitPrice(), 1000.0)

	assert(t, tradesArray[0].GetBuyer().GetUserId(), "lily")
	assert(t, tradesArray[0].GetSeller().GetUserId(), "jim")
	assert(t, tradesArray[0].GetPrice(), 900.0)
	assert(t, tradesArray[0].GetSize(), 1.0)
}

func TestPlaceMarketOrderBuyOnePartialFill(t *testing.T) {
	ob := &usecases.Orderbook{}

	// 1000(*)
	incomingOrder := domain.NewOrder("john", "ticker", false, domain.LimitOrderType, 1, 1000)
	ob.PlaceLimitOrder(*incomingOrder)

	// 900 < 1000(*)
	incomingOrder = domain.NewOrder("jim", "ticker", false, domain.LimitOrderType, 1, 900)
	ob.PlaceLimitOrder(*incomingOrder)

	// 900 < 1000(*) < 1100
	incomingOrder = domain.NewOrder("jane", "ticker", false, domain.LimitOrderType, 4, 1100)
	ob.PlaceLimitOrder(*incomingOrder)

	// 900 < 1000(*) < 1005 < 1100
	incomingOrder = domain.NewOrder("jun", "ticker", false, domain.LimitOrderType, 9, 1005)
	ob.PlaceLimitOrder(*incomingOrder)

	// 900 < 1000(*) < 1005[2] < 1100
	incomingOrder = domain.NewOrder("jack", "ticker", false, domain.LimitOrderType, 9, 1005)
	ob.PlaceLimitOrder(*incomingOrder)
	assert(t, ob.GetTotalVolumeAllSells(), 24.0)

	incomingOrder = domain.NewOrder("lily", "ticker", true, domain.MarketOrderType, 0.5, 0)
	tradesArray := ob.PlaceMarketOrder(*incomingOrder)
	assert(t, ob.GetTotalVolumeAllSells(), 23.5)
	assert(t, ob.LowestSell.GetLimitPrice(), 900.0)

	assert(t, tradesArray[0].GetBuyer().GetUserId(), "lily")
	assert(t, tradesArray[0].GetSeller().GetUserId(), "jim")
	assert(t, tradesArray[0].GetPrice(), 900.0)
	assert(t, tradesArray[0].GetSize(), 0.5)
}

func TestPlaceMarketOrderBuyMultiFill(t *testing.T) {
	ob := &usecases.Orderbook{}

	// 1000(*)
	incomingOrder := domain.NewOrder("john", "ticker", false, domain.LimitOrderType, 1, 1000)
	ob.PlaceLimitOrder(*incomingOrder)

	// 900 < 1000(*)
	incomingOrder = domain.NewOrder("jim", "ticker", false, domain.LimitOrderType, 1, 900)
	ob.PlaceLimitOrder(*incomingOrder)

	// 900 < 1000(*) < 1100
	incomingOrder = domain.NewOrder("jane", "ticker", false, domain.LimitOrderType, 4, 1100)
	ob.PlaceLimitOrder(*incomingOrder)

	// 900 < 1000(*) < 1005 < 1100
	incomingOrder = domain.NewOrder("jun", "ticker", false, domain.LimitOrderType, 9, 1005)
	ob.PlaceLimitOrder(*incomingOrder)

	// 900 < 1000(*) < 1005[2] < 1100
	incomingOrder = domain.NewOrder("jack", "ticker", false, domain.LimitOrderType, 9, 1005)
	ob.PlaceLimitOrder(*incomingOrder)
	assert(t, ob.GetTotalVolumeAllSells(), 24.0)

	incomingOrder = domain.NewOrder("lily", "ticker", true, domain.MarketOrderType, 23.5, 0)
	tradesArray := ob.PlaceMarketOrder(*incomingOrder)
	assert(t, ob.GetTotalVolumeAllSells(), 0.5)
	assert(t, ob.LowestSell.GetLimitPrice(), 1100.0)

	// check price priority
	assert(t, tradesArray[0].GetSeller().GetUserId(), "jim")
	assert(t, tradesArray[1].GetSeller().GetUserId(), "john")
	// check time priority jun > jack
	assert(t, tradesArray[2].GetSeller().GetUserId(), "jun")
	assert(t, tradesArray[3].GetSeller().GetUserId(), "jack")
	assert(t, tradesArray[4].GetSeller().GetUserId(), "jane")

	incomingOrder = domain.NewOrder("lily", "ticker", true, domain.MarketOrderType, 0.5, 0)
	tradesArray = ob.PlaceMarketOrder(*incomingOrder)
	assert(t, ob.GetTotalVolumeAllSells(), 0.0)
	assert(t, tradesArray[0].GetSeller().GetUserId(), "jane")
	assert(t, tradesArray[0].GetSize(), 0.5)
	assert(t, tradesArray[0].GetPrice(), 1100.0)

	assert(t, ob.GetLastTrades()[0].GetSeller().GetUserId(), "jim")
	assert(t, ob.GetLastTrades()[1].GetSeller().GetUserId(), "john")
	assert(t, ob.GetLastTrades()[2].GetSeller().GetUserId(), "jun")
	assert(t, ob.GetLastTrades()[3].GetSeller().GetUserId(), "jack")
	assert(t, ob.GetLastTrades()[4].GetSeller().GetUserId(), "jane")
	assert(t, ob.GetLastTrades()[5].GetSeller().GetUserId(), "jane")
	assert(t, ob.LowestSell == nil, true)
	assert(t, ob.SellTree == nil, true)
}

func TestPlaceMarketOrderSellMultiFill(t *testing.T) {
	ob := &usecases.Orderbook{}

	// 1000(*)
	incomingOrder := domain.NewOrder("john", "ticker", true, domain.LimitOrderType, 1, 1000)
	ob.PlaceLimitOrder(*incomingOrder)

	// 900 < 1000(*)
	incomingOrder = domain.NewOrder("jim", "ticker", true, domain.LimitOrderType, 1, 900)
	ob.PlaceLimitOrder(*incomingOrder)

	// 900 < 1000(*) < 1100
	incomingOrder = domain.NewOrder("jane", "ticker", true, domain.LimitOrderType, 4, 1100)
	ob.PlaceLimitOrder(*incomingOrder)

	// 900 < 1000(*) < 1005 < 1100
	incomingOrder = domain.NewOrder("jun", "ticker", true, domain.LimitOrderType, 9, 1005)
	ob.PlaceLimitOrder(*incomingOrder)

	// 900 < 1000(*) < 1005[2] < 1100
	incomingOrder = domain.NewOrder("jack", "ticker", true, domain.LimitOrderType, 9, 1005)
	ob.PlaceLimitOrder(*incomingOrder)
	assert(t, ob.GetTotalVolumeAllBuys(), 24.0)

	incomingOrder = domain.NewOrder("lily", "ticker", false, domain.MarketOrderType, 23.5, 0)
	tradesArray := ob.PlaceMarketOrder(*incomingOrder)
	assert(t, ob.GetTotalVolumeAllBuys(), 0.5)
	assert(t, ob.HighestBuy.GetLimitPrice(), 900.0)

	// check price priority
	assert(t, tradesArray[0].GetBuyer().GetUserId(), "jane")
	assert(t, tradesArray[1].GetBuyer().GetUserId(), "jun")
	// 	// check time priority jun > jack
	assert(t, tradesArray[2].GetBuyer().GetUserId(), "jack")
	assert(t, tradesArray[3].GetBuyer().GetUserId(), "john")
	assert(t, tradesArray[4].GetBuyer().GetUserId(), "jim")

	incomingOrder = domain.NewOrder("lily", "ticker", false, domain.MarketOrderType, 0.5, 0)
	tradesArray = ob.PlaceMarketOrder(*incomingOrder)
	assert(t, ob.GetTotalVolumeAllSells(), 0.0)
	assert(t, tradesArray[0].GetBuyer().GetUserId(), "jim")
	assert(t, tradesArray[0].GetSize(), 0.5)
	assert(t, tradesArray[0].GetPrice(), 900.0)

	assert(t, ob.GetLastTrades()[0].GetBuyer().GetUserId(), "jane")
	assert(t, ob.GetLastTrades()[1].GetBuyer().GetUserId(), "jun")
	assert(t, ob.GetLastTrades()[2].GetBuyer().GetUserId(), "jack")
	assert(t, ob.GetLastTrades()[3].GetBuyer().GetUserId(), "john")
	assert(t, ob.GetLastTrades()[4].GetBuyer().GetUserId(), "jim")
	assert(t, ob.GetLastTrades()[5].GetBuyer().GetUserId(), "jim")
	assert(t, ob.HighestBuy == nil, true)
	assert(t, ob.BuyTree == nil, true)
}
