package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/trandinhkhoa/crypto-exchange/domain"
	"github.com/trandinhkhoa/crypto-exchange/server"
	"github.com/trandinhkhoa/crypto-exchange/usecases"
)

const exchangeDomain = "http://localhost:3000"

type PlaceLimitOrderResponseBody struct {
	Msg   string
	Order domain.Order
}
type PlaceMarketOrderResponseBody struct {
	Matches []usecases.Trade
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
	reqPrice, _ := http.NewRequest(http.MethodGet, exchangeDomain+"/book/ETHUSD/currentPrice", nil)
	respPrice, _ := http.DefaultClient.Do(reqPrice)
	decodedRespPrice := &CurrentPriceResponseBody{}
	if err := json.NewDecoder(respPrice.Body).Decode(decodedRespPrice); err != nil {
		return 0, err
	}
	fmt.Println("Current Price ", decodedRespPrice.CurrentPrice)
	return decodedRespPrice.CurrentPrice, nil
}

func GetBestAskPrice() (float64, error) {
	reqPrice, _ := http.NewRequest(http.MethodGet, exchangeDomain+"/book/ETHUSD/bestAsk", nil)
	respPrice, _ := http.DefaultClient.Do(reqPrice)
	decodedRespPrice := &BestAskPriceResponseBody{}
	if err := json.NewDecoder(respPrice.Body).Decode(decodedRespPrice); err != nil {
		return 0, err
	}
	return decodedRespPrice.BestAskPrice, nil
}

func GetBestBidPrice() (float64, error) {
	reqPrice, _ := http.NewRequest(http.MethodGet, exchangeDomain+"/book/ETHUSD/bestBid", nil)
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
			UserId:    "maker123",
			OrderType: "LIMIT",
			IsBid:     false,
			Size:      1,
			Price:     bestAskPrice + 1,
			Ticker:    "ETHUSD",
		}
		bidBody := server.PlaceOrderRequest{
			UserId:    "maker123",
			OrderType: "LIMIT",
			IsBid:     true,
			Size:      1,
			Price:     bestBidPrice - 1,
			Ticker:    "ETHUSD",
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
	if order.OrderType == "LIMIT" {
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
	ticker := time.NewTicker(150 * time.Millisecond)
	for {
		isBid := true
		if int(rand.Intn(9)) < 5 {
			isBid = false
		}
		askBody := server.PlaceOrderRequest{
			UserId:    "traderJoe123",
			OrderType: "MARKET",
			IsBid:     isBid,
			Size:      1,
			Ticker:    "ETHUSD",
		}
		PlaceOrder(askBody)
		<-ticker.C
	}
}
