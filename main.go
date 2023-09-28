package main

import (
	"time"

	"github.com/sirupsen/logrus"
	"github.com/trandinhkhoa/crypto-exchange/client"
	"github.com/trandinhkhoa/crypto-exchange/server"
)

func init() {
	// Customize the default logger
	logrus.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05.000",
		FullTimestamp:   true,
	})
}

func main() {
	go server.StartServer()
	time.Sleep(1 * time.Second)
	go client.MakeMarket()
	// go client.PlaceLimitFromFile()
	go client.PlaceMarketRepeat()
	select {}
}
