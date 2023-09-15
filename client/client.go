package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/trandinhkhoa/crypto-exchange/orderbook"
	"github.com/trandinhkhoa/crypto-exchange/server"
)

const exchangeDomain = "http://localhost:3000"

type PlaceLimitOrderResponseBody struct {
	Msg   string
	Order orderbook.Order
}
type PlaceMarketOrderResponseBody struct {
	Matches []orderbook.Match
}

type CurrentPriceResponseBody struct {
	CurrentPrice float64
}
type BestAskPriceResponseBody struct {
	BestAskPrice float64
}

type BestBidPriceResponseBody struct {
	BestBidPrice float64
}

func GetCurrentPrice() (float64, error) {
	reqPrice, _ := http.NewRequest(http.MethodGet, exchangeDomain+"/book/ETH/currentPrice", nil)
	respPrice, _ := http.DefaultClient.Do(reqPrice)
	decodedRespPrice := &CurrentPriceResponseBody{}
	if err := json.NewDecoder(respPrice.Body).Decode(decodedRespPrice); err != nil {
		return 0, err
	}
	fmt.Println("Current Price ", decodedRespPrice.CurrentPrice)
	return decodedRespPrice.CurrentPrice, nil
}

func GetBestAskPrice() (float64, error) {
	reqPrice, _ := http.NewRequest(http.MethodGet, exchangeDomain+"/book/ETH/bestAsk", nil)
	respPrice, _ := http.DefaultClient.Do(reqPrice)
	decodedRespPrice := &BestAskPriceResponseBody{}
	if err := json.NewDecoder(respPrice.Body).Decode(decodedRespPrice); err != nil {
		return 0, err
	}
	return decodedRespPrice.BestAskPrice, nil
}

func GetBestBidPrice() (float64, error) {
	reqPrice, _ := http.NewRequest(http.MethodGet, exchangeDomain+"/book/ETH/bestBid", nil)
	respPrice, _ := http.DefaultClient.Do(reqPrice)
	decodedRespPrice := &BestBidPriceResponseBody{}
	if err := json.NewDecoder(respPrice.Body).Decode(decodedRespPrice); err != nil {
		return 0, err
	}
	return decodedRespPrice.BestBidPrice, nil
}

func simulateFetchPriceFromOtherExchange() float64 {
	return 1000
}

func MakeMarket() {
	ticker := time.NewTicker(100 * time.Millisecond)
	for {
		bestAskPrice, _ := GetBestAskPrice()
		if bestAskPrice == 0 {
			bestAskPrice = simulateFetchPriceFromOtherExchange()
		}

		bestBidPrice, _ := GetBestBidPrice()
		if bestBidPrice == 0 {
			bestBidPrice = simulateFetchPriceFromOtherExchange()
		}

		askBody := server.PlaceOrderRequest{
			Type:   "LIMIT",
			IsBid:  false,
			Size:   1,
			Price:  bestAskPrice + 1,
			Market: "ETH",
		}
		bidBody := server.PlaceOrderRequest{
			Type:   "LIMIT",
			IsBid:  true,
			Size:   1,
			Price:  bestBidPrice - 1,
			Market: "ETH",
		}
		PlaceOrder(askBody)
		PlaceOrder(bidBody)

		<-ticker.C
	}
}

func PlaceOrder(order server.PlaceOrderRequest) error {
	url := exchangeDomain + "/order"
	orderBody, err := json.Marshal(order)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader([]byte(orderBody)))
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("somethings is wrong")
		return err
	}

	var decodedResp any
	if order.Type == "LIMIT" {
		decodedResp = &PlaceLimitOrderResponseBody{}
	} else {
		decodedResp = &PlaceMarketOrderResponseBody{}
	}
	if err := json.NewDecoder(resp.Body).Decode(decodedResp); err != nil {
		fmt.Println("OH NO")
		return err
	}
	fmt.Println(decodedResp)

	return nil
}

func PlaceMarketRepeat() {
	timer := time.NewTimer(1500 * time.Millisecond)
	<-timer.C
	ticker := time.NewTicker(1500 * time.Millisecond)
	for {
		isBid := true
		if int(rand.Intn(9)) < 3 {
			isBid = false
		}
		askBody := server.PlaceOrderRequest{
			Type:   "MARKET",
			IsBid:  isBid,
			Size:   1,
			Market: "ETH",
		}
		PlaceOrder(askBody)
		<-ticker.C
	}
}
