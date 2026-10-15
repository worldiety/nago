// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package gostrict translates a strict subset of Go into JavaScript. A program of the subset is valid Go, so it
// runs natively, in goja and in a browser with the same results, and the compiler proves, that it can not fail
// at runtime: there is no nil dereference, no index out of range, no division by zero and no panic. The subset
// is meant to be written by language models, so every rejected construct is reported with the reason and a fix,
// see [Diagnostics].
//
// # The subset
//
// Types:
//   - bool, string, float64, int32 (alias rune) and uint32. There is no int: it has 64 bits in Go, while
//     JavaScript computes with float64. int32 and uint32 wrap around in both. Write int32(0) or var n int32 for
//     an integer constant.
//   - named struct types, pointers to them, created only by &T{...}, functions and interfaces with methods.
//   - list.List[T] and dict.Map[K, V] from go.wdy.de/nago/pkg/gostrict/std instead of slices and maps. A dict
//     iterates in the order of insertion.
//   - generic functions, types and methods, including methods with own type parameters.
//   - sealed interfaces as sum types: an interface with an unexported method, e.g. interface{ isShape() }, is
//     implemented only by the types of its package, its variants. A type switch over its variants must be
//     exhaustive or have a default case, like the union types of golang/go#76920, whose notation can replace it
//     once Go has them. A type switch over another interface and type assertions are not allowed.
//   - enums: a named int32, uint32 or string type with constants of it. A switch over it must cover all constants
//     or have a default case.
//
// There is no nil except for error. Every variable, field and element must have a value: var p *T without a
// value, a struct literal without its pointer, function, interface, list or map fields and list.Make of such
// elements are rejected.
//
// Functions, which may have no result, return (T, bool), e.g. l.Get(i) and m.Get(k). If T may contain nil, the
// value must be used only where ok is true:
//
//	if v, ok := m.Get(k); ok { use(v) }
//	v, ok := l.Get(i)
//	if !ok { return }
//	use(v)
//
// The same guards prove, that an error is not nil, before its method is called, and that an integer divisor is
// not zero:
//
//	if err != nil { return err.Error() }
//	if d == 0 { return 0, errors.New("no values") }
//	return sum / d, nil
//
// A guarded variable must not be assigned between the guard and its use, and not by a function literal.
//
// Package variables are bound once: they are initialized, when the program is loaded, never change and hold
// only values, which can not change: basic values, structs of such values, handles, functions and list.Frozen
// for tables. The state, which changes, is a struct, which an exported function returns as a pointer and the host
// passes back, e.g. func NewGame() *Game and func Step(g *Game, dt float64). So a call depends only on its
// arguments and the instances of a program are interchangeable.
//
// Statements: if, for in all forms, range over int32, list.List.All, Values, dict.Map.All, Keys and iterators
// (iter.Seq, iter.Seq2), switch without fallthrough, labels with break and continue, return. A labeled break or
// continue must not leave a range over an iterator, because its body becomes a function.
//
// Not part of the subset: goroutines, channels, select, defer, recover, panic, goto, fallthrough, type switches
// and assertions, embedded fields, variadic parameters (calling list.Of(a, b) is fine), init functions, any as
// the type of a value, comparing interfaces except an error with nil, taking the address of a variable or field,
// arrays, slices, maps, len, append, make, new, indexing, slicing or ranging over strings, converting float64 to
// an integer (use num.Trunc32 and num.Round32), shifts by a variable count, method expressions and types declared
// inside functions.
//
// Packages of the standard library: math, strings (Contains, HasPrefix, HasSuffix, ReplaceAll, TrimPrefix,
// TrimSpace, TrimSuffix), errors (New) and iter. Anything else comes from the packages list, dict, text and num
// of go.wdy.de/nago/pkg/gostrict/std, whose functions never panic.
//
// # Translation
//
// Every function is translated once for each combination of type arguments, so the translated code knows the
// type of every value. Structs are JavaScript objects, which are copied on assignment like in Go. An assignment
// to an existing struct copies the fields into it, so that a pointer to it sees the change. Interfaces box the
// value with the methods of its type.
//
// # Loading
//
// [LoadModules] reads the packages from trees of files, e.g. an [embed.FS] of an application or the files, which a
// language model has written into a database. It needs no Go installation: the library of the subset and the
// allowed functions of the standard library are built in as stubs, see TestStubs. [Load] reads a module from
// disk.
//
// # Header packages
//
// A package with the directive //gostrict:header in its documentation is trusted: its bodies are not checked and
// run natively in Go, the loader uses only its declarations. Its boundary is checked: it exports only functions,
// whose parameters are bool, int32, uint32, float64 or string and whose results may also be an error, and each
// declares its JavaScript binding as a function expression with the same number of parameters:
//
//	//gostrict:js (a, b) => a * b + 1
//	func Mix(a, b float64) float64 { return float64(a*b) + 1 }
//
// A header package may export handles, which refer to objects of the host, see package std/js, e.g. a THREE.Mesh,
// and constraints, which unite handles or basic types. A function may be generic over such a constraint, e.g.
// func SetPosition[T Mesh | Group](o T, x, y, z float64), and keeps a single binding, while each instantiation
// gets its own native function. A Go file of the package with //go:build !gostrict is ignored by the subset and
// may provide a Go API for native hosts, see package lib/three.
//
// A binding returns several results as an array and an error as null or $errors_New(message). [Generate] writes
// the functions of a header package as [Funcs], which an engine calls instead of the bindings, and [GenerateTS] the
// TypeScript declaration for a JavaScript host. gojaengine.CheckBindings calls both implementations with edge
// cases and random values and reports, where they differ.
//
// # Messages
//
// A program exchanges messages with its backend, e.g. a three.js program in a browser, which synchronizes its
// state. A message is a record. The program does not send messages itself: an exported function returns a msg.Out,
// whose messages the host sends after the call. An exported function, which takes the state and a record,
// handles the messages of this type, which the backend sends at any time, see [Bundle.Handlers]. A [Session]
// couples an instance with the backend and serializes the calls like the event loop of a browser. The generated
// code of a program encodes the messages for it and dispatches its messages to a typed interface of the backend.
// Backend and program import the same package of messages, so they can not disagree about them.
//
// # Caching
//
// A [Cache] compiles a program once per hash of its sources and keeps the most recently used bundles. A bundle
// contains only the parts of the runtime, which it uses, and has a [Bundle.Hash], under which an engine caches
// its compiled form and a browser may cache it forever.
//
// # Engines
//
// An [Engine] executes a [Bundle] behind a neutral interface, so an application may replace goja by another
// engine. The values, which cross the boundary, are those of [ValueKind]. The native engine executes the Go code
// itself through the generated Exports of a package and is the reference in tests. Exported functions take and
// return records, which are structs of exported fields of these values, lists and maps of them, sealed interfaces,
// whose variants are records, and states, which the host holds as opaque tokens. [Bundle.Records] and
// [Bundle.Unions] describe them, e.g. for a form of the input of a use case. A handler of a sealed interface
// receives each of its variants as a message. The generated Client calls the
// exported functions with the types of Go in any engine, [Pool] spreads concurrent calls over instances and
// [Bundle.Expect] checks the signature, which an application requires of a program. In goja, many bundles may
// share one runtime, because a program of the subset can not reach the globals of JavaScript.
//
// Floating point is not exact across engines: JavaScript rounds every operation, while Go may fuse a
// multiplication and an addition into one instruction, e.g. on arm64, and other targets like a GPU may differ
// more. The native performance matters more than equal bits, so floats agree only approximately, see
// [CloseValues]. Integers, strings and the control flow agree exactly.
package gostrict
