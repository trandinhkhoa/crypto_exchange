package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/trandinhkhoa/crypto-exchange/orderbook"
)

const exchangeDomain = "http://localhost:3000"

type PlaceOrderResponseBody struct {
	Msg   string
	Order orderbook.Order
}

func PlaceLimitOrder() error {
	url := exchangeDomain + "/order"
	orderBody := `{
		"Type": "LIMIT",
		"IsBid": true,
		"Size": 1,
		"Price": 10000,
		"Market": "ETH"
	}`
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader([]byte(orderBody)))
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("somethings is wrong")
		return err
	}

	decodedResp := &PlaceOrderResponseBody{}
	if err := json.NewDecoder(resp.Body).Decode(decodedResp); err != nil {
		return err
	}

	// logrus.WithFields(logrus.Fields{
	// 	"msg":       decodedResp.Msg,
	// 	"id":        decodedResp.Order.ID,
	// 	"isBid":     decodedResp.Order.IsBid,
	// 	"size":      decodedResp.Order.Size,
	// 	"timestamp": decodedResp.Order.Timestamp,
	// }).Info("Respone Received")

	return nil
}

func PlaceMarketOrder() error {
	url := exchangeDomain + "/order"
	orderBody := `{
		"Type": "MARKET",
		"IsBid": false,
		"Size": 1,
		"Market": "ETH"
	}`
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader([]byte(orderBody)))
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("somethings is wrong")
		return err
	}

	decodedResp := &PlaceOrderResponseBody{}
	if err := json.NewDecoder(resp.Body).Decode(decodedResp); err != nil {
		return err
	}

	return nil
}

func PlaceLimitRepeat() {
	ticker := time.NewTicker(1000 * time.Millisecond)
	for {
		PlaceLimitOrder()
		<-ticker.C
	}
}

func PlaceMarketRepeat() {
	ticker := time.NewTicker(1500 * time.Millisecond)
	for {
		PlaceMarketOrder()
		<-ticker.C
	}
}
