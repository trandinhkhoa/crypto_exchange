package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
	"github.com/trandinhkhoa/crypto-exchange/domain"
	"github.com/trandinhkhoa/crypto-exchange/usecases"
	"golang.org/x/net/websocket"
)

type OrderBookData struct {
	TotalAsksVolume float64
	TotalBidsVolume float64
	Asks            []*OrderData
	Bids            []*OrderData
}

// TODO: perhaps user'id should not come from the body
type OrderData struct {
	ID        int
	UserId    string
	IsBid     bool
	Size      float64
	Price     float64
	Timestamp int64
}

type TradeData struct {
	Price     float64
	Size      float64
	Timestamp int64
}

// fields need to be visible to outer packages since this struct will be used by package json
type PlaceOrderRequest struct {
	UserId    string
	OrderType domain.OrderType // limit or ticker
	IsBid     bool
	Size      float64
	Price     float64
	Ticker    string
}

type WebServiceHandler struct {
	ex *usecases.Exchange
}

func NewWebServiceHandler(ex *usecases.Exchange) *WebServiceHandler {
	handler := WebServiceHandler{}
	handler.ex = ex
	return &handler
}

func (handler WebServiceHandler) HandlePlaceOrder(c echo.Context) error {
	var placeOrderData PlaceOrderRequest
	// TODO: check if msg body has all the required fields
	if err := json.NewDecoder(c.Request().Body).Decode(&placeOrderData); err != nil {
		return err
	}
	incomingOrder := domain.NewOrder(
		placeOrderData.UserId,
		placeOrderData.Ticker,
		placeOrderData.IsBid,
		placeOrderData.OrderType,
		placeOrderData.Size,
		placeOrderData.Price,
	)

	// TODO: check if userid exist
	_, ok := handler.ex.GetUsersMap()[placeOrderData.UserId]
	if !ok {
		msg := fmt.Sprintf("userId %s does not exist", placeOrderData.UserId)
		return c.JSON(400, map[string]interface{}{"msg": msg})
	}

	if placeOrderData.OrderType == domain.MarketOrderType {
		trades := handler.ex.PlaceMarketOrder(*incomingOrder)
		tradesDataArray := make([]TradeData, 0)
		for _, trade := range trades {
			tradeData := &TradeData{
				Timestamp: trade.GetTimeStamp(),
				Price:     trade.GetPrice(),
				Size:      trade.GetSize(),
			}
			tradesDataArray = append(tradesDataArray, *tradeData)
		}
		return c.JSON(200, map[string]interface{}{"matches": tradesDataArray})
	} else {
		handler.ex.PlaceLimitOrder(*incomingOrder)
		return c.JSON(200, map[string]interface{}{
			"msg": "limit order placed",
			"order": OrderData{
				ID:        int(incomingOrder.GetId()),
				UserId:    incomingOrder.GetUserId(),
				IsBid:     incomingOrder.GetIsBid(),
				Size:      incomingOrder.Size,
				Price:     incomingOrder.GetLimitPrice(),
				Timestamp: incomingOrder.GetTimeStamp(),
			},
		})
	}
}

func (handler WebServiceHandler) HandleGetBook(c echo.Context) error {
	// TODO: "ticker" -> "ticker"
	// TODO: should not need to convert to usescase.TIcker
	ticker := usecases.Ticker(c.Param("ticker"))
	orderBookData := OrderBookData{
		TotalAsksVolume: 0.0,
		TotalBidsVolume: 0.0,
		Asks:            make([]*OrderData, 0),
		Bids:            make([]*OrderData, 0),
	}

	buybook, bVolume, sellbook, sVolume := handler.ex.GetBook(string(ticker))
	// TODO: controller layer should not depend/know the implementation of the book like this
	// (the fact that each price level (limit) is implemented as a linked list)
	for _, limit := range buybook {
		order := limit.HeadOrder
		for order != nil {
			orderData := &OrderData{
				ID:        int(order.GetId()),
				IsBid:     order.GetIsBid(),
				Size:      order.Size,
				Price:     order.GetLimitPrice(),
				Timestamp: order.GetTimeStamp(),
			}
			orderBookData.Bids = append(orderBookData.Bids, orderData)
			order = order.NextOrder
		}
	}
	for _, limit := range sellbook {
		order := limit.HeadOrder
		for order != nil {
			orderData := &OrderData{
				ID:        int(order.GetId()),
				IsBid:     order.GetIsBid(),
				Size:      order.Size,
				Price:     order.GetLimitPrice(),
				Timestamp: order.GetTimeStamp(),
			}
			orderBookData.Asks = append(orderBookData.Asks, orderData)
			order = order.NextOrder
		}
	}
	orderBookData.TotalAsksVolume = sVolume
	orderBookData.TotalBidsVolume = bVolume

	return c.JSON(200, orderBookData)
}

func (handler WebServiceHandler) HandleGetCurrentPrice(c echo.Context) error {
	ticker := c.Param("ticker")
	lastTrades := handler.ex.GetLastTrades(ticker, 1)
	if len(lastTrades) == 0 {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"msg": "no trades yet",
		})
	}
	currentPrice := lastTrades[0].GetPrice()
	return c.JSON(http.StatusOK, map[string]interface{}{
		"currentPrice": currentPrice,
	})
}

func (handler WebServiceHandler) HandleGetBestAsk(c echo.Context) error {
	ticker := c.Param("ticker")
	bestAskPrice := handler.ex.GetBestSell(ticker)
	return c.JSON(http.StatusOK, map[string]interface{}{
		"bestAskPrice": bestAskPrice,
	})
}

func (handler WebServiceHandler) HandleGetBestBid(c echo.Context) error {
	ticker := c.Param("ticker")
	bestBidPrice := handler.ex.GetBestBuy(ticker)
	return c.JSON(http.StatusOK, map[string]interface{}{
		"bestBidPrice": bestBidPrice,
	})
}

func (handler WebServiceHandler) HandleCancelOrder(c echo.Context) error {
	orderId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(400, map[string]interface{}{
			"msg": "id not numeric",
		})
	}
	ticker := c.Param("ticker")
	handler.ex.CancelOrder(int64(orderId), ticker)
	return c.JSON(200, map[string]interface{}{
		"msg": "order cancelled",
	})
}

// TODO: dont always return 400
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

// this function is called everytime a client connect to the websocket
func (handler WebServiceHandler) WebSocketHandler(ws *websocket.Conn) {
	// TODO: let client choose ticker
	ticker := "ETHUSD"
	lastTrades := handler.ex.GetLastTrades(ticker, 1)
	lastCurrentPrice := lastTrades[0].GetPrice()
	fmt.Println("Hi I am WebSocketHandler")

	for {
		lastTrades = handler.ex.GetLastTrades(ticker, 1)
		currentPrice := lastTrades[0].GetPrice()
		if currentPrice == 0 {
			fmt.Println("YOYOYO", lastTrades)
		}
		if currentPrice != lastCurrentPrice {
			lastCurrentPrice = currentPrice
			msg := fmt.Sprint(currentPrice)
			// Send the received message back to the client

			if err := websocket.Message.Send(ws, msg); err != nil {
				fmt.Println("Can't send:", err)
				break
			} else {
				logrus.WithFields(logrus.Fields{
					"msg": msg,
				}).Info("Sent to client")
			}
			// time.Sleep(1 * time.Second)
		}
	}
}

// TODO:
func (handler WebServiceHandler) registerUser(c echo.Context) {
}

// TODO: dont return all details about users ?
// or maybe check the right of the requester
func (handler WebServiceHandler) HandleGetUsers(c echo.Context) error {
	return c.JSON(200, handler.ex.GetUsersMap())
}

// TODO: dont return all details about users ?
func (handler WebServiceHandler) HandleGetUser(c echo.Context) error {
	userId := c.Param("userId")
	user, ok := handler.ex.GetUsersMap()[userId]
	if !ok {
		return c.JSON(404, fmt.Sprintf("UserId %s does not exist", userId))
	}
	return c.JSON(200, user)
}

// TODO: this is infrastructure code. need to separate it from the controller above
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

	handler := WebServiceHandler{}
	ex := usecases.NewExchange()

	handler.ex = ex

	handler.ex.RegisterUserWithBalance("maker123",
		map[string]float64{
			"ETH": 10000.0,
			"USD": 1000000.0,
		},
	)
	handler.ex.RegisterUserWithBalance("traderJoe123",
		map[string]float64{
			"ETH": 10.0,
			"USD": 1000.0,
		},
	)
	handler.ex.RegisterUserWithBalance("me",
		map[string]float64{
			"ETH": 0.0,
			"USD": 1000.0,
		},
	)

	e.POST("/order", handler.HandlePlaceOrder)

	e.GET("/users", handler.HandleGetUsers)
	e.GET("/users/:userId", handler.HandleGetUser)
	e.GET("/book/:ticker", handler.HandleGetBook)
	// TODO: handle error when ticker does not exist
	e.GET("/book/:ticker/currentPrice", handler.HandleGetCurrentPrice)
	// TODO: handle error when this is called while no bid/ask is in the book
	e.GET("/book/:ticker/bestAsk", handler.HandleGetBestAsk)
	e.GET("/book/:ticker/bestBid", handler.HandleGetBestBid)

	e.DELETE("/order/:ticker/:id", handler.HandleCancelOrder)

	e.GET("/ws/currentPrice", echo.WrapHandler(websocket.Handler(handler.WebSocketHandler)))

	e.Start(":3000")
}
