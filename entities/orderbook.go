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
	LastTradedPrice float64
}

// TODO: hide all the pointers, make sure if &Orderbook{} is used it would be useless
func NewOrderbook() *Orderbook {
	return &Orderbook{
		idToOrderMap: make(map[int64]*Order),
	}
}

// TODO: dont expose the pointer
func TreeToArray(node *Limit) []*Limit {
	array := make([]*Limit, 0)
	treeToArrayHelper(node, &array)
	return array
}

func travelLimitTreeAndAddOrderToLimit(node *Limit, incomingOrder *Order) *Limit {
	if incomingOrder.GetLimitPrice() == node.HeadOrder.GetLimitPrice() {
		node.AddOrder(incomingOrder)
		return nil
	} else if incomingOrder.IsBetter(node.HeadOrder) {
		if node.LeftChild != nil {
			return travelLimitTreeAndAddOrderToLimit(node.LeftChild, incomingOrder)
		} else {
			newLimit := setupNewLimit(incomingOrder)
			node.LeftChild = newLimit
			newLimit.Parent = node
			return newLimit
		}
	} else {
		if node.RightChild != nil {
			return travelLimitTreeAndAddOrderToLimit(node.RightChild, incomingOrder)
		} else {
			newLimit := setupNewLimit(incomingOrder)
			node.RightChild = newLimit
			newLimit.Parent = node
			return newLimit
		}
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
		existingOrder := bestLimit.HeadOrder
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
			bestLimit.DeleteOrder(existingOrder)
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

		bestLimit.TotalVolume -= sizeFilled

		// if current limit is out of liquidity, remove it and move on to next limit
		// this check for empty limit, YIKES
		// if bestLimit.TotalVolume == 0 {
		if bestLimit.HeadOrder == nil {
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
	ob.LastTradedPrice = tradesArray[len(tradesArray)-1].GetPrice()
	return tradesArray
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

func (ob *Orderbook) CancelOrder(orderId int64) (string, bool, float64, float64) {
	order, ok := ob.idToOrderMap[orderId]
	if !ok {
		return "", false, 0, 0
	}

	limit := order.ParentLimit
	limit.DeleteOrder(order)
	if limit.HeadOrder == nil {
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

func (ob *Orderbook) GetBestLimits(tree *Limit, k int) []*Limit {
	array := make([]*Limit, 0)
	ob.dfTraversal(tree, k, &array)
	return array
}

func (ob *Orderbook) clearLimit(limit *Limit, isBid bool) {

	parent := limit.Parent
	rightChild := limit.RightChild
	leftMostOfRightSide := findLeftMost(rightChild)

	// replace current node with leftMostOfRightSide
	if leftMostOfRightSide != nil {
		// detach leftMostOfRightSide from its parent
		// leftMostOfRightSide.Parent == nil ?? -> leftMostOfRightSide == root
		// ; but leftMostleftMostOfRightSide == current.Child -> impossible
		if leftMostOfRightSide == leftMostOfRightSide.Parent.LeftChild {
			leftMostOfRightSide.Parent.LeftChild = nil
		}
		if leftMostOfRightSide == leftMostOfRightSide.Parent.RightChild {
			leftMostOfRightSide.Parent.RightChild = nil
		}
		leftMostOfRightSide.Parent = nil

		// connect current parent to leftMostOfRightSide
		if parent != nil && parent.LeftChild == limit {
			parent.LeftChild = leftMostOfRightSide
		} else if parent != nil && parent.RightChild == limit {
			parent.RightChild = leftMostOfRightSide
		}
		leftMostOfRightSide.Parent = parent

		// connect leftMostOfRightSide to current node's childs
		leftMostOfRightSide.LeftChild = limit.LeftChild
		if leftMostOfRightSide != limit.RightChild {
			leftMostOfRightSide.RightChild = limit.RightChild
		}
		if limit.LeftChild != nil {
			limit.LeftChild.Parent = leftMostOfRightSide
		}
		if limit.RightChild != nil {
			limit.RightChild.Parent = leftMostOfRightSide
		}

		// detach current node
		limit.Parent = nil
		limit.LeftChild = nil
		limit.RightChild = nil
	} else {
		// if right side empty, just move up left side
		if parent != nil {
			// if not root
			if parent.LeftChild == limit {
				parent.LeftChild = limit.LeftChild
			} else {
				parent.RightChild = limit.LeftChild
			}
		} else {
			if isBid {
				ob.BuyTree = limit.LeftChild
			} else {
				ob.SellTree = limit.LeftChild
			}
		}
	}
}
func sumTree(node *Limit) float64 {
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

func (ob *Orderbook) dfTraversal(node *Limit, k int, arr *[]*Limit) {
	if node == nil {
		return
	}
	ob.dfTraversal(node.LeftChild, k, arr)
	if len(*arr) < k {
		*arr = append(*arr, node)
	}
	if len(*arr) >= k {
		return
	}
	ob.dfTraversal(node.RightChild, k, arr)
}
func findLeftMost(node *Limit) *Limit {
	if node != nil && node.LeftChild != nil {
		return findLeftMost(node.LeftChild)
	} else {
		return node
	}
}
func setupNewLimit(incomingOrder *Order) *Limit {
	newLimit := NewLimit(incomingOrder.GetLimitPrice())
	newLimit.Parent = nil
	newLimit.LeftChild = nil
	newLimit.RightChild = nil
	newLimit.AddOrder(incomingOrder)
	return newLimit
}

func treeToArrayHelper(node *Limit, array *[]*Limit) {
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
