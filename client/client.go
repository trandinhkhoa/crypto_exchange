package client

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/trandinhkhoa/crypto-exchange/controllers"
)

type Client struct {
	ExchangeServer string
}

type PlaceLimitOrderResponseBody struct {
	Msg   string
	Order controllers.OrderResponse
}
type PlaceMarketOrderResponseBody struct {
	Matches []controllers.TradeResponse
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

func (client Client) GetCurrentPrice() (float64, error) {
	reqPrice, err := http.NewRequest(http.MethodGet, client.ExchangeServer+"/book/ETHUSD/currentPrice", nil)
	if err != nil {
		logrus.Error(err)
	}
	respPrice, err := http.DefaultClient.Do(reqPrice)
	if err != nil {
		logrus.Error(err)
	}
	decodedRespPrice := &CurrentPriceResponseBody{}
	if err := json.NewDecoder(respPrice.Body).Decode(decodedRespPrice); err != nil {
		return 0, err
	}
	fmt.Println("Current Price ", decodedRespPrice.CurrentPrice)
	return decodedRespPrice.CurrentPrice, nil
}

func (client Client) GetBestAskPrice() (float64, error) {
	reqPrice, err := http.NewRequest(http.MethodGet, client.ExchangeServer+"/book/ETHUSD/bestAsk", nil)
	if err != nil {
		logrus.Error(err)
	}
	respPrice, err := http.DefaultClient.Do(reqPrice)
	if err != nil {
		logrus.Error(err)
	}
	decodedRespPrice := &BestAskPriceResponseBody{}
	if err := json.NewDecoder(respPrice.Body).Decode(decodedRespPrice); err != nil {
		return 0, err
	}
	return decodedRespPrice.BestAskPrice, nil
}

func (client Client) GetBestBidPrice() (float64, error) {
	reqPrice, err := http.NewRequest(http.MethodGet, client.ExchangeServer+"/book/ETHUSD/bestBid", nil)
	if err != nil {
		logrus.Error(err)
	}
	respPrice, err := http.DefaultClient.Do(reqPrice)
	if err != nil {
		logrus.Error(err)
	}
	decodedRespPrice := &BestBidPriceResponseBody{}
	if err := json.NewDecoder(respPrice.Body).Decode(decodedRespPrice); err != nil {
		return 0, err
	}
	return decodedRespPrice.BestBidPrice, nil
}

func simulateFetchPriceFromOtherExchange() float64 {
	return 1000
}

type PlaceOrderRequest struct {
	UserId    string  `json:"UserId"`
	OrderType string  `json:"OrderType"`
	IsBid     bool    `json:"IsBid"`
	Size      float64 `json:"Size"`
	Price     float64 `json:"Price"`
	Ticker    string  `json:"Ticker"`
}

func (client Client) PlaceLimitFromFile() {
	file, err := os.Open("Coinbase_BTCUSD_ob_10_2017_09_05.csv")
	if err != nil {
		panic(fmt.Sprintf("Could not open the csv file: %s", err))
	}

	r := csv.NewReader(file)

	// Read each record from csv
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(fmt.Sprintf("Could not read the csv file: %s", err))
		}

		price, _ := strconv.ParseFloat(record[2], 64)
		size, _ := strconv.ParseFloat(record[3], 64)
		isBid := record[1] == "a"

		order := controllers.PlaceOrderRequest{
			UserId:    "maker123",
			OrderType: "LIMIT",
			IsBid:     isBid,
			Size:      size,
			Price:     price,
			Ticker:    "ETHUSD",
		}

		client.PlaceOrder(order)
		time.Sleep(1 * time.Millisecond)
	}
}

func (client Client) MakeMarket() {
	ticker := time.NewTicker(75 * time.Millisecond)
	const spread = 0.2
	const halfSpread = spread / 2.0

	for {
		<-ticker.C

		fmt.Println("MAKEING")
		lastTradedPrice, err := client.GetCurrentPrice()
		if err != nil || lastTradedPrice == 0 {
			lastTradedPrice = simulateFetchPriceFromOtherExchange()
		}

		// Calculate bid and ask prices centered around last traded price
		bidPrice := lastTradedPrice - halfSpread
		askPrice := lastTradedPrice + halfSpread

		// Place bid order
		bidBody := controllers.PlaceOrderRequest{
			UserId:    "maker123",
			OrderType: "LIMIT",
			IsBid:     true,
			Size:      1,
			Price:     bidPrice,
			Ticker:    "ETHUSD",
		}
		client.PlaceOrder(bidBody)

		// Place ask order
		askBody := controllers.PlaceOrderRequest{
			UserId:    "maker123",
			OrderType: "LIMIT",
			IsBid:     false,
			Size:      1,
			Price:     askPrice,
			Ticker:    "ETHUSD",
		}
		client.PlaceOrder(askBody)
	}
}

func (client Client) PlaceOrder(order controllers.PlaceOrderRequest) error {
	url := client.ExchangeServer + "/order"
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
	// fmt.Println(order.OrderType, decodedResp)

	return nil
}

func (client Client) PlaceMarketRepeat() {
	timer := time.NewTimer(1500 * time.Millisecond)
	<-timer.C
	ticker := time.NewTicker(100 * time.Millisecond)
	trendTicker := time.NewTicker(10 * time.Second) // To switch trend every X seconds

	isUpwardTrend := true // Initialize as upward trend

	go func() { // Goroutine to switch trend direction
		for {
			if rand.Intn(9) < 5 {
				isUpwardTrend = !isUpwardTrend // Flip the trend direction
			}
			<-trendTicker.C
		}
	}()

	for {
		isBid := isUpwardTrend // Use the current trend direction
		if isUpwardTrend {
			if rand.Intn(9) < 3 {
				isBid = false
			}
		} else {
			if rand.Intn(9) >= 3 {
				isBid = false
			}
		}

		orderBody := controllers.PlaceOrderRequest{
			UserId:    "traderJoe123",
			OrderType: "MARKET",
			IsBid:     isBid,
			Size:      1,
			Ticker:    "ETHUSD",
		}
		client.PlaceOrder(orderBody)
		<-ticker.C
	}
}
