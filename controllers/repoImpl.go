package controllers

import (
	"fmt"

	"github.com/trandinhkhoa/crypto-exchange/entities"
)

type SqlDbHandler interface {
	Exec(string) error
	Close()
	// Query(string)
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
