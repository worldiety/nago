// Package hdr is a header package, whose host calls functions on another goroutine.
//
//gostrict:header
package hdr

import "go.wdy.de/nago/pkg/gostrict/std/js"

// Conn is a handle.
type Conn struct{ ref js.Ref }

// Done is a record, which the function takes.
type Done struct {
	Text string
}

// Send calls done, when the request ended.
//
//gostrict:host
//gostrict:calls concurrent
func Send(done func(d Done) error) string { return "" }

// SendConst has a constant, too.
//
//gostrict:host
//gostrict:const name
//gostrict:calls concurrent
func SendConst(name string, done func(d Done) error) string { return "" }

// NoFunc takes no function.
//
//gostrict:host
//gostrict:calls concurrent // want GS028
func NoFunc() {}

// Bad is a wrong value.
//
//gostrict:host
//gostrict:calls sometimes // want GS028
func Bad(done func()) {}

// Open returns a handle.
//
//gostrict:host
func Open() Conn { return Conn{} }

// Later calls fn after it returned.
//
//gostrict:host
//gostrict:calls concurrent // want GS028
func (c Conn) Later(fn func()) {}
