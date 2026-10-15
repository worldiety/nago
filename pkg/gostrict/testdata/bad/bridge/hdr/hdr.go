// Package hdr is a header package with mistakes at its boundary.
//
//gostrict:header
package hdr

type Handle struct{ id int32 } // want GS034

//gostrict:js (h) => h
func (h Handle) ID() int32 { return h.id } // want GS034

//gostrict:js (v) => v
func Same[T any](v T) T { return v } // want GS034

//gostrict:js (xs) => xs.length
func Count(xs []int32) int32 { return int32(len(xs)) } // want GS034

func Missing(a int32) int32 { return a } // want GS028

//gostrict:js (a => a // want GS035
func Syntax(a int32) int32 { return a }

//gostrict:js (a) => a + b // want GS035
func Unbound(a, b int32) int32 { return a + b }

//gostrict:js async (a) => a // want GS035
func Async(a int32) int32 { return a }

//gostrict:js Math.abs // want GS035
func Abs(x float64) float64 { return x }

//gostrict:js (a, b) => a + b
func Add(a, b int32) int32 { return a + b }
