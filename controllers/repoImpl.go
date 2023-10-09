package controllers

import (
	"fmt"

	"github.com/trandinhkhoa/crypto-exchange/entities"
)

type SqlDbHandler interface {
	Exec(string) error
	Close()
	Query(string) Row
}
type Row interface {
	Scan(dest ...interface{})
	Next() bool
}

type OrdersRepoImpl struct {
	sqlDbHandler SqlDbHandler
}

func NewOrdersRepoImpl(sqlDbHandler SqlDbHandler) *OrdersRepoImpl {
	return &OrdersRepoImpl{
		sqlDbHandler: sqlDbHandler,
	}
}

const (
	id        = "id"
	userid    = "userid"
	size      = "size"
	price     = "price"
	timestamp = "timestamp"
)

// TODO: dont use Fatal

func (ordersRepoImpl OrdersRepoImpl) Create(order entities.Order) {
	var tableName string
	if order.GetIsBid() {
		tableName = "buyOrders"
	} else {
		tableName = "sellOrders"
	}
	queryStr := fmt.Sprintf("INSERT INTO %s (%s, %s, %s, %s, %s) VALUES (%d,'%s',%f,%f,%d)",
		tableName,
		id, userid, size, price, timestamp,
		order.GetId(), order.GetUserId(), order.GetSize(), order.GetLimitPrice(), order.GetTimeStamp(),
	)

	ordersRepoImpl.sqlDbHandler.Exec(queryStr)
}

func (ordersRepoImpl OrdersRepoImpl) Update(order entities.Order) {
	var tableName string
	if order.GetIsBid() {
		tableName = "buyOrders"
	} else {
		tableName = "sellOrders"
	}
	queryStr := fmt.Sprintf("UPDATE %s SET %s = %f WHERE %s = %d",
		tableName,
		size, order.GetSize(),
		id, order.GetId())

	ordersRepoImpl.sqlDbHandler.Exec(queryStr)
}

func (ordersRepoImpl OrdersRepoImpl) Delete(order entities.Order) {
	var tableName string
	if order.GetIsBid() {
		tableName = "buyOrders"
	} else {
		tableName = "sellOrders"
	}
	queryStr := fmt.Sprintf("DELETE FROM %s WHERE %s = %d",
		tableName,
		id, order.GetId())

	ordersRepoImpl.sqlDbHandler.Exec(queryStr)
}

func (ordersRepoImpl OrdersRepoImpl) ReadAll(side string) []entities.Order {
	var tableName string
	var isBid bool
	if side == "buy" {
		tableName = "buyOrders"
		isBid = true
	} else {
		tableName = "sellOrders"
		isBid = false
	}

	queryStr := fmt.Sprintf("SELECT * FROM %s", tableName)

	ordersRepoImpl.sqlDbHandler.Exec(queryStr)
	rows := ordersRepoImpl.sqlDbHandler.Query(queryStr)

	buyOrders := make([]entities.Order, 0)
	for rows.Next() {
		// TODO: put this somewhere else ??
		var id int64
		var userId string
		var size float64
		var price float64
		var timestamp int64
		rows.Scan(&id, &userId, &size, &price, &timestamp)
		order := entities.NewOrder(userId, "ETHUSD", isBid, entities.LimitOrderType, size, price)
		buyOrders = append(buyOrders, *order)
	}

	return buyOrders
}

type UsersRepoImpl struct {
	sqlDbHandler SqlDbHandler
}

func NewUsersRepoImpl(sqlDbHandler SqlDbHandler) *UsersRepoImpl {
	return &UsersRepoImpl{
		sqlDbHandler: sqlDbHandler,
	}
}

func (usersRepoImpl UsersRepoImpl) Create(user entities.User) {
	tableName := "users"
	queryStr := fmt.Sprintf("INSERT INTO %s (%s, %s, %s) VALUES ('%s',%f,%f)",
		tableName,
		userid, "ETH", "USD",
		user.GetUserId(), user.Balance["ETH"], user.Balance["USD"])

	usersRepoImpl.sqlDbHandler.Exec(queryStr)
}

func (usersRepoImpl UsersRepoImpl) Update(user entities.User) {
	tableName := "users"
	queryStr := fmt.Sprintf("UPDATE %s SET %s = %f, %s = %f WHERE %s = '%s'",
		tableName,
		"ETH", user.Balance["ETH"],
		"USD", user.Balance["USD"],
		userid, user.GetUserId())

	usersRepoImpl.sqlDbHandler.Exec(queryStr)
}

func (userRepoImpl UsersRepoImpl) ReadAll() []entities.User {
	tableName := "users"

	queryStr := fmt.Sprintf("SELECT * FROM %s", tableName)

	userRepoImpl.sqlDbHandler.Exec(queryStr)
	rows := userRepoImpl.sqlDbHandler.Query(queryStr)

	usersList := make([]entities.User, 0)
	for rows.Next() {
		// TODO: put this somewhere else ??
		var userId string
		var ethBalance float64
		var usdBalance float64
		rows.Scan(&userId, &ethBalance, &usdBalance)
		user := entities.NewUser(userId, map[string]float64{
			"ETH": ethBalance,
			"USD": usdBalance,
		})
		usersList = append(usersList, *user)
	}

	return usersList
}

type LastTradesRepoImpl struct {
	sqlDbHandler SqlDbHandler
}

func NewLastTradesRepoImpl(sqlDbHandler SqlDbHandler) *LastTradesRepoImpl {
	return &LastTradesRepoImpl{
		sqlDbHandler: sqlDbHandler,
	}
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func (tradeRepoImpl LastTradesRepoImpl) Create(trade entities.Trade) {
	tableName := "lastTrades"
	queryStr := fmt.Sprintf("INSERT INTO %s (price, size, isBuyerMaker, timestamp) VALUES ('%f',%f,%d,%d)",
		tableName,
		trade.GetPrice(), trade.GetSize(), boolToInt(trade.GetIsBuyerMaker()), trade.GetTimeStamp())

	tradeRepoImpl.sqlDbHandler.Exec(queryStr)
}

func (tradeRepoImpl LastTradesRepoImpl) ReadAll() []entities.Trade {
	tableName := "lastTrades"

	queryStr := fmt.Sprintf("SELECT * FROM %s", tableName)

	tradeRepoImpl.sqlDbHandler.Exec(queryStr)
	rows := tradeRepoImpl.sqlDbHandler.Query(queryStr)

	tradesList := make([]entities.Trade, 0)
	for rows.Next() {
		// TODO: put this somewhere else ??
		var id int64
		var price float64
		var size float64
		var isBuyerMaker bool
		var timestamp int64
		rows.Scan(&id, &price, &size, &isBuyerMaker, &timestamp)
		user := entities.NewTradeWithTimeStamp(nil, nil, price, size, isBuyerMaker, timestamp)
		tradesList = append(tradesList, *user)
	}

	return tradesList
}
