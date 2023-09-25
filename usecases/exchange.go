package usecases

import (
	"github.com/trandinhkhoa/crypto-exchange/domain"
)

type Ticker string

const (
	ETHUSD Ticker = "ETHUSD"
)

type Exchange struct {
	usersMap      map[string]*domain.User
	orderbooksMap map[Ticker]*Orderbook
}

func NewExchange() *Exchange {
	newExchange := &Exchange{}
	newExchange.usersMap = make(map[string]*domain.User, 0)
	ethusdOrderbook := &Orderbook{}
	newExchange.orderbooksMap = map[Ticker]*Orderbook{}
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

func (ex Exchange) GetLastTrades(ticker string, k int) []Trade {
	length := len(ex.orderbooksMap[Ticker(ticker)].lastTrades)
	if k > length {
		return ex.orderbooksMap[Ticker(ticker)].lastTrades
	}
	return ex.orderbooksMap[Ticker(ticker)].lastTrades[length-k:]
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
	if o.GetIsBid() {
		user.Balance[ticker2] -= o.Size * o.GetLimitPrice()
	} else {
		user.Balance[ticker1] -= o.Size
	}

	ex.orderbooksMap[ticker].PlaceLimitOrder(o)
}

func (ex *Exchange) PlaceMarketOrder(o domain.Order) []Trade {
	// TODO: volume check
	ticker := Ticker(o.GetTicker())
	ticker1 := string(ticker[:3])
	ticker2 := string(ticker[3:])

	// match
	tradesArray := ex.orderbooksMap[ticker].PlaceMarketOrder(o)
	//execute
	for _, trade := range tradesArray {
		buyer := ex.usersMap[trade.buyer.GetUserId()]
		seller := ex.usersMap[trade.seller.GetUserId()]

		buyer.Balance[ticker1] += trade.size
		if trade.buyer.GetOrderType() == domain.MarketOrderType {
			buyer.Balance[ticker2] -= trade.size * trade.price
		}

		if trade.seller.GetOrderType() == domain.MarketOrderType {
			seller.Balance[ticker1] -= trade.size
		}
		seller.Balance[ticker2] += trade.size * trade.price
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
	buybook := TreeToArray(ex.orderbooksMap[Ticker(ticker)].BuyTree)
	buyVolume := ex.orderbooksMap[Ticker(ticker)].GetTotalVolumeAllBuys()
	sellbook := TreeToArray(ex.orderbooksMap[Ticker(ticker)].SellTree)
	sellVolume := ex.orderbooksMap[Ticker(ticker)].GetTotalVolumeAllSells()
	return buybook, buyVolume, sellbook, sellVolume
}

func (ex *Exchange) GetBestBuy(ticker string) float64 {
	return ex.orderbooksMap[Ticker(ticker)].HighestBuy.GetLimitPrice()
}

func (ex *Exchange) GetBestSell(ticker string) float64 {
	return ex.orderbooksMap[Ticker(ticker)].LowestSell.GetLimitPrice()
}
