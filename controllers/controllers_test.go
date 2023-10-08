package controllers_test

import (
	"bytes"
	"database/sql"
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

func setup() (string, *sql.DB) {

	ex = usecases.NewExchange()
	// TODO: not pretty but i dont think the dependency rule is violated here. Unlike `controllers`, package `controllers_test` is not really inside package `server`
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
		logrus.Error("File does not exist, skipping")
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
		logrus.Info("File does not exist, skipping")
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
