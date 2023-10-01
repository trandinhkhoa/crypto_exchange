package entities_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/trandinhkhoa/crypto-exchange/entities"
)

func TestLimit(t *testing.T) {
	o := entities.NewOrder("john", "ticker", true, "LIMIT", 1, 1000)
	l := entities.NewLimit(1000)
	assert.Equal(t, len(l.GetAllOrders()), 0)
	assert.Equal(t, l.GetTotalVolume(), 0.0)

	l.AddOrder(o)
	assert.Equal(t, len(l.GetAllOrders()), 1)
	assert.Equal(t, l.GetAllOrders()[0].GetUserId(), "john")
	assert.Equal(t, l.GetTotalVolume(), 1.0)

	o = entities.NewOrder("jane", "ticker", true, "LIMIT", 1, 1000)
	l.AddOrder(o)
	assert.Equal(t, len(l.GetAllOrders()), 2)
	assert.Equal(t, l.GetAllOrders()[0].GetUserId(), "john")
	assert.Equal(t, l.GetAllOrders()[1].GetUserId(), "jane")
	assert.Equal(t, l.GetTotalVolume(), 2.0)

	o = entities.NewOrder("jim", "ticker", true, "LIMIT", 1, 1000)
	l.AddOrder(o)
	assert.Equal(t, len(l.GetAllOrders()), 3)
	assert.Equal(t, l.GetAllOrders()[0].GetUserId(), "john")
	assert.Equal(t, l.GetAllOrders()[1].GetUserId(), "jane")
	assert.Equal(t, l.GetAllOrders()[2].GetUserId(), "jim")
	assert.Equal(t, l.GetTotalVolume(), 3.0)

	// delete middle (cancel)
	l.DeleteOrderById(l.GetAllOrders()[1].GetId())
	assert.Equal(t, len(l.GetAllOrders()), 2)
	assert.Equal(t, l.GetAllOrders()[0].GetUserId(), "john")
	assert.Equal(t, l.GetAllOrders()[1].GetUserId(), "jim")
	assert.Equal(t, l.GetTotalVolume(), 2.0)

	// delete head (cancel/match)
	l.DeleteOrderById(l.GetAllOrders()[0].GetId())
	assert.Equal(t, len(l.GetAllOrders()), 1)
	assert.Equal(t, l.GetAllOrders()[0].GetUserId(), "jim")
	assert.Equal(t, l.GetTotalVolume(), 1.0)

	// delete the rest
	l.DeleteOrderById(l.GetAllOrders()[0].GetId())
	assert.Equal(t, len(l.GetAllOrders()), 0)
	assert.Equal(t, l.GetTotalVolume(), 0.0)
}
