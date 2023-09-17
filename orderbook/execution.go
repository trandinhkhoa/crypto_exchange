package orderbook

func execute(matchArray []Match, ob *OrderBook) {
	for _, match := range matchArray {
		// get seller id
		// increase USD balance of seller
		askOrder := match.AskOrder
		// askOrder := ob.IDToOrderMap[as]
		seller := askOrder.UserId
		ob.Users[seller].Balance["USD"] += match.SizeFilled * match.Price
		if askOrder.OrderType == "MARKET" {
			ob.Users[seller].Balance["ETH"] -= match.SizeFilled
		}

		// decrease USD balance of buyer
		// ob.UsersBalances[]
		// bidOrderID := match.BidID
		bidOrder := match.BidOrder
		// !!!! if order  is market it will not be in the map
		// bidOrder := ob.IDToOrderMap[bidOrderID]
		buyer := bidOrder.UserId
		ob.Users[buyer].Balance["ETH"] += match.SizeFilled
		if bidOrder.OrderType == "MARKET" {
			ob.Users[seller].Balance["USD"] -= match.SizeFilled * match.Price
		}

	}

}
