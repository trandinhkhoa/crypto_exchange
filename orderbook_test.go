package main

import (
	"fmt"
	"reflect"
	"testing"
)

func assert(t *testing.T, a, b any) {
	if !reflect.DeepEqual(a, b) {
		t.Errorf("%+v != %+v", a, b)
	}
}

func TestLimit(t *testing.T) {
	l := NewLimit(10_000)
	buyOrderA := NewOrder(true, 5)
	buyOrderB := NewOrder(true, 8)
	buyOrderC := NewOrder(true, 10)

	l.AddOrder(buyOrderA)
	l.AddOrder(buyOrderB)
	l.AddOrder(buyOrderC)

	l.DeleteOrder(buyOrderB)

	a := 2
	b := 2
	assert(t, a, b)
	assert(t, l.price, float64(10_000))
	assert(t, l.totalVolume, float64(15))
	assert(t, len(l.Orders), 2)
	assert(t, l.Orders[0].size, float64(5))
	assert(t, l.Orders[1].size, float64(10))
	fmt.Println(l)
}
