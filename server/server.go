package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
	"github.com/trandinhkhoa/crypto-exchange/entities"
	"github.com/trandinhkhoa/crypto-exchange/usecases"
	"golang.org/x/net/websocket"
)

type OrderBookResponse struct {
	TotalAsksVolume float64
	TotalBidsVolume float64
	Asks            []*OrderResponse
	Bids            []*OrderResponse
}

// TODO: perhaps user'id should not come from the body
type OrderResponse struct {
	ID        int
	UserId    string
	IsBid     bool
	Size      float64
	Price     float64
	Timestamp int64
}

type TradeResponse struct {
	Price        float64
	Size         float64
	IsBuyerMaker bool
	Timestamp    int64
}
type LimitResponse struct {
	Price  float64
	Volume float64
}

// fields need to be visible to outer packages since this struct will be used by package json
type PlaceOrderRequest struct {
	UserId    string
	OrderType entities.OrderType // limit or ticker
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
	incomingOrder := entities.NewOrder(
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

	if placeOrderData.OrderType == entities.MarketOrderType {
		trades := handler.ex.PlaceMarketOrder(*incomingOrder)
		tradesDataArray := make([]TradeResponse, 0)
		for _, trade := range trades {
			tradeData := &TradeResponse{
				Timestamp:    trade.GetTimeStamp(),
				Price:        trade.GetPrice(),
				Size:         trade.GetSize(),
				IsBuyerMaker: trade.GetIsBuyerMaker(),
			}
			tradesDataArray = append(tradesDataArray, *tradeData)
		}
		return c.JSON(200, map[string]interface{}{"matches": tradesDataArray})
	} else {
		handler.ex.PlaceLimitOrder(*incomingOrder)
		return c.JSON(200, map[string]interface{}{
			"msg": "limit order placed",
			"order": OrderResponse{
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
	orderBookData := OrderBookResponse{
		TotalAsksVolume: 0.0,
		TotalBidsVolume: 0.0,
		Asks:            make([]*OrderResponse, 0),
		Bids:            make([]*OrderResponse, 0),
	}

	buybook, bVolume, sellbook, sVolume := handler.ex.GetBook(string(ticker))
	// TODO: controller layer should not depend/know the implementation of the book like this
	// (the fact that each price level (limit) is implemented as a linked list)
	for _, limit := range buybook {
		order := limit.HeadOrder
		for order != nil {
			orderData := &OrderResponse{
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
			orderData := &OrderResponse{
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
func (handler WebServiceHandler) WebSocketHandlerCurrentPrice(ws *websocket.Conn) {
	// TODO: let client choose ticker
	ticker := "ETHUSD"
	lastCurrentPrice := 0.0
	currentPrice := lastCurrentPrice

	for {
		// reading a field (float64) instead of slice as a dirty work around for the issue described below
		// still racy but float64 will hide the issue
		// with a slice, it's more obvious. because i might be reading a newly allocated entry that has not been filled yet
		// TODO: use channel
		currentPrice = handler.ex.GetLastPrice(ticker)
		if currentPrice != lastCurrentPrice {
			lastCurrentPrice = currentPrice
			msg := fmt.Sprint(currentPrice)

			if err := websocket.Message.Send(ws, msg); err != nil {
				fmt.Println("Can't send:", err)
				break
			} else {
				logrus.WithFields(logrus.Fields{
					"msg": msg,
				}).Info("Sent to client")
			}
		}
		// TODO: w/o sleep -> currentPrice == 0/weird value. The book might be queried too fast and even when it is empty ???
		// when this happen , the orderbook is not empty, lastTrades is not empty either, lastTrades when debugged even show the correct value
		// w/o low sleep time (100 ns) the issue is also more prevalent if i increase the number of socket listeners (1 listener ok, 3 listener not ok)
		// another dirty "workaround" , faster market maker ?? -> No --> the issue only happen when this function is called
		// even after creating a function that return  float64as the last price did not help
		// this function is being ran concurrently (each HTTP request has its own goroutine )
		//, it might trigger read access to trades[] at the same time trades is being written into , hence the weird value
		// time.Sleep(100 * time.Nanosecond)
	}
}

func (handler WebServiceHandler) WebSocketHandlerLastTrade(ws *websocket.Conn) {
	ticker := "ETHUSD"

	for {
		arr := handler.ex.GetLastTrades(ticker, 15)
		responsesArr := make([]TradeResponse, 0)
		for _, trade := range arr {
			response := TradeResponse{
				Price:        trade.GetPrice(),
				Size:         trade.GetSize(),
				IsBuyerMaker: trade.GetIsBuyerMaker(),
				Timestamp:    trade.GetTimeStamp(),
			}
			responsesArr = append(responsesArr, response)
		}

		arrayJSON, _ := json.Marshal(responsesArr)

		if err := websocket.Message.Send(ws, string(arrayJSON)); err != nil {
			fmt.Println("Can't send:", err)
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func (handler WebServiceHandler) WebSocketHandlerBestBuys(ws *websocket.Conn) {
	ticker := "ETHUSD"

	for {
		arr := handler.ex.GetBestBuys(ticker, 15)
		responsesArr := make([]LimitResponse, 0)
		for _, limit := range arr {
			response := LimitResponse{
				Price:  limit.GetLimitPrice(),
				Volume: limit.TotalVolume,
			}
			responsesArr = append(responsesArr, response)
		}

		arrayJSON, _ := json.Marshal(responsesArr)

		if err := websocket.Message.Send(ws, string(arrayJSON)); err != nil {
			fmt.Println("Can't send:", err)
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func (handler WebServiceHandler) WebSocketHandlerBestSells(ws *websocket.Conn) {
	ticker := "ETHUSD"

	for {
		arr := handler.ex.GetBestSells(ticker, 15)
		responsesArr := make([]LimitResponse, 0)
		for _, limit := range arr {
			response := LimitResponse{
				Price:  limit.GetLimitPrice(),
				Volume: limit.TotalVolume,
			}
			responsesArr = append(responsesArr, response)
		}

		arrayJSON, _ := json.Marshal(responsesArr)

		if err := websocket.Message.Send(ws, string(arrayJSON)); err != nil {
			fmt.Println("Can't send:", err)
			break
		}
		time.Sleep(500 * time.Millisecond)
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

	e.GET("/ws/currentPrice", echo.WrapHandler(websocket.Handler(handler.WebSocketHandlerCurrentPrice)))
	e.GET("/ws/lastTrades", echo.WrapHandler(websocket.Handler(handler.WebSocketHandlerLastTrade)))
	e.GET("/ws/bestSells", echo.WrapHandler(websocket.Handler(handler.WebSocketHandlerBestSells)))
	e.GET("/ws/bestBuys", echo.WrapHandler(websocket.Handler(handler.WebSocketHandlerBestBuys)))

	e.Start(":3000")
}
