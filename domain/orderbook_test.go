package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/trandinhkhoa/crypto-exchange/domain"
)

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
	ob := domain.NewOrderbook()

	// Root: 1000
	// L--- 1005
	//     L--- 1100
	// R--- 900
	incomingOrder := domain.NewOrder("john", "ticker", true, domain.LimitOrderType, 1, 1000)
	ob.PlaceLimitOrder(*incomingOrder)

	assert.Equal(t, ob.HighestBuy.TotalVolume, 1.0)
	assert.Equal(t, ob.HighestBuy.GetLimitPrice(), 1000.0)
	assert.Equal(t, ob.HighestBuy.HeadOrder.GetUserId(), "john")

	incomingOrder = domain.NewOrder("jim", "ticker", true, domain.LimitOrderType, 1, 900)
	ob.PlaceLimitOrder(*incomingOrder)
	assert.Equal(t, ob.HighestBuy.TotalVolume, 1.0)
	assert.Equal(t, ob.HighestBuy.GetLimitPrice(), 1000.0)
	assert.Equal(t, ob.HighestBuy.HeadOrder.GetUserId(), "john")

	assert.Equal(t, ob.BuyTree.RightChild.TotalVolume, 1.0)
	assert.Equal(t, ob.BuyTree.RightChild.GetLimitPrice(), 900.0)
	assert.Equal(t, ob.BuyTree.RightChild.HeadOrder.GetUserId(), "jim")

	incomingOrder = domain.NewOrder("jane", "ticker", true, domain.LimitOrderType, 4, 1100)
	ob.PlaceLimitOrder(*incomingOrder)
	assert.Equal(t, ob.HighestBuy.TotalVolume, 4.0)
	assert.Equal(t, ob.HighestBuy.GetLimitPrice(), 1100.0)
	assert.Equal(t, ob.HighestBuy.HeadOrder.GetUserId(), "jane")

	assert.Equal(t, ob.BuyTree.RightChild.TotalVolume, 1.0)
	assert.Equal(t, ob.BuyTree.RightChild.GetLimitPrice(), 900.0)
	assert.Equal(t, ob.BuyTree.RightChild.HeadOrder.GetUserId(), "jim")

	assert.Equal(t, ob.BuyTree.LeftChild.TotalVolume, 4.0)
	assert.Equal(t, ob.BuyTree.LeftChild.GetLimitPrice(), 1100.0)
	assert.Equal(t, ob.BuyTree.LeftChild.HeadOrder.GetUserId(), "jane")

	incomingOrder = domain.NewOrder("jun", "ticker", true, domain.LimitOrderType, 9, 1005)
	ob.PlaceLimitOrder(*incomingOrder)
	assert.Equal(t, ob.HighestBuy.TotalVolume, 4.0)
	assert.Equal(t, ob.HighestBuy.GetLimitPrice(), 1100.0)
	assert.Equal(t, ob.HighestBuy.HeadOrder.GetUserId(), "jane")

	assert.Equal(t, ob.BuyTree.RightChild.TotalVolume, 1.0)
	assert.Equal(t, ob.BuyTree.RightChild.GetLimitPrice(), 900.0)
	assert.Equal(t, ob.BuyTree.RightChild.HeadOrder.GetUserId(), "jim")

	assert.Equal(t, ob.BuyTree.LeftChild.TotalVolume, 4.0)
	assert.Equal(t, ob.BuyTree.LeftChild.GetLimitPrice(), 1100.0)
	assert.Equal(t, ob.BuyTree.LeftChild.HeadOrder.GetUserId(), "jane")

	assert.Equal(t, ob.BuyTree.LeftChild.RightChild.TotalVolume, 9.0)
	assert.Equal(t, ob.BuyTree.LeftChild.RightChild.GetLimitPrice(), 1005.0)
	assert.Equal(t, ob.BuyTree.LeftChild.RightChild.HeadOrder.GetUserId(), "jun")

	incomingOrder = domain.NewOrder("jack", "ticker", true, domain.LimitOrderType, 9, 1005)
	ob.PlaceLimitOrder(*incomingOrder)
	assert.Equal(t, ob.HighestBuy.TotalVolume, 4.0)
	assert.Equal(t, ob.HighestBuy.GetLimitPrice(), 1100.0)
	assert.Equal(t, ob.HighestBuy.HeadOrder.GetUserId(), "jane")

	assert.Equal(t, ob.BuyTree.RightChild.TotalVolume, 1.0)
	assert.Equal(t, ob.BuyTree.RightChild.GetLimitPrice(), 900.0)
	assert.Equal(t, ob.BuyTree.RightChild.HeadOrder.GetUserId(), "jim")

	assert.Equal(t, ob.BuyTree.LeftChild.TotalVolume, 4.0)
	assert.Equal(t, ob.BuyTree.LeftChild.GetLimitPrice(), 1100.0)
	assert.Equal(t, ob.BuyTree.LeftChild.HeadOrder.GetUserId(), "jane")

	assert.Equal(t, ob.BuyTree.LeftChild.RightChild.TotalVolume, 18.0)
	assert.Equal(t, ob.BuyTree.LeftChild.RightChild.GetLimitPrice(), 1005.0)
	assert.Equal(t, ob.BuyTree.LeftChild.RightChild.HeadOrder.GetUserId(), "jun")
	assert.Equal(t, ob.BuyTree.LeftChild.RightChild.TailOrder.GetUserId(), "jack")

	arr := domain.TreeToArray(ob.BuyTree)
	for index := 0; index < len(arr)-1; index++ {
		if arr[index].GetLimitPrice() <= arr[index+1].GetLimitPrice() {
			t.Errorf("Buy Limit not sorted in descending order ")
			break
		}
	}
}

func TestPlaceLimitOrderSell(t *testing.T) {
	ob := domain.NewOrderbook()

	// Root: 1000
	// L--- 900
	// R--- 1005
	//     R--- 1100

	incomingOrder := domain.NewOrder("john", "ticker", false, domain.LimitOrderType, 1, 1000)
	ob.PlaceLimitOrder(*incomingOrder)

	assert.Equal(t, ob.LowestSell.TotalVolume, 1.0)
	assert.Equal(t, ob.LowestSell.GetLimitPrice(), 1000.0)
	assert.Equal(t, ob.LowestSell.HeadOrder.GetUserId(), "john")

	incomingOrder = domain.NewOrder("jim", "ticker", false, domain.LimitOrderType, 1, 900)
	ob.PlaceLimitOrder(*incomingOrder)
	assert.Equal(t, ob.LowestSell.TotalVolume, 1.0)
	assert.Equal(t, ob.LowestSell.GetLimitPrice(), 900.0)
	assert.Equal(t, ob.LowestSell.HeadOrder.GetUserId(), "jim")

	assert.Equal(t, ob.SellTree.LeftChild.TotalVolume, 1.0)
	assert.Equal(t, ob.SellTree.LeftChild.GetLimitPrice(), 900.0)
	assert.Equal(t, ob.SellTree.LeftChild.HeadOrder.GetUserId(), "jim")

	incomingOrder = domain.NewOrder("jane", "ticker", false, domain.LimitOrderType, 4, 1100)
	ob.PlaceLimitOrder(*incomingOrder)
	assert.Equal(t, ob.LowestSell.TotalVolume, 1.0)
	assert.Equal(t, ob.LowestSell.GetLimitPrice(), 900.0)
	assert.Equal(t, ob.LowestSell.HeadOrder.GetUserId(), "jim")

	assert.Equal(t, ob.SellTree.LeftChild.TotalVolume, 1.0)
	assert.Equal(t, ob.SellTree.LeftChild.GetLimitPrice(), 900.0)
	assert.Equal(t, ob.SellTree.LeftChild.HeadOrder.GetUserId(), "jim")

	assert.Equal(t, ob.SellTree.RightChild.TotalVolume, 4.0)
	assert.Equal(t, ob.SellTree.RightChild.GetLimitPrice(), 1100.0)
	assert.Equal(t, ob.SellTree.RightChild.HeadOrder.GetUserId(), "jane")

	incomingOrder = domain.NewOrder("jun", "ticker", false, domain.LimitOrderType, 9, 1005)
	ob.PlaceLimitOrder(*incomingOrder)
	assert.Equal(t, ob.LowestSell.TotalVolume, 1.0)
	assert.Equal(t, ob.LowestSell.GetLimitPrice(), 900.0)
	assert.Equal(t, ob.LowestSell.HeadOrder.GetUserId(), "jim")

	assert.Equal(t, ob.SellTree.LeftChild.TotalVolume, 1.0)
	assert.Equal(t, ob.SellTree.LeftChild.GetLimitPrice(), 900.0)
	assert.Equal(t, ob.SellTree.LeftChild.HeadOrder.GetUserId(), "jim")

	assert.Equal(t, ob.SellTree.RightChild.TotalVolume, 4.0)
	assert.Equal(t, ob.SellTree.RightChild.GetLimitPrice(), 1100.0)
	assert.Equal(t, ob.SellTree.RightChild.HeadOrder.GetUserId(), "jane")

	assert.Equal(t, ob.SellTree.RightChild.LeftChild.TotalVolume, 9.0)
	assert.Equal(t, ob.SellTree.RightChild.LeftChild.GetLimitPrice(), 1005.0)
	assert.Equal(t, ob.SellTree.RightChild.LeftChild.HeadOrder.GetUserId(), "jun")

	incomingOrder = domain.NewOrder("jack", "ticker", false, domain.LimitOrderType, 9, 1005)
	ob.PlaceLimitOrder(*incomingOrder)
	assert.Equal(t, ob.LowestSell.TotalVolume, 1.0)
	assert.Equal(t, ob.LowestSell.GetLimitPrice(), 900.0)
	assert.Equal(t, ob.LowestSell.HeadOrder.GetUserId(), "jim")

	assert.Equal(t, ob.SellTree.LeftChild.TotalVolume, 1.0)
	assert.Equal(t, ob.SellTree.LeftChild.GetLimitPrice(), 900.0)
	assert.Equal(t, ob.SellTree.LeftChild.HeadOrder.GetUserId(), "jim")

	assert.Equal(t, ob.SellTree.RightChild.TotalVolume, 4.0)
	assert.Equal(t, ob.SellTree.RightChild.GetLimitPrice(), 1100.0)
	assert.Equal(t, ob.SellTree.RightChild.HeadOrder.GetUserId(), "jane")

	assert.Equal(t, ob.SellTree.RightChild.LeftChild.TotalVolume, 18.0)
	assert.Equal(t, ob.SellTree.RightChild.LeftChild.GetLimitPrice(), 1005.0)
	assert.Equal(t, ob.SellTree.RightChild.LeftChild.HeadOrder.GetUserId(), "jun")
	assert.Equal(t, ob.SellTree.RightChild.LeftChild.TailOrder.GetUserId(), "jack")

	arr := domain.TreeToArray(ob.SellTree)
	for index := 0; index < len(arr)-1; index++ {
		if arr[index].GetLimitPrice() >= arr[index+1].GetLimitPrice() {
			t.Errorf("Sell Limit not sorted in ascending order ")
			break
		}
	}
	// fmt.Println(limitArrToJson(arr))
}
func TestPlaceMarketOrderBuyOneFill(t *testing.T) {
	// Root: 1000
	// L--- 900
	// R--- 1005
	//     R--- 1100
	ob := domain.NewOrderbook()

	incomingOrder := domain.NewOrder("john", "ticker", false, domain.LimitOrderType, 1, 1000)
	ob.PlaceLimitOrder(*incomingOrder)

	incomingOrder = domain.NewOrder("jim", "ticker", false, domain.LimitOrderType, 1, 900)
	ob.PlaceLimitOrder(*incomingOrder)

	incomingOrder = domain.NewOrder("jane", "ticker", false, domain.LimitOrderType, 4, 1100)
	ob.PlaceLimitOrder(*incomingOrder)

	incomingOrder = domain.NewOrder("jun", "ticker", false, domain.LimitOrderType, 9, 1005)
	ob.PlaceLimitOrder(*incomingOrder)

	incomingOrder = domain.NewOrder("jack", "ticker", false, domain.LimitOrderType, 9, 1005)
	ob.PlaceLimitOrder(*incomingOrder)
	assert.Equal(t, ob.GetTotalVolumeAllSells(), 24.0)

	incomingOrder = domain.NewOrder("lily", "ticker", true, domain.MarketOrderType, 1, 0)
	tradesArray := ob.PlaceMarketOrder(*incomingOrder)
	assert.Equal(t, ob.GetTotalVolumeAllSells(), 23.0)
	assert.Equal(t, ob.LowestSell.GetLimitPrice(), 1000.0)

	assert.Equal(t, tradesArray[0].GetBuyer().GetUserId(), "lily")
	assert.Equal(t, tradesArray[0].GetSeller().GetUserId(), "jim")
	assert.Equal(t, tradesArray[0].GetPrice(), 900.0)
	assert.Equal(t, tradesArray[0].GetSize(), 1.0)
}

func TestPlaceMarketOrderBuyOnePartialFill(t *testing.T) {
	ob := domain.NewOrderbook()

	// Root: 1000
	// L--- 900
	// R--- 1005
	//     R--- 1100
	incomingOrder := domain.NewOrder("john", "ticker", false, domain.LimitOrderType, 1, 1000)
	ob.PlaceLimitOrder(*incomingOrder)

	incomingOrder = domain.NewOrder("jim", "ticker", false, domain.LimitOrderType, 1, 900)
	ob.PlaceLimitOrder(*incomingOrder)

	incomingOrder = domain.NewOrder("jane", "ticker", false, domain.LimitOrderType, 4, 1100)
	ob.PlaceLimitOrder(*incomingOrder)

	incomingOrder = domain.NewOrder("jun", "ticker", false, domain.LimitOrderType, 9, 1005)
	ob.PlaceLimitOrder(*incomingOrder)

	incomingOrder = domain.NewOrder("jack", "ticker", false, domain.LimitOrderType, 9, 1005)
	ob.PlaceLimitOrder(*incomingOrder)
	assert.Equal(t, ob.GetTotalVolumeAllSells(), 24.0)

	incomingOrder = domain.NewOrder("lily", "ticker", true, domain.MarketOrderType, 0.5, 0)
	tradesArray := ob.PlaceMarketOrder(*incomingOrder)
	assert.Equal(t, ob.GetTotalVolumeAllSells(), 23.5)
	assert.Equal(t, ob.LowestSell.GetLimitPrice(), 900.0)

	assert.Equal(t, tradesArray[0].GetBuyer().GetUserId(), "lily")
	assert.Equal(t, tradesArray[0].GetSeller().GetUserId(), "jim")
	assert.Equal(t, tradesArray[0].GetPrice(), 900.0)
	assert.Equal(t, tradesArray[0].GetSize(), 0.5)
}

func TestPlaceMarketOrderBuyMultiFill(t *testing.T) {
	ob := domain.NewOrderbook()

	// Root: 1000
	// L--- 900
	// R--- 1005
	//     R--- 1100

	incomingOrder := domain.NewOrder("john", "ticker", false, domain.LimitOrderType, 1, 1000)
	ob.PlaceLimitOrder(*incomingOrder)

	incomingOrder = domain.NewOrder("jim", "ticker", false, domain.LimitOrderType, 1, 900)
	ob.PlaceLimitOrder(*incomingOrder)

	incomingOrder = domain.NewOrder("jane", "ticker", false, domain.LimitOrderType, 4, 1100)
	ob.PlaceLimitOrder(*incomingOrder)

	incomingOrder = domain.NewOrder("jun", "ticker", false, domain.LimitOrderType, 9, 1005)
	ob.PlaceLimitOrder(*incomingOrder)

	incomingOrder = domain.NewOrder("jack", "ticker", false, domain.LimitOrderType, 9, 1005)
	ob.PlaceLimitOrder(*incomingOrder)
	assert.Equal(t, ob.GetTotalVolumeAllSells(), 24.0)

	incomingOrder = domain.NewOrder("lily", "ticker", true, domain.MarketOrderType, 23.5, 0)
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

	incomingOrder = domain.NewOrder("lily", "ticker", true, domain.MarketOrderType, 0.5, 0)
	tradesArray = ob.PlaceMarketOrder(*incomingOrder)
	assert.Equal(t, ob.GetTotalVolumeAllSells(), 0.0)
	assert.Equal(t, tradesArray[0].GetSeller().GetUserId(), "jane")
	assert.Equal(t, tradesArray[0].GetSize(), 0.5)
	assert.Equal(t, tradesArray[0].GetPrice(), 1100.0)

	assert.Equal(t, ob.GetLastTrades()[0].GetSeller().GetUserId(), "jim")
	assert.Equal(t, ob.GetLastTrades()[1].GetSeller().GetUserId(), "john")
	assert.Equal(t, ob.GetLastTrades()[2].GetSeller().GetUserId(), "jun")
	assert.Equal(t, ob.GetLastTrades()[3].GetSeller().GetUserId(), "jack")
	assert.Equal(t, ob.GetLastTrades()[4].GetSeller().GetUserId(), "jane")
	assert.Equal(t, ob.GetLastTrades()[5].GetSeller().GetUserId(), "jane")
	assert.Equal(t, ob.LowestSell == nil, true)
	assert.Equal(t, ob.SellTree == nil, true)
}

func TestPlaceMarketOrderSellMultiFill(t *testing.T) {
	// Root: 1000
	// L--- 1005
	//     L--- 1100
	// R--- 900
	ob := domain.NewOrderbook()

	incomingOrder := domain.NewOrder("john", "ticker", true, domain.LimitOrderType, 1, 1000)
	ob.PlaceLimitOrder(*incomingOrder)

	incomingOrder = domain.NewOrder("jim", "ticker", true, domain.LimitOrderType, 1, 900)
	ob.PlaceLimitOrder(*incomingOrder)

	incomingOrder = domain.NewOrder("jane", "ticker", true, domain.LimitOrderType, 4, 1100)
	ob.PlaceLimitOrder(*incomingOrder)

	incomingOrder = domain.NewOrder("jun", "ticker", true, domain.LimitOrderType, 9, 1005)
	ob.PlaceLimitOrder(*incomingOrder)

	incomingOrder = domain.NewOrder("jack", "ticker", true, domain.LimitOrderType, 9, 1005)
	ob.PlaceLimitOrder(*incomingOrder)
	assert.Equal(t, ob.GetTotalVolumeAllBuys(), 24.0)

	incomingOrder = domain.NewOrder("lily", "ticker", false, domain.MarketOrderType, 23.5, 0)
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

	incomingOrder = domain.NewOrder("lily", "ticker", false, domain.MarketOrderType, 0.5, 0)
	tradesArray = ob.PlaceMarketOrder(*incomingOrder)
	assert.Equal(t, ob.GetTotalVolumeAllBuys(), 0.0)
	assert.Equal(t, tradesArray[0].GetBuyer().GetUserId(), "jim")
	assert.Equal(t, tradesArray[0].GetSize(), 0.5)
	assert.Equal(t, tradesArray[0].GetPrice(), 900.0)

	assert.Equal(t, ob.GetLastTrades()[0].GetBuyer().GetUserId(), "jane")
	assert.Equal(t, ob.GetLastTrades()[1].GetBuyer().GetUserId(), "jun")
	assert.Equal(t, ob.GetLastTrades()[2].GetBuyer().GetUserId(), "jack")
	assert.Equal(t, ob.GetLastTrades()[3].GetBuyer().GetUserId(), "john")
	assert.Equal(t, ob.GetLastTrades()[4].GetBuyer().GetUserId(), "jim")
	assert.Equal(t, ob.GetLastTrades()[5].GetBuyer().GetUserId(), "jim")
	assert.Equal(t, ob.HighestBuy == nil, true)
	assert.Equal(t, ob.BuyTree == nil, true)
}

func TestCancelOrderSimple(t *testing.T) {
	ob := domain.NewOrderbook()

	// Root: 1000

	incomingOrder := domain.NewOrder("john", "ticker", true, domain.LimitOrderType, 1, 1000)
	ob.PlaceLimitOrder(*incomingOrder)

	assert.Equal(t, 1.0, ob.GetTotalVolumeAllBuys(), 1.0)
	assert.Equal(t, 1000.0, ob.HighestBuy.GetLimitPrice())

	ob.CancelOrder(incomingOrder.GetId())

	assert.Equal(t, 0.0, ob.GetTotalVolumeAllBuys())
	assert.True(t, ob.HighestBuy == nil)
}

func TestCancelOrderBigTree(t *testing.T) {
	ob := domain.NewOrderbook()

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
	johnOrder := domain.NewOrder("john", "ticker", false, domain.LimitOrderType, 1, 7)
	ob.PlaceLimitOrder(*johnOrder)

	jimOrder := domain.NewOrder("jim", "ticker", false, domain.LimitOrderType, 1, 3)
	ob.PlaceLimitOrder(*jimOrder)

	janeOrder := domain.NewOrder("jane", "ticker", false, domain.LimitOrderType, 4, 11)
	ob.PlaceLimitOrder(*janeOrder)

	junOrder := domain.NewOrder("jun", "ticker", false, domain.LimitOrderType, 9, 9)
	ob.PlaceLimitOrder(*junOrder)

	jackOrder := domain.NewOrder("jack", "ticker", false, domain.LimitOrderType, 9, 9)
	ob.PlaceLimitOrder(*jackOrder)

	jerryOrder := domain.NewOrder("jerry", "ticker", false, domain.LimitOrderType, 2, 1)
	ob.PlaceLimitOrder(*jerryOrder)

	jessicaOrder := domain.NewOrder("jessica", "ticker", false, domain.LimitOrderType, 2, 2)
	ob.PlaceLimitOrder(*jessicaOrder)

	jessicaTwinOrder := domain.NewOrder("jessicaTwin", "ticker", false, domain.LimitOrderType, 3, 2)
	ob.PlaceLimitOrder(*jessicaTwinOrder)

	jillOrder := domain.NewOrder("jill", "ticker", false, domain.LimitOrderType, 3, 4)
	ob.PlaceLimitOrder(*jillOrder)

	jeffOrder := domain.NewOrder("jeff", "ticker", false, domain.LimitOrderType, 3, 5)
	ob.PlaceLimitOrder(*jeffOrder)

	jacobOrder := domain.NewOrder("jacob", "ticker", false, domain.LimitOrderType, 4, 6)
	ob.PlaceLimitOrder(*jacobOrder)

	julieOrder := domain.NewOrder("julie", "ticker", false, domain.LimitOrderType, 4, 8)
	ob.PlaceLimitOrder(*julieOrder)

	jamesOrder := domain.NewOrder("james", "ticker", false, domain.LimitOrderType, 5, 10)
	ob.PlaceLimitOrder(*jamesOrder)

	joanOrder := domain.NewOrder("joan", "ticker", false, domain.LimitOrderType, 5, 12)
	ob.PlaceLimitOrder(*joanOrder)

	jamieOrder := domain.NewOrder("jamie", "ticker", false, domain.LimitOrderType, 6, 13)
	ob.PlaceLimitOrder(*jamieOrder)

	jodieOrder := domain.NewOrder("jodie", "ticker", false, domain.LimitOrderType, 6, 14)
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

	arr := domain.TreeToArray(ob.SellTree)
	for index := 0; index < len(arr)-1; index++ {
		if arr[index].GetLimitPrice() >= arr[index+1].GetLimitPrice() {
			t.Errorf("Sell Limit not sorted in ascending order ")
			break
		}
	}
}

func TestGetKBestBuys(t *testing.T) {
	ob := domain.NewOrderbook()

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
	johnOrder := domain.NewOrder("john", "ticker", false, domain.LimitOrderType, 1, 7)
	ob.PlaceLimitOrder(*johnOrder)

	jimOrder := domain.NewOrder("jim", "ticker", false, domain.LimitOrderType, 1, 3)
	ob.PlaceLimitOrder(*jimOrder)

	janeOrder := domain.NewOrder("jane", "ticker", false, domain.LimitOrderType, 4, 11)
	ob.PlaceLimitOrder(*janeOrder)

	junOrder := domain.NewOrder("jun", "ticker", false, domain.LimitOrderType, 9, 9)
	ob.PlaceLimitOrder(*junOrder)

	jackOrder := domain.NewOrder("jack", "ticker", false, domain.LimitOrderType, 9, 9)
	ob.PlaceLimitOrder(*jackOrder)

	jerryOrder := domain.NewOrder("jerry", "ticker", false, domain.LimitOrderType, 2, 1)
	ob.PlaceLimitOrder(*jerryOrder)

	jessicaOrder := domain.NewOrder("jessica", "ticker", false, domain.LimitOrderType, 2, 2)
	ob.PlaceLimitOrder(*jessicaOrder)

	jillOrder := domain.NewOrder("jill", "ticker", false, domain.LimitOrderType, 3, 4)
	ob.PlaceLimitOrder(*jillOrder)

	jeffOrder := domain.NewOrder("jeff", "ticker", false, domain.LimitOrderType, 3, 5)
	ob.PlaceLimitOrder(*jeffOrder)

	jacobOrder := domain.NewOrder("jacob", "ticker", false, domain.LimitOrderType, 4, 6)
	ob.PlaceLimitOrder(*jacobOrder)

	julieOrder := domain.NewOrder("julie", "ticker", false, domain.LimitOrderType, 4, 8)
	ob.PlaceLimitOrder(*julieOrder)

	jamesOrder := domain.NewOrder("james", "ticker", false, domain.LimitOrderType, 5, 10)
	ob.PlaceLimitOrder(*jamesOrder)

	joanOrder := domain.NewOrder("joan", "ticker", false, domain.LimitOrderType, 5, 12)
	ob.PlaceLimitOrder(*joanOrder)

	jamieOrder := domain.NewOrder("jamie", "ticker", false, domain.LimitOrderType, 6, 13)
	ob.PlaceLimitOrder(*jamieOrder)

	jodieOrder := domain.NewOrder("jodie", "ticker", false, domain.LimitOrderType, 6, 14)
	ob.PlaceLimitOrder(*jodieOrder)

	assert.Equal(t, 64.0, ob.GetTotalVolumeAllSells())

	assert.Equal(t, 1.0, ob.LowestSell.GetLimitPrice())

	arr := ob.GetBestLimits(ob.SellTree, 3)
	assert.Equal(t, 3, len(arr))
	assert.Equal(t, 1.0, arr[0].GetLimitPrice())
	assert.Equal(t, 2.0, arr[1].GetLimitPrice())
	assert.Equal(t, 3.0, arr[2].GetLimitPrice())
}
