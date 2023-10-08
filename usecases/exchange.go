package usecases

import (
	"errors"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/trandinhkhoa/crypto-exchange/entities"
)

type Ticker string

const (
	ETHUSD Ticker = "ETHUSD"
)

type Exchange struct {
	usersMap      map[string]*entities.User
	orderbooksMap map[Ticker]*entities.Orderbook
	mu            sync.Mutex

	// uppercase for now for quick injection
	// TODO: pass these as constructor args ??
	UsersRepo  UsersRepository
	OrdersRepo OrdersRepository
}

func NewExchange() *Exchange {
	newExchange := &Exchange{}
	newExchange.usersMap = make(map[string]*entities.User, 0)
	ethusdOrderbook := entities.NewOrderbook()
	newExchange.orderbooksMap = map[Ticker]*entities.Orderbook{}
	newExchange.orderbooksMap[ETHUSD] = ethusdOrderbook

	return newExchange
}

// there is a lock inside Exchange so the pointer is the receiver
func (ex *Exchange) GetUsersMap() map[string]entities.User {
	usersMap := make(map[string]entities.User, 0)

	for k, v := range ex.usersMap {
		usersMap[k] = *v
	}
	return usersMap
}

func (ex *Exchange) GetLastTrades(ticker string, k int) []entities.Trade {
	length := len(ex.orderbooksMap[Ticker(ticker)].GetLastTrades())
	if k > length {
		return ex.orderbooksMap[Ticker(ticker)].GetLastTrades()
	}
	return ex.orderbooksMap[Ticker(ticker)].GetLastTrades()[length-k:]
}

func (ex *Exchange) GetBestBuys(ticker string, k int) []*entities.Limit {
	book := ex.orderbooksMap[Ticker(ticker)]
	return book.GetBestLimits(book.BuyTree, k)
}

func (ex *Exchange) GetBestSells(ticker string, k int) []*entities.Limit {
	book := ex.orderbooksMap[Ticker(ticker)]
	return book.GetBestLimits(book.SellTree, k)
}

func (ex *Exchange) GetLastPrice(ticker string) float64 {
	return ex.orderbooksMap[Ticker(ticker)].GetLastTradedPrice()
}

func (ex *Exchange) PlaceLimitOrder(o entities.Order) {
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

	go ex.persistAfterLimitOrder(o)
}

func (ex *Exchange) PlaceMarketOrder(o entities.Order) []entities.Trade {
	// TODO: volume check
	ticker := Ticker(o.GetTicker())
	ticker1 := string(ticker[:3])
	ticker2 := string(ticker[3:])

	ex.mu.Lock()
	defer ex.mu.Unlock()
	// match
	// TODO: PlaceMarketOrder() should not modify the orderbook.
	// Market Buyer/Seller might not have sufficient balance and there is no way to check it before calling PlaceMarketOrder

	// TODO: bubble up the error
	tradesArray, err := ex.orderbooksMap[ticker].PlaceMarketOrder(o)
	if err != nil {
		var noLiquidError *entities.NoLiquidityError
		switch {
		case errors.As(err, &noLiquidError):
			logrus.Error(noLiquidError.Error())
		default:
			logrus.Errorf("Unexpected error placing market order id: %d, error: %s \n", o.GetId(), err)
		}
		return nil
	}

	//execute
	for _, trade := range tradesArray {
		buyer := ex.usersMap[trade.GetBuyer().GetUserId()]
		seller := ex.usersMap[trade.GetSeller().GetUserId()]

		buyer.Balance[ticker1] += trade.GetSize()
		if trade.GetBuyer().GetOrderType() == entities.MarketOrderType {
			buyer.Balance[ticker2] -= trade.GetSize() * trade.GetPrice()
		}

		if trade.GetSeller().GetOrderType() == entities.MarketOrderType {
			seller.Balance[ticker1] -= trade.GetSize()
		}
		// TODO: john's limit order might be filled (here) at the same time as he is placing a new limit order
		// -> concurrent write
		seller.Balance[ticker2] += trade.GetSize() * trade.GetPrice()
		logrus.WithFields(logrus.Fields{
			"trade": trade,
		}).Info("Order Executed")
	}

	go ex.persistAfterMarketOrder(tradesArray)

	return tradesArray
}

func (ex *Exchange) RegisterUser(userId string) {
	// TODO: should have an array of tickers, iterate it and set their balances to zeros
	newUser := entities.NewUser(userId, make(map[string]float64))
	newUser.Balance[string(ETHUSD)[:3]] = 0
	newUser.Balance[string(ETHUSD)[3:]] = 0
	ex.usersMap[userId] = newUser

	//persist the creation
	ex.UsersRepo.Create(*newUser)
}

func (ex *Exchange) RegisterUserWithBalance(userId string, balance map[string]float64) {
	// TODO: should have an array of tickers, iterate it and set their balances to zeros
	newUser := entities.NewUser(userId, balance)
	ex.usersMap[userId] = newUser

	//persist the creation
	ex.UsersRepo.Create(*newUser)
}

func (ex *Exchange) GetBook(ticker string) ([]*entities.Limit, float64, []*entities.Limit, float64) {
	buybook := entities.TreeToArray(ex.orderbooksMap[Ticker(ticker)].BuyTree)
	buyVolume := ex.orderbooksMap[Ticker(ticker)].GetTotalVolumeAllBuys()
	sellbook := entities.TreeToArray(ex.orderbooksMap[Ticker(ticker)].SellTree)
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

func (ex *Exchange) persistAfterLimitOrder(order entities.Order) {
	// persist users balance
	ex.UsersRepo.Update(ex.GetUsersMap()[order.GetUserId()])
	// PlaceLimitOrder only adds for now so persist the creation
	ex.OrdersRepo.Create(order)
}

func (ex *Exchange) persistAfterMarketOrder(tradesArray []entities.Trade) {
	for _, trade := range tradesArray {
		buyer := trade.GetBuyer()
		seller := trade.GetSeller()

		// persist users balance
		ex.UsersRepo.Update(ex.GetUsersMap()[buyer.GetUserId()])
		ex.UsersRepo.Update(ex.GetUsersMap()[seller.GetUserId()])

		ticker := Ticker(trade.GetBuyer().GetTicker())
		// order does not exist == deleted => persist the deletion else persist current state
		_, err := ex.orderbooksMap[ticker].GetOrderbyId(trade.GetBuyer().GetId())
		if err != nil {
			ex.OrdersRepo.Delete(trade.GetBuyer())
		} else {
			ex.OrdersRepo.Update(trade.GetBuyer())
		}

		_, err = ex.orderbooksMap[ticker].GetOrderbyId(trade.GetSeller().GetId())
		if err != nil {
			ex.OrdersRepo.Delete(trade.GetSeller())
		} else {
			ex.OrdersRepo.Update(trade.GetSeller())
		}
	}
}
