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