package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/trandinhkhoa/crypto-exchange/orderbook"
)

type OrderBookData struct {
	TotalAsksVolume float64
	TotalBidsVolume float64
	Asks            []*OrderData
	Bids            []*OrderData
}

type OrderData struct {
	ID        int
	IsBid     bool
	Size      float64
	Price     float64
	Timestamp int64
}

type MatchData struct {
	Ask        OrderData
	Bid        OrderData
	SizeFilled float64
	Price      float64
}

type OrderType string

const (
	MarketOrderType OrderType = "MARKET"
	LimitOrderType  OrderType = "LIMIT"
)

type MarketType string

const (
	ETHMarketType MarketType = "ETH"
)

// fields need to be visible to outer packages since this struct will be used by package json
type PlaceOrderRequest struct {
	Type   OrderType // limit or market
	IsBid  bool
	Size   float64
	Price  float64
	Market MarketType
}

type Exchange struct {
	orderbooks map[MarketType]*orderbook.OrderBook
}

func NewExchange() *Exchange {
	aMap := make(map[MarketType]*orderbook.OrderBook)
	aMap[ETHMarketType] = orderbook.NewOrderbook()
	return &Exchange{
		orderbooks: aMap,
	}
}

func (ex *Exchange) handlePlaceOrder(c echo.Context) error {
	var placeOrderData PlaceOrderRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&placeOrderData); err != nil {
		return err
	}
	incomingOrder := orderbook.NewOrder(
		placeOrderData.IsBid,
		placeOrderData.Size,
	)
	if placeOrderData.Type == MarketOrderType {
		matches := ex.orderbooks[placeOrderData.Market].PlaceMarketOrder(incomingOrder)
		return c.JSON(200, map[string]interface{}{"matches": matches})
	} else {
		ex.orderbooks[placeOrderData.Market].PlaceLimitOrder(placeOrderData.Price, incomingOrder)
		return c.JSON(200, map[string]interface{}{
			"msg": "limit order placed",
			"order": OrderData{
				ID:        incomingOrder.ID,
				IsBid:     incomingOrder.IsBid,
				Size:      incomingOrder.Size,
				Price:     incomingOrder.Price,
				Timestamp: incomingOrder.Timestamp,
			},
		})
	}
}

func (ex *Exchange) handleGetBook(c echo.Context) error {
	marketType := MarketType(c.Param("market"))
	orderBookData := OrderBookData{
		TotalAsksVolume: 0.0,
		TotalBidsVolume: 0.0,
		Asks:            make([]*OrderData, 0),
		Bids:            make([]*OrderData, 0),
	}

	for _, iterator := range ex.orderbooks[marketType].AskLimits {
		for _, order := range iterator.Orders {
			orderData := &OrderData{
				ID:        order.ID,
				IsBid:     order.IsBid,
				Size:      order.Size,
				Price:     order.Price,
				Timestamp: order.Timestamp,
			}
			orderBookData.Asks = append(orderBookData.Asks, orderData)
		}
	}
	for _, iterator := range ex.orderbooks[marketType].BidLimits {
		for _, order := range iterator.Orders {
			orderData := &OrderData{
				ID:        order.ID,
				IsBid:     order.IsBid,
				Size:      order.Size,
				Price:     order.Price,
				Timestamp: order.Timestamp,
			}
			orderBookData.Bids = append(orderBookData.Bids, orderData)
		}
	}
	orderBookData.TotalAsksVolume += ex.orderbooks[marketType].GetTotalVolumeAllAsks()
	orderBookData.TotalBidsVolume += ex.orderbooks[marketType].GetTotalVolumeAllBids()

	return c.JSON(200, orderBookData)
}

func (ex *Exchange) handleGetCurrentPrice(c echo.Context) error {
	marketType := MarketType(c.Param("market"))
	currentPrice := ex.orderbooks[marketType].CurrentPrice
	return c.JSON(http.StatusOK, map[string]interface{}{
		"currentPrice": currentPrice,
	})
}

func (ex *Exchange) handleGetBestAsk(c echo.Context) error {
	marketType := MarketType(c.Param("market"))
	bestAskPrice := ex.orderbooks[marketType].GetBestAsk().Price
	return c.JSON(http.StatusOK, map[string]interface{}{
		"bestAskPrice": bestAskPrice,
	})
}

func (ex *Exchange) handleGetBestBid(c echo.Context) error {
	marketType := MarketType(c.Param("market"))
	bestBidPrice := ex.orderbooks[marketType].GetBestBid().Price
	return c.JSON(http.StatusOK, map[string]interface{}{
		"bestBidPrice": bestBidPrice,
	})
}

func (ex *Exchange) handleCancelOrder(c echo.Context) error {
	resp := "handleCancelOrder"
	cancelledOrderID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		// ... handle error
		// TODO: check if this is executed
		return err
	}
	// for now, assuming we will only ever have 1 market ETH
	order, ok := ex.orderbooks[ETHMarketType].IDToOrderMap[cancelledOrderID]
	if ok {
		ex.orderbooks[ETHMarketType].CancelOrder(order)
	} else {
		// TODO: check if this is executed
		panic("order not found")
	}
	return c.JSON(200, resp)
}

// not a good idea to always return 400
func httpErrorHandler(err error, c echo.Context) {
	fmt.Println(err)
	c.JSON(http.StatusBadRequest, err)
}

func PanicRecoveryMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					err, ok := r.(error)
					var panicMsg string
					if ok {
						panicMsg = err.Error()
					} else {
						panicMsg = fmt.Sprintf("%v", r)
					}
					// log the panic message and stack trace here.
					errStr := fmt.Errorf("panic recovered: %s\n%s", panicMsg, string(debug.Stack()))
					fmt.Println(errStr)

					// Return the panic message to the client.
					// For security reasons, might need to replace the detailed message with a generic one in production.
					c.JSON(http.StatusInternalServerError, map[string]interface{}{
						"message": panicMsg,
					})
				}
			}()
			return next(c)
		}
	}
}

func StartServer() {
	e := echo.New()
	// Recover middleware to catch panics
	e.Use(middleware.Recover())
	e.Use(PanicRecoveryMiddleware())

	e.HTTPErrorHandler = httpErrorHandler

	// client, err := ethclient.Dial("http://localhost:8545")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	ex := NewExchange()
	// ex, err := NewExchange(exchangePrivateKey, client)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// ex.registerUser("829e924fdf021ba3dbbc4225edfece9aca04b929d6e75613329ca6f1d31c0bb4", 8)
	// ex.registerUser("a453611d9419d0e56f499079478fd72c37b251a94bfde4d19872c44cf65386e3", 7)
	// ex.registerUser("e485d098507f54e7733a205420dfddbe58db035fa577fc294ebd14db90767a52", 666)

	e.POST("/order", ex.handlePlaceOrder)

	e.GET("/book/:market", ex.handleGetBook)
	e.GET("/book/:market/currentPrice", ex.handleGetCurrentPrice)
	e.GET("/book/:market/bestAsk", ex.handleGetBestAsk)
	e.GET("/book/:market/bestBid", ex.handleGetBestBid)

	e.DELETE("/order/:id", ex.handleCancelOrder)

	e.Start(":3000")
}
