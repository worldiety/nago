// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package std

// PROPOSAL - do not use
// There is a lot of discussion, why Go don't need a Result type and why it may cause more harm than benefits.
// They are right, if a Result type is used alone, today. Actually a result does not make much sense at all,
// even not much sense for Rust. A language, which has proper sum type support, would just declare that tuple type.
// So why has Rust that? Because of all that helper and monads for Result, which allows writing composable functions
// more easily. A different pattern matching approach or build-in types may be an alternative, but that becomes
// esoteric.
// So what is the problem with the (T,error) tuple in Go? By definition, it is overloaded with a lot of zero value
// initializations, which is a subset of the problem of creating valid instances of types. The zero value is just
// a special case, which may be fine or not, which totally depends on your domain. Often, there are more invalid
// type states than valid ones, including the zero value to be invalid. One consequence is the dreaded nil pointer
// issue, but we don't need to solve that in Go, because a nil pointer to type may be fine or not, but another
// non-nil literal instance may be not. This can only be mitigated by enforcing a constructor usage.
// Thus as long, as we cannot express the usage of constructors, this Result type is totally useless and will
// create more signal-to-noise ratio than the conventional multi-return.
//
// Therefore, we plan to introduce a linter, which allows marking types to be only valid when initalized with a
// constructor. E.g.
//
//			// ![constructor]
//			type ID string
//
//			// ![constructor]
//			type Person struct {
//		     ...
//	      _ std.EnforceConstructor // could be an alternative, but deriving from base types are not expressible
//		    }
//
// This notation disallows all zero value initializations in code:
//
//	var p Person // invalid
//	func x() (p Person) {} // invalid
//	p:=Person{} // invalid
//	p:=Person{...} // invalid
//
// Instead, only allowed, is a constructor usage:
//
//	p := NewPerson()
//
// A constructor is defined as a function within the same Package, which returns at least one value of the required type.
// That is easy to check, no fuss with naming or multiple returns.
//
// Limitations:
//   - probably we can't do generic resolving and/or must ignore that
//   - TODO: big problem: what if we want a sum type like below, but the either type enforces a constructor and we cannot use it, due to generics?
type Result[T, E any] struct {
	// this is not a sum type and cannot be properly modelled and is also more wasteful than the tagged union of rust.
	// but our goal is to express either-or
	ok    T
	err   E
	state sumState
}

type sumState uint8

const (
	zero sumState = iota
	ok
	err
)

func Ok[T, E any](v T) Result[T, E] {
	return Result[T, E]{
		ok:    v,
		state: ok,
		// TODO we cannot avoid the zero value of E here
	}
}

func Err[T, E any](e E) Result[T, E] {
	return Result[T, E]{
		err:   e,
		state: err,
		// TODO we cannot avoid the zero value of T here
	}
}

func (r Result[T, E]) Unwrap() T {
	if r.state != ok {
		panic("result is not ok")
	}

	return r.ok
}

func (r Result[T, E]) UnwrapOr(e T) T {
	if r.state != ok {
		return e
	}

	return r.ok
}

func (r Result[T, E]) UnwrapErr() E {
	if r.state != err {
		panic("result is not err")
	}

	return r.err
}

func (r Result[T, E]) Ok() bool {
	return r.state == ok
}

func (r Result[T, E]) Err() bool {
	return r.state == err
}
