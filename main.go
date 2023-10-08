package main

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/trandinhkhoa/crypto-exchange/client"
	"github.com/trandinhkhoa/crypto-exchange/controllers"
	"github.com/trandinhkhoa/crypto-exchange/infrastructure"
	"github.com/trandinhkhoa/crypto-exchange/usecases"
	"golang.org/x/net/websocket"
)

func init() {
	// Customize the default logger
	logrus.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05.000",
		DisableQuote:    true,
		FullTimestamp:   true,
	})
}

func StartServer() {
	e := echo.New()

	// client, err := ethclient.Dial("http://localhost:8545")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	dbHandler := infrastructure.SetupDatabase("./real.db")
	defer dbHandler.Close()

	handler := controllers.WebServiceHandler{}
	ex := usecases.NewExchange()

	// injections of implementations
	ordersRepoImpl := controllers.NewOrdersRepoImpl(dbHandler)
	ex.OrdersRepo = ordersRepoImpl
	usersRepoImpl := controllers.NewUsersRepoImpl(dbHandler)
	ex.UsersRepo = usersRepoImpl

	handler.Ex = ex

	handler.Ex.RegisterUserWithBalance("maker123",
		map[string]float64{
			"ETH": 10000.0,
			"USD": 1000000.0,
		},
	)
	handler.Ex.RegisterUserWithBalance("traderJoe123",
		map[string]float64{
			"ETH": 10.0,
			"USD": 1000.0,
		},
	)
	handler.Ex.RegisterUserWithBalance("me",
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

func main() {
	go StartServer()
	time.Sleep(1 * time.Second)
	go client.MakeMarket()
	// go client.PlaceLimitFromFile()
	go client.PlaceMarketRepeat()
	select {}
}
