package usecases_test

import (
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

var ex *usecases.Exchange

func deleteDb(filePath string) {
	if _, err := os.Stat(filePath); err == nil {
		err := os.Remove(filePath)
		if err != nil {
			logrus.Error("Error cleaning up test db: ", err)
			return
		}
	} else if os.IsNotExist(err) {
		logrus.Info("File does not exist, skipping")
	} else {
		logrus.Error("Error cleaning up test db: ", err)
		return
	}
}

func setup() (string, controllers.SqlDbHandler) {

	ex = usecases.NewExchange()
	// TODO: Mock the implementation of OrdersRepository and UsersRepository
	// short on time for now so i just use sqlite handler instead
	filePath := "./test.db"
	deleteDb(filePath)

	// injections of implementations
	dbHandler := infrastructure.NewSqliteDbHandler("./test.db")
	ordersRepoImpl := controllers.NewOrdersRepoImpl(dbHandler)
	ex.OrdersRepo = ordersRepoImpl
	usersRepoImpl := controllers.NewUsersRepoImpl(dbHandler)
	ex.UsersRepo = usersRepoImpl
	logrus.SetOutput(io.Discard)

	return filePath, dbHandler
}

// TODO: should hide even the fact that it is a SQL db
func teardown(filePath string, dbHandler controllers.SqlDbHandler) {
	dbHandler.Close()
	deleteDb(filePath)
}

func setupTest() func() {
	filePath, dbHandler := setup()

	return func() {
		teardown(filePath, dbHandler)
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
	ex.PlaceLimitOrderAndPersist(*incomingOrder)

	// 900 < 1000(*)
	incomingOrder = entities.NewOrder("jim", "ETHUSD", false, entities.LimitOrderType, 1, 90)
	ex.PlaceLimitOrderAndPersist(*incomingOrder)

	// 900 < 1000(*) < 1100
	incomingOrder = entities.NewOrder("jane", "ETHUSD", false, entities.LimitOrderType, 4, 110)
	ex.PlaceLimitOrderAndPersist(*incomingOrder)

	// 900 < 1000(*) < 1005 < 1100
	incomingOrder = entities.NewOrder("jun", "ETHUSD", false, entities.LimitOrderType, 9, 105)
	ex.PlaceLimitOrderAndPersist(*incomingOrder)

	// 900 < 1000(*) < 1005[2] < 1100
	incomingOrder = entities.NewOrder("jack", "ETHUSD", false, entities.LimitOrderType, 9, 105)
	ex.PlaceLimitOrderAndPersist(*incomingOrder)

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
	ex.PlaceLimitOrderAndPersist(*johnOrder)

	jimOrder := entities.NewOrder("jim", "ETHUSD", false, entities.LimitOrderType, 1, 90)
	ex.PlaceLimitOrderAndPersist(*jimOrder)

	janeOrder := entities.NewOrder("jane", "ETHUSD", false, entities.LimitOrderType, 4, 110)
	ex.PlaceLimitOrderAndPersist(*janeOrder)

	junOrder := entities.NewOrder("jun", "ETHUSD", false, entities.LimitOrderType, 9, 105)
	ex.PlaceLimitOrderAndPersist(*junOrder)

	jackOrder := entities.NewOrder("jack", "ETHUSD", false, entities.LimitOrderType, 9, 105)
	ex.PlaceLimitOrderAndPersist(*jackOrder)

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
