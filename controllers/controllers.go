package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
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
	Ex *usecases.Exchange
}

func NewWebServiceHandler(ex *usecases.Exchange) *WebServiceHandler {
	handler := WebServiceHandler{}
	handler.Ex = ex
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
	_, ok := handler.Ex.GetUsersMap()[placeOrderData.UserId]
	if !ok {
		msg := fmt.Sprintf("userId %s does not exist", placeOrderData.UserId)
		return c.JSON(400, map[string]interface{}{"msg": msg})
	}

	if placeOrderData.OrderType == entities.MarketOrderType {
		trades := handler.Ex.PlaceMarketOrder(*incomingOrder)
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
		handler.Ex.PlaceLimitOrderAndPersist(*incomingOrder)
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
	// TODO: should not need to convert to usescase.TIcker
	ticker := usecases.Ticker(c.Param("ticker"))
	orderBookData := OrderBookResponse{
		TotalAsksVolume: 0.0,
		TotalBidsVolume: 0.0,
		Asks:            make([]*OrderResponse, 0),
		Bids:            make([]*OrderResponse, 0),
	}

	buybook, bVolume, sellbook, sVolume := handler.Ex.GetBook(string(ticker))
	for _, limit := range buybook {
		ordersList := limit.GetAllOrders()
		for _, order := range ordersList {
			orderData := &OrderResponse{
				ID:        int(order.GetId()),
				IsBid:     order.GetIsBid(),
				Size:      order.GetSize(),
				Price:     order.GetLimitPrice(),
				Timestamp: order.GetTimeStamp(),
			}
			orderBookData.Bids = append(orderBookData.Bids, orderData)
		}
	}
	for _, limit := range sellbook {
		ordersList := limit.GetAllOrders()
		for _, order := range ordersList {
			orderData := &OrderResponse{
				ID:        int(order.GetId()),
				IsBid:     order.GetIsBid(),
				Size:      order.GetSize(),
				Price:     order.GetLimitPrice(),
				Timestamp: order.GetTimeStamp(),
			}
			orderBookData.Asks = append(orderBookData.Asks, orderData)
		}
	}
	orderBookData.TotalAsksVolume = sVolume
	orderBookData.TotalBidsVolume = bVolume

	return c.JSON(200, orderBookData)
}

func (handler WebServiceHandler) HandleGetCurrentPrice(c echo.Context) error {
	ticker := c.Param("ticker")
	lastTrades := handler.Ex.GetLastTrades(ticker, 1)
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
	bestAskPrice := handler.Ex.GetBestSell(ticker)
	return c.JSON(http.StatusOK, map[string]interface{}{
		"bestAskPrice": bestAskPrice,
	})
}

func (handler WebServiceHandler) HandleGetBestBid(c echo.Context) error {
	ticker := c.Param("ticker")
	bestBidPrice := handler.Ex.GetBestBuy(ticker)
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
	handler.Ex.CancelOrder(int64(orderId), ticker)
	return c.JSON(200, map[string]interface{}{
		"msg": "order cancelled",
	})
}

// this function is called everytime a client connect to the websocket
func (handler WebServiceHandler) WebSocketHandlerCurrentPrice(ws *websocket.Conn) {
	// TODO: let client choose ticker
	ticker := "ETHUSD"
	lastCurrentPrice := 0.0
	currentPrice := lastCurrentPrice

	for {
		// NOTE: reading the last value of the slice orderbook.lastTrades introduce race condition.
		// because i might be reading an   entry that has just been allocated by the other goroutine for orderprocessing and the it has yet to be filled
		// TODO: use channel
		currentPrice = handler.Ex.GetLastPrice(ticker)
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
	}
}

func (handler WebServiceHandler) WebSocketHandlerLastTrade(ws *websocket.Conn) {
	ticker := "ETHUSD"

	for {
		arr := handler.Ex.GetLastTrades(ticker, 15)
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
		arr := handler.Ex.GetBestBuys(ticker, 15)
		responsesArr := make([]LimitResponse, 0)
		for _, limit := range arr {
			response := LimitResponse{
				Price:  limit.GetLimitPrice(),
				Volume: limit.GetTotalVolume(),
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
		arr := handler.Ex.GetBestSells(ticker, 15)
		responsesArr := make([]LimitResponse, 0)
		for _, limit := range arr {
			response := LimitResponse{
				Price:  limit.GetLimitPrice(),
				Volume: limit.GetTotalVolume(),
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
	return c.JSON(200, handler.Ex.GetUsersMap())
}

// TODO: dont return all details about users ?
func (handler WebServiceHandler) HandleGetUser(c echo.Context) error {
	userId := c.Param("userId")
	user, ok := handler.Ex.GetUsersMap()[userId]
	if !ok {
		return c.JSON(404, fmt.Sprintf("UserId %s does not exist", userId))
	}
	return c.JSON(200, user)
}
