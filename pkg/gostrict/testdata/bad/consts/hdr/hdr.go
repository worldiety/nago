// Package hdr is a header package, whose host reads arguments at compile time.
//
//gostrict:header
package hdr

import "go.wdy.de/nago/pkg/gostrict/std/js"

// Kind is an enum.
type Kind int32

const (
	KindA Kind = 1
	KindB Kind = 2
)

// Point is a record.
type Point struct {
	X, Y float64
}

// Activity names a step.
//
//gostrict:host
//gostrict:const name
func Activity(name string) {}

// Pair takes two constants and a value.
//
//gostrict:host
//gostrict:const a, b
func Pair(a string, b int32, c float64) {}

// Mark takes a constant of an enum.
//
//gostrict:host
//gostrict:const k
func Mark(k Kind) {}

//gostrict:host
//gostrict:const nope // want GS028
func Unknown(name string) {}

//gostrict:host
//gostrict:const // want GS028
func Empty(name string) {}

//gostrict:host
//gostrict:const p // want GS028
func Record(p Point) {}

//gostrict:host
//gostrict:const names // want GS028
func Names(names ...string) {}

//gostrict:host
//gostrict:const v // want GS028
func Generic[T int32 | float64](v T) {}

// Step is a handle.
type Step struct{ ref js.Ref }

//gostrict:host
//gostrict:const name // want GS028
func (s Step) Label(name string) Step { return s }
