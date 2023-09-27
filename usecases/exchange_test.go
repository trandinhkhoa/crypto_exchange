package usecases_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
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

	assert.Equal(t, ex.GetUsersMap()["john"].Balance["ETH"], 1999.0)
	assert.Equal(t, ex.GetUsersMap()["john"].Balance["USD"], 2000.0)

	assert.Equal(t, ex.GetUsersMap()["jim"].Balance["ETH"], 1999.0)
	assert.Equal(t, ex.GetUsersMap()["jim"].Balance["USD"], 2090.0)

	assert.Equal(t, ex.GetUsersMap()["jane"].Balance["ETH"], 1996.0)
	assert.Equal(t, ex.GetUsersMap()["jane"].Balance["USD"], 2000.0)

	assert.Equal(t, ex.GetUsersMap()["jun"].Balance["ETH"], 1991.0)
	assert.Equal(t, ex.GetUsersMap()["jun"].Balance["USD"], 2000.0)

	assert.Equal(t, ex.GetUsersMap()["jack"].Balance["ETH"], 1991.0)
	assert.Equal(t, ex.GetUsersMap()["jack"].Balance["USD"], 2000.0)

	// TODO: assert.Equal should not hide the line with the error
	assert.Equal(t, ex.GetUsersMap()["lily"].Balance["ETH"], 2001.0)
	assert.Equal(t, ex.GetUsersMap()["lily"].Balance["USD"], 1910.0)
}

func TestCancelOrderExchange(t *testing.T) {
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

	// Root: 1000
	// L--- 900
	// R--- 1005
	//     R--- 1100
	johnOrder := domain.NewOrder("john", "ETHUSD", false, domain.LimitOrderType, 1, 100)
	ex.PlaceLimitOrder(*johnOrder)

	jimOrder := domain.NewOrder("jim", "ETHUSD", false, domain.LimitOrderType, 1, 90)
	ex.PlaceLimitOrder(*jimOrder)

	janeOrder := domain.NewOrder("jane", "ETHUSD", false, domain.LimitOrderType, 4, 110)
	ex.PlaceLimitOrder(*janeOrder)

	junOrder := domain.NewOrder("jun", "ETHUSD", false, domain.LimitOrderType, 9, 105)
	ex.PlaceLimitOrder(*junOrder)

	jackOrder := domain.NewOrder("jack", "ETHUSD", false, domain.LimitOrderType, 9, 105)
	ex.PlaceLimitOrder(*jackOrder)

	ex.CancelOrder(jimOrder.GetId(), "ETHUSD")
	assert.Equal(t, 2000.0, ex.GetUsersMap()["jim"].Balance["ETH"])
	assert.Equal(t, 2000.0, ex.GetUsersMap()["jim"].Balance["USD"])

	lilyOrder := domain.NewOrder("lily", "ETHUSD", true, domain.MarketOrderType, 1, 0)
	ex.PlaceMarketOrder(*lilyOrder)

	assert.Equal(t, 1996.0, ex.GetUsersMap()["jane"].Balance["ETH"])
	assert.Equal(t, 2000.0, ex.GetUsersMap()["jane"].Balance["USD"])

	assert.Equal(t, 1991.0, ex.GetUsersMap()["jun"].Balance["ETH"])
	assert.Equal(t, 2000.0, ex.GetUsersMap()["jun"].Balance["USD"])

	assert.Equal(t, 1991.0, ex.GetUsersMap()["jack"].Balance["ETH"])
	assert.Equal(t, 2000.0, ex.GetUsersMap()["jack"].Balance["USD"])

	// jim's balance is restored
	assert.Equal(t, 2000.0, ex.GetUsersMap()["jim"].Balance["ETH"])
	assert.Equal(t, 2000.0, ex.GetUsersMap()["jim"].Balance["USD"])
	// john matched with lily
	assert.Equal(t, 1999.0, ex.GetUsersMap()["john"].Balance["ETH"])
	assert.Equal(t, 2100.0, ex.GetUsersMap()["john"].Balance["USD"])

	assert.Equal(t, 2001.0, ex.GetUsersMap()["lily"].Balance["ETH"])
	assert.Equal(t, 1900.0, ex.GetUsersMap()["lily"].Balance["USD"])
}
