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
	o := domain.NewOrder("john", true, "LIMIT", 1, 1000)
	l := domain.NewLimit(1000)
	assert(t, l.HeadOrder == nil, true)
	assert(t, l.TailOrder == nil, true)
	assert(t, l.TotalVolume == 0, true)

	l.AddOrder(o)
	assert(t, l.HeadOrder.GetUserId(), "john")
	assert(t, l.TailOrder.GetUserId(), "john")

	o = domain.NewOrder("jane", true, "LIMIT", 1, 1000)
	l.AddOrder(o)
	assert(t, l.HeadOrder.GetUserId(), "john")
	assert(t, l.TailOrder.GetUserId(), "jane")

	o = domain.NewOrder("jim", true, "LIMIT", 1, 1000)
	l.AddOrder(o)
	assert(t, l.HeadOrder.GetUserId(), "john")
	assert(t, l.HeadOrder.NextOrder.GetUserId(), "jane")
	assert(t, l.TailOrder.GetUserId(), "jim")

	// delete middle (cancel)
	l.DeleteOrder(l.HeadOrder.NextOrder.GetId())
	assert(t, l.HeadOrder.GetUserId(), "john")
	assert(t, l.TailOrder.GetUserId(), "jim")

	// delete head (cancel/match)
	l.DeleteOrder(l.HeadOrder.GetId())
	assert(t, l.HeadOrder.GetUserId(), "jim")
	assert(t, l.TailOrder.GetUserId(), "jim")

	// delete the rest
	l.DeleteOrder(l.HeadOrder.GetId())
	assert(t, l.HeadOrder == nil, true)
	assert(t, l.TailOrder == nil, true)
}
