package main

import (
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
	go client.MakeMarket()
	go client.PlaceMarketRepeat()
	select {}
}
