package server

import (
	"database/sql"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/trandinhkhoa/crypto-exchange/entities"
)

type OrdersRepoImpl struct {
	dbHandler *sql.DB
}

func NewOrdersRepoImpl(dbHandler *sql.DB) *OrdersRepoImpl {
	return &OrdersRepoImpl{
		dbHandler: dbHandler,
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

	_, err := ordersRepoImpl.dbHandler.Exec(queryStr)
	if err != nil {
		logrus.Fatal("Unable to create order: ", err)
	}
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

	_, err := ordersRepoImpl.dbHandler.Exec(queryStr)
	if err != nil {
		logrus.Fatal("Unable to update order: ", err)
	}
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

	_, err := ordersRepoImpl.dbHandler.Exec(queryStr)
	if err != nil {
		logrus.Fatal("Unable to delete order: ", err)
	}
}

type UsersRepoImpl struct {
	dbHandler *sql.DB
}

func NewUsersRepoImpl(dbHandler *sql.DB) *UsersRepoImpl {
	return &UsersRepoImpl{
		dbHandler: dbHandler,
	}
}

func (usersRepoImpl UsersRepoImpl) Create(user entities.User) {
	tableName := "users"
	queryStr := fmt.Sprintf("INSERT INTO %s (%s, %s, %s) VALUES ('%s',%f,%f)",
		tableName,
		userid, "ETH", "USD",
		user.GetUserId(), user.Balance["ETH"], user.Balance["USD"])

	_, err := usersRepoImpl.dbHandler.Exec(queryStr)
	if err != nil {
		logrus.Fatal("Unable to create user: ", err)
	}
}

func (usersRepoImpl UsersRepoImpl) Update(user entities.User) {
	tableName := "users"
	queryStr := fmt.Sprintf("UPDATE %s SET %s = %f, %s = %f WHERE %s = '%s'",
		tableName,
		"ETH", user.Balance["ETH"],
		"USD", user.Balance["USD"],
		userid, user.GetUserId())

	_, err := usersRepoImpl.dbHandler.Exec(queryStr)
	if err != nil {
		logrus.Fatal("Unable to update user: ", err, queryStr)
	}
}
