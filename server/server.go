package server

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
	"github.com/trandinhkhoa/crypto-exchange/orderbook"
	"github.com/trandinhkhoa/crypto-exchange/users"
	"golang.org/x/net/websocket"
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
	UserId string
	Type   OrderType // limit or market
	IsBid  bool
	Size   float64
	Price  float64
	Market MarketType
}

type Exchange struct {
	orderbooks  map[MarketType]*orderbook.OrderBook
	idToUserMap map[string]*users.User
	hotWallet   *users.Wallet
}

func NewExchange() *Exchange {
	aMap := make(map[MarketType]*orderbook.OrderBook)
	usersMap := make(map[string]*users.User)
	aMap[ETHMarketType] = orderbook.NewOrderbook((*users.Users)(&usersMap))
	return &Exchange{
		orderbooks:  aMap,
		idToUserMap: usersMap,
		// TODO: what if no &
		hotWallet: &users.Wallet{
			PublicKey:  "",
			PrivateKey: "",
			Address:    "",
		},
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
		placeOrderData.UserId,
		string(placeOrderData.Type),
	)
	if placeOrderData.Type == MarketOrderType {
		matches := ex.orderbooks[placeOrderData.Market].PlaceMarketOrder(incomingOrder)
		return c.JSON(200, map[string]interface{}{"matches": matches})
	} else {
		user := ex.idToUserMap[incomingOrder.UserId]
		if incomingOrder.IsBid {
			user.Balance["USD"] -= incomingOrder.Price * incomingOrder.Size
		} else {
			user.Balance["ETH"] -= incomingOrder.Size
		}
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

// this function is called everytime a client connect to the websocket
func (ex *Exchange) WebSocketHandler(ws *websocket.Conn) {
	lastCurrentPrice := ex.orderbooks["ETH"].CurrentPrice
	fmt.Println("Hi I am WebSocketHandler")

	for {
		currentPrice := ex.orderbooks["ETH"].CurrentPrice
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

func generateRandomString() string {
	prefix := "id"
	length := 10
	source := rand.NewSource(time.Now().UnixNano())
	randomizer := rand.New(source)

	letterRunes := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	b := strings.Builder{}
	b.WriteString(prefix)

	for i := 0; i < length-len(prefix); i++ {
		randomIndex := randomizer.Intn(len(letterRunes))
		b.WriteRune(letterRunes[randomIndex])
	}

	return b.String()
}

func (ex *Exchange) RegisterUser(newUser *users.User) {
	// TODO: check new id uniqueness
	ex.idToUserMap[newUser.Id] = newUser
}

func (ex *Exchange) handleGetUsers(c echo.Context) error {
	return c.JSON(200, ex.idToUserMap)
}

func (ex *Exchange) handleGetUser(c echo.Context) error {
	userId := c.Param("userId")
	user, ok := ex.idToUserMap[userId]
	if !ok {
		return c.JSON(404, fmt.Sprintf("UserId %s does not exist", userId))
	}
	return c.JSON(200, user)
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
	ex.RegisterUser(&users.User{
		// Id:         generateRandomString(),
		Id:         "maker123",
		PrivateKey: "",
		Balance: map[string]float64{
			"ETH": 10000.0,
			"USD": 1000000.0,
		},
	})
	ex.RegisterUser(&users.User{
		Id:         "traderJoe123",
		PrivateKey: "",
		Balance: map[string]float64{
			"ETH": 10.0,
			"USD": 1010.0,
		},
	})
	ex.RegisterUser(&users.User{
		Id:         "mememe",
		PrivateKey: "",
		Balance: map[string]float64{
			"ETH": 0.0,
			"USD": 1010.0,
		},
	})

	e.POST("/order", ex.handlePlaceOrder)

	e.GET("/users", ex.handleGetUsers)
	e.GET("/users/:userId", ex.handleGetUser)
	e.GET("/book/:market", ex.handleGetBook)
	e.GET("/book/:market/currentPrice", ex.handleGetCurrentPrice)
	e.GET("/book/:market/bestAsk", ex.handleGetBestAsk)
	e.GET("/book/:market/bestBid", ex.handleGetBestBid)

	e.DELETE("/order/:id", ex.handleCancelOrder)

	e.GET("/ws/currentPrice", echo.WrapHandler(websocket.Handler(ex.WebSocketHandler)))

	e.Start(":3000")
}
