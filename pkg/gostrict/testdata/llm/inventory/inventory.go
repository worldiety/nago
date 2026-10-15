// Package inventory manages a small stock of items.
package inventory

import (
	"errors"

	"go.wdy.de/nago/pkg/gostrict/std/dict"
	"go.wdy.de/nago/pkg/gostrict/std/list"
	"go.wdy.de/nago/pkg/gostrict/std/num"
	"go.wdy.de/nago/pkg/gostrict/std/text"
)

type Item struct {
	Name  string
	Price float64
	Qty   int32
}

type Inventory struct {
	items dict.Map[string, *Item]
}

func NewInventory() *Inventory {
	return &Inventory{items: dict.New[string, *Item]()}
}

func (inv *Inventory) Add(name string, price float64, qty int32) error {
	if name == "" {
		return errors.New("empty name")
	}
	if qty < 0 {
		return errors.New("negative quantity for " + name)
	}
	if it, ok := inv.items.Get(name); ok {
		it.Qty += qty
		return nil
	}
	inv.items.Set(name, &Item{Name: name, Price: price, Qty: qty})
	return nil
}

func (inv *Inventory) Remove(name string, qty int32) error {
	it, ok := inv.items.Get(name)
	if !ok {
		return errors.New("no item " + name)
	}
	if it.Qty < qty {
		return errors.New("not enough " + name + ": have " + text.FromInt(it.Qty) + ", want " + text.FromInt(qty))
	}
	it.Qty -= qty
	if it.Qty == 0 {
		inv.items.Delete(name)
	}
	return nil
}

func (inv *Inventory) Total() float64 {
	var total float64
	for _, it := range inv.items.All() {
		total += it.Price * float64(it.Qty)
	}
	return total
}

func formatMoney(f float64) string {
	cents := num.Round32(f * 100)
	sign := ""
	if cents < 0 {
		sign = "-"
		cents = -cents
	}
	frac := cents % 100
	fs := text.FromInt(frac)
	if frac < 10 {
		fs = "0" + fs
	}
	return sign + text.FromInt(cents/100) + "." + fs
}

func Run() string {
	inv := NewInventory()
	errs := list.New[string]()
	record := func(err error) {
		if err != nil {
			errs.Append(err.Error())
		}
	}
	record(inv.Add("widget", 2.5, 10))
	record(inv.Add("bolt", 0.1, 100))
	record(inv.Add("apple", 0.75, 4))
	record(inv.Add("widget", 2.5, 5))
	record(inv.Add("", 1, 1))
	record(inv.Add("nut", 0.05, -3))
	record(inv.Remove("bolt", 40))
	record(inv.Remove("gear", 1))
	record(inv.Remove("apple", 9))

	names := list.New[string]()
	for name := range inv.items.Keys() {
		names.Append(name)
	}
	list.SortFunc(names, text.Compare)

	lines := list.New[string]()
	for name := range names.Values() {
		if it, ok := inv.items.Get(name); ok {
			lines.Append(it.Name + " x" + text.FromInt(it.Qty) + " @ " + formatMoney(it.Price))
		}
	}
	lines.Append("total: " + formatMoney(inv.Total()))
	for e := range errs.Values() {
		lines.Append("error: " + e)
	}
	return text.Join(lines, "\n")
}
