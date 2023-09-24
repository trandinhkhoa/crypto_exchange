package domain_test

import (
	"reflect"
	"testing"

	"github.com/trandinhkhoa/crypto-exchange/domain"
)

func assert(t *testing.T, a, b any) {
	if !reflect.DeepEqual(a, b) {
		t.Errorf("%+v != %+v", a, b)
	}
}

func TestLimit(t *testing.T) {
	o := domain.NewOrder("john", "ticker", true, "LIMIT", 1, 1000)
	l := domain.NewLimit(1000)
	assert(t, l.HeadOrder == nil, true)
	assert(t, l.TailOrder == nil, true)
	assert(t, l.TotalVolume, 0.0)

	l.AddOrder(o)
	assert(t, l.HeadOrder.GetUserId(), "john")
	assert(t, l.TailOrder.GetUserId(), "john")
	assert(t, l.TotalVolume, 1.0)

	o = domain.NewOrder("jane", "ticker", true, "LIMIT", 1, 1000)
	l.AddOrder(o)
	assert(t, l.HeadOrder.GetUserId(), "john")
	assert(t, l.TailOrder.GetUserId(), "jane")
	assert(t, l.TotalVolume, 2.0)

	o = domain.NewOrder("jim", "ticker", true, "LIMIT", 1, 1000)
	l.AddOrder(o)
	assert(t, l.HeadOrder.GetUserId(), "john")
	assert(t, l.HeadOrder.NextOrder.GetUserId(), "jane")
	assert(t, l.TailOrder.GetUserId(), "jim")
	assert(t, l.TotalVolume, 3.0)

	// delete middle (cancel)
	l.DeleteOrder(l.HeadOrder.NextOrder.GetId())
	assert(t, l.HeadOrder.GetUserId(), "john")
	assert(t, l.TailOrder.GetUserId(), "jim")
	assert(t, l.TotalVolume, 2.0)

	// delete head (cancel/match)
	l.DeleteOrder(l.HeadOrder.GetId())
	assert(t, l.HeadOrder.GetUserId(), "jim")
	assert(t, l.TailOrder.GetUserId(), "jim")
	assert(t, l.TotalVolume, 1.0)

	// delete the rest
	l.DeleteOrder(l.HeadOrder.GetId())
	assert(t, l.HeadOrder == nil, true)
	assert(t, l.TailOrder == nil, true)
	assert(t, l.TotalVolume, 0.0)
}
