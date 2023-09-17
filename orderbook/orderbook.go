package orderbook

import (
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/trandinhkhoa/crypto-exchange/users"
)

type Order struct {
	// false = ask
	UserId    string
	ID        int
	IsBid     bool
	OrderType string
	Size      float64
	Price     float64
	limit     *Limit
	Timestamp int64
}

// TODO: make sure ID is unique among orderbooks of all markets
func NewOrder(isBid bool, size float64, userId string, orderType string) *Order {
	return &Order{
		UserId:    userId,
		ID:        int(rand.Int31()),
		IsBid:     isBid,
		OrderType: orderType,
		Size:      size,
		Timestamp: time.Now().UnixNano(),
	}
}

// implement Stringer interface
func (o *Order) String() string {
	return fmt.Sprintf("userId: %s size: %.2f", o.UserId, o.Size)
}

func (o *Order) isFilled() bool {
	return o.Size == float64(0)
}

// for each price level(limit) we need to know total volume and the corresponding orders
type Limit struct {
	isBid        bool
	orderBookPtr *OrderBook
	Price        float64
	TotalVolume  float64
	// uppercase O for Order Stringer interface
	Orders []*Order
}

func NewLimit(price float64, orderBookPtr *OrderBook, isBid bool) *Limit {
	return &Limit{
		isBid:        isBid,
		orderBookPtr: orderBookPtr,
		Price:        price,
		Orders:       []*Order{},
	}
}

func (l *Limit) AddOrder(o *Order) {
	o.limit = l
	o.Price = l.Price
	l.Orders = append(l.Orders, o)
	l.TotalVolume += float64(o.Size)
}

func (l *Limit) DeleteOrder(o *Order) {
	// delete by swapping the deleted with the last entry and re-slicing
	for index, value := range l.Orders {
		if value == o {
			l.Orders[index] = l.Orders[len(l.Orders)-1]
			l.Orders = l.Orders[:len(l.Orders)-1]
		}
	}
	// remove the size of the deleted order from the total volume
	l.TotalVolume -= float64(o.Size)
}

// for now incoming always == MARKET
func (l *Limit) fill(incomingOrder *Order) []Match {
	matchArray := make([]Match, 0)

	for _, existingOrder := range l.Orders {
		if incomingOrder.isFilled() {
			break
		}
		var bid *Order
		var ask *Order

		if incomingOrder.IsBid {
			bid = incomingOrder
			ask = existingOrder
		} else {
			ask = incomingOrder
			bid = existingOrder
		}
		var smallerOrder *Order
		var biggerOrder *Order

		if existingOrder.Size >= incomingOrder.Size {
			biggerOrder = existingOrder
			smallerOrder = incomingOrder
		} else {
			smallerOrder = existingOrder
			biggerOrder = incomingOrder
		}

		match := Match{
			// BidID:      bid.ID,
			BidOrder: bid,
			AskOrder: ask,
			// AskID:      ask.ID,
			Price:      l.Price,
			SizeFilled: smallerOrder.Size,
		}
		// fmt.Println(match)
		matchArray = append(matchArray, match)

		l.TotalVolume = l.TotalVolume - smallerOrder.Size
		biggerOrder.Size = biggerOrder.Size - smallerOrder.Size
		smallerOrder.Size = 0

		if existingOrder.isFilled() {
			l.DeleteOrder(existingOrder)
		}
	}
	return matchArray
}

// ob.askLimits should be sorted according to limit price
type AskLimitsInterface []*Limit

func (ls AskLimitsInterface) Less(a int, b int) bool {
	if ls[a].Price < ls[b].Price {
		return true
	} else {
		return false
	}
}

func (ls AskLimitsInterface) Swap(a int, b int) {
	ls[a], ls[b] = ls[b], ls[a]
}

func (ls AskLimitsInterface) Len() int {
	return len(ls)
}

type BidLimitsInterface []*Limit

func (ls BidLimitsInterface) Less(a int, b int) bool {
	if ls[a].Price > ls[b].Price {
		return true
	} else {
		return false
	}
}

func (ls BidLimitsInterface) Swap(a int, b int) {
	ls[a], ls[b] = ls[b], ls[a]
}

func (ls BidLimitsInterface) Len() int {
	return len(ls)
}

type OrderBook struct {
	Users         users.Users
	AskLimits     AskLimitsInterface
	BidLimits     BidLimitsInterface
	PriceToAskMap map[float64]*Limit
	PriceToBidMap map[float64]*Limit
	IDToOrderMap  map[int]*Order
	CurrentPrice  float64
	mu            sync.Mutex
}

func NewOrderbook(idToUserMap *users.Users) *OrderBook {
	// w/o make :  assignment to entry in nil map
	return &OrderBook{
		Users:         *idToUserMap,
		PriceToAskMap: make(map[float64]*Limit),
		PriceToBidMap: make(map[float64]*Limit),
		IDToOrderMap:  make(map[int]*Order),
	}
}

// fill at `price`
func (ob *OrderBook) PlaceLimitOrder(price float64, o *Order) {
	ob.mu.Lock()
	defer ob.mu.Unlock()
	var limit *Limit

	// find the limit object with the corresponding price
	if o.IsBid {
		limit = ob.PriceToBidMap[price]
	} else {
		limit = ob.PriceToAskMap[price]
	}
	if limit == nil {
		limit = NewLimit(price, ob, o.IsBid)
		if o.IsBid {
			ob.BidLimits = append(ob.BidLimits, limit)
			ob.PriceToBidMap[price] = limit
		} else {
			ob.AskLimits = append(ob.AskLimits, limit)
			ob.PriceToAskMap[price] = limit
		}
	}

	logrus.WithFields(logrus.Fields{
		"isBid":     o.IsBid,
		"size":      o.Size,
		"price":     limit.Price,
		"timestamp": o.Timestamp,
	}).Info("limit order placed")

	ob.IDToOrderMap[o.ID] = o
	limit.AddOrder(o)
}

func (ob *OrderBook) GetBestAsk() Limit {
	// sort.Sort takes in as argument that implement a certain interface.
	sort.Sort(ob.AskLimits)
	if len(ob.BidLimits) == 0 {
		return *NewLimit(0, nil, false)
	}
	return *ob.AskLimits[0]
}

func (ob *OrderBook) GetBestBid() Limit {
	sort.Sort(ob.BidLimits)
	if len(ob.BidLimits) == 0 {
		return *NewLimit(0, nil, true)
	}
	return *ob.BidLimits[0]
}

func (ob *OrderBook) clearLimit(limit *Limit) {
	var limits []*Limit
	if limit.isBid {
		delete(ob.PriceToBidMap, limit.Price)
		limits = ob.BidLimits
	} else {
		delete(ob.PriceToAskMap, limit.Price)
		limits = ob.AskLimits
	}
	for index, toBeDeletedLimit := range limits {
		if toBeDeletedLimit.Price == limit.Price {
			limits[index] = limits[len(limits)-1]
			if limit.isBid {
				ob.BidLimits = ob.BidLimits[:len(ob.BidLimits)-1]
			} else {
				ob.AskLimits = ob.AskLimits[:len(ob.AskLimits)-1]
			}
			// horribly wrong here
			// limits = limits[:len(limits)-1]
			break
		}
	}
}

// fill at best price
func (ob *OrderBook) PlaceMarketOrder(incomingOrder *Order) []Match {
	ob.mu.Lock()
	defer ob.mu.Unlock()
	// check if there is enough liquidity
	if incomingOrder.IsBid && incomingOrder.Size > ob.GetTotalVolumeAllAsks() {
		panic("Not enough ask liquidity")
	} else if !incomingOrder.IsBid && incomingOrder.Size > ob.GetTotalVolumeAllBids() {
		panic("Not enough ask liquidity")
	}

	matchArray := make([]Match, 0)
	var limits []*Limit

	// if bid, search for asks, starting from best price
	// AskLimits/BidLimits should be sorted according to limit price
	if incomingOrder.IsBid {
		limits = ob.AskLimits
		sort.Sort(AskLimitsInterface(limits))
	} else {
		limits = ob.BidLimits
		sort.Sort(BidLimitsInterface(limits))
	}

	for _, limit := range limits {
		// inside a limit, orders should be sorted according to timestamp
		returnedArray := limit.fill(incomingOrder)
		matchArray = append(matchArray, returnedArray...)
		// clear limit if there is no orders left inside
		if len(limit.Orders) == 0 {
			ob.clearLimit(limit)
		}

		ob.CurrentPrice = limit.Price
		logrus.WithFields(logrus.Fields{
			"currentPrice": ob.CurrentPrice,
		}).Info("------")

		if incomingOrder.isFilled() {
			break
		}
	}
	execute(matchArray, ob)

	// logrus.Info("market order filled")
	return matchArray
}

func (ob *OrderBook) GetTotalVolumeAllBids() float64 {
	total := float64(0)
	for _, limit := range ob.BidLimits {
		total += limit.TotalVolume
	}
	return total
}

func (ob *OrderBook) GetTotalVolumeAllAsks() float64 {
	total := float64(0)
	for _, limit := range ob.AskLimits {
		total += limit.TotalVolume
	}
	return total
}

func (ob *OrderBook) CancelOrder(o *Order) {
	o.limit.DeleteOrder(o)
	o.limit = nil
}

type Match struct {
	AskOrder   *Order
	BidOrder   *Order
	SizeFilled float64
	Price      float64
}
