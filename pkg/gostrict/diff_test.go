// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict_test

import (
	"context"
	"fmt"
	"math"
	"strings"
	"sync"
	"testing"

	"go.wdy.de/nago/pkg/gostrict"
	"go.wdy.de/nago/pkg/gostrict/engine/gojaengine"
	"go.wdy.de/nago/pkg/gostrict/engine/native"
	"go.wdy.de/nago/pkg/gostrict/engine/vm"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/alias"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/arena"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/binary"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/core"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/draw"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/errs"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/funcs"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/generics"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/handlevalues"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/host"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/hostuse"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/ints"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/iters"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/jsons"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/longs"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/money"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/orders"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/physics"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/shapes"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/spin"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/stdvalues"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/texts"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/tparams"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/trees"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/values"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/viewhost"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/widgets"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/widgetsuse"
	"go.wdy.de/nago/pkg/gostrict/std/big"
)

type diffCase struct {
	pkg  string
	name string
	args []any
}

// jsonDoc has every type, which json.Decode reads, unknown keys, a duplicate key and a key of a field without JSON.
const jsonDoc = `{"name":"Grüße ä😀","id":-9223372036854775808,"max":18446744073709551615,"Count":-7,
	"mask":4294967295,"ratio":-0,"ok":true,"tags":["a","b"],"points":[{"x":1.5,"y":-2},{"x":0,"y":1e-3,"extra":[1,{"z":null}]}],
	"frozen":[1,2,3],"index":{"b":2,"a":1,"b":3},"note":"hello","shape":{"x":3,"y":4},"data":"AAEC/w==",
	"big":-123456789012345678901234567890,
	"tree":{"name":"root","children":[{"name":"a","children":[{"name":"a1"}]},{"name":"b","children":null}]},
	"Skip":"no","-":"no","hidden":[1],"unknown":{"deep":[true,false,null,"x",-1.5e3]}}`

// jsonErrors are documents, which json.Decode refuses with the same message at the same offset in every engine.
var jsonErrors = []string{
	``, `{`, `{"id":1.5}`, `{"id":"1"}`, `{"max":-1}`, `{"id":9223372036854775808}`, `{"ratio":1e999}`, `{"name":1}`,
	"{\"name\":\"a\x01\"}", `{"name":"\x"}`, `{"name":"\ud83d\u12"}`, `{"tags":["a",1]}`, `{"tags":["a",]}`, `{"a":1,}`,
	`{"ok":tru}`, `{"data":"!!!!"}`, `{} x`, `{"id":01}`, `{"id":-}`, `{"ratio":1.}`, `{"ratio":1e}`, `{"tags":{}}`,
	`{"index":[]}`, `{"points":[{"x":"1"}]}`, `{"Count":2147483648}`, `{"mask":4294967296}`, `[1]`, `"x"`, `{"a" 1}`,
	`{1:2}`, `{"name":"abc`, "{\"name\":\"\xff\"}", "\xef\xbb\xbf{}", `{"big":1.5}`, `{"note":5}`, `{"shape":[]}`,
	`{"x":` + strings.Repeat("[", 600) + `}`,
}

// goEngine is shared by the tests, so that they use its cache of compiled programs.
var goEngine = &gojaengine.Engine{}

// exports are the exported functions of the examples for the native engine.
var exports = map[string]gostrict.Funcs{
	"alias":        native.Reflect(map[string]any{"Order": alias.Order, "Run": alias.Run}),
	"arena":        native.Reflect(map[string]any{"Frame": arena.Frame, "NewClient": arena.NewClient, "OnHit": arena.OnHit, "OnWorld": arena.OnWorld, "Status": arena.Status}),
	"core":         native.Reflect(map[string]any{"Control": core.Control, "Equal": core.Equal, "Lists": core.Lists, "Maps": core.Maps, "Options": core.Options, "Pairs": core.Pairs, "Ranges": core.Ranges, "Results": core.Results}),
	"draw":         native.Reflect(map[string]any{"Frame": draw.Frame, "New": draw.New}),
	"errs":         native.Reflect(map[string]any{"Average": errs.Average, "Run": errs.Run}),
	"funcs":        native.Reflect(map[string]any{"Branches": funcs.Branches, "Captures": funcs.Captures, "Inlined": funcs.Inlined, "Methods": funcs.Methods}),
	"generics":     native.Reflect(map[string]any{"Run": generics.Run, "SumFloats": generics.SumFloats}),
	"handlevalues": native.Reflect(map[string]any{"ImageFrame": handlevalues.ImageFrame, "PassFrame": handlevalues.PassFrame, "TextFrame": handlevalues.TextFrame, "WithFrame": handlevalues.WithFrame}),
	"hostuse":      native.Reflect(map[string]any{"Callbacks": hostuse.Callbacks, "Records": hostuse.Records, "Run": hostuse.Run}),
	"ints":         native.Reflect(map[string]any{"Bits": ints.Bits, "Conv": ints.Conv, "Div": ints.Div, "DivU": ints.DivU, "Hash": ints.Hash, "Wrap": ints.Wrap, "Xorshift": ints.Xorshift}),
	"binary":       native.Reflect(map[string]any{"Bits": binary.Bits, "Decode": binary.Decode, "Encode": binary.Encode, "Floats": binary.Floats, "Invalid": binary.Invalid, "LittleBits": binary.LittleBits, "Grow": binary.Grow, "Offsets": binary.Offsets, "Overrun": binary.Overrun, "Records": binary.Records, "Text": binary.Text}),
	"jsons":        native.Reflect(map[string]any{"Containers": jsons.Containers, "Encoded": jsons.Encoded, "Floats": jsons.Floats, "Nodes": jsons.Nodes, "Numbers": jsons.Numbers, "Points": jsons.Points, "Raw": jsons.Raw, "Scalars": jsons.Scalars, "Small": jsons.Small, "Sum": jsons.Sum, "Text": jsons.Text, "Values": jsons.Values, "Written": jsons.Written}),
	"longs":        native.Reflect(map[string]any{"Bits": longs.Bits, "Containers": longs.Containers, "Conv": longs.Conv, "Div": longs.Div, "DivU": longs.DivU, "Floats": longs.Floats, "MinMax": longs.MinMax, "Parse": longs.Parse, "Ranges": longs.Ranges, "Shifts": longs.Shifts, "Text": longs.Text, "Wrap": longs.Wrap}),
	"iters":        native.Reflect(map[string]any{"FirstOver": iters.FirstOver, "Live": iters.Live, "Nested": iters.Nested, "Sum": iters.Sum, "Switch": iters.Switch}),
	"money":        native.Reflect(map[string]any{"Calc": money.Calc, "Factorial": money.Factorial, "Ints": money.Ints, "Invoice": money.Invoice, "Limits": money.Limits, "Parse": money.Parse, "Round": money.Round}),
	"orders":       native.Reflect(map[string]any{"NewGame": orders.NewGame, "Snapshot": orders.Snapshot, "Step": orders.Step, "Summarize": orders.Summarize}),
	"physics":      native.Reflect(map[string]any{"Simulate": physics.Simulate, "Speed": physics.Speed}),
	"shapes":       native.Reflect(map[string]any{"Area": shapes.Area, "ColorName": shapes.ColorName, "Largest": shapes.Largest, "NewCanvas": shapes.NewCanvas, "OnEvent": shapes.OnEvent, "Scale": shapes.Scale, "Summary": shapes.Summary}),
	"spin":         native.Reflect(map[string]any{"Frame": spin.Frame, "Orbit": spin.Orbit, "Setup": spin.Setup}),
	"stdvalues":    native.Reflect(map[string]any{"DictAll": stdvalues.DictAll, "DictGet": stdvalues.DictGet, "DictIter": stdvalues.DictIter, "DictKeys": stdvalues.DictKeys, "FrozenValues": stdvalues.FrozenValues, "IterAll": stdvalues.IterAll, "IterValues": stdvalues.IterValues, "ListAll": stdvalues.ListAll, "ListValues": stdvalues.ListValues, "DictNew": stdvalues.DictNew, "Freeze": stdvalues.Freeze, "ListLen": stdvalues.ListLen, "ListMake": stdvalues.ListMake, "ListNew": stdvalues.ListNew, "ListOf": stdvalues.ListOf, "OptArg": stdvalues.OptArg, "OptInfer": stdvalues.OptInfer, "OptNone": stdvalues.OptNone, "OptOr": stdvalues.OptOr, "OptSome": stdvalues.OptSome, "SortFunc": stdvalues.SortFunc, "SortInfer": stdvalues.SortInfer, "SortVar": stdvalues.SortVar, "TextRepeat": stdvalues.TextRepeat}),
	"texts":        native.Reflect(map[string]any{"Misc": texts.Misc, "Order": texts.Order, "Parse": texts.Parse, "Reverse": texts.Reverse, "Stats": texts.Stats, "Words": texts.Words}),
	"tparams":      native.Reflect(map[string]any{"Args": tparams.Args, "Call": tparams.Call, "Closure": tparams.Closure, "Field": tparams.Field, "Pass": tparams.Pass, "Pointer": tparams.Pointer, "Two": tparams.Two, "Value": tparams.Value}),
	"trees":        native.Reflect(map[string]any{"Run": trees.Run, "Values": trees.Values}),
	"values":       native.Reflect(map[string]any{"Run": values.Run}),
	"viewhost":     native.Reflect(map[string]any{"Arg": viewhost.Arg, "Empty": viewhost.Empty, "Later": viewhost.Later, "Nested": viewhost.Nested, "Own": viewhost.Own, "Variadic": viewhost.Variadic}),
	"widgetsuse":   native.Reflect(map[string]any{"Run": widgetsuse.Run}),
}

// vmEngine is the interpreter, which the tests share.
var vmEngine = &vm.Engine{}

// vmPending are the examples, which the vm can not compile yet, with the reason.
var vmPending = map[string]string{}

// vmHosts are the functions of the header packages of the examples, which the vm needs, while goja uses their
// bindings.
var vmHosts = map[string]gostrict.Funcs{"hostuse": host.Host(), "widgetsuse": widgets.Host()}

// checkVM compares the result of the vm with the native one, unless the example is pending.
func checkVM(t *testing.T, b *gostrict.Bundle, pkg string, host gostrict.Funcs, name string, args []any, want []any) {
	t.Helper()
	if _, err := b.VM(); err != nil {
		if why, ok := vmPending[pkg]; ok {
			t.Logf("vm pending (%s): %v", why, err)
			return
		}
		t.Fatal(err)
	}
	if why, ok := vmPending[pkg]; ok {
		t.Errorf("%s is pending (%s), but compiles for the vm", pkg, why)
	}
	if host == nil {
		// the vm needs the functions of header packages in Go
		host = vmHosts[pkg]
	}
	got := call(t, vmEngine, b, host, name, args)
	if !gostrict.CloseValues(want, got, gojaengine.Tolerance) {
		t.Fatalf("native and vm differ\n  native: %#v\n  vm:     %#v", want, got)
	}
}

// nativeEngine returns the native engine of an example.
func nativeEngine(pkg string) native.Engine {
	return native.Engine{Exports: exports[pkg]}
}

// TestDifferential runs each case natively and translated in goja, both behind the same engine interface, and
// compares the results, floats up to the tolerance.
func TestDifferential(t *testing.T) {
	cases := []diffCase{
		{"funcs", "Captures", []any{int32(4)}},
		{"funcs", "Inlined", []any{int32(3)}},
		{"funcs", "Branches", []any{int32(3)}},
		{"funcs", "Branches", []any{int32(6)}},
		{"funcs", "Methods", []any{int32(2)}},
		{"funcs", "Methods", []any{int32(6)}},
		{"core", "Pairs", []any{int32(21)}},
		{"core", "Ranges", []any{int32(10)}},
		{"core", "Equal", []any{1.5, 2.5}},
		{"core", "Equal", []any{2.0, 2.0}},
		{"core", "Equal", []any{math.NaN(), 1.0}},
		{"core", "Equal", []any{0.0, math.Copysign(0, -1)}},
		{"core", "Results", []any{int32(47)}},
		{"core", "Results", []any{int32(-123)}},
		{"core", "Lists", []any{int32(6)}},
		{"core", "Lists", []any{int32(12)}},
		{"core", "Maps", []any{int32(10)}},
		{"core", "Maps", []any{int32(0)}},
		{"core", "Options", []any{int32(5)}},
		{"core", "Options", []any{int32(0)}},
		{"core", "Control", []any{int32(20)}},
		{"generics", "Run", nil},
		{"generics", "SumFloats", []any{0.1, 0.2, 0.3}},
		{"physics", "Simulate", []any{int32(600), int32(0)}},
		{"physics", "Simulate", []any{int32(600), int32(45)}},
		{"physics", "Speed", []any{3.0, -4.0}},
		{"texts", "Reverse", []any{"Grüße 😀!"}},
		{"texts", "Stats", []any{"Möhre 😀 ö"}},
		{"texts", "Words", []any{"  der Hund und der Hase und der  Igel "}},
		{"texts", "Order", []any{"\uFFFD", "😀"}},
		{"texts", "Order", []any{"abc", "abd"}},
		{"texts", "Parse", []any{"-2147483648"}},
		{"texts", "Parse", []any{"2147483648"}},
		{"texts", "Parse", []any{"12x"}},
		{"texts", "Parse", []any{"1073741824"}},
		{"texts", "Misc", []any{"\u0085 ab😀a \uFEFF"}},
		{"jsons", "Scalars", []any{jsonDoc}},
		{"jsons", "Containers", []any{jsonDoc}},
		{"jsons", "Values", []any{jsonDoc}},
		{"jsons", "Nodes", []any{jsonDoc}},
		{"jsons", "Scalars", []any{`{}`}},
		{"jsons", "Containers", []any{`null`}},
		{"jsons", "Values", []any{` {} `}},
		{"jsons", "Nodes", []any{`{"tree":null}`}},
		{"jsons", "Text", []any{`"aä😀\ud800x\udc00\"\\\/\b\f\n\r\t"`}},
		{"jsons", "Raw", []any{"2261fffe5c6eff22"}},
		{"jsons", "Raw", []any{"22e282ac2061e28222"}},
		{"jsons", "Numbers", []any{`[0, -0, 9223372036854775807, -9223372036854775808]`}},
		{"jsons", "Numbers", []any{`[9223372036854775808]`}},
		{"jsons", "Floats", []any{`[0, -0, 1.5e300, 4.9e-324, 1e-400, 0.1, 123456789012345678901234567890, -1E+2]`}},
		{"jsons", "Floats", []any{`[1e309]`}},
		{"jsons", "Small", []any{`[2147483647,-2147483648]`, `[4294967295,0,-0]`}},
		{"jsons", "Small", []any{`[2147483648]`, `[-1]`}},
		{"jsons", "Sum", []any{`[{"x":1,"y":2},{"y":0.5},null]`}},
		{"jsons", "Encoded", []any{jsonDoc}},
		{"jsons", "Encoded", []any{`{}`}},
		{"jsons", "Written", []any{1.5, "a\"\\\n\x01\x7fä😀\u2028"}},
		{"jsons", "Written", []any{0.0, ""}},
		{"jsons", "Written", []any{math.Copysign(0, -1), "x"}},
		{"jsons", "Written", []any{5e-324, "x"}},
		{"jsons", "Written", []any{1.2345678901234567e-7, "x"}},
		{"jsons", "Written", []any{math.Pi * 1e20, "x"}},
		{"jsons", "Written", []any{math.MaxFloat64, "x"}},
		{"jsons", "Written", []any{math.NaN(), "x"}},
		{"jsons", "Written", []any{math.Inf(-1), "x"}},
		{"jsons", "Points", []any{int32(3)}},
		{"binary", "Encode", []any{uint32(7), int64(-42), 1.5, "Grüße", int32(5)}},
		{"binary", "Decode", []any{"", int64(0), int32(0)}},
		{"binary", "Decode", []any{"xx", int64(0), int32(0)}},
		{"binary", "Decode", []any{"00112233", int64(0), int32(0)}},
		{"binary", "Decode", []any{"574b444108000701a0ffffffffffffd63fc00000000000000000f83f4772c3bcc39f6500f9cfebfa1f90000000000000", int64(8), int32(5)}},
		{"binary", "Text", []any{"Grüße 😀!"}},
		{"binary", "Text", []any{""}},
		{"binary", "Invalid", nil},
		{"binary", "Bits", []any{uint32(0xDEADBEEF), uint64(0xDEADBEEFCAFEBABE), uint32(7)}},
		{"binary", "Bits", []any{uint32(0), uint64(0), uint32(0)}},
		{"binary", "Bits", []any{uint32(1), uint64(1 << 40), uint32(100)}},
		{"binary", "Floats", []any{1.5, uint32(0xC0490FDB), uint32(12)}},
		{"binary", "Floats", []any{math.Inf(-1), uint32(0xFFF), uint32(12)}},
		{"binary", "Floats", []any{-0.1, uint32(0x7F), uint32(0)}},
		{"binary", "Floats", []any{1e300, uint32(0x80000000), uint32(40)}},
		{"binary", "Records", []any{int32(1000)}},
		{"binary", "Overrun", []any{int64(0)}},
		{"binary", "Overrun", []any{int64(2)}},
		{"binary", "Overrun", []any{int64(-1)}},
		{"binary", "Overrun", []any{int64(math.MaxInt64)}},
		{"binary", "Grow", []any{int64(0)}},
		{"binary", "Grow", []any{int64(1)}},
		{"binary", "Grow", []any{int64(3000)}},
		{"binary", "Offsets", []any{int64(2), int64(5), int64(3)}},
		{"binary", "Offsets", []any{int64(-4), int64(100), int64(9)}},
		{"binary", "Offsets", []any{int64(7), int64(3), int64(10)}},
		{"binary", "Offsets", []any{int64(0), int64(4), int64(-1)}},
		{"binary", "Offsets", []any{int64(math.MinInt64), int64(math.MaxInt64), int64(math.MaxInt64)}},
		{"binary", "Offsets", []any{int64(1) << 53, int64(1)<<53 + 1, int64(1)<<53 + 1}},
		{"binary", "LittleBits", []any{uint64(0xABCDEF), uint32(21)}},
		{"binary", "LittleBits", []any{uint64(0xFFFFFFFFFFFFFFFF), uint32(64)}},
		{"longs", "Wrap", []any{int64(math.MaxInt64), int64(7)}},
		{"longs", "Wrap", []any{int64(math.MinInt64), int64(-1)}},
		{"longs", "Div", []any{int64(-7), int64(2)}},
		{"longs", "Div", []any{int64(math.MinInt64), int64(-1)}},
		{"longs", "Div", []any{int64(5), int64(0)}},
		{"longs", "DivU", []any{uint64(math.MaxUint64), uint64(7)}},
		{"longs", "Bits", []any{uint64(0xDEADBEEFCAFEBABE)}},
		{"longs", "Shifts", []any{int32(-5), uint32(0xF0000001), int64(-5), uint64(0xF000000000000001), uint32(4)}},
		{"longs", "Shifts", []any{int32(-5), uint32(0xF0000001), int64(-5), uint64(0xF000000000000001), uint32(31)}},
		{"longs", "Shifts", []any{int32(-5), uint32(0xF0000001), int64(-5), uint64(0xF000000000000001), uint32(32)}},
		{"longs", "Shifts", []any{int32(-5), uint32(0xF0000001), int64(-5), uint64(0xF000000000000001), uint32(63)}},
		{"longs", "Shifts", []any{int32(-5), uint32(0xF0000001), int64(-5), uint64(0xF000000000000001), uint32(64)}},
		{"longs", "Shifts", []any{int32(-5), uint32(0xF0000001), int64(-5), uint64(0xF000000000000001), uint32(4000000000)}},
		{"longs", "Conv", []any{int32(-1), uint32(math.MaxUint32), int64(-1), uint64(math.MaxUint64)}},
		{"longs", "Conv", []any{int32(math.MinInt32), uint32(1 << 31), int64(1 << 40), uint64(1 << 63)}},
		{"longs", "Floats", []any{-2.5}},
		{"longs", "Floats", []any{math.NaN()}},
		{"longs", "Floats", []any{1e30}},
		{"longs", "Floats", []any{math.Inf(-1)}},
		{"longs", "Floats", []any{1e15 + 0.5}},
		{"longs", "Text", []any{int64(math.MinInt64), uint64(math.MaxUint64)}},
		{"longs", "Parse", []any{"-9223372036854775808"}},
		{"longs", "Parse", []any{"9223372036854775808"}},
		{"longs", "Parse", []any{"18446744073709551615"}},
		{"longs", "Parse", []any{"18446744073709551616"}},
		{"longs", "Parse", []any{"+12"}},
		{"longs", "Parse", []any{"1x"}},
		{"longs", "Ranges", []any{int64(5), uint64(4)}},
		{"longs", "Containers", []any{int64(1 << 53), uint64(1<<64 - 1)}},
		{"longs", "MinMax", []any{int64(-3), int64(9), uint64(1 << 63), uint64(4)}},
		{"ints", "Hash", []any{"Grüße 😀"}},
		{"ints", "Wrap", []any{int32(math.MaxInt32), int32(7)}},
		{"ints", "Wrap", []any{int32(math.MinInt32), int32(-1)}},
		{"ints", "Div", []any{int32(-7), int32(2)}},
		{"ints", "Div", []any{int32(math.MinInt32), int32(-1)}},
		{"ints", "Div", []any{int32(5), int32(0)}},
		{"ints", "DivU", []any{uint32(4000000000), uint32(7)}},
		{"ints", "Bits", []any{uint32(0xDEADBEEF)}},
		{"ints", "Conv", []any{-2.5}},
		{"ints", "Conv", []any{math.NaN()}},
		{"ints", "Conv", []any{1e12}},
		{"ints", "Conv", []any{math.Inf(-1)}},
		{"ints", "Xorshift", []any{uint32(2463534242), int32(1000)}},
		{"values", "Run", nil},
		{"iters", "FirstOver", []any{int32(20), int32(7)}},
		{"iters", "FirstOver", []any{int32(4), int32(7)}},
		{"iters", "Sum", []any{int32(30)}},
		{"iters", "Nested", []any{int32(10)}},
		{"iters", "Switch", []any{"hallo-welt-xyz"}},
		{"iters", "Live", nil},
		{"errs", "Run", []any{int32(40)}},
		{"errs", "Run", []any{int32(0)}},
		{"errs", "Run", []any{int32(-3)}},
		{"trees", "Run", []any{int32(0)}},
		{"trees", "Run", []any{int32(200)}},
		{"trees", "Values", nil},
		{"hostuse", "Callbacks", []any{int32(5), 100.0}},
		{"hostuse", "Callbacks", []any{int32(5), 3.0}},
		{"alias", "Run", nil},
		{"alias", "Order", nil},
		{"money", "Ints", []any{"123456789012345678901234567890", "-987654321"}},
		{"money", "Ints", []any{"-7", "2"}},
		{"money", "Ints", []any{"5", "0"}},
		{"money", "Factorial", []any{int32(0)}},
		{"money", "Factorial", []any{int32(40)}},
		{"money", "Parse", []any{"-0012.3400"}},
		{"money", "Parse", []any{"+7"}},
		{"money", "Parse", []any{"1e3"}},
		{"money", "Parse", []any{".5"}},
		{"money", "Parse", []any{"5."}},
		{"money", "Parse", []any{"--1"}},
		{"money", "Parse", []any{"1.2.3"}},
		{"money", "Parse", []any{" 1"}},
		{"money", "Parse", []any{""}},
		{"money", "Calc", []any{"10", "3", int32(4), int32(0)}},
		{"money", "Calc", []any{"-10", "3", int32(2), int32(1)}},
		{"money", "Calc", []any{"0.1", "0.2", int32(10), int32(2)}},
		{"money", "Calc", []any{"1.005", "0.001", int32(0), int32(3)}},
		{"money", "Calc", []any{"1", "0", int32(2), int32(0)}},
		{"money", "Calc", []any{"1.00", "1", int32(-3), int32(0)}},
		{"money", "Round", []any{[]any{"2.5", "-2.5", "3.5", "-3.5", "2.51", "-2.49", "0.004", "-0.005", "12.345", "7"}, int32(0)}},
		{"money", "Round", []any{[]any{"12.345", "-12.345", "12.355", "0.005", "x"}, int32(2)}},
		{"money", "Invoice", []any{[]any{"19.99", "0.333", "1234567890.12"}, "19"}},
		{"money", "Invoice", []any{[]any{"0.10", "0.10", "0.10"}, "7"}},
		{"money", "Limits", []any{"-2147483648", uint32(4294967295), int32(-5)}},
		{"money", "Limits", []any{"2147483648", uint32(0), int32(30)}},
		{"money", "Limits", []any{"9007199254740993", uint32(7), int32(5000)}},
		{"money", "Limits", []any{"-" + strings.Repeat("9", 400), uint32(1), int32(0)}},
		{"money", "Limits", []any{"18446744073709553664", uint32(1), int32(0)}},
		{"money", "Limits", []any{"-" + strings.Repeat("7", 300), uint32(1), int32(0)}},
	}
	for _, src := range jsonErrors {
		cases = append(cases, diffCase{"jsons", "Scalars", []any{src}})
	}

	ctx := context.Background()
	for _, tc := range cases {
		t.Run(fmt.Sprintf("%s.%s%v", tc.pkg, tc.name, tc.args), func(t *testing.T) {
			b := compile(t, "internal/examples/"+tc.pkg)
			want := call(t, native.Engine{Exports: exports[tc.pkg]}, b, nil, tc.name, tc.args)
			got := call(t, goEngine, b, nil, tc.name, tc.args)
			if !gostrict.CloseValues(want, got, gojaengine.Tolerance) {
				t.Fatalf("the engines differ\n  native: %#v\n  goja:   %#v", want, got)
			}
			checkVM(t, b, tc.pkg, nil, tc.name, tc.args, want)
			t.Logf("%v", want)
		})
	}
	_ = ctx
}

// TestHost calls the native functions of a header package in goja, checks, that they agree with their bindings,
// and runs the program with the bindings.
func TestHost(t *testing.T) {
	b := compile(t, "internal/examples/hostuse")

	calls := 0
	h := gostrict.Funcs{}
	for name, f := range host.Host() {
		h[name] = func(args []any) []any {
			calls++
			return f(args)
		}
	}

	for _, args := range [][]any{{3.0, 4.0}, {3.0, 0.0}, {200.0, math.Copysign(0, -1)}} {
		want, _ := hostuse.Run(args[0].(float64), args[1].(float64))
		viaHost := call(t, goEngine, b, h, "Run", args)
		viaJS := call(t, goEngine, b, nil, "Run", args)
		if viaHost[0] != want || !gostrict.SameValues(viaHost, viaJS) {
			t.Fatalf("%v: host %v, bindings %v, want %v", args, viaHost, viaJS, want)
		}
	}
	for _, d := range []float64{0, 1.5, -2} {
		w1, w2, w3 := hostuse.Records(d)
		want := []any{w1, w2, w3}
		viaHost := call(t, goEngine, b, h, "Records", []any{d})
		viaJS := call(t, goEngine, b, nil, "Records", []any{d})
		if !gostrict.SameValues(viaHost, want) || !gostrict.SameValues(viaJS, want) {
			t.Fatalf("records %v: host %v, bindings %v, want %v", d, viaHost, viaJS, want)
		}
	}
	if calls == 0 {
		t.Fatal("the host functions were not called")
	}

	// the callbacks are called during the call and the kept one later, between two calls
	for _, e := range []gostrict.Engine{goEngine, vmEngine} {
		testKept(t, e, b, h)
	}
}

// testKept checks the callbacks, which a host function keeps, in an engine: they are called during the call and
// later, between two calls, and report an error once.
func testKept(t *testing.T, e gostrict.Engine, b *gostrict.Bundle, h gostrict.Funcs) {
	var kept gostrict.Callback
	keep := "go.wdy.de/nago/pkg/gostrict/internal/examples/host.Keep"
	withKeep := gostrict.Funcs{}
	for name, f := range h {
		withKeep[name] = f
	}
	withKeep[keep] = func(args []any) []any {
		kept = args[0].(gostrict.Callback)
		return h[keep](args)
	}
	in, err := e.Load(context.Background(), b, withKeep)
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	for _, limit := range []float64{100, 5} {
		w1, w2, w3 := hostuse.Callbacks(6, limit)
		rects, err := gostrict.ToNeutral(w1)
		if err != nil {
			t.Fatal(err)
		}
		want := []any{rects, w2, w3}
		viaHost, err := in.Call(context.Background(), "Callbacks", int32(6), limit)
		if err != nil {
			t.Fatal(err)
		}
		viaJS := call(t, goEngine, b, nil, "Callbacks", []any{int32(6), limit})
		if !gostrict.SameValues(viaHost, want) || !gostrict.SameValues(viaJS, want) {
			t.Fatalf("callbacks %v: host %v, bindings %v, want %v", limit, viaHost, viaJS, want)
		}
	}
	if got, err := kept.Call(1.5); err != nil || got[0] != 9.0 {
		t.Fatalf("the kept callback returned %v, %v", got, err)
	}

	// an error of a callback is reported by the next call
	if _, err := kept.Call("x"); err == nil {
		t.Fatal("a callback took a wrong argument")
	}
	if _, err := in.Call(context.Background(), "Callbacks", int32(1), 1.0); err == nil || !strings.Contains(err.Error(), "callback") {
		t.Fatalf("the error of the callback was not reported: %v", err)
	}
	if _, err := in.Call(context.Background(), "Callbacks", int32(1), 1.0); err != nil {
		t.Fatalf("the error of the callback was reported twice: %v", err)
	}

}

// TestHostBindings checks the bindings of the example host.
func TestHostBindings(t *testing.T) {
	b := compile(t, "internal/examples/hostuse")
	mismatches, err := gojaengine.CheckBindings(context.Background(), b, host.Host(), 200)
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range mismatches {
		t.Error(m)
	}
}

// TestWidgets builds views of a header package like lib/ui natively, with the bindings and with the host in Go: a
// sealed interface of handles, builder methods, a record with a method, an enum, a variable and a callback.
func TestWidgets(t *testing.T) {
	b := compile(t, "internal/examples/widgetsuse")
	for _, n := range []int32{1, 5} {
		want := call(t, nativeEngine("widgetsuse"), b, nil, "Run", []any{n})
		viaJS := call(t, goEngine, b, nil, "Run", []any{n})
		viaHost := call(t, goEngine, b, widgets.Host(), "Run", []any{n})
		viaVM := call(t, vmEngine, b, widgets.Host(), "Run", []any{n})
		if !gostrict.SameValues(want, viaJS) || !gostrict.SameValues(want, viaHost) || !gostrict.SameValues(want, viaVM) {
			t.Fatalf("the views differ\n  native: %v\n  js:     %v\n  host:   %v\n  vm:     %v", want, viaJS, viaHost, viaVM)
		}
		t.Log(want)
	}
}

// TestBigBindings checks the bindings of std/big, which the money example uses, against math/big.
func TestBigBindings(t *testing.T) {
	b := compile(t, "internal/examples/money")
	mismatches, err := gojaengine.CheckBindings(context.Background(), b, big.Host(), 500)
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range mismatches {
		t.Error(m)
	}
	checked := 0
	for _, x := range b.Externs {
		if _, ok := big.Host()[x.Name]; ok {
			checked++
		}
	}
	if checked < 10 {
		t.Fatalf("only %d functions of std/big were checked", checked)
	}
}

// TestBrokenBinding expects the check of the bindings to find, that Math.round is not math.Round.
func TestBrokenBinding(t *testing.T) {
	b := compile(t, "testdata/headers/brokenuse")
	mismatches, err := gojaengine.CheckBindings(context.Background(), b, gostrict.Funcs{
		"go.wdy.de/nago/pkg/gostrict/testdata/headers/broken.Round": func(args []any) []any {
			return []any{math.Round(args[0].(float64))}
		},
	}, 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(mismatches) == 0 {
		t.Fatal("the broken binding was not found")
	}
	t.Logf("found %d mismatches, e.g. %s", len(mismatches), mismatches[0])
}

func call(t *testing.T, e gostrict.Engine, b *gostrict.Bundle, host gostrict.Funcs, name string, args []any) []any {
	t.Helper()
	ctx := context.Background()
	in, err := e.Load(ctx, b, host)
	if err != nil {
		t.Fatalf("%v\n%s", err, numbered(b.JS))
	}
	defer in.Close()
	res, err := in.Call(ctx, name, args...)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

var (
	bundlesMu sync.Mutex
	bundles   = map[string]*gostrict.Bundle{}
)

func compile(t *testing.T, dir string) *gostrict.Bundle {
	t.Helper()
	bundlesMu.Lock()
	defer bundlesMu.Unlock()
	if b, ok := bundles[dir]; ok {
		return b
	}

	b, diags, err := gostrict.Compile(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(diags) > 0 {
		t.Fatalf("%s is not in the subset:\n%s", dir, diags.Text())
	}
	bundles[dir] = b
	return b
}

func numbered(js string) string {
	var sb strings.Builder
	for i, l := range strings.Split(js, "\n") {
		fmt.Fprintf(&sb, "%4d %s\n", i+1, l)
	}
	return sb.String()
}
