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
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"go.wdy.de/nago/pkg/gostrict"
	"go.wdy.de/nago/pkg/gostrict/lib/ui"
)

// The tests of this file check, that a program, which the checker accepts, loads and runs natively, in goja and in
// the vm alike, or that the checker rejects it with a finding of its own, never the translation with GS099. They
// cover three constructs, which loaded in one engine and not in the other or passed the checker and failed in the
// translation: a view of lib/ui, which crosses to the host as a ui.View without an exported function with the
// result ui.View (goja), the method of a type parameter (vm) and a function of a standard package as a value (GS099).
// Each program is a file of an example package together with its doc.go, so that it runs natively as the Go code of
// the package, next to a counter row, which differs only in the construct and ran in both engines before.

// gapProgram returns the module llm.local/<dir> of the files of the example package in internal/examples/<dir>.
// A build constraint //go:build !gostrict is removed, edits replace an old text, which must occur exactly once in
// the files, by a new one, and extra adds files.
func gapProgram(t *testing.T, dir string, files []string, edits [][2]string, extra map[string]string) (string, gostrict.Module) {
	t.Helper()
	fsys := fstest.MapFS{}
	used := make([]int, len(edits))
	for _, name := range files {
		buf, err := os.ReadFile(filepath.Join("internal/examples", dir, name))
		if err != nil {
			t.Fatal(err)
		}
		src := strings.ReplaceAll(string(buf), "//go:build !gostrict\n", "")
		for i, e := range edits {
			used[i] += strings.Count(src, e[0])
			src = strings.ReplaceAll(src, e[0], e[1])
		}
		fsys[name] = &fstest.MapFile{Data: []byte(src)}
	}
	for i, n := range used {
		if n != 1 {
			t.Fatalf("the edit %q occurs %d times in %v", edits[i][0], n, files)
		}
	}
	for name, src := range extra {
		fsys[name] = &fstest.MapFile{Data: []byte(src)}
	}
	path := "llm.local/" + dir
	return path, gostrict.Module{Path: path, FS: fsys}
}

// gapCompile compiles the module and fails, if there is any diagnostic.
func gapCompile(t *testing.T, path string, m gostrict.Module) *gostrict.Bundle {
	t.Helper()
	b, diags, err := gostrict.NewCache(1).Compile(path, m)
	if err != nil {
		t.Fatal(err)
	}
	if len(diags) > 0 {
		t.Fatalf("the program does not compile:\n%s", diags.Text())
	}
	return b
}

// gapEngines are goja and the vm, the engines besides the native one.
var gapEngines = []struct {
	name string
	e    gostrict.Engine
}{{"goja", goEngine}, {"vm", vmEngine}}

// gapRun loads the bundle in the engine and calls the function. after runs before the instance is closed.
func gapRun(e gostrict.Engine, b *gostrict.Bundle, host gostrict.Funcs, fn string, args []any, after func()) ([]any, error) {
	ctx := context.Background()
	in, err := e.Load(ctx, b, host)
	if err != nil {
		return nil, fmt.Errorf("load: %w", err)
	}
	defer in.Close()
	res, err := in.Call(ctx, fn, args...)
	if err != nil {
		return nil, fmt.Errorf("call: %w", err)
	}
	if after != nil {
		after()
	}
	return res, nil
}

// ---- views of lib/ui as a ui.View to the host ----

// viewRecorder wraps the functions of lib/ui and records the types of the views, which cross to the host as a
// ui.View, a variant, also in a list. It keeps the functions, which TabsPage and SizeClass take, and calls them
// later, like a window, which renders the page, and records the type of the view, which they return.
type viewRecorder struct {
	crossed []string
	kept    []gostrict.Callback
}

func (r *viewRecorder) walk(v any) {
	switch v := v.(type) {
	case gostrict.Variant:
		r.crossed = append(r.crossed, strings.TrimPrefix(v.Tag, recModule+"lib/ui."))
		for _, f := range v.Fields {
			r.walk(f)
		}
	case []any:
		for _, x := range v {
			r.walk(x)
		}
	}
}

func (r *viewRecorder) host() gostrict.Funcs {
	h := gostrict.Funcs{}
	for name, f := range ui.Host() {
		later := name == recModule+"lib/ui.TabsPage" || name == recModule+"lib/ui.SizeClass"
		h[name] = func(args []any) []any {
			r.walk(args)
			if later {
				r.kept = append(r.kept, args[1].(gostrict.Callback))
			}
			return f(args)
		}
	}
	return h
}

// callLater calls the kept functions and returns the types of the views, which they return.
func (r *viewRecorder) callLater() ([]string, error) {
	var res []string
	for _, cb := range r.kept {
		out, err := cb.Call()
		if err != nil {
			return nil, err
		}
		v, ok := out[0].(gostrict.Variant)
		if !ok {
			return nil, fmt.Errorf("the function returned %T, not a ui.View", out[0])
		}
		res = append(res, strings.TrimPrefix(v.Tag, recModule+"lib/ui."))
	}
	return res, nil
}

// viewHostCases are the programs of internal/examples/viewhost with the types of the views, which cross to the host
// as a ui.View, during the call and from the functions, which the host calls later.
var viewHostCases = []struct {
	file, fn       string
	crossed, later []string
}{
	{"own.go", "Own", nil, nil},
	{"arg.go", "Arg", []string{"TText"}, nil},
	{"variadic.go", "Variadic", []string{"TAvatar", "TDivider", "TImage", "TQrCode", "TSpacer", "TText"}, nil},
	{"nested.go", "Nested", []string{"TAlertBanner", "TCard", "TGrid", "TList", "TScrollView"}, nil},
	{"later.go", "Later", nil, []string{"TCodeEditor", "THero"}},
	{"empty.go", "Empty", nil, nil},
}

// viewHostV is the counter row: the same program with an exported function with the result ui.View, with which goja
// loaded such a program before.
const viewHostV = "package viewhost\n\nimport \"go.wdy.de/nago/pkg/gostrict/lib/ui\"\n\n// V returns a view.\nfunc V() ui.View { return ui.Text(\"v\") }\n"

// TestViewHost runs each program of internal/examples/viewhost natively, in goja and in the vm, without and with an
// exported function with the result ui.View: the results are the same, the views of the types of the row cross to
// the host as a ui.View in both engines, and the functions, which the host calls later, return their views.
func TestViewHost(t *testing.T) {
	for _, c := range viewHostCases {
		for _, counter := range []bool{false, true} {
			name := c.fn + " without an exported ui.View"
			extra := map[string]string{}
			if counter {
				name = c.fn + " counter: with an exported func V() ui.View"
				extra["v.go"] = viewHostV
			}
			t.Run(name, func(t *testing.T) {
				path, m := gapProgram(t, "viewhost", []string{"doc.go", c.file}, nil, extra)
				b := gapCompile(t, path, m)
				for _, dp := range []float64{3, 12} {
					args := []any{dp}
					want := call(t, nativeEngine("viewhost"), compile(t, "internal/examples/viewhost"), nil, c.fn, args)
					if s, _ := want[0].(string); !strings.HasSuffix(s, "dp;") {
						t.Fatalf("native: %s(%v) = %v, want the widths of the frames", c.fn, dp, want)
					}
					for _, e := range gapEngines {
						r := &viewRecorder{}
						var later []string
						var laterErr error
						got, err := gapRun(e.e, b, r.host(), c.fn, args, func() { later, laterErr = r.callLater() })
						if err != nil {
							t.Errorf("%s: %s(%v): %v", e.name, c.fn, dp, err)
							continue
						}
						if !gostrict.SameValues(got, want) {
							t.Errorf("%s: %s(%v) = %v, native %v", e.name, c.fn, dp, got, want)
						}
						slices.Sort(r.crossed)
						if !slices.Equal(r.crossed, c.crossed) {
							t.Errorf("%s: views crossed as a ui.View %v, want %v", e.name, r.crossed, c.crossed)
						}
						slices.Sort(later)
						if laterErr != nil || !slices.Equal(later, c.later) {
							t.Errorf("%s: the functions called later returned %v, %v, want %v", e.name, later, laterErr, c.later)
						}
					}
				}
			})
		}
	}
}

// TestViewHostWhole runs the whole package internal/examples/viewhost as one program, whose views of all types cross
// to the host, natively, in goja and in the vm.
func TestViewHostWhole(t *testing.T) {
	b := compile(t, "internal/examples/viewhost")
	if _, ok := b.Exports["V"]; ok || len(b.Exports) != len(viewHostCases) {
		t.Fatalf("the package exports %d functions, want the %d of the cases and no V", len(b.Exports), len(viewHostCases))
	}
	for _, c := range viewHostCases {
		args := []any{5.0}
		want := call(t, nativeEngine("viewhost"), b, nil, c.fn, args)
		for _, e := range gapEngines {
			got, err := gapRun(e.e, b, ui.Host(), c.fn, args, nil)
			if err != nil {
				t.Errorf("%s: %s: %v", e.name, c.fn, err)
				continue
			}
			if !gostrict.SameValues(got, want) {
				t.Errorf("%s: %s = %v, native %v", e.name, c.fn, got, want)
			}
		}
	}
}

// ---- methods of type parameters ----

// tparamCases are the programs of internal/examples/tparams. iface are the edits of the counter row, which replace
// the type parameter by its constraint, an interface, whose method value the vm loaded before.
var tparamCases = []struct {
	file, fn string
	iface    [][2]string
}{
	{"value.go", "Value", [][2]string{{"valueArea[S shape](s S)", "valueArea(s shape)"}}},
	{"pointer.go", "Pointer", [][2]string{{"pointerArea[S shape](s S)", "pointerArea(s shape)"}}},
	{"args.go", "Args", [][2]string{{"argsScale[S scaler](s S, k int32)", "argsScale(s scaler, k int32)"}}},
	{"pass.go", "Pass", [][2]string{{"passArea[S shape](s S)", "passArea(s shape)"}}},
	{"closure.go", "Closure", [][2]string{{"closureArea[S shape](s S)", "closureArea(s shape)"}}},
	{"field.go", "Field", [][2]string{
		{"type holder[S shape] struct{ s S }", "type holder struct{ s shape }"},
		{"func (h holder[S]) area()", "func (h holder) area()"},
		{"holder[sq]{", "holder{"},
	}},
	{"two.go", "Two", [][2]string{{"twoArea[S shape](s S)", "twoArea(s shape)"}}},
	{"call.go", "Call", [][2]string{{"callArea[S shape](s S)", "callArea(s shape)"}}},
}

// TestTypeParamMethods runs each program of internal/examples/tparams with the method of a type parameter natively,
// in goja and in the vm, next to its counter row with an interface instead of the type parameter: the results are
// the native ones.
func TestTypeParamMethods(t *testing.T) {
	for _, c := range tparamCases {
		for _, counter := range []bool{false, true} {
			name := c.fn + " through a type parameter"
			var edits [][2]string
			if counter {
				name = c.fn + " counter: through an interface"
				edits = c.iface
			}
			t.Run(name, func(t *testing.T) {
				path, m := gapProgram(t, "tparams", []string{"doc.go", c.file}, edits, nil)
				b := gapCompile(t, path, m)
				for _, n := range []int32{2, 5} {
					args := []any{n}
					want := call(t, nativeEngine("tparams"), compile(t, "internal/examples/tparams"), nil, c.fn, args)
					for _, e := range gapEngines {
						got, err := gapRun(e.e, b, nil, c.fn, args, nil)
						if err != nil {
							t.Errorf("%s: %s(%d): %v", e.name, c.fn, n, err)
							continue
						}
						if !gostrict.SameValues(got, want) {
							t.Errorf("%s: %s(%d) = %v, native %v", e.name, c.fn, n, got, want)
						}
					}
				}
			})
		}
	}
}

// TestTypeParamMethodsWhole runs the whole package internal/examples/tparams as one program natively, in goja and in
// the vm.
func TestTypeParamMethodsWhole(t *testing.T) {
	b := compile(t, "internal/examples/tparams")
	if len(b.Exports) != len(tparamCases) {
		t.Fatalf("the package exports %d functions, want %d", len(b.Exports), len(tparamCases))
	}
	for _, c := range tparamCases {
		args := []any{int32(4)}
		want := call(t, nativeEngine("tparams"), b, nil, c.fn, args)
		for _, e := range gapEngines {
			got, err := gapRun(e.e, b, nil, c.fn, args, nil)
			if err != nil {
				t.Errorf("%s: %s: %v", e.name, c.fn, err)
				continue
			}
			if !gostrict.SameValues(got, want) {
				t.Errorf("%s: %s = %v, native %v", e.name, c.fn, got, want)
			}
		}
	}
}

// ---- functions and methods of standard packages as values ----

// stdValueCases are the programs of internal/examples/stdvalues. value is the text of the line, which uses the
// function as a value, and direct are the edits of the counter row, which call the function directly instead.
var stdValueCases = []struct {
	file, fn, value string
	direct          [][2]string
}{
	{"sortfunc.go", "SortFunc", "s := list.SortFunc[int32]", [][2]string{{"s := list.SortFunc[int32]\n\ts(l,", "list.SortFunc(l,"}}},
	{"sortinfer.go", "SortInfer", "= list.SortFunc\n", [][2]string{{"var s func(list.List[int32], func(x, y int32) int32) = list.SortFunc\n\ts(l,", "list.SortFunc(l,"}}},
	{"sortvar.go", "SortVar", "var sorter = list.SortFunc[int32]", [][2]string{{"var sorter = list.SortFunc[int32]", "func sorter(l list.List[int32], cmp func(x, y int32) int32) { list.SortFunc(l, cmp) }"}}},
	{"listof.go", "ListOf", "f := list.Of[int32]", [][2]string{{"f := list.Of[int32]\n\tl := f(a, b, a)", "l := list.Of(a, b, a)"}}},
	{"listnew.go", "ListNew", "f := list.New[int32]", [][2]string{{"f := list.New[int32]\n\tl := f()", "l := list.New[int32]()"}}},
	{"listmake.go", "ListMake", "f := list.Make[int32]", [][2]string{{"f := list.Make[int32]\n\tl := f(a)", "l := list.Make[int32](a)"}}},
	{"freeze.go", "Freeze", "f := list.Freeze[int32]", [][2]string{{"f := list.Freeze[int32]\n\tl := f(list.Of(a, b))", "l := list.Freeze(list.Of(a, b))"}}},
	{"dictnew.go", "DictNew", "f := dict.New[string, int32]", [][2]string{{"f := dict.New[string, int32]\n\tm := f()", "m := dict.New[string, int32]()"}}},
	{"optsome.go", "OptSome", "f := opt.Some[int32]", [][2]string{{"f := opt.Some[int32]\n\treturn f(a)", "return opt.Some(a)"}}},
	{"optnone.go", "OptNone", "f := opt.None[int32]", [][2]string{{"f := opt.None[int32]\n\treturn f()", "return opt.None[int32]()"}}},
	{"optarg.go", "OptArg", "optArgCall(opt.Some[int32], a, b)", [][2]string{{"opt.Some[int32], a, b", "func(v int32) opt.Option[int32] { return opt.Some(v) }, a, b"}}},
	{"optinfer.go", "OptInfer", "optInferCall(opt.Some, a, b)", [][2]string{{"opt.Some, a, b", "func(v int32) opt.Option[int32] { return opt.Some(v) }, a, b"}}},
	{"textrepeat.go", "TextRepeat", "f := text.Repeat", [][2]string{{"f := text.Repeat\n\treturn text.RuneCount(f(", "return text.RuneCount(text.Repeat("}}},
	{"listlen.go", "ListLen", "f := l.Len", [][2]string{{"f := l.Len\n\treturn 10*f()", "return 10*l.Len()"}}},
	{"optor.go", "OptOr", "f := opt.Some(a).Or", [][2]string{{"f := opt.Some(a).Or\n\treturn f(b)", "return opt.Some(a).Or(b)"}}},
	{"dictget.go", "DictGet", "f := m.Get", [][2]string{{"f := m.Get\n\tv, ok := f(\"a\")", "v, ok := m.Get(\"a\")"}}},
}

// stdValueLine returns the line of the file, which contains the text, which must occur exactly once.
func stdValueLine(t *testing.T, m gostrict.Module, file, text string) int {
	t.Helper()
	src := string(m.FS.(fstest.MapFS)[file].Data)
	if strings.Count(src, text) != 1 {
		t.Fatalf("%q occurs %d times in %s", text, strings.Count(src, text), file)
	}
	return strings.Count(src[:strings.Index(src, text)], "\n") + 1
}

// stdValueRun runs the program of the row in goja and in the vm and compares the results with the native ones.
func stdValueRun(t *testing.T, b *gostrict.Bundle, fn string) {
	t.Helper()
	for _, args := range [][]any{{int32(2), int32(5)}, {int32(7), int32(1)}} {
		want := call(t, nativeEngine("stdvalues"), b, nil, fn, args)
		for _, e := range gapEngines {
			got, err := gapRun(e.e, b, nil, fn, args, nil)
			if err != nil {
				t.Errorf("%s: %s%v: %v", e.name, fn, args, err)
				continue
			}
			if !gostrict.SameValues(got, want) {
				t.Errorf("%s: %s%v = %v, native %v", e.name, fn, args, got, want)
			}
		}
	}
}

// TestStdValues checks each program of internal/examples/stdvalues, which uses a function or a method of a standard
// package as a value: if the checker accepts it, it translates without any diagnostic and runs natively, in goja and
// in the vm alike; if the checker rejects it, it does so with a finding of its own at the line of the value and
// never with GS099, and the translation adds nothing. The counter row calls the function directly: the checker
// accepts it and it runs, so that a rejection is one of the value.
func TestStdValues(t *testing.T) {
	for _, c := range stdValueCases {
		t.Run(c.fn, func(t *testing.T) {
			t.Run("counter: called directly", func(t *testing.T) {
				path, m := gapProgram(t, "stdvalues", []string{"doc.go", c.file}, c.direct, nil)
				prog, err := gostrict.LoadModules(path, m)
				if err != nil {
					t.Fatal(err)
				}
				if diags := gostrict.Check(prog); len(diags) > 0 {
					t.Fatalf("the checker rejects the direct call:\n%s", diags.Text())
				}
				stdValueRun(t, gapCompile(t, path, m), c.fn)
			})
			t.Run("as a value", func(t *testing.T) {
				path, m := gapProgram(t, "stdvalues", []string{"doc.go", c.file}, nil, nil)
				line := stdValueLine(t, m, c.file, c.value)
				prog, err := gostrict.LoadModules(path, m)
				if err != nil {
					t.Fatal(err)
				}
				checked := gostrict.Check(prog)
				b, diags, err := gostrict.NewCache(1).Compile(path, m)
				if err != nil {
					t.Fatal(err)
				}
				for _, d := range diags {
					if d.Code == gostrict.CodeInternal {
						t.Errorf("GS099 at %s:%d: %s", d.File, d.Line, d.Msg)
					}
				}
				if len(checked) == 0 {
					if len(diags) > 0 {
						t.Fatalf("the checker accepts the program, the translation rejects it:\n%s", diags.Text())
					}
					stdValueRun(t, b, c.fn)
					return
				}
				own := false
				for _, d := range checked {
					if d.Code != gostrict.CodeInternal && d.File == path+"/"+c.file && d.Line == line {
						own = true
					}
				}
				if !own {
					t.Errorf("the checker rejects the program, but not with a finding of its own at line %d:\n%s", line, checked.Text())
				}
				if diags.Text() != checked.Text() || b != nil {
					t.Errorf("the translation changes the findings of the checker:\n%s\nchecker:\n%s", diags.Text(), checked.Text())
				}
				t.Logf("rejected by the checker:\n%s", checked.Text())
			})
		})
	}
}

// TestStdValuesExcluded checks, that the package internal/examples/stdvalues as a whole contains none of the
// programs, so that it is in the subset, whichever way the checker decides about a value, and that each program of
// the table is a file of the package.
func TestStdValuesExcluded(t *testing.T) {
	b := compile(t, "internal/examples/stdvalues")
	if len(b.Exports) != 0 {
		t.Fatalf("the package exports %d functions without the excluded files", len(b.Exports))
	}
	files, err := filepath.Glob("internal/examples/stdvalues/*.go")
	if err != nil {
		t.Fatal(err)
	}
	var want []string
	for _, c := range stdValueCases {
		want = append(want, c.file)
	}
	var got []string
	for _, f := range files {
		if name := filepath.Base(f); name != "doc.go" {
			got = append(got, name)
		}
	}
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Fatalf("the files of the package %v are not the files of the table %v", got, want)
	}
}
