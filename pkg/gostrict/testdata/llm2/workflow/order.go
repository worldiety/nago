// Package workflow keeps the orders of a shop in the browser and synchronizes them with the backend.
package workflow

import (
	"go.wdy.de/nago/pkg/gostrict/std/list"
	"go.wdy.de/nago/pkg/gostrict/std/num"
	"go.wdy.de/nago/pkg/gostrict/std/opt"
	"go.wdy.de/nago/pkg/gostrict/testdata/llm2/workflow/proto"
)

// Status names the state of an order.
type Status string

const (
	StatusDraft     Status = "draft"
	StatusSubmitted Status = "submitted"
	StatusPaid      Status = "paid"
	StatusShipped   Status = "shipped"
	StatusCancelled Status = "cancelled"
)

// Coupon is a known discount code.
type Coupon string

const (
	Save10   Coupon = "SAVE10"
	Save20   Coupon = "SAVE20"
	FreeShip Coupon = "FREESHIP"
)

// ShippingFee is charged for orders below FreeShippingFrom.
const ShippingFee = 4.95

// FreeShippingFrom is the subtotal, from which shipping is free.
const FreeShippingFrom = 50.0

// Order is the state of an order.
type Order interface{ isOrder() }

// Draft is an order, which the customer still edits.
type Draft struct {
	ID       string
	Customer string
	Items    list.List[proto.Item]
	Coupon   opt.Option[Coupon]
}

// Submitted is an order, which waits for its payment.
type Submitted struct {
	ID       string
	Customer string
	Items    list.List[proto.Item]
	Coupon   opt.Option[Coupon]
	Amount   float64
}

// Paid is an order, which waits for its shipment.
type Paid struct {
	ID       string
	Customer string
	Items    list.List[proto.Item]
	Amount   float64
	Method   proto.PayMethod
}

// Shipped is an order on its way.
type Shipped struct {
	ID       string
	Customer string
	Amount   float64
	Method   proto.PayMethod
	Carrier  string
	Tracking string
}

// Cancelled is an order, which will not be delivered.
type Cancelled struct {
	ID       string
	Customer string
	Reason   string
	From     Status
	Refund   opt.Option[float64]
}

func (Draft) isOrder()     {}
func (Submitted) isOrder() {}
func (Paid) isOrder()      {}
func (Shipped) isOrder()   {}
func (Cancelled) isOrder() {}

// StatusOf returns the status of an order.
func StatusOf(o Order) Status {
	switch o.(type) {
	case Draft:
		return StatusDraft
	case Submitted:
		return StatusSubmitted
	case Paid:
		return StatusPaid
	case Shipped:
		return StatusShipped
	case Cancelled:
		return StatusCancelled
	}
	return StatusDraft
}

// parseCoupon returns the coupon of a code.
func parseCoupon(code string) (Coupon, bool) {
	switch Coupon(code) {
	case Save10:
		return Save10, true
	case Save20:
		return Save20, true
	case FreeShip:
		return FreeShip, true
	default:
		return Save10, false
	}
}

// discountRate returns the part of the subtotal, which a coupon takes off.
func discountRate(c Coupon) float64 {
	switch c {
	case Save10:
		return 0.1
	case Save20:
		return 0.2
	case FreeShip:
		return 0
	}
	return 0
}

// freeShipping tells, whether a coupon waives the shipping fee.
func freeShipping(c Coupon) bool {
	switch c {
	case Save10, Save20:
		return false
	case FreeShip:
		return true
	}
	return false
}

// subtotal sums the positions.
func subtotal(items list.List[proto.Item]) float64 {
	sum := 0.0
	for _, it := range items.All() {
		sum += float64(it.Qty) * it.Price
	}
	return sum
}

// amountOf returns the amount, which the customer pays for the positions with an optional coupon.
func amountOf(items list.List[proto.Item], coupon opt.Option[Coupon]) float64 {
	sub := subtotal(items)
	rate := 0.0
	free := sub >= FreeShippingFrom
	if c, ok := coupon.Get(); ok {
		rate = discountRate(c)
		free = free || freeShipping(c)
	}
	amount := sub * (1 - rate)
	if !free {
		amount += ShippingFee
	}
	return roundCents(amount)
}

// roundCents rounds an amount to whole cents.
func roundCents(f float64) float64 {
	return float64(num.Round32(f*100)) / 100
}
