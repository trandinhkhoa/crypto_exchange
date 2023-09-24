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

func (ex Exchange) GetLastTrades(ticker string) []Trade {
	return ex.orderbooksMap[Ticker(ticker)].lastTrades
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

func (ex *Exchange) PlaceMarketOrder(o domain.Order) {
	// TODO: volume check
	ticker := Ticker(o.GetTicker())
	ticker1 := string(ticker[:3])
	ticker2 := string(ticker[3:])

	// match
	tradesArray := ex.orderbooksMap[ticker].PlaceMarketOrder(o)
	//execute
	for _, value := range tradesArray {
		buyer := ex.usersMap[value.buyer.GetUserId()]
		seller := ex.usersMap[value.seller.GetUserId()]
		buyer.Balance[ticker1] += value.size
		buyer.Balance[ticker2] -= value.size * value.price

		seller.Balance[ticker1] -= value.size
		seller.Balance[ticker2] += value.size * value.price
	}
}

func (ex *Exchange) RegisterUser(userId string) domain.User {
	// TODO: should have an array of tickers, iterate it and set their balances to zeros
	newUser := domain.User{
		UserId:  userId,
		Balance: make(map[string]float64),
	}
	newUser.Balance[string(ETHUSD)[:3]] = 0
	newUser.Balance[string(ETHUSD)[3:]] = 0
	return newUser
}

func (ex *Exchange) RegisterUserWithBalance(userId string, balance map[string]float64) {
	// TODO: should have an array of tickers, iterate it and set their balances to zeros
	newUser := domain.User{
		UserId:  userId,
		Balance: balance,
	}
	ex.usersMap[userId] = &newUser
}
