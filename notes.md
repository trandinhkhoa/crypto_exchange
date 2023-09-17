# Go
```
go mod init github.com/trandinhkhoa/crypto-exchange
```

# Exchange
Ask = Sell
Bid = Buy

Limit = a group of orders at a certain price level
    - add order
    - delete order
    - fill
        - if bid then iterate through ask orders and fill them. vice versa
            - 1 of the order will be completely fill = the smaller (size) one
                - return a Match
                    - bid
                    - ask
                    - size filled
                    - which price
    - total volume

Order = has size, which limit it is set, timestamp, type(bid or ask)

orderbook = list of ask + list of bids + need a way to see which order is at a certain price
    - e.g. (match : 10 btc ask against 4 btc bid ->  rest: need to find match for the remaining 6)
    - (API) place limit order
        - the rest of the order to the books
        - limit.fill
    - (API) bitTotalVolume = sum of all bid
    - (API) askTotalVolume = sum of all ask
    - (API) place market order
        - test place market order
            - test number of matches
            - test number of ask/bid remaining
            - test remaining volumn of ask/bid
            - test if matches are what were expected (ask = sellOrder, bid = buyOrder, sizeFilled = smaller, price)
            - test if ask/bid is filled
            - test multi fill scenario

- limit order
- market order = fill at the best price
    - thin order book = not a lot of size at a certain price level
    - always minimum of 1 match (the best price)
    - assumption: exchange has enough volume (liquidity)
        - exchange pay market maker based on the total volume provided over a period of time
        - has to check if there is enough volume -> bidTotalVolume
            - e.g. order with 20 BTC but there is only 10 BTC

- market maker: provide liquidity at all times to an exchange
match = match ask against a bid, keep track of the size being filled (10 btc ask against 4 btc bid -> need to find match for the remaining 6)
- how to determine price when your exchange boot up the first time ?
    - aggregate the orderbook of other exchanges


- vscode: debug running server:
    - choose debug: "create launch.json" -> "Attach to Process:
    - click on debug -> search for process name (not necessarily the process id) .e.g "exchange"
- Go :
    - FAQ: Should I check go.sum into git?
        - https://twitter.com/FiloSottile/status/1029404663358087173
        - Generally yes. With it, anyone with your sources doesn't have to trust other GitHub repositories and custom import path owners. Something better is coming, but in the meantime it's the same model as hashes in lock files.
    - goroutines
        - channel vs routine
            - channels: for communication between goroutines
            - mutex: for case where you dont care about communication and just want to make sure only one goroutine can access a variable at a time to avoid conflicts
    - in Go, when you create a new instance of a struct using &OrderData{...}, you're allocating memory for that struct on the heap. This memory will remain valid and won't be garbage collected as long as there's a reference to it. In your case, the reference is maintained in the orderBookData.Bids slice.
        - Once you exit the code block, the local variable orderData goes out of scope, but the memory it points to is still valid because the orderBookData.Bids slice holds a reference to it.
    - if inner struct has a pointer to outer struct, be careful passing around the instance of the outer struct. If pass by value , the new arg will contain an instance of the inner class that is pointing to the "original" outer instance -> pass the pointer instead
    - similarities:
        - go get <> == npm install
            - go.mod == package.json
```
	for _, iterator := range ex.orderbooks[marketType].BidLimits {
		for _, order := range iterator.Orders {
			orderData := &OrderData{
				ID:        order.ID,
				IsBid:     order.IsBid,
				Size:      order.Size,
				Price:     order.Limit.Price,
				Timestamp: order.Timestamp,
			}
			orderBookData.Bids = append(orderBookData.Bids, orderData)
		}
	}
	orderBookData.TotalAsksVolume += ex.orderbooks[marketType].GetTotalVolumeAllAsks()
	orderBookData.TotalBidsVolume += ex.orderbooks[marketType].GetTotalVolumeAllBids()

	return c.JSON(200, orderBookData)
```


- makers:
    - "make" liquidity by providing orders for others to trade against.

- misc.:
    - bad code architecture -> I find it hard to write tests
        - I should be able to test Limit without initializing Orderbook, users
    - https://stackoverflow.com/questions/1222392/what-is-domain-driven-design-ddd/1222488#1222488
        - Rob Knight's answer
    - composition over inheritance ???
        - top answer: https://www.reddit.com/r/AskProgramming/comments/lv7m7a/i_still_dont_understand_the_prefer_composition/
    - https://www.reddit.com/r/golang/comments/9ls611/when_to_use_methods/?utm_source=reddit-android
    - https://www.reddit.com/r/golang/comments/104m27v/how_to_avoid_oo_minded/
    - https://www.gobeyond.dev/standard-package-layout/
    - https://blog.cleancoder.com/
    - https://manuel.kiessling.net/2012/09/28/applying-the-clean-architecture-to-go-applications/
- current bugs:
    - fixed: users place sell market order, matched got usd back, but eth not deducted
    - place limit order high (sell), try to clear out buys to push the price, to see if the limit order is eventually filled
        - multiple 10000 price matched -> because for a moement bestAsk is 10000, market making logic then keep hanging around this value
