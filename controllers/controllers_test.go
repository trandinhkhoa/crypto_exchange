package controllers_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/trandinhkhoa/crypto-exchange/controllers"
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

func TestControllersHandlePlaceOrder(t *testing.T) {
	defer setupTest()()
	// Setting up the Echo controllers for testing
	e := echo.New()

	orderBody := `{
		"UserId" : "jane",
		"OrderType": "LIMIT",
		"IsBid": true,
		"Size": 1,
		"Price": 10000,
		"Ticker": "ETHUSD"
	}`

	req := httptest.NewRequest(http.MethodPost, "/order", bytes.NewReader([]byte(orderBody)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	// Record the response
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	ex.RegisterUserWithBalance("jane",
		map[string]float64{
			"ETH": 2000.0,
			"USD": 2000.0,
		})

	handler := controllers.NewWebServiceHandler(ex)
	// TODO: return error if request body is not in correct format .e.g wrong json field name
	if assert.NoError(t, handler.HandlePlaceOrder(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		pattern := `^{"msg":"limit order placed","order":{"ID":\d+,"UserId":"jane","IsBid":true,"Size":1,"Price":10000,"Timestamp":\d+}}\n$`

		re, err := regexp.Compile(pattern)
		assert.NoError(t, err)

		assert.True(t, re.MatchString(rec.Body.String()), "\nExpected: %s \nActual: %s", pattern, rec.Body.String())
	}
}
