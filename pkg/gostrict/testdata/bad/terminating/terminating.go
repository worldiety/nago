// Package terminating uses the terminating profile.
//
//gostrict:terminating
package terminating

import (
	"iter"

	"go.wdy.de/nago/pkg/gostrict/std/fuel"
	"go.wdy.de/nago/pkg/gostrict/std/list"
)

func Forever() int32 {
	n := int32(0)
	for { // want GS040
		n++
		if n > 10 {
			break
		}
	}
	return n
}

func While(x float64) float64 {
	for x > 1 { // want GS040
		x /= 2
	}
	return x
}

func Fuel(x float64) float64 {
	f := fuel.New(100)
	for f.Burn() && x > 1 {
		x /= 2
	}
	return x
}

func Counted(n int32) int32 {
	s := int32(0)
	for i := int32(0); i < n; i++ {
		s += i
	}
	for i := n; i >= 0; i -= 2 {
		s += i
	}
	for i := int32(0); i <= 10; i++ {
		s += i
	}
	return s
}

func Changed(n int32) int32 {
	s := int32(0)
	for i := int32(0); i < n; i++ { // want GS040
		n++
		s += i
	}
	for i := int32(0); i <= n; i++ { // want GS040
		s += i
	}
	for i := int32(0); i < n; i++ { // want GS040
		i--
	}
	return s
}

func Grow(l list.List[int32]) int32 {
	for _, v := range l.All() {
		l.Append(v)
	}
	other := list.New[int32]()
	for _, v := range l.All() {
		other.Append(v)
	}
	return other.Len()
}

func Fact(n int32) int32 {
	if n <= 1 {
		return 1
	}
	return n * Fact(n-1)
}

func Halve(n int32) int32 {
	if n <= 0 {
		return 0
	}
	return 1 + Halve(n/2)
}

func Fib(n int32) int32 { // want GS040
	return Fib(n-1) + Fib(n-2)
}

func ping(n int32) int32 { // want GS040
	return pong(n)
}

func pong(n int32) int32 { // want GS040
	return ping(n)
}

func apply(f func() int32) int32 { // want GS040
	return f()
}

func Indirect() int32 { // want GS040
	return apply(Indirect)
}

type Shape interface{ Area() float64 }

type Twice struct{ S Shape }

func (t Twice) Area() float64 { // want GS040
	return 2 * t.S.Area()
}

func seq(n int32) iter.Seq[int32] {
	return func(yield func(int32) bool) {
		for i := range n {
			if !yield(i) {
				return
			}
		}
	}
}

func Iterate() int32 {
	s := int32(0)
	for v := range seq(5) {
		s += v
	}
	return s
}
