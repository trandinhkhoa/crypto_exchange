package usecases_test

import (
	"database/sql"
	"io"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/trandinhkhoa/crypto-exchange/controllers"
	"github.com/trandinhkhoa/crypto-exchange/entities"
	"github.com/trandinhkhoa/crypto-exchange/infrastructure"
	"github.com/trandinhkhoa/crypto-exchange/usecases"
)

// do extra setup or teardown before or after a test executes.
// It is also sometimes necessary to control which code runs on the main thread.
var ex *usecases.Exchange

func setup() (string, *sql.DB) {

	ex = usecases.NewExchange()
	// TODO: not pretty but i dont think the dependency rule is violated here. As package `usecases_test` is not really inside package `server`
	filePath := "./test.db"
	if _, err := os.Stat(filePath); err == nil {
		// File exists, proceed to delete
		err := os.Remove(filePath)
		if err != nil {
			logrus.Error("Error cleaning up test db: ", err)
			return filePath, nil
		}
	} else if os.IsNotExist(err) {
		// File does not exist, do nothing
		logrus.Info("File does not exist, skipping")
	} else {
		// Some other error occurred
		logrus.Error("Error cleaning up test db: ", err)
		return filePath, nil
	}
	dbHandler := infrastructure.SetupDatabase("./test.db")
	// injections of implementations
	ordersRepoImpl := controllers.NewOrdersRepoImpl(dbHandler)
	ex.OrdersRepo = ordersRepoImpl
	usersRepoImpl := controllers.NewUsersRepoImpl(dbHandler)
	ex.UsersRepo = usersRepoImpl
	logrus.SetOutput(io.Discard)

	return filePath, dbHandler
}

func teardown(filePath string, dbHandler *sql.DB) {
	//tear down
	dbHandler.Close()
	if _, err := os.Stat(filePath); err == nil {
		// File exists, proceed to delete
		err := os.Remove(filePath)
		if err != nil {
			logrus.Error("Error cleaning up test db: ", err)
			return
		}
	} else if os.IsNotExist(err) {
		// File does not exist, do nothing
		logrus.Error("File does not exist, skipping")
	} else {
		// Some other error occurred
		logrus.Error("Error cleaning up test db: ", err)
		return
	}
}

func setupTest() func() {
	// Setup code here
	filePath, dbHandler := setup()

	// tear down later
	return func() {
		teardown(filePath, dbHandler)
		// tear-down code here
	}
}

func TestPlaceLimitOrderExchange(t *testing.T) {
	defer setupTest()()

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
	incomingOrder := entities.NewOrder("john", "ETHUSD", false, entities.LimitOrderType, 1, 100)
	ex.PlaceLimitOrder(*incomingOrder)

	// 900 < 1000(*)
	incomingOrder = entities.NewOrder("jim", "ETHUSD", false, entities.LimitOrderType, 1, 90)
	ex.PlaceLimitOrder(*incomingOrder)

	// 900 < 1000(*) < 1100
	incomingOrder = entities.NewOrder("jane", "ETHUSD", false, entities.LimitOrderType, 4, 110)
	ex.PlaceLimitOrder(*incomingOrder)

	// 900 < 1000(*) < 1005 < 1100
	incomingOrder = entities.NewOrder("jun", "ETHUSD", false, entities.LimitOrderType, 9, 105)
	ex.PlaceLimitOrder(*incomingOrder)

	// 900 < 1000(*) < 1005[2] < 1100
	incomingOrder = entities.NewOrder("jack", "ETHUSD", false, entities.LimitOrderType, 9, 105)
	ex.PlaceLimitOrder(*incomingOrder)

	incomingOrder = entities.NewOrder("lily", "ETHUSD", true, entities.MarketOrderType, 1, 0)
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
	defer setupTest()()

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
	johnOrder := entities.NewOrder("john", "ETHUSD", false, entities.LimitOrderType, 1, 100)
	ex.PlaceLimitOrder(*johnOrder)

	jimOrder := entities.NewOrder("jim", "ETHUSD", false, entities.LimitOrderType, 1, 90)
	ex.PlaceLimitOrder(*jimOrder)

	janeOrder := entities.NewOrder("jane", "ETHUSD", false, entities.LimitOrderType, 4, 110)
	ex.PlaceLimitOrder(*janeOrder)

	junOrder := entities.NewOrder("jun", "ETHUSD", false, entities.LimitOrderType, 9, 105)
	ex.PlaceLimitOrder(*junOrder)

	jackOrder := entities.NewOrder("jack", "ETHUSD", false, entities.LimitOrderType, 9, 105)
	ex.PlaceLimitOrder(*jackOrder)

	ex.CancelOrder(jimOrder.GetId(), "ETHUSD")
	assert.Equal(t, 2000.0, ex.GetUsersMap()["jim"].Balance["ETH"])
	assert.Equal(t, 2000.0, ex.GetUsersMap()["jim"].Balance["USD"])

	lilyOrder := entities.NewOrder("lily", "ETHUSD", true, entities.MarketOrderType, 1, 0)
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
