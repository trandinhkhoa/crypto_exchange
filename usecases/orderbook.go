package usecases

import (
	"time"

	"github.com/trandinhkhoa/crypto-exchange/domain"
)

type Trade struct {
	buyer     *domain.Order
	seller    *domain.Order
	price     float64
	size      float64
	timestamp int64
}

// TODO: dont return order as it contains pointer,
// as w/ the pointer, the content of h field can be modified
// defeat the point of encapsulation
func (t Trade) GetBuyer() domain.Order {
	return *t.buyer
}

func (t Trade) GetSeller() domain.Order {
	return *t.seller
}

func (t Trade) GetPrice() float64 {
	return t.price
}

func (t Trade) GetSize() float64 {
	return t.size
}

func (t Trade) GetTimeStamp() int64 {
	return t.timestamp
}

func NewTrade(
	buyer *domain.Order,
	seller *domain.Order,
	price float64,
	size float64,
) *Trade {
	return &Trade{
		buyer:     buyer,
		seller:    seller,
		price:     price,
		size:      size,
		timestamp: time.Now().UnixNano(),
	}
}

// Book
//
//	Limit *buyTree;
//	Limit *sellTree;
//	Limit *lowestSell;
//	Limit *highestBuy;
type Orderbook struct {
	// TODO: limitation: this way the "interface" of Orderbook is tied to its implementation
	// e.g. switch from BST to heap will be costly
	// domain.Limit has the same issue
	BuyTree    *domain.Limit
	SellTree   *domain.Limit
	LowestSell *domain.Limit
	HighestBuy *domain.Limit
	lastTrades []Trade
}

func setupNewLimit(incomingOrder *domain.Order) *domain.Limit {
	newLimit := domain.NewLimit(incomingOrder.GetLimitPrice())
	newLimit.Parent = nil
	newLimit.LeftChild = nil
	newLimit.RightChild = nil
	newLimit.AddOrder(incomingOrder)
	return newLimit
}

func treeToArrayHelper(node *domain.Limit, array *[]*domain.Limit) {
	if node == nil {
		return
	}
	if node.LeftChild == nil && node.RightChild == nil {
		*array = append(*array, node)
		return
	}
	if node.LeftChild != nil {
		treeToArrayHelper(node.LeftChild, array)
	}
	*array = append(*array, node)
	if node.RightChild != nil {
		treeToArrayHelper(node.RightChild, array)
	}
}

// TODO: dont expose the pointer
func TreeToArray(node *domain.Limit) []*domain.Limit {
	array := make([]*domain.Limit, 0)
	treeToArrayHelper(node, &array)
	return array
}

func travelLimitTreeAndAddNode(node *domain.Limit, incomingOrder domain.Order) *domain.Limit {
	if incomingOrder.GetLimitPrice() == node.HeadOrder.GetLimitPrice() {
		node.AddOrder(&incomingOrder)
		return nil
	} else if incomingOrder.IsBetter(node.HeadOrder) {
		if node.LeftChild != nil {
			return travelLimitTreeAndAddNode(node.LeftChild, incomingOrder)
		} else {
			newLimit := setupNewLimit(&incomingOrder)
			node.LeftChild = newLimit
			newLimit.Parent = node
			return newLimit
		}
	} else {
		if node.RightChild != nil {
			return travelLimitTreeAndAddNode(node.RightChild, incomingOrder)
		} else {
			newLimit := setupNewLimit(&incomingOrder)
			node.RightChild = newLimit
			newLimit.Parent = node
			return newLimit
		}
	}
}

func (ob *Orderbook) PlaceLimitOrder(incomingOrder domain.Order) {
	// check if price level is in buyTree/sellTree
	if incomingOrder.GetIsBid() {
		if ob.BuyTree == nil {
			// if empty tree
			newLimit := setupNewLimit(&incomingOrder)
			ob.BuyTree = newLimit
			ob.HighestBuy = newLimit
		} else {
			limit := travelLimitTreeAndAddNode(ob.BuyTree, incomingOrder)
			if limit != nil && limit.GetLimitPrice() > ob.HighestBuy.GetLimitPrice() {
				ob.HighestBuy = limit
			}
		}
	} else {
		if ob.SellTree == nil {
			// if empty tree
			newLimit := setupNewLimit(&incomingOrder)
			ob.SellTree = newLimit
			ob.LowestSell = newLimit
		} else {
			limit := travelLimitTreeAndAddNode(ob.SellTree, incomingOrder)
			if limit != nil && limit.GetLimitPrice() < ob.LowestSell.GetLimitPrice() {
				ob.LowestSell = limit
			}
		}
	}
}

func findLeftMost(node *domain.Limit) *domain.Limit {
	if node.LeftChild != nil {
		return findLeftMost(node.LeftChild)
	} else {
		return node
	}
}

func (ob *Orderbook) PlaceMarketOrder(incomingOrder domain.Order) []Trade {
	// TODO: check volume somewhere else ? dont use panic ?
	if (incomingOrder.GetIsBid() && ob.GetTotalVolumeAllSells() < incomingOrder.Size) || (!incomingOrder.GetIsBid() && ob.GetTotalVolumeAllBuys() < incomingOrder.Size) {
		panic("Not enough volume")
	}
	// check if price level is in buyTree/sellTree
	tradesArray := make([]Trade, 0)

	var smallerOrder *domain.Order
	var biggerOrder *domain.Order

	var makerTree *domain.Limit
	var bestLimit *domain.Limit

	if incomingOrder.GetIsBid() {
		makerTree = ob.SellTree
		bestLimit = ob.LowestSell
	} else {
		makerTree = ob.BuyTree
		bestLimit = ob.HighestBuy
	}

	for incomingOrder.Size > 0 {
		existingOrder := bestLimit.HeadOrder
		if existingOrder.Size < incomingOrder.Size {
			smallerOrder = existingOrder
			biggerOrder = &incomingOrder
		} else {
			smallerOrder = &incomingOrder
			biggerOrder = existingOrder
		}

		sizeFilled := smallerOrder.Size
		biggerOrder.Size = biggerOrder.Size - sizeFilled
		smallerOrder.Size = 0
		if existingOrder.Size == 0 {
			bestLimit.DeleteOrder(existingOrder.GetId())
		}

		var buy *domain.Order
		var sell *domain.Order
		// update trades and ob volume
		if incomingOrder.GetIsBid() {
			buy = &incomingOrder
			sell = existingOrder
		} else {
			buy = existingOrder
			sell = &incomingOrder
		}
		tradesArray = append(tradesArray, *NewTrade(
			buy,
			sell,
			existingOrder.GetLimitPrice(),
			sizeFilled))

		bestLimit.TotalVolume -= sizeFilled

		// if current limit is out of liquidity, remove it and move on to next limit
		if bestLimit.TotalVolume == 0 {
			if bestLimit.Parent == nil && bestLimit.RightChild == nil {
				bestLimit = nil
				makerTree = nil
				break
			}
			if bestLimit.Parent != nil {
				parent := bestLimit.Parent
				newChild := bestLimit.RightChild
				// cut link parent to current lowest sell both ways
				// link parent - child both ways
				// current lowest has to be the left child of its parent
				parent.LeftChild = newChild
				if newChild != nil {
					newChild.Parent = parent
				}
				bestLimit.Parent = nil

				bestLimit = findLeftMost(parent)
			} else {
				// if root
				makerTree = bestLimit.RightChild
				makerTree.Parent = nil
				bestLimit = findLeftMost(makerTree)
			}
		}
	}

	// TODO: how to check all deleted nodes has references to them all removed ?
	//  so that the garbage collector can do its thing ??
	if incomingOrder.GetIsBid() {
		ob.SellTree = makerTree
		ob.LowestSell = bestLimit
	} else {
		ob.BuyTree = makerTree
		ob.HighestBuy = bestLimit
	}
	ob.lastTrades = append(ob.lastTrades, tradesArray...)
	return tradesArray
}

func sumTree(node *domain.Limit) float64 {
	if node == nil {
		return 0
	}
	if node.LeftChild == nil && node.RightChild == nil {
		return node.TotalVolume
	}
	sum := 0.0
	sum += sumTree(node.LeftChild)
	sum += node.TotalVolume
	sum += sumTree(node.RightChild)
	return sum
}

func (ob *Orderbook) GetTotalVolumeAllSells() float64 {
	return sumTree(ob.SellTree)
}

func (ob *Orderbook) GetTotalVolumeAllBuys() float64 {
	return sumTree(ob.BuyTree)
}

func (ob *Orderbook) GetLastTrades() []Trade {
	return ob.lastTrades
}
