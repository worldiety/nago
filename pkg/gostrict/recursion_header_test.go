// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict_test

import (
	"context"
	"go/types"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"go.wdy.de/nago/pkg/gostrict"
	"go.wdy.de/nago/pkg/gostrict/lib/ui"
)

// The tests of this file check the classification of the functions of the header packages, which take a function,
// after L-N2-2 of the contract R-SCRIPT-GS-RECURSION: each calls its function arguments at once or later, e.g. the
// callback of a button, and a function without classification fails the test. The functions are found here from
// the sources of the header packages, not from the list of the checker, and a sample of them is run, so that the
// classification is observed and not only claimed.

const recModule = "go.wdy.de/nago/pkg/gostrict/"

// recHeaderPackages returns the import paths of the header packages of nago, the directories below std and lib,
// whose files have the directive //gostrict:header.
func recHeaderPackages(t *testing.T) []string {
	t.Helper()
	var res []string
	for _, root := range []string{"std", "lib"} {
		dirs, err := os.ReadDir(root)
		if err != nil {
			t.Fatal(err)
		}
		for _, d := range dirs {
			if !d.IsDir() {
				continue
			}
			files, _ := filepath.Glob(filepath.Join(root, d.Name(), "*.go"))
			for _, f := range files {
				if strings.HasSuffix(f, "_test.go") {
					continue
				}
				buf, err := os.ReadFile(f)
				if err != nil {
					t.Fatal(err)
				}
				if slices.Contains(strings.Split(string(buf), "\n"), "//gostrict:header") {
					res = append(res, recModule+root+"/"+d.Name())
					break
				}
			}
		}
	}
	return res
}

// takesFunction tells, whether a parameter of the function is a function.
func takesFunction(fn *types.Func) bool {
	params := fn.Signature().Params()
	for i := range params.Len() {
		if _, ok := params.At(i).Type().Underlying().(*types.Signature); ok {
			return true
		}
	}
	return false
}

// recHeaderFuncs returns the exported functions and methods of the header packages of nago, which take a function.
func recHeaderFuncs(t *testing.T) []*types.Func {
	t.Helper()
	var res []*types.Func
	for _, path := range recHeaderPackages(t) {
		prog, err := gostrict.LoadModules(path)
		if err != nil {
			t.Fatal(err)
		}
		p := prog.Package(path)
		if p == nil || p.Kind != gostrict.KindHeader {
			t.Fatalf("%s is not loaded as a header package", path)
		}
		scope := p.Types.Scope()
		for _, name := range scope.Names() {
			switch o := scope.Lookup(name).(type) {
			case *types.Func:
				if o.Exported() && takesFunction(o) {
					res = append(res, o)
				}
			case *types.TypeName:
				named, ok := o.Type().(*types.Named)
				if !ok {
					continue
				}
				for i := range named.NumMethods() {
					if m := named.Method(i); m.Exported() && takesFunction(m) {
						res = append(res, m)
					}
				}
			}
		}
	}
	return res
}

// recObserved are header functions, whose call is run with a function argument, which counts its calls. The
// argument runs at once, if the count is not zero after the call. The program exports a function with the result
// ui.View, without which goja does not load a program, which passes a view to a function of the host.
var recObserved = []struct {
	fn, probe string
	now       bool
}{
	{"(" + recModule + "lib/ui.TBox).WithFrame", `_ = ui.BoxAlign(ui.Center, ui.Text("")).WithFrame(func(fr ui.Frame) ui.Frame {
		h.n++
		return fr
	})`, true},
	{"(" + recModule + "lib/ui.TStack).With", `_ = ui.VStack().With(func(s ui.TStack) ui.TStack {
		h.n++
		return s
	})`, true},
	{"(" + recModule + "lib/ui.TText).WithFrame", `_ = ui.Text("").WithFrame(func(fr ui.Frame) ui.Frame {
		h.n++
		return fr
	})`, true},
	{recModule + "std/list.SortFunc", `list.SortFunc(list.Of(int32(2), 1), func(a, b int32) int32 {
		h.n++
		return a - b
	})`, true},
	{recModule + "lib/ui.PrimaryButton", `_ = ui.PrimaryButton(func() { h.n++ })`, false},
	{"(" + recModule + "lib/ui.TStack).Action", `_ = ui.VStack().Action(func() { h.n++ })`, false},
	{"(" + recModule + "lib/ui.TText).Action", `_ = ui.Text("").Action(func() { h.n++ })`, false},
	{recModule + "lib/ui.TabsPage", `_ = ui.Tabs(ui.TabsPage("a", func() ui.View {
		h.n++
		return ui.Text("")
	}))`, false},
	{recModule + "lib/ui.SizeClass", `_ = ui.SizeClass(ui.SizeClassSmall, func() ui.View {
		h.n++
		return ui.Text("")
	})`, false},
}

// recObserve runs the probe in goja and in the vm and returns, whether the function argument ran during the call.
func recObserve(t *testing.T, probe string) bool {
	t.Helper()
	src := `package s

import (
	"go.wdy.de/nago/pkg/gostrict/lib/ui"
	"go.wdy.de/nago/pkg/gostrict/std/list"
)

type hit struct{ n int32 }

func V() ui.View { return ui.Text("") }

func L() int32 { return list.Of(int32(1)).Len() }

func Probe() int32 {
	h := &hit{}
	` + probe + `
	return h.n
}
`
	b, diags, err := gostrict.NewCache(1).Compile("llm.local/s", script(src))
	if err != nil || len(diags) > 0 {
		t.Fatalf("the probe does not compile: %v\n%s", err, diags.Text())
	}
	var res []bool
	for _, e := range recEngines() {
		in, err := e.e.Load(context.Background(), b, ui.Host())
		if err != nil {
			t.Fatalf("%s: %v", e.name, err)
		}
		out, err := in.Call(context.Background(), "Probe")
		_ = in.Close()
		if err != nil {
			t.Fatalf("%s: %v", e.name, err)
		}
		res = append(res, out[0].(int32) > 0)
	}
	if res[0] != res[1] {
		t.Fatalf("goja and the vm differ: %v", res)
	}
	return res[0]
}

// TestRecursionHeaderObserved runs the sample: each function calls its argument at once or not at all during the
// call, as the sample says, both kinds are in it and each function of the sample is one of the header functions,
// which take a function.
func TestRecursionHeaderObserved(t *testing.T) {
	all := map[string]bool{}
	for _, fn := range recHeaderFuncs(t) {
		all[fn.FullName()] = true
	}
	var now, later int
	for _, o := range recObserved {
		t.Run(o.fn, func(t *testing.T) {
			if !all[o.fn] {
				t.Fatalf("%s is no header function, which takes a function", o.fn)
			}
			if got := recObserve(t, o.probe); got != o.now {
				t.Fatalf("the argument ran during the call: %v, want %v", got, o.now)
			}
		})
		if o.now {
			now++
		} else {
			later++
		}
	}
	if now == 0 || later == 0 {
		t.Fatalf("the sample needs both kinds: at once %d, later %d", now, later)
	}
}

// recHeaderName returns the name of the function, as [gostrict.HeaderCallbacks] names it: the import path, the type
// of a method and the name.
func recHeaderName(fn *types.Func) string {
	if recv := fn.Signature().Recv(); recv != nil {
		t := recv.Type()
		if p, ok := t.(*types.Pointer); ok {
			t = p.Elem()
		}
		if n, ok := t.(*types.Named); ok {
			return fn.Pkg().Path() + "." + n.Obj().Name() + "." + fn.Name()
		}
	}
	return fn.Pkg().Path() + "." + fn.Name()
}

// recUnclassified returns the names of the functions without classification, which a test of a host reports.
func recUnclassified(cbs []gostrict.HeaderCallback) []string {
	var res []string
	for _, cb := range cbs {
		if cb.Calls != gostrict.CallsNow && cb.Calls != gostrict.CallsLater {
			res = append(res, cb.Name)
		}
	}
	return res
}

// TestRecursionHeaderClassified fails, if a function of a header package of nago, which takes a function, has no
// classification: the list of gostrict.HeaderCallbacks has the same functions as the sources and each is now or
// later, and the classification of the sample is the observed one.
func TestRecursionHeaderClassified(t *testing.T) {
	got := map[string]gostrict.CallTime{}
	for _, path := range recHeaderPackages(t) {
		prog, err := gostrict.LoadModules(path)
		if err != nil {
			t.Fatal(err)
		}
		var own []gostrict.HeaderCallback
		for _, cb := range gostrict.HeaderCallbacks(prog) {
			if strings.HasPrefix(cb.Name, path+".") {
				own = append(own, cb)
				got[cb.Name] = cb.Calls
			}
		}
		if missing := recUnclassified(own); len(missing) > 0 {
			t.Errorf("%s: functions without //gostrict:calls now or later:\n  %s", path, strings.Join(missing, "\n  "))
		}
	}
	want := map[string]bool{}
	for _, fn := range recHeaderFuncs(t) {
		name := recHeaderName(fn)
		want[name] = true
		if _, ok := got[name]; !ok {
			t.Errorf("HeaderCallbacks misses %s", name)
		}
	}
	for name := range got {
		if !want[name] {
			t.Errorf("HeaderCallbacks lists %s, which takes no function or is no exported function of a header", name)
		}
	}
	if len(want) == 0 {
		t.Fatal("no header function takes a function")
	}
	for _, o := range recObserved {
		name := strings.NewReplacer("(", "", ").", ".").Replace(o.fn)
		calls, ok := got[name]
		switch {
		case !ok:
			t.Errorf("%s is not listed", name)
		case o.now && calls != gostrict.CallsNow, !o.now && calls != gostrict.CallsLater:
			t.Errorf("%s is classified %q, but the run observed now=%v", name, calls, o.now)
		}
	}
	t.Logf("%d functions of the header packages of nago take a function", len(want))
}

// TestRecursionHeaderClassifiedCounter is the counter test with an invented header function, which calls its
// argument, without the directive, with now and with later: without one it is reported as unclassified and counts as
// now, so the cycle through it is reported, as with now, while later has no edge.
func TestRecursionHeaderClassifiedCounter(t *testing.T) {
	for _, row := range []struct {
		name, directive string
		calls           gostrict.CallTime
		finding         bool
	}{
		{"without directive", "", "", true},
		{"now", "//gostrict:calls now\n", gostrict.CallsNow, true},
		{"later", "//gostrict:calls later\n", gostrict.CallsLater, false},
	} {
		t.Run(row.name, func(t *testing.T) {
			fsys := fstest.MapFS{
				"h/h.go": {Data: []byte("// Package h is an invented header.\n//\n//gostrict:header\npackage h\n\n// Each calls f.\n//\n//gostrict:host\n" + row.directive + "func Each(f func()) {}\n")},
				"app.go": {Data: []byte("// Package app uses the invented header.\n//\n//gostrict:terminating\npackage app\n\nimport \"llm.local/app/h\"\n\nfunc F() int32 {\n\th.Each(func() { _ = F() })\n\treturn 1\n}\n")},
			}
			prog, err := gostrict.LoadModules("llm.local/app", gostrict.Module{Path: "llm.local/app", FS: fsys, Trusted: true})
			if err != nil {
				t.Fatal(err)
			}
			cbs := gostrict.HeaderCallbacks(prog)
			var each []gostrict.HeaderCallback
			for _, cb := range cbs {
				if cb.Name == "llm.local/app/h.Each" {
					each = append(each, cb)
				}
			}
			if len(each) != 1 || each[0].Calls != row.calls {
				t.Fatalf("HeaderCallbacks lists %+v, want Each with %q", each, row.calls)
			}
			if missing := recUnclassified(each); (len(missing) > 0) != (row.calls == "") {
				t.Errorf("unclassified %v with the directive %q", missing, row.directive)
			}
			diags := gostrict.Check(prog)
			found := false
			for _, d := range diags {
				switch {
				case d.Code == gostrict.CodeTermination && d.File == "llm.local/app/app.go" && d.Line == 8:
					found = true
				default:
					t.Errorf("unexpected diagnostic %s:%d %s %s", d.File, d.Line, d.Code, d.Msg)
				}
			}
			if found != row.finding {
				t.Errorf("GS040 at F: %v, want %v:\n%s", found, row.finding, diags.Text())
			}
		})
	}
}
