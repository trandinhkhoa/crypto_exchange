package entities_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/trandinhkhoa/crypto-exchange/entities"
)

func TestLimit(t *testing.T) {
	o := entities.NewOrder("john", "ticker", true, "LIMIT", 1, 1000)
	l := entities.NewLimit(1000)
	assert.Equal(t, l.HeadOrder == nil, true)
	assert.Equal(t, l.TailOrder == nil, true)
	assert.Equal(t, l.TotalVolume, 0.0)

	l.AddOrder(o)
	assert.Equal(t, l.HeadOrder.GetUserId(), "john")
	assert.Equal(t, l.TailOrder.GetUserId(), "john")
	assert.Equal(t, l.TotalVolume, 1.0)

	o = entities.NewOrder("jane", "ticker", true, "LIMIT", 1, 1000)
	l.AddOrder(o)
	assert.Equal(t, l.HeadOrder.GetUserId(), "john")
	assert.Equal(t, l.TailOrder.GetUserId(), "jane")
	assert.Equal(t, l.TotalVolume, 2.0)

	o = entities.NewOrder("jim", "ticker", true, "LIMIT", 1, 1000)
	l.AddOrder(o)
	assert.Equal(t, l.HeadOrder.GetUserId(), "john")
	assert.Equal(t, l.HeadOrder.NextOrder.GetUserId(), "jane")
	assert.Equal(t, l.TailOrder.GetUserId(), "jim")
	assert.Equal(t, l.TotalVolume, 3.0)

	// delete middle (cancel)
	l.DeleteOrder(l.HeadOrder.NextOrder)
	assert.Equal(t, l.HeadOrder.GetUserId(), "john")
	assert.Equal(t, l.TailOrder.GetUserId(), "jim")
	assert.Equal(t, l.TotalVolume, 2.0)

	// delete head (cancel/match)
	l.DeleteOrder(l.HeadOrder)
	assert.Equal(t, l.HeadOrder.GetUserId(), "jim")
	assert.Equal(t, l.TailOrder.GetUserId(), "jim")
	assert.Equal(t, l.TotalVolume, 1.0)

	// delete the rest
	l.DeleteOrder(l.HeadOrder)
	assert.Equal(t, l.HeadOrder == nil, true)
	assert.Equal(t, l.TailOrder == nil, true)
	assert.Equal(t, l.TotalVolume, 0.0)
}
