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
//   - opt.Option[T] for a value, which may be missing, e.g. the next node of a list, Next opt.Option[*Node]. Its
//     zero value is empty, so it may be left out of a struct literal. opt.Some(v) and opt.None[T]() create it,
//     o.Get() returns (T, bool) and o.Or(def) the value or def.
//   - generic functions, types and methods, including methods with own type parameters.
//   - sealed interfaces as sum types: an interface with an unexported method, e.g. interface{ isShape() }, is
//     implemented only by the types of its package, its variants. A type switch over its variants must be
//     exhaustive or have a default case, like the union types of golang/go#76920, whose notation can replace it
//     once Go has them. A variant may implement it with a pointer receiver, then *T is the variant, and a generic
//     sealed interface, e.g. Result[T], has the generic types of its package as variants, e.g. Ok[T]. Two values
//     compare like in Go, if all variants are comparable. A type switch over another interface and type
//     assertions are not allowed.
//   - enums: a named int32, uint32 or string type with constants of it. A switch over it must cover all constants
//     or have a default case.
//   - big.Int of go.wdy.de/nago/pkg/gostrict/std/big for integers of any size and decimal.Dec of std/decimal for
//     exact decimals, e.g. amounts of money. Both never change, functions return new values, and == is not
//     allowed for them, compare with big.Cmp(a, b) == 0 or decimal.Cmp(a, b) == 0. Create them with
//     big.FromInt32, big.Parse, decimal.FromInt32 or decimal.Parse("19.99"). Addition, subtraction and
//     multiplication of decimals are exact, a division and rounding take the scale and an explicit rounding:
//     decimal.Quo(net, decimal.FromInt32(3), 2, decimal.HalfUp) and decimal.Round(d, 2, decimal.HalfEven).
//     Format them with big.String and decimal.String, never through float64.
//
// There is no nil except for error. Every variable, field and element must have a value, unless it is an
// opt.Option: var p *T without a value, a struct literal without its pointer, function, interface, list or map
// fields and list.Make of such elements are rejected.
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
// (iter.Seq, iter.Seq2), switch without fallthrough, labels with break and continue, return.
//
// Not part of the subset: goroutines, channels, select, defer, recover, panic, goto, fallthrough, type switches
// and assertions, embedded fields, variadic parameters (calling list.Of(a, b) is fine), init functions, any as
// the type of a value, comparing interfaces except an error with nil and sealed interfaces, whose variants are
// all comparable, taking the address of a variable or field,
// arrays, slices, maps, len, append, make, new, indexing, slicing or ranging over strings, converting float64 to
// an integer (use num.Trunc32 and num.Round32), shifts by a variable count, method expressions and types declared
// inside functions.
//
// Numbers become strings with text.FromInt(n) and text.FromFloat(f, decimals), which rounds the same in Go and
// JavaScript.
//
// Packages of the standard library: math (without Frexp, Ldexp and the functions of int; a constant argument like
// the sign of math.IsInf(x, 0) is fine), strings (Contains, HasPrefix, HasSuffix, ReplaceAll, TrimPrefix,
// TrimSpace, TrimSuffix), errors (New) and iter. Anything else comes from the packages list, dict, opt, text,
// num, fuel, big and decimal of go.wdy.de/nago/pkg/gostrict/std, whose functions never panic.
//
// # Terminating profile
//
// A package with the directive //gostrict:terminating in its documentation, or every package, if the host sets
// [Program.Terminating], must also prove, that each call ends, e.g. a script, which runs on a server:
//
//   - a for loop counts with a counter, which only its post statement changes by a constant, up to a bound, which
//     is a constant or a variable, which the loop does not change: for i := int32(0); i < n; i++.
//   - a range over a number, a dict.Map, a list.List, to which its body does not append, or an iterator.
//   - a loop with another condition burns fuel of package go.wdy.de/nago/pkg/gostrict/std/fuel:
//     f := fuel.New(1000); for f.Burn() && x > eps { ... }.
//   - a function calls itself only with a decreasing integer parameter, which its first statement checks:
//     if n <= 1 { return 1 }; return n * Fact(n-1). A call of a function value or of a method of an interface
//     may call any function, which is used as a value, or any method of its name, so such cycles are reported too.
//
// # Translation
//
// Every function is translated once for each combination of type arguments, so the translated code knows the
// type of every value. Structs are JavaScript objects, which are copied on assignment like in Go. An assignment
// to an existing struct copies the fields into it, so that a pointer to it sees the change. Interfaces box the
// value with the methods of its type. A struct type, whose values the whole program never changes in place, is
// immutable: its values are shared instead of copied and all its zero values are one object, e.g. a style in a
// package variable, which many cells of a table refer to. So it is cheap to write sheet.Cell{Text: t, Style: header}
// for each cell.
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
// run natively in Go, the loader uses only its declarations. Only a trusted module may contain header packages,
// see [Module.Trusted], so a program, which a language model writes, uses the built-in libraries, e.g.
// go.wdy.de/nago/pkg/gostrict/lib/three, and can not bind JavaScript itself. Its boundary is checked: it exports
// only functions, whose parameters are bool, int32, uint32, float64, string, big.Int, records of another package,
// lists and maps of them or functions of such values, and whose results may also be an error, and each declares
// its JavaScript binding as a function expression with the same number of parameters. A record is a plain object
// with its fields in JavaScript:
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
// A header package may also export the values of its API: records, enums, constants and variables, whose binding is
// their value in JavaScript, e.g. //gostrict:js "100%" for var Full Length. Its types may have methods with a value
// receiver, whose binding takes the receiver first, e.g. the builder methods of a view, and a function may be
// variadic, whose binding takes the arguments as an array. A sealed interface of its handles, e.g.
// type View interface{ isView() }, unites them, so that a container takes any of them, like ui.VStack(ui.Text("a"),
// ui.PrimaryButton(f)); the host receives the handle and its type. lib/ui mirrors the views of nago this way.
//
// A binding returns several results as an array and an error as null or $errors_New(message). [Generate] writes
// the functions of a header package as [Funcs], which an engine calls instead of the bindings, and [GenerateTS] the
// TypeScript declaration for a JavaScript host. Only header packages need generated code, because reflection can
// neither instantiate their generic functions nor construct their handles; a program never does. gojaengine.CheckBindings calls both implementations with edge
// cases and random values and reports, where they differ.
//
// A function of a header package may take a function of the program, e.g. the function, which renders a cell of a
// sheet. A binding receives the JavaScript function, a host in Go a [Callback], which the generated code turns into
// a Go function. The host may call it during the call, e.g. in the host function, or later between two calls of the
// instance, e.g. when a window renders the sheet, but always on the goroutine, which uses the instance. A function
// with the directive //gostrict:host instead of a binding is implemented only by a host in Go, e.g. the views of
// nago in lib/ui, so its program runs only on a server, whose engine refuses to load the bundle without it. Such a
// program may return a handle from an exported function, e.g. the ui.View, which the server renders.
//
// # Messages
//
// A program exchanges messages with its backend, e.g. a three.js program in a browser, which synchronizes its
// state. A message is a record. The program does not send messages itself: an exported function returns a msg.Out,
// whose messages the host sends after the call. An exported function, which takes the state and a record,
// handles the messages of this type, which the backend sends at any time, see [Bundle.Handlers]. A [Session]
// couples an instance with the backend and serializes the calls like the event loop of a browser. The backend
// sends a record or a variant as it is, see [Session.Send], and receives the messages of the program with typed
// functions, see [Router] and [Handle]. Backend and program import the same package of messages, so they can not
// disagree about them. The variants of a sealed interface, which the backend decodes, are registered once with
// [Variants].
//
// The fields of a record are bool, int32, uint32, float64, string, enums of them, list.List and dict.Map of such
// values and other records; opt.Option, pointers, functions and interfaces stay inside the program. A message may
// also be a sealed interface, whose variants are records: msg.Out[messages.Event] sends any variant, and a handler
// of the interface receives each of them:
//
//	func OnMessage(s *Store, m messages.Incoming) msg.Out[messages.Event] {
//		switch m := m.(type) {
//		case messages.Payment, messages.Refund:
//			...
//		}
//		return msg.New[messages.Event]()
//	}
//
// An exported function, which is no handler, may take any parameters of these types besides the state, e.g.
// func AddItem(s *Store, sku string, qty int32). Like in Go, a function still ends with a return after an
// exhaustive switch.
//
// # Programs in the browser
//
// The components threejs and canvas2 of nago run a program in the browser, see package presentation/ui/program.
// It draws with the libraries go.wdy.de/nago/pkg/gostrict/lib/three or lib/canvas, which bind three.js and the 2D
// context of a canvas. The component calls the exported functions, which it is told, e.g. Init, which returns the
// state, Resize with the size of the canvas and Frame with the time in seconds in each animation frame:
//
//	func Init(n int32) *Game
//	func Resize(g *Game, width, height float64)
//	func Frame(g *Game, t float64) msg.Out[messages.Score]
//
// Events of the host are records, which go to handlers like the messages of the backend: input.Event of package
// lib/input for the pointer, the wheel and the keys, and asset.Loaded and asset.Failed of package lib/asset for
// images, textures and models, which canvas.LoadImage, three.LoadTexture and three.LoadModel load. A handler runs
// between two frames, so it changes the state, which the next frame draws:
//
//	func OnInput(g *Game, e input.Event) {
//		if e.Type == input.KeyDown && e.Code == "Space" {
//			g.Jump = true
//		}
//	}
//
// A handler is found by its signature, not by its name: an exported function, which takes the state and a record,
// handles this record, and only one function may handle a record. A handler, like any exported function, may
// return msg.Out results, e.g. OnInput sends a finished stroke, also several of different types. A record may
// have no fields, e.g. type Clear struct{}. The state may hold handles, e.g. a canvas.Image or a three.Mesh, and
// a message package is a plain package of records, which the backend imports as well.
//
// A table, which never changes, is a package variable of list.Frozen:
//
//	var palette = list.Freeze(list.Of[uint32](0xff0000ff, 0x00ff00ff))
//
// A program, which needs the state of its backend, asks for it in Init, e.g. by returning (*Game,
// msg.Out[messages.Hello]), because the messages of a broken connection are lost.
//
// # Whole programs
//
// [LoadRoots] loads many packages as one program, e.g. all scripts of an application, which call each other by
// importing each other. The exported functions of the roots are its entry points, e.g. "example.com/scripts/crm/
// read.Main", which take and return records, lists, maps, basic values and states, but no sealed interfaces, whose
// methods would bring code of the caller into the callee. The compiler calls an [Aspect] of the host around each
// entry point: when the host calls it and when another package calls it, also through a function value, but not
// within its own package. So the host keeps a frame of each call, e.g. with the rights of the callee and a
// transaction of its area, which its header functions use, while the program neither sees nor bypasses it. A
// denied call or another failure of the host, see [Abort], ends the whole call, which the program can not catch,
// and the aspect exits each frame with the error. Imports have no cycles, so the depth of the frames is bounded by
// the program itself.
//
// [CompileWhole] checks the whole program once, and [Whole.Bundle] translates a bundle for some entry points, which
// contains only what they reach. A changed script is a new whole program, while running calls keep their bundles.
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
// itself, whose functions native.Reflect adapts, and is the reference in tests. Exported functions take and
// return records, which are structs of exported fields of these values, lists and maps of them, sealed interfaces,
// whose variants are records, and states, which the host holds as opaque tokens. [Bundle.Records] and
// [Bundle.Unions] describe them, e.g. for a form of the input of a use case. A handler of a sealed interface
// receives each of its variants as a message. [Call] calls the exported functions with the values of Go in any
// engine and converts them by reflection, see [ToNeutral], [Pool] spreads concurrent calls over instances and
// [Bundle.Expect] checks the signature, which an application requires of a program. In goja, many bundles may
// share one runtime, because a program of the subset can not reach the globals of JavaScript.
//
// Floating point is not exact across engines: JavaScript rounds every operation, while Go may fuse a
// multiplication and an addition into one instruction, e.g. on arm64, and other targets like a GPU may differ
// more. The native performance matters more than equal bits, so floats agree only approximately, see
// [CloseValues]. Integers, strings and the control flow agree exactly.
package gostrict
