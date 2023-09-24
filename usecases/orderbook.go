package usecases

import "github.com/trandinhkhoa/crypto-exchange/domain"

// Book
//
//	Limit *buyTree;
//	Limit *sellTree;
//	Limit *lowestSell;
//	Limit *highestBuy;
type Orderbook struct {
	buyTree    *domain.Limit
	sellTree   *domain.Limit
	lowestSell *domain.Limit
	highestBuy *domain.Limit
}
