package concurrent

import (
	"errors"

	"go.wdy.de/nago/pkg/gostrict/std/big"
	"go.wdy.de/nago/pkg/gostrict/std/decimal"
	"go.wdy.de/nago/pkg/gostrict/std/dict"
	"go.wdy.de/nago/pkg/gostrict/std/list"
	"go.wdy.de/nago/pkg/gostrict/std/opt"
	"go.wdy.de/nago/pkg/gostrict/testdata/bad/concurrent/hdr"
)

type Kind string

type Plain struct {
	ID    string
	Count int32
	Kind  Kind
	Sum   big.Int
	Cost  decimal.Dec
	Next  opt.Option[string]
	Tags  list.Frozen[string]
}

type WithList struct {
	Items list.List[string]
}

type Box struct {
	N int32
}

func (b *Box) Inc() {}

func handler(d hdr.Done) error {
	return errors.New(d.Text)
}

const limit = "x"

var table = list.Freeze(list.Of("a"))

func Values(in Plain, n int32, s string, k Kind) string {
	id := in.ID + s
	return hdr.Send(func(d hdr.Done) error {
		local := d.Text + id + in.ID + limit + string(k)
		if n > 0 {
			local = local + "!"
		}
		if v, ok := table.Get(0); ok {
			local = local + v
		}
		return errors.New(local)
	})
}

func Named() {
	hdr.Send(handler)
	hdr.Send((handler))
}

func Ranged(l list.List[string]) {
	for _, v := range l.All() {
		hdr.Send(func(d hdr.Done) error { return errors.New(v) })
	}
}

func Mutable(l list.List[string], m dict.Map[string, string], p *Plain, w WithList, f func() string, e error, c hdr.Conn) {
	hdr.Send(func(d hdr.Done) error {
		l.Append(d.Text) // want GS046
		return nil
	})
	hdr.Send(func(d hdr.Done) error {
		m.Set("a", d.Text) // want GS046
		return nil
	})
	hdr.Send(func(d hdr.Done) error {
		return errors.New(p.ID) // want GS046
	})
	hdr.Send(func(d hdr.Done) error {
		return errors.New(w.Items.GetOr(0, "")) // want GS046
	})
	hdr.Send(func(d hdr.Done) error {
		return errors.New(f()) // want GS046
	})
	hdr.Send(func(d hdr.Done) error {
		return e // want GS046
	})
	hdr.Send(func(d hdr.Done) error {
		c.Later(func() {}) // want GS046
		return nil
	})
}

func Generic[T any](v T) {
	hdr.Send(func(d hdr.Done) error {
		use(v) // want GS046
		return nil
	})
}

func use[T any](v T) {}

func Reassigned(in Plain) {
	x := "a"
	x = "b"
	hdr.Send(func(d hdr.Done) error {
		return errors.New(x) // want GS046
	})
	y := "a"
	hdr.Send(func(d hdr.Done) error {
		y = d.Text // want GS046
		return nil
	})
	use(y)
	in.ID = "c"
	hdr.Send(func(d hdr.Done) error {
		return errors.New(in.ID) // want GS046
	})
	for i := int32(0); i < 3; i = i + 1 {
		hdr.Send(func(d hdr.Done) error {
			use(i) // want GS046
			return nil
		})
	}
	var b Box
	b.Inc()
	hdr.Send(func(d hdr.Done) error {
		use(b.N) // want GS046
		return nil
	})
}

func NamedResult() (r string) {
	hdr.Send(func(d hdr.Done) error {
		return errors.New(r) // want GS046
	})
	return "a"
}

func NotDirect(b Box) {
	f := func(d hdr.Done) error { return nil }
	hdr.Send(f)        // want GS046
	hdr.Send(pick())   // want GS046
	hdr.Send(b.Handle) // want GS046
	g := hdr.Send      // want GS046
	g(handler)
	h := hdr.SendConst // want GS044
	h("a", handler)
}

func (b Box) Handle(d hdr.Done) error { return nil }

func pick() func(d hdr.Done) error { return handler }

func Nested(id string, l list.List[string]) {
	hdr.Send(func(d hdr.Done) error {
		inner := list.New[string]()
		fn := func() {
			inner.Append(id)
			l.Append(id) // want GS046
		}
		fn()
		return nil
	})
}
