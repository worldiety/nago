// Package proto declares the messages, which the order workflow exchanges with its backend.
package proto

import "go.wdy.de/nago/pkg/gostrict/std/list"

// PayMethod is the way, in which an order was paid.
type PayMethod int32

const (
	Card PayMethod = iota
	Invoice
	PayPal
)

// ErrorCode classifies a rejected request.
type ErrorCode string

const (
	UnknownOrder      ErrorCode = "unknown_order"
	DuplicateOrder    ErrorCode = "duplicate_order"
	InvalidTransition ErrorCode = "invalid_transition"
	InvalidInput      ErrorCode = "invalid_input"
	AmountMismatch    ErrorCode = "amount_mismatch"
)

// Item is a position of an order.
type Item struct {
	SKU   string
	Qty   int32
	Price float64
}

// Incoming is a message, which the backend sends at any time.
type Incoming interface{ isIncoming() }

// Payment reports, that the customer paid an order.
type Payment struct {
	OrderID string
	Amount  float64
	Method  PayMethod
}

// Shipment reports, that the warehouse shipped an order.
type Shipment struct {
	OrderID  string
	Carrier  string
	Tracking string
}

// CancelRequest asks to cancel an order, e.g. by the support.
type CancelRequest struct {
	OrderID string
	Reason  string
}

func (Payment) isIncoming()       {}
func (Shipment) isIncoming()      {}
func (CancelRequest) isIncoming() {}

// Event is a message, which the program sends to the backend.
type Event interface{ isEvent() }

// Created tells, that a draft was created.
type Created struct {
	OrderID  string
	Customer string
}

// ItemAdded tells, that a draft got another position.
type ItemAdded struct {
	OrderID  string
	Item     Item
	Subtotal float64
}

// Submitted tells, that an order waits for its payment.
type Submitted struct {
	OrderID string
	Items   list.List[Item]
	Amount  float64
	Coupon  string
}

// Paid tells, that the payment of an order was accepted.
type Paid struct {
	OrderID string
	Amount  float64
	Method  PayMethod
}

// Shipped tells, that an order is on its way.
type Shipped struct {
	OrderID  string
	Carrier  string
	Tracking string
}

// Cancelled tells, that an order was cancelled and how much is refunded.
type Cancelled struct {
	OrderID string
	Reason  string
	Refund  float64
}

// Rejected tells, that a request or message could not be applied.
type Rejected struct {
	OrderID string
	Code    ErrorCode
	State   string
	Action  string
	Message string
}

func (Created) isEvent()   {}
func (ItemAdded) isEvent() {}
func (Submitted) isEvent() {}
func (Paid) isEvent()      {}
func (Shipped) isEvent()   {}
func (Cancelled) isEvent() {}
func (Rejected) isEvent()  {}
