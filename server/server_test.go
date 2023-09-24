package server

// import (
// 	"bytes"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"

// 	"github.com/labstack/echo/v4"
// 	"github.com/stretchr/testify/assert"
// )

// // func assert(t *testing.T, a, b any) {
// // 	if !reflect.DeepEqual(a, b) {
// // 		t.Errorf("%+v != %+v", a, b)
// // 	}
// // }

// func TestServerHandlePlaceOrder(t *testing.T) {
// 	// Setting up the Echo server for testing
// 	e := echo.New()

// 	// Create the request body
// 	orderBody := `{
// 		"Type": "LIMIT",
// 		"IsBid": true,
// 		"Size": 1,
// 		"Price": 10000,
// 		"Market": "ETH"
// 	}`

// 	req := httptest.NewRequest(http.MethodPost, "/order", bytes.NewReader([]byte(orderBody)))
// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

// 	// Record the response
// 	rec := httptest.NewRecorder()

// 	c := e.NewContext(req, rec)

// 	// Execute the handler
// 	if assert.NoError(t, NewExchange().handlePlaceOrder(c)) {
// 		assert.Equal(t, http.StatusOK, rec.Code)
// 		// assert.Equal(t, rec.Body.String(), "{\"msg\":\"limit order placed\"}")
// 	}
// }
