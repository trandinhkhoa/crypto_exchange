package entities

import (
	"fmt"
)

type Orderbook struct {
	// TODO: limitation: this way the "interface" of Orderbook is tied to its implementation
	// e.g. switch from BST to heap will be costly
	// entities.Limit has the same issue
	BuyTree         *Limit
	SellTree        *Limit
	LowestSell      *Limit
	HighestBuy      *Limit
	lastTrades      []Trade
	idToOrderMap    map[int64]*Order
	lastTradedPrice float64
}

// TODO: hide all the pointers, make sure if &Orderbook{} is used it would be useless
func NewOrderbook() *Orderbook {
	return &Orderbook{
		idToOrderMap: make(map[int64]*Order),
	}
}

func (ob *Orderbook) PlaceLimitOrder(incomingOrder Order) {
	// check if price level is in buyTree/sellTree
	if incomingOrder.GetIsBid() {
		if ob.BuyTree == nil {
			// if empty tree
			newLimit := setupNewLimit(&incomingOrder)
			ob.BuyTree = newLimit
			ob.HighestBuy = newLimit
		} else {
			limit := travelLimitTreeAndAddOrderToLimit(ob.BuyTree, &incomingOrder)
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
			limit := travelLimitTreeAndAddOrderToLimit(ob.SellTree, &incomingOrder)
			if limit != nil && limit.GetLimitPrice() < ob.LowestSell.GetLimitPrice() {
				ob.LowestSell = limit
			}
		}
	}
	ob.idToOrderMap[incomingOrder.GetId()] = &incomingOrder
}

func (ob *Orderbook) PlaceMarketOrder(incomingOrder Order) []Trade {
	// TODO: check volume somewhere else ? dont use panic ?
	if (incomingOrder.GetIsBid() && ob.GetTotalVolumeAllSells() < incomingOrder.Size) || (!incomingOrder.GetIsBid() && ob.GetTotalVolumeAllBuys() < incomingOrder.Size) {
		panic("Not enough volume")
	}
	// check if price level is in buyTree/sellTree
	tradesArray := make([]Trade, 0)

	var smallerOrder *Order
	var biggerOrder *Order

	var makerTree *Limit
	var bestLimit *Limit

	if incomingOrder.GetIsBid() {
		makerTree = ob.SellTree
		bestLimit = ob.LowestSell
	} else {
		makerTree = ob.BuyTree
		bestLimit = ob.HighestBuy
	}

	for incomingOrder.Size > 0 {
		existingOrder := bestLimit.headOrder
		if existingOrder == nil {
			fmt.Println("HELLO  existingOrder == nil")
		}
		if bestLimit == nil {
			fmt.Println("HELLO  bestLimit == nil")
		}
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
			bestLimit.deleteOrder(existingOrder)
		}

		var buy *Order
		var sell *Order
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
			sizeFilled,
			existingOrder.isBid))

		bestLimit.totalVolume -= sizeFilled

		// if current limit is out of liquidity, remove it and move on to next limit
		// this check for empty limit, YIKES
		// if bestLimit.TotalVolume == 0 {
		if bestLimit.headOrder == nil {
			if bestLimit.parent == nil && bestLimit.rightChild == nil {
				bestLimit = nil
				makerTree = nil
				break
			}
			if bestLimit.parent != nil {
				parent := bestLimit.parent
				newChild := bestLimit.rightChild
				// cut link parent to current lowest sell both ways
				// link parent - child both ways
				// current lowest has to be the left child of its parent
				parent.leftChild = newChild
				if newChild != nil {
					newChild.parent = parent
				}
				bestLimit.parent = nil

				bestLimit = findLeftMost(parent)
			} else {
				// if root
				makerTree = bestLimit.rightChild
				makerTree.parent = nil
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
	ob.lastTradedPrice = tradesArray[len(tradesArray)-1].GetPrice()
	return tradesArray
}

func (ob Orderbook) GetTotalVolumeAllSells() float64 {
	return sumTree(ob.SellTree)
}

func (ob Orderbook) GetTotalVolumeAllBuys() float64 {
	return sumTree(ob.BuyTree)
}

func (ob Orderbook) GetLastTrades() []Trade {
	return ob.lastTrades
}

func (ob Orderbook) GetLastTradedPrice() float64 {
	return ob.lastTradedPrice
}

func (ob *Orderbook) CancelOrder(orderId int64) (string, bool, float64, float64) {
	order, ok := ob.idToOrderMap[orderId]
	if !ok {
		return "", false, 0, 0
	}

	limit := order.parentLimit
	limit.deleteOrder(order)
	if limit.headOrder == nil {
		ob.clearLimit(limit, order.GetIsBid())
		if limit == ob.HighestBuy {
			ob.HighestBuy = findLeftMost(ob.BuyTree)
		} else if limit == ob.LowestSell {
			ob.LowestSell = findLeftMost(ob.SellTree)
		}
	}
	// TODO: remove order from map
	delete(ob.idToOrderMap, order.GetId())

	return order.GetUserId(), order.GetIsBid(), order.GetLimitPrice(), order.Size
}

func (ob Orderbook) GetBestLimits(tree *Limit, k int) []*Limit {
	array := make([]*Limit, 0)
	ob.dfTraversal(tree, k, &array)
	return array
}

func (ob *Orderbook) clearLimit(limit *Limit, isBid bool) {

	parent := limit.parent
	rightChild := limit.rightChild
	leftMostOfRightSide := findLeftMost(rightChild)

	// replace current node with leftMostOfRightSide
	if leftMostOfRightSide != nil {
		// detach leftMostOfRightSide from its parent
		// leftMostOfRightSide.Parent == nil ?? -> leftMostOfRightSide == root
		// ; but leftMostleftMostOfRightSide == current.Child -> impossible
		if leftMostOfRightSide == leftMostOfRightSide.parent.leftChild {
			leftMostOfRightSide.parent.leftChild = nil
		}
		if leftMostOfRightSide == leftMostOfRightSide.parent.rightChild {
			leftMostOfRightSide.parent.rightChild = nil
		}
		leftMostOfRightSide.parent = nil

		// connect current parent to leftMostOfRightSide
		if parent != nil && parent.leftChild == limit {
			parent.leftChild = leftMostOfRightSide
		} else if parent != nil && parent.rightChild == limit {
			parent.rightChild = leftMostOfRightSide
		}
		leftMostOfRightSide.parent = parent

		// connect leftMostOfRightSide to current node's childs
		leftMostOfRightSide.leftChild = limit.leftChild
		if leftMostOfRightSide != limit.rightChild {
			leftMostOfRightSide.rightChild = limit.rightChild
		}
		if limit.leftChild != nil {
			limit.leftChild.parent = leftMostOfRightSide
		}
		if limit.rightChild != nil {
			limit.rightChild.parent = leftMostOfRightSide
		}

		// detach current node
		limit.parent = nil
		limit.leftChild = nil
		limit.rightChild = nil
	} else {
		// if right side empty, just move up left side
		if parent != nil {
			// if not root
			if parent.leftChild == limit {
				parent.leftChild = limit.leftChild
			} else {
				parent.rightChild = limit.leftChild
			}
		} else {
			if isBid {
				ob.BuyTree = limit.leftChild
			} else {
				ob.SellTree = limit.leftChild
			}
		}
	}
}
func sumTree(node *Limit) float64 {
	if node == nil {
		return 0
	}
	if node.leftChild == nil && node.rightChild == nil {
		return node.totalVolume
	}
	sum := 0.0
	sum += sumTree(node.leftChild)
	sum += node.totalVolume
	sum += sumTree(node.rightChild)
	return sum
}

func (ob *Orderbook) dfTraversal(node *Limit, k int, arr *[]*Limit) {
	if node == nil {
		return
	}
	ob.dfTraversal(node.leftChild, k, arr)
	if len(*arr) < k {
		*arr = append(*arr, node)
	}
	if len(*arr) >= k {
		return
	}
	ob.dfTraversal(node.rightChild, k, arr)
}
func findLeftMost(node *Limit) *Limit {
	if node != nil && node.leftChild != nil {
		return findLeftMost(node.leftChild)
	} else {
		return node
	}
}
func setupNewLimit(incomingOrder *Order) *Limit {
	newLimit := NewLimit(incomingOrder.GetLimitPrice())
	newLimit.parent = nil
	newLimit.leftChild = nil
	newLimit.rightChild = nil
	newLimit.AddOrder(incomingOrder)
	return newLimit
}

func treeToArrayHelper(node *Limit, array *[]*Limit) {
	if node == nil {
		return
	}
	if node.leftChild == nil && node.rightChild == nil {
		*array = append(*array, node)
		return
	}
	if node.leftChild != nil {
		treeToArrayHelper(node.leftChild, array)
	}
	*array = append(*array, node)
	if node.rightChild != nil {
		treeToArrayHelper(node.rightChild, array)
	}
}

// TODO: dont expose the pointer
func TreeToArray(node *Limit) []*Limit {
	array := make([]*Limit, 0)
	treeToArrayHelper(node, &array)
	return array
}

func travelLimitTreeAndAddOrderToLimit(node *Limit, incomingOrder *Order) *Limit {
	if incomingOrder.GetLimitPrice() == node.headOrder.GetLimitPrice() {
		node.AddOrder(incomingOrder)
		return nil
	} else if incomingOrder.IsBetter(node.headOrder) {
		if node.leftChild != nil {
			return travelLimitTreeAndAddOrderToLimit(node.leftChild, incomingOrder)
		} else {
			newLimit := setupNewLimit(incomingOrder)
			node.leftChild = newLimit
			newLimit.parent = node
			return newLimit
		}
	} else {
		if node.rightChild != nil {
			return travelLimitTreeAndAddOrderToLimit(node.rightChild, incomingOrder)
		} else {
			newLimit := setupNewLimit(incomingOrder)
			node.rightChild = newLimit
			newLimit.parent = node
			return newLimit
		}
	}
}
