package main

import (
	"flag"
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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
	logrus.SetLevel(logrus.DebugLevel)
}

func initialTablesSetup(db controllers.SqlDbHandler) {
	// TODO: avoid hardcoding all currencies
	createTableSQL := `CREATE TABLE IF NOT EXISTS users (
		"userid" TEXT PRIMARY KEY,
		"ETH" FLOAT,
		"USD" FLOAT
	);`
	if err := db.Exec(createTableSQL); err != nil {
		panic("Unable to create table users")
	}

	createTableSQL = `CREATE TABLE IF NOT EXISTS buyOrders (
		"id" INTEGER PRIMARY KEY,
		"userid" TEXT,
		"size" INTEGER,
		"price" INTEGER,
		"timestamp" INTEGER
	);`

	if err := db.Exec(createTableSQL); err != nil {
		panic("Unable to create table buyOrders")
	}

	createTableSQL = `CREATE TABLE IF NOT EXISTS sellOrders (
		"id" INTEGER PRIMARY KEY,
		"userid" TEXT,
		"size" INTEGER,
		"price" INTEGER,
		"timestamp" INTEGER
	);`

	if err := db.Exec(createTableSQL); err != nil {
		panic("Unable to create table sellOrders")
	}

	createTableSQL = `CREATE TABLE IF NOT EXISTS lastTrades (
		"id" INTEGER PRIMARY KEY AUTOINCREMENT,
		"price" FLOAT,
		"size" FLOAT,
		"isBuyerMaker" BOOLEAN,
		"timestamp" INTEGER
	);`
	if err := db.Exec(createTableSQL); err != nil {
		panic("Unable to create table lastTrades")
	}
}

func createSomeUsers(apiHandler *controllers.WebServiceHandler) {
	apiHandler.Ex.RegisterUserWithBalance("maker123",
		map[string]float64{
			"ETH": 10000.0,
			"USD": 1000000.0,
		},
	)
	apiHandler.Ex.RegisterUserWithBalance("traderJoe123",
		map[string]float64{
			"ETH": 10.0,
			"USD": 1000.0,
		},
	)
	apiHandler.Ex.RegisterUserWithBalance("me",
		map[string]float64{
			"ETH": 0.0,
			"USD": 1000.0,
		},
	)

}

func StartServer(freshstart bool, port int, serverStarted chan bool) {
	e := echo.New()
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.HEAD, echo.PUT, echo.PATCH, echo.POST, echo.DELETE},
	}))

	// client, err := ethclient.Dial("http://localhost:8545")

	// injections of implementations
	ex := usecases.NewExchange()

	dbHandler := infrastructure.NewSqliteDbHandler("./real.db")
	defer dbHandler.Close()

	ordersRepoImpl := controllers.NewOrdersRepoImpl(dbHandler)
	ex.OrdersRepo = ordersRepoImpl
	usersRepoImpl := controllers.NewUsersRepoImpl(dbHandler)
	ex.UsersRepo = usersRepoImpl
	lastTradeRepoImpl := controllers.NewLastTradesRepoImpl(dbHandler)
	ex.LastTradesRepo = lastTradeRepoImpl

	notifyUserImpl := controllers.NewNotifyUserImpl()
	ex.Notifier = notifyUserImpl
	apiHandler := controllers.NewWebServiceHandler(ex, notifyUserImpl)

	// initial database setup
	if freshstart {
		initialTablesSetup(dbHandler)
		createSomeUsers(apiHandler)
	} else {
		ex.Recover()
	}
	close(serverStarted)

	e.POST("/order", apiHandler.HandlePlaceOrder)

	e.GET("/users", apiHandler.HandleGetUsers)
	e.GET("/users/:userId", apiHandler.HandleGetUser)
	e.GET("/book/:ticker", apiHandler.HandleGetBook)
	// TODO: handle error when ticker does not exist
	e.GET("/book/:ticker/currentPrice", apiHandler.HandleGetCurrentPrice)
	// TODO: handle error when this is called while no bid/ask is in the book
	e.GET("/book/:ticker/bestAsk", apiHandler.HandleGetBestAsk)
	e.GET("/book/:ticker/bestBid", apiHandler.HandleGetBestBid)

	e.DELETE("/order/:ticker/:id", apiHandler.HandleCancelOrder)

	// in practice, you would have 1 websocket URL.
	// each usecase below would be represented by  an event specified in the payload
	e.GET("/ws/currentPrice", echo.WrapHandler(websocket.Handler(apiHandler.WebSocketHandlerCurrentPrice)))
	e.GET("/ws/lastTrades", echo.WrapHandler(websocket.Handler(apiHandler.WebSocketHandlerLastTrade)))
	e.GET("/ws/bestSells", echo.WrapHandler(websocket.Handler(apiHandler.WebSocketHandlerBestSells)))
	e.GET("/ws/bestBuys", echo.WrapHandler(websocket.Handler(apiHandler.WebSocketHandlerBestBuys)))
	e.GET("/ws/userInfo", echo.WrapHandler(websocket.Handler(apiHandler.WebSocketHandlerUserInfo)))

	e.Start(fmt.Sprintf(":%d", port))
}

func main() {
	// Define flags
	var freshstart bool
	var port int

	flag.BoolVar(&freshstart, "freshstart", true, "Indicate whether it's a fresh start or not")
	flag.IntVar(&port, "port", 3000, "Port to run the application on")

	// Parse the flags
	flag.Parse()

	serverStarted := make(chan bool)
	go StartServer(freshstart, port, serverStarted)

	<-serverStarted

	marketMaker := &client.Client{
		ExchangeServer: "http://localhost:" + fmt.Sprintf("%d", port),
	}
	go marketMaker.MakeMarket()
	// go client.PlaceLimitFromFile()
	marketParticipant := &client.Client{
		ExchangeServer: "http://localhost:" + fmt.Sprintf("%d", port),
	}
	go marketParticipant.PlaceMarketRepeat()
	select {}
}
