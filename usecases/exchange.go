package usecases

import (
	"fmt"
	"sync"

	"github.com/trandinhkhoa/crypto-exchange/domain"
)

type Ticker string

const (
	ETHUSD Ticker = "ETHUSD"
)

type Exchange struct {
	usersMap      map[string]*domain.User
	orderbooksMap map[Ticker]*domain.Orderbook
	mu            sync.Mutex
}

func NewExchange() *Exchange {
	newExchange := &Exchange{}
	newExchange.usersMap = make(map[string]*domain.User, 0)
	ethusdOrderbook := domain.NewOrderbook()
	newExchange.orderbooksMap = map[Ticker]*domain.Orderbook{}
	newExchange.orderbooksMap[ETHUSD] = ethusdOrderbook

	return newExchange
}

func (ex Exchange) GetUsersMap() map[string]domain.User {
	// return a copy, not a reference
	usersMap := make(map[string]domain.User, 0)

	for k, v := range ex.usersMap {
		usersMap[k] = *v
	}
	return usersMap
}

func (ex Exchange) GetLastTrades(ticker string, k int) []domain.Trade {
	length := len(ex.orderbooksMap[Ticker(ticker)].GetLastTrades())
	if k > length {
		return ex.orderbooksMap[Ticker(ticker)].GetLastTrades()
	}
	return ex.orderbooksMap[Ticker(ticker)].GetLastTrades()[length-k:]
}

func (ex Exchange) GetBestBuys(ticker string, k int) []*domain.Limit {
	book := ex.orderbooksMap[Ticker(ticker)]
	return book.GetBestLimits(book.BuyTree, k)
}

func (ex Exchange) GetBestSells(ticker string, k int) []*domain.Limit {
	book := ex.orderbooksMap[Ticker(ticker)]
	return book.GetBestLimits(book.SellTree, k)
}

func (ex Exchange) GetLastPrice(ticker string) float64 {
	return ex.orderbooksMap[Ticker(ticker)].LastTradedPrice
}

func (ex *Exchange) PlaceLimitOrder(o domain.Order) {
	ticker := Ticker(o.GetTicker())
	userId := o.GetUserId()
	user := ex.usersMap[userId]
	// TODO: check ticker values is valid ?? asset's name (part of ticker) should be an enum too ??
	ticker1 := string(ticker[:3])
	ticker2 := string(ticker[3:])

	// block user balance
	// TODO: check user's balance
	ex.mu.Lock()
	defer ex.mu.Unlock()
	if o.GetIsBid() {
		user.Balance[ticker2] -= o.Size * o.GetLimitPrice()
	} else {
		user.Balance[ticker1] -= o.Size
	}

	ex.orderbooksMap[ticker].PlaceLimitOrder(o)
}

func (ex *Exchange) PlaceMarketOrder(o domain.Order) []domain.Trade {
	// TODO: volume check
	ticker := Ticker(o.GetTicker())
	ticker1 := string(ticker[:3])
	ticker2 := string(ticker[3:])

	ex.mu.Lock()
	defer ex.mu.Unlock()
	// match
	// TODO: PlaceMarketOrder() should not modify the orderbook.
	// Market Buyer/Seller might not have sufficient balance and there is no way to check it before calling PlaceMarketOrder
	tradesArray := ex.orderbooksMap[ticker].PlaceMarketOrder(o)
	//execute
	for _, trade := range tradesArray {
		buyer := ex.usersMap[trade.GetBuyer().GetUserId()]
		seller := ex.usersMap[trade.GetSeller().GetUserId()]

		buyer.Balance[ticker1] += trade.Size
		if trade.GetBuyer().GetOrderType() == domain.MarketOrderType {
			buyer.Balance[ticker2] -= trade.Size * trade.Price
		}

		if trade.GetSeller().GetOrderType() == domain.MarketOrderType {
			seller.Balance[ticker1] -= trade.Size
		}
		// TODO: john's limit order might be filled (here) at the same time as he is placing a new limit order
		// -> concurrent write
		seller.Balance[ticker2] += trade.Size * trade.Price
		fmt.Println(trade)
	}
	return tradesArray
}

func (ex *Exchange) RegisterUser(userId string) {
	// TODO: should have an array of tickers, iterate it and set their balances to zeros
	newUser := domain.User{
		UserId:  userId,
		Balance: make(map[string]float64),
	}
	newUser.Balance[string(ETHUSD)[:3]] = 0
	newUser.Balance[string(ETHUSD)[3:]] = 0
	ex.usersMap[userId] = &newUser
}

func (ex *Exchange) RegisterUserWithBalance(userId string, balance map[string]float64) {
	// TODO: should have an array of tickers, iterate it and set their balances to zeros
	newUser := domain.User{
		UserId:  userId,
		Balance: balance,
	}
	ex.usersMap[userId] = &newUser
}

func (ex *Exchange) GetBook(ticker string) ([]*domain.Limit, float64, []*domain.Limit, float64) {
	buybook := domain.TreeToArray(ex.orderbooksMap[Ticker(ticker)].BuyTree)
	buyVolume := ex.orderbooksMap[Ticker(ticker)].GetTotalVolumeAllBuys()
	sellbook := domain.TreeToArray(ex.orderbooksMap[Ticker(ticker)].SellTree)
	sellVolume := ex.orderbooksMap[Ticker(ticker)].GetTotalVolumeAllSells()
	return buybook, buyVolume, sellbook, sellVolume
}

func (ex *Exchange) GetBestBuy(ticker string) float64 {
	if ex.orderbooksMap[Ticker(ticker)].HighestBuy == nil {
		// TODO: return error instead
		return 0
	}
	return ex.orderbooksMap[Ticker(ticker)].HighestBuy.GetLimitPrice()
}

func (ex *Exchange) GetBestSell(ticker string) float64 {
	if ex.orderbooksMap[Ticker(ticker)].LowestSell == nil {
		// TODO: return error instead
		return 0
	}
	return ex.orderbooksMap[Ticker(ticker)].LowestSell.GetLimitPrice()
}

func (ex *Exchange) CancelOrder(orderId int64, ticker string) {
	orderbook := ex.orderbooksMap[Ticker(ticker)]
	userId, isBid, price, size := orderbook.CancelOrder(orderId)
	user := ex.usersMap[userId]

	ticker1 := string(ticker[:3])
	ticker2 := string(ticker[3:])
	if isBid {
		user.Balance[ticker2] += size * price
	} else {
		user.Balance[ticker1] += size
	}
}
