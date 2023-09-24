package usecases_test

import (
	"testing"

	"github.com/trandinhkhoa/crypto-exchange/domain"
	"github.com/trandinhkhoa/crypto-exchange/usecases"
)

func TestPlaceLimitOrderExchange(t *testing.T) {
	ex := usecases.NewExchange()

	ex.RegisterUserWithBalance("john",
		map[string]float64{
			"ETH": 2000.0,
			"USD": 2000.0,
		})
	ex.RegisterUserWithBalance("jim",
		map[string]float64{
			"ETH": 2000.0,
			"USD": 2000.0,
		})
	ex.RegisterUserWithBalance("jane",
		map[string]float64{
			"ETH": 2000.0,
			"USD": 2000.0,
		})
	ex.RegisterUserWithBalance("jun",
		map[string]float64{
			"ETH": 2000.0,
			"USD": 2000.0,
		})
	ex.RegisterUserWithBalance("jack",
		map[string]float64{
			"ETH": 2000.0,
			"USD": 2000.0,
		})
	ex.RegisterUserWithBalance("lily",
		map[string]float64{
			"ETH": 2000.0,
			"USD": 2000.0,
		})

	// 1000(*)
	incomingOrder := domain.NewOrder("john", "ETHUSD", false, domain.LimitOrderType, 1, 100)
	ex.PlaceLimitOrder(*incomingOrder)

	// 900 < 1000(*)
	incomingOrder = domain.NewOrder("jim", "ETHUSD", false, domain.LimitOrderType, 1, 90)
	ex.PlaceLimitOrder(*incomingOrder)

	// 900 < 1000(*) < 1100
	incomingOrder = domain.NewOrder("jane", "ETHUSD", false, domain.LimitOrderType, 4, 110)
	ex.PlaceLimitOrder(*incomingOrder)

	// 900 < 1000(*) < 1005 < 1100
	incomingOrder = domain.NewOrder("jun", "ETHUSD", false, domain.LimitOrderType, 9, 105)
	ex.PlaceLimitOrder(*incomingOrder)

	// 900 < 1000(*) < 1005[2] < 1100
	incomingOrder = domain.NewOrder("jack", "ETHUSD", false, domain.LimitOrderType, 9, 105)
	ex.PlaceLimitOrder(*incomingOrder)

	incomingOrder = domain.NewOrder("lily", "ETHUSD", true, domain.MarketOrderType, 1, 0)
	ex.PlaceMarketOrder(*incomingOrder)

	assert(t, ex.GetUsersMap()["john"].Balance["ETH"], 1999.0)
	assert(t, ex.GetUsersMap()["john"].Balance["USD"], 2000.0)

	assert(t, ex.GetUsersMap()["jim"].Balance["ETH"], 1999.0)
	assert(t, ex.GetUsersMap()["jim"].Balance["USD"], 2090.0)

	assert(t, ex.GetUsersMap()["jane"].Balance["ETH"], 1996.0)
	assert(t, ex.GetUsersMap()["jane"].Balance["USD"], 2000.0)

	assert(t, ex.GetUsersMap()["jun"].Balance["ETH"], 1991.0)
	assert(t, ex.GetUsersMap()["jun"].Balance["USD"], 2000.0)

	assert(t, ex.GetUsersMap()["jack"].Balance["ETH"], 1991.0)
	assert(t, ex.GetUsersMap()["jack"].Balance["USD"], 2000.0)

	assert(t, ex.GetUsersMap()["lily"].Balance["ETH"], 2001.0)
	assert(t, ex.GetUsersMap()["lily"].Balance["USD"], 1910.0)
}
