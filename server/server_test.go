package server_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/trandinhkhoa/crypto-exchange/server"
	"github.com/trandinhkhoa/crypto-exchange/usecases"
)

func TestServerHandlePlaceOrder(t *testing.T) {
	// Setting up the Echo server for testing
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

	ex := usecases.NewExchange()

	// injections of implementations
	ordersRepoImpl := server.OrdersRepoImpl{}
	ex.OrdersRepo = ordersRepoImpl
	usersRepoImpl := server.UsersRepoImpl{}
	ex.UsersRepo = usersRepoImpl

	ex.RegisterUserWithBalance("jane",
		map[string]float64{
			"ETH": 2000.0,
			"USD": 2000.0,
		})
	handler := server.NewWebServiceHandler(ex)
	// TODO: return error if request body is not in correct format .e.g wrong json field name
	if assert.NoError(t, handler.HandlePlaceOrder(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		pattern := `^{"msg":"limit order placed","order":{"ID":\d+,"UserId":"jane","IsBid":true,"Size":1,"Price":10000,"Timestamp":\d+}}\n$`

		re, err := regexp.Compile(pattern)
		assert.NoError(t, err)

		assert.True(t, re.MatchString(rec.Body.String()), "\nExpected: %s \nActual: %s", pattern, rec.Body.String())
	}
}
