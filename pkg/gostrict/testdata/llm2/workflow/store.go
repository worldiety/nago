package workflow

import (
	"math"
	"strings"

	"go.wdy.de/nago/pkg/gostrict/std/dict"
	"go.wdy.de/nago/pkg/gostrict/std/list"
	"go.wdy.de/nago/pkg/gostrict/std/msg"
	"go.wdy.de/nago/pkg/gostrict/std/opt"
	"go.wdy.de/nago/pkg/gostrict/std/text"
	"go.wdy.de/nago/pkg/gostrict/testdata/llm2/workflow/proto"
)

// Store holds all orders by their id.
type Store struct {
	Orders dict.Map[string, Order]
}

// New returns an empty store.
func New() *Store {
	return &Store{Orders: dict.New[string, Order]()}
}

// Summary describes the store.
type Summary struct {
	Drafts    int32
	Submitted int32
	Paid      int32
	Shipped   int32
	Cancelled int32
	Refunds   int32
	TotalPaid string
	Refunded  string
}

func reject(id string, code proto.ErrorCode, state Status, action, message string) proto.Event {
	return proto.Rejected{OrderID: id, Code: code, State: string(state), Action: action, Message: message}
}

func invalid(id string, o Order, action string) proto.Event {
	st := StatusOf(o)
	return reject(id, proto.InvalidTransition, st, action, "can not "+action+" an order, which is "+string(st))
}

func unknown(id, action string) proto.Event {
	return reject(id, proto.UnknownOrder, "", action, "there is no order "+id)
}

// CreateDraft starts a new order of the customer with an optional coupon code, which may be empty.
func CreateDraft(s *Store, id, customer, coupon string) msg.Out[proto.Event] {
	out := msg.New[proto.Event]()
	id = strings.TrimSpace(id)
	if id == "" {
		out.Send(reject(id, proto.InvalidInput, "", "create", "the id is empty"))
		return out
	}
	if s.Orders.Has(id) {
		out.Send(reject(id, proto.DuplicateOrder, "", "create", "the order "+id+" exists already"))
		return out
	}
	c := opt.None[Coupon]()
	code := strings.TrimSpace(coupon)
	if code != "" {
		parsed, ok := parseCoupon(text.UpperASCII(code))
		if !ok {
			out.Send(reject(id, proto.InvalidInput, "", "create", "unknown coupon "+code))
			return out
		}
		c = opt.Some(parsed)
	}
	s.Orders.Set(id, Draft{ID: id, Customer: customer, Items: list.New[proto.Item](), Coupon: c})
	out.Send(proto.Created{OrderID: id, Customer: customer})
	return out
}

// AddItem adds a position to a draft.
func AddItem(s *Store, id, sku string, qty int32, price float64) msg.Out[proto.Event] {
	out := msg.New[proto.Event]()
	if qty <= 0 || price < 0 || math.IsNaN(price) || math.Abs(price) > math.MaxFloat64 || sku == "" {
		out.Send(reject(id, proto.InvalidInput, "", "add item", "the position is not valid"))
		return out
	}
	o, ok := s.Orders.Get(id)
	if !ok {
		out.Send(unknown(id, "add item"))
		return out
	}
	switch v := o.(type) {
	case Draft:
		item := proto.Item{SKU: sku, Qty: qty, Price: price}
		v.Items.Append(item)
		s.Orders.Set(id, v)
		out.Send(proto.ItemAdded{OrderID: id, Item: item, Subtotal: subtotal(v.Items)})
	case Submitted, Paid, Shipped, Cancelled:
		out.Send(invalid(id, v, "add item"))
	}
	return out
}

// Submit closes a draft and waits for its payment.
func Submit(s *Store, id string) msg.Out[proto.Event] {
	out := msg.New[proto.Event]()
	o, ok := s.Orders.Get(id)
	if !ok {
		out.Send(unknown(id, "submit"))
		return out
	}
	switch v := o.(type) {
	case Draft:
		if v.Items.Len() == 0 {
			out.Send(reject(id, proto.InvalidInput, StatusDraft, "submit", "the order has no items"))
			return out
		}
		amount := amountOf(v.Items, v.Coupon)
		s.Orders.Set(id, Submitted{ID: id, Customer: v.Customer, Items: v.Items, Coupon: v.Coupon, Amount: amount})
		code := ""
		if c, ok := v.Coupon.Get(); ok {
			code = string(c)
		}
		out.Send(proto.Submitted{OrderID: id, Items: v.Items.Clone(), Amount: amount, Coupon: code})
	case Submitted, Paid, Shipped, Cancelled:
		out.Send(invalid(id, v, "submit"))
	}
	return out
}

// Cancel cancels an order on behalf of the customer.
func Cancel(s *Store, id, reason string) msg.Out[proto.Event] {
	out := msg.New[proto.Event]()
	cancel(s, id, reason, out)
	return out
}

func cancel(s *Store, id, reason string, out msg.Out[proto.Event]) {
	o, ok := s.Orders.Get(id)
	if !ok {
		out.Send(unknown(id, "cancel"))
		return
	}
	switch v := o.(type) {
	case Draft:
		s.Orders.Set(id, Cancelled{ID: id, Customer: v.Customer, Reason: reason, From: StatusDraft})
		out.Send(proto.Cancelled{OrderID: id, Reason: reason})
	case Submitted:
		s.Orders.Set(id, Cancelled{ID: id, Customer: v.Customer, Reason: reason, From: StatusSubmitted})
		out.Send(proto.Cancelled{OrderID: id, Reason: reason})
	case Paid:
		s.Orders.Set(id, Cancelled{ID: id, Customer: v.Customer, Reason: reason, From: StatusPaid, Refund: opt.Some(v.Amount)})
		out.Send(proto.Cancelled{OrderID: id, Reason: reason, Refund: v.Amount})
	case Shipped, Cancelled:
		out.Send(invalid(id, v, "cancel"))
	}
}

// OnMessage handles the messages of the backend.
func OnMessage(s *Store, m proto.Incoming) msg.Out[proto.Event] {
	out := msg.New[proto.Event]()
	switch v := m.(type) {
	case proto.Payment:
		pay(s, v, out)
	case proto.Shipment:
		ship(s, v, out)
	case proto.CancelRequest:
		cancel(s, v.OrderID, v.Reason, out)
	}
	return out
}

func pay(s *Store, p proto.Payment, out msg.Out[proto.Event]) {
	o, ok := s.Orders.Get(p.OrderID)
	if !ok {
		out.Send(unknown(p.OrderID, "pay"))
		return
	}
	switch v := o.(type) {
	case Submitted:
		if math.IsNaN(p.Amount) || math.Abs(p.Amount-v.Amount) >= 0.005 {
			out.Send(reject(p.OrderID, proto.AmountMismatch, StatusSubmitted, "pay",
				"expected "+text.FromFloat(v.Amount, 2)+" but got "+text.FromFloat(p.Amount, 2)))
			return
		}
		s.Orders.Set(p.OrderID, Paid{ID: v.ID, Customer: v.Customer, Items: v.Items, Amount: v.Amount, Method: p.Method})
		out.Send(proto.Paid{OrderID: p.OrderID, Amount: v.Amount, Method: p.Method})
	case Draft, Paid, Shipped, Cancelled:
		out.Send(invalid(p.OrderID, v, "pay"))
	}
}

func ship(s *Store, sh proto.Shipment, out msg.Out[proto.Event]) {
	o, ok := s.Orders.Get(sh.OrderID)
	if !ok {
		out.Send(unknown(sh.OrderID, "ship"))
		return
	}
	switch v := o.(type) {
	case Paid:
		if strings.TrimSpace(sh.Tracking) == "" {
			out.Send(reject(sh.OrderID, proto.InvalidInput, StatusPaid, "ship", "the tracking number is empty"))
			return
		}
		s.Orders.Set(sh.OrderID, Shipped{ID: v.ID, Customer: v.Customer, Amount: v.Amount, Method: v.Method,
			Carrier: sh.Carrier, Tracking: sh.Tracking})
		out.Send(proto.Shipped{OrderID: sh.OrderID, Carrier: sh.Carrier, Tracking: sh.Tracking})
	case Draft, Submitted, Shipped, Cancelled:
		out.Send(invalid(sh.OrderID, v, "ship"))
	}
}

// Summarize counts the orders per state and sums the paid and refunded amounts.
func Summarize(s *Store) Summary {
	var sum Summary
	paid := 0.0
	refunded := 0.0
	for _, o := range s.Orders.All() {
		switch v := o.(type) {
		case Draft:
			sum.Drafts++
		case Submitted:
			sum.Submitted++
		case Paid:
			sum.Paid++
			paid += v.Amount
		case Shipped:
			sum.Shipped++
			paid += v.Amount
		case Cancelled:
			sum.Cancelled++
			if r, ok := v.Refund.Get(); ok {
				sum.Refunds++
				refunded += r
			}
		}
	}
	sum.TotalPaid = text.FromFloat(paid, 2)
	sum.Refunded = text.FromFloat(refunded, 2)
	return sum
}
