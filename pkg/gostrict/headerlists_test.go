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
	"go/types"
	"slices"
	"sort"
	"strings"
	"sync"
	"testing"
	"testing/fstest"
	"time"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/pkg/gostrict"
	"go.wdy.de/nago/pkg/gostrict/engine/gojaengine"
	"go.wdy.de/nago/pkg/gostrict/engine/vm"
	"go.wdy.de/nago/pkg/gostrict/lib/canvas"
	"go.wdy.de/nago/pkg/gostrict/lib/three"
	"go.wdy.de/nago/pkg/gostrict/lib/ui"
	"go.wdy.de/nago/pkg/gostrict/std/big"
	"go.wdy.de/nago/presentation/core"
	nui "go.wdy.de/nago/presentation/ui"
)

// The tests of this file check the contract R-SCRIPT-GS-HEADER-LISTS of wokoda (L-N11-2, L-N11-9) for the trusted
// header packages of nago, std/* and lib/*: each function, which the enumeration over go/types finds
// (headerlists_enum_test.go), stands in hlTable, a new one fails; and each one is observed in goja and in the vm,
// with the functions of the host of lib/ui, lib/canvas, lib/three and std/big, in the window of an application of
// nago, in one generated program (headerlists_gen_test.go):
//   - the lengths of each list and dict, which the program holds (the arguments, the earlier results of the same
//     handle and a list and a dict of its own), are the same after the call (row 1 of the table);
//   - a list or dict, which the function returns, is new: the program calls it twice and deletes from the first; the
//     second keeps its length (row 2; a list.Frozen never shrinks, it is not deleted from);
//   - a recording function, registered through each enumerated header with //gostrict:calls later or now at the start,
//     runs during no other call, and a function of the program, which a header with later gets, does not run during
//     its call; counter row: ui.RadioStateGroup.Notify with //gostrict:calls earlier runs the observer of Observe
//     (row 3);
//   - Delete and Pop of list.List and Delete of dict.Map shorten their receiver (row 4, in the same program);
//   - //gostrict:calls earlier at a function with a function parameter and now at one without are findings of the
//     header; counter row: ui.RadioStateGroup.Notify carries earlier without a finding (row 5,
//     TestHeaderListsDirectives).
// The generator writes a value of each type by itself; hlFixed gives the values of the types, which it can not write,
// and hlSkip names each function, which the generated program does not call, with the reason. Lists and dicts in a
// record (e.g. a struct with a list field) are held through the record only, not each on its own.

// hlFixed are the values of types, which the generator does not find by a constructor.
var hlFixed = map[string]string{
	"go.wdy.de/nago/pkg/gostrict/lib/ui.Window": "wnd",
	"go.wdy.de/nago/pkg/gostrict/lib/ui.View":   `ui.Text("v")`,
}

// hlSkip are the functions, which the generated program does not call, with the reason. A function there, which the
// enumeration does not find any more, fails the test as well.
var hlSkip = map[string]string{}

// hlTable is the table of the contract for nago: each enumerated function with its kinds, l (lists), h (a handle with
// such a method), c (callbacks), and the directive. It is written down, so that a new function fails the test, until
// the table names it.
var hlTable = map[string]string{}

// hlEnumerated returns the enumerated functions of nago.
func hlEnumerated(t *testing.T) (*gostrict.Program, []hlFunc) {
	t.Helper()
	prog := hlLoad(t, hlPackages)
	return prog, hlEnumerate(prog, "go.wdy.de/nago/pkg/gostrict/std/", "go.wdy.de/nago/pkg/gostrict/lib/")
}

// hlRow is the entry of the table of a function: its kinds and its directive.
func hlRow(f hlFunc) string {
	k := []byte(f.Why)
	sort.Slice(k, func(i, j int) bool { return k[i] < k[j] })
	if f.Calls == "" {
		return string(k)
	}
	return string(k) + " " + f.Calls
}

// TestHeaderListsTable checks, that the table names each enumerated function with its kinds and directive and
// nothing else.
func TestHeaderListsTable(t *testing.T) {
	_, fs := hlEnumerated(t)
	if len(fs) < 100 {
		t.Fatalf("premise: the enumeration finds only %d functions", len(fs))
	}
	seen := map[string]bool{}
	var missing []string
	for _, f := range fs {
		seen[f.Name] = true
		want, ok := hlTable[f.Name]
		switch {
		case !ok:
			missing = append(missing, fmt.Sprintf("%q: %q,", f.Name, hlRow(f)))
		case want != hlRow(f):
			t.Errorf("%s: the table says %q, the enumeration %q", f.Name, want, hlRow(f))
		}
	}
	if len(missing) > 0 {
		t.Errorf("%d enumerated functions are not in the table:\n%s", len(missing), strings.Join(missing, "\n"))
	}
	for n := range hlTable {
		if !seen[n] {
			t.Errorf("the table names %s, which the enumeration does not find", n)
		}
	}
	for n := range hlSkip {
		if !seen[n] {
			t.Errorf("hlSkip names %s, which the enumeration does not find", n)
		}
	}
}

// hlProgram is the generated program and the functions, which it calls, by the index of their block.
type hlProgram struct {
	src    string
	blocks []string
	skip   map[string]string // the functions, which it does not call, with the reason
	// earlier are the indices of the blocks of functions with //gostrict:calls earlier
	earlier []int
	regs    int // the recorders registered through a header with later or now
}

// hlBuild writes the program, which calls the functions.
func hlBuild(prog *gostrict.Program, fs []hlFunc) hlProgram {
	var pkgs []*types.Package
	for _, p := range prog.Packages {
		if p.Kind != gostrict.KindUser {
			pkgs = append(pkgs, p.Types)
		}
	}
	g := &hlGen{ctors: hlCtors(pkgs), imps: map[string]bool{}, fixed: hlFixed}
	res := hlProgram{skip: map[string]string{}}
	var regs, blocks []string
	var counters []string
	block := func(f hlFunc, i int) (string, error) {
		var recv *hlVal
		var b strings.Builder
		var held []hlHeld
		sig := f.Obj.Type().(*types.Signature)
		if f.Recv != nil {
			rt := types.Type(f.Recv)
			if f.Recv.TypeParams().Len() > 0 {
				var targs []types.Type
				for range f.Recv.TypeParams().Len() {
					targs = append(targs, types.Typ[types.Int32])
				}
				inst, err := types.Instantiate(nil, f.Recv, targs, false)
				if err != nil {
					return "", err
				}
				rt = inst
			}
			rv, err := g.value(rt, 0)
			if err != nil {
				return "", err
			}
			n := g.name("h")
			rv.pre = append(rv.pre, n+" := "+rv.expr)
			switch hlFull(rt) {
			case "go.wdy.de/nago/pkg/gostrict/std/list.List", "go.wdy.de/nago/pkg/gostrict/std/dict.Map", "go.wdy.de/nago/pkg/gostrict/std/list.Frozen":
				rv.pre = append(rv.pre, "n"+n+" := "+n+".Len()")
				held = append(held, hlHeld{n, -1})
			}
			// the lists and dicts, which the same handle returned before
			if named, ok := types.Unalias(rt).(*types.Named); ok && hlIsHandle(named) {
				for j := range named.NumMethods() {
					m := named.Method(j)
					ms := m.Type().(*types.Signature)
					if !m.Exported() || ms.Params().Len() > 0 || ms.Results().Len() != 1 {
						continue
					}
					switch hlFull(ms.Results().At(0).Type()) {
					case "go.wdy.de/nago/pkg/gostrict/std/list.List", "go.wdy.de/nago/pkg/gostrict/std/dict.Map", "go.wdy.de/nago/pkg/gostrict/std/list.Frozen":
						e := g.name("e")
						rv.pre = append(rv.pre, e+" := "+n+"."+m.Name()+"()", "n"+e+" := "+e+".Len()")
						held = append(held, hlHeld{e, -1})
					}
				}
			}
			rv.expr = n
			if obj, _, _ := types.LookupFieldOrMethod(rt, false, f.Obj.Pkg(), f.Obj.Name()); obj != nil {
				rv.sig, _ = obj.Type().(*types.Signature)
			}
			recv = &rv
		}
		c, err := g.call(f.Obj, recv, 1, "own")
		if err != nil {
			return "", err
		}
		held = append(held, c.held...)
		for _, p := range c.pre {
			b.WriteString("\t" + p + "\n")
		}
		b.WriteString("\ts := tot()\n")
		name := f.Name
		res := sig.Results()
		last := ""
		if res.Len() > 0 {
			lt := res.At(res.Len() - 1).Type()
			if types.Identical(lt, types.Universe.Lookup("error").Type()) {
				last = "error"
			} else if b, ok := lt.Underlying().(*types.Basic); ok && b.Kind() == types.Bool && res.Len() > 1 {
				last = "bool"
			}
		}
		fresh := ""
		switch {
		case res.Len() == 0:
			b.WriteString("\t" + c.expr + "\n")
		case last == "error":
			blank := strings.Repeat("_, ", res.Len()-1)
			b.WriteString("\tif " + blank + "e := " + c.expr + "; e != nil {\n\t\tout.Append(" + fmt.Sprintf("%q", name+"|error ") + " + e.Error())\n\t}\n")
		case last == "bool":
			b.WriteString("\t" + strings.TrimSuffix(strings.Repeat("_, ", res.Len()), ", ") + " = " + c.expr + "\n")
		case res.Len() == 1:
			switch hlFull(res.At(0).Type()) {
			case "go.wdy.de/nago/pkg/gostrict/std/list.List", "go.wdy.de/nago/pkg/gostrict/std/dict.Map":
				fresh = hlFull(res.At(0).Type())
				b.WriteString("\ta := " + c.expr + "\n\tz := " + c.expr + "\n\tnz := z.Len()\n")
			default:
				b.WriteString("\t_ = " + c.expr + "\n")
			}
		default:
			b.WriteString("\t" + strings.TrimSuffix(strings.Repeat("_, ", res.Len()), ", ") + " = " + c.expr + "\n")
		}
		earlier := f.Calls == "earlier"
		if earlier {
			b.WriteString("\tif tot() == s {\n\t\tout.Append(" + fmt.Sprintf("%q", name+"|earlier: no recorder ran") + ")\n\t}\n")
		} else {
			b.WriteString("\tif tot() != s {\n\t\tout.Append(" + fmt.Sprintf("%q", name+"|a recorder ran") + ")\n\t}\n")
		}
		if f.Calls == "later" {
			b.WriteString("\tif own != 0 {\n\t\tout.Append(" + fmt.Sprintf("%q", name+"|later: the argument ran") + ")\n\t}\n")
		}
		// row 1 and row 4: no held list or dict is shorter afterwards, except the receiver of Delete and Pop, which is
		shortens := f.Recv != nil && (f.Obj.Name() == "Delete" || f.Obj.Name() == "Pop") &&
			(hlFull(f.Recv) == "go.wdy.de/nago/pkg/gostrict/std/list.List" || hlFull(f.Recv) == "go.wdy.de/nago/pkg/gostrict/std/dict.Map")
		if shortens {
			b.WriteString("\tif " + recv.expr + ".Len() >= n" + recv.expr + " {\n\t\tout.Append(" + fmt.Sprintf("%q", name+"|does not shorten its receiver") + ")\n\t}\n")
		}
		for _, h := range held {
			cond := h.name + ".Len() < " + fmt.Sprint(h.n)
			if h.n < 0 {
				cond = h.name + ".Len() < n" + h.name
			}
			if shortens && (h.name == recv.expr || slices.ContainsFunc(recv.held, func(r hlHeld) bool { return r.name == h.name })) {
				continue
			}
			b.WriteString("\tif " + cond + " {\n\t\tout.Append(" + fmt.Sprintf("%q", name+"|held "+h.name+" shorter") + ")\n\t}\n")
		}
		switch fresh {
		case "go.wdy.de/nago/pkg/gostrict/std/list.List":
			b.WriteString("\ta.Delete(0)\n\tif z.Len() != nz {\n\t\tout.Append(" + fmt.Sprintf("%q", name+"|result not new") + ")\n\t}\n")
		case "go.wdy.de/nago/pkg/gostrict/std/dict.Map":
			b.WriteString("\tfor k := range a.Keys() {\n\t\ta.Delete(k)\n\t}\n\tif z.Len() != nz {\n\t\tout.Append(" + fmt.Sprintf("%q", name+"|result not new") + ")\n\t}\n")
		}
		return b.String(), nil
	}
	for i, f := range fs {
		if r, ok := hlSkip[f.Name]; ok {
			res.skip[f.Name] = r
			continue
		}
		body, err := block(f, i)
		if err != nil {
			res.skip[f.Name] = "generator: " + err.Error()
			continue
		}
		idx := len(blocks)
		blocks = append(blocks, "// "+f.Name+"\nfunc b"+fmt.Sprint(idx)+"(wnd ui.Window, out list.List[string], tot func() int32) {\n\town := int32(0)\n\t_ = own\n"+body+"}\n")
		res.blocks = append(res.blocks, f.Name)
		if f.Calls == "earlier" {
			res.earlier = append(res.earlier, idx)
		}
		// row 3: a recorder registered through each header with later or now
		if f.Calls == "later" || f.Calls == "now" {
			cn := g.name("c")
			counters = append(counters, cn)
			var recv *hlVal
			if f.Recv != nil {
				rv, err := g.value(f.Recv, 0)
				if err != nil {
					continue
				}
				if obj, _, _ := types.LookupFieldOrMethod(f.Recv, false, f.Obj.Pkg(), f.Obj.Name()); obj != nil {
					rv.sig, _ = obj.Type().(*types.Signature)
				}
				recv = &rv
			}
			c, err := g.call(f.Obj, recv, 1, cn)
			if err != nil {
				continue
			}
			var r strings.Builder
			for _, p := range c.pre {
				r.WriteString("\t" + p + "\n")
			}
			sig := f.Obj.Type().(*types.Signature)
			switch {
			case sig.Results().Len() == 0:
				r.WriteString("\t" + c.expr + "\n")
			default:
				r.WriteString("\t" + strings.TrimSuffix(strings.Repeat("_, ", sig.Results().Len()), ", ") + " = " + c.expr + "\n")
			}
			regs = append(regs, "\t// register through "+f.Name+"\n"+r.String())
		}
	}
	g.imps["go.wdy.de/nago/pkg/gostrict/std/list"] = true
	g.imps["go.wdy.de/nago/pkg/gostrict/std/text"] = true
	g.imps["go.wdy.de/nago/pkg/gostrict/lib/ui"] = true
	var imps []string
	for p := range g.imps {
		imps = append(imps, "\t\""+p+"\"")
	}
	sort.Strings(imps)
	var src strings.Builder
	src.WriteString("package hl\n\nimport (\n" + strings.Join(imps, "\n") + "\n)\n\n")
	src.WriteString("// Run calls each enumerated function and returns the deviations, one per line.\nfunc Run(wnd ui.Window) string {\n\tout := list.New[string]()\n")
	for _, c := range counters {
		src.WriteString("\t" + c + " := int32(0)\n")
	}
	for _, r := range regs {
		src.WriteString(r)
	}
	sum := "int32(0)"
	if len(counters) > 0 {
		sum = strings.Join(counters, " + ")
	}
	src.WriteString("\ttot := func() int32 { return " + sum + " }\n")
	for i := range blocks {
		src.WriteString(fmt.Sprintf("\tb%d(wnd, out, tot)\n", i))
	}
	src.WriteString("\treturn text.Join(out, \"\\n\")\n}\n\n")
	for _, b := range blocks {
		src.WriteString(b + "\n")
	}
	res.src = src.String()
	res.regs = len(regs)
	return res
}

// hlHost are the functions of the hosts of the header packages.
func hlHost() gostrict.Funcs {
	f := gostrict.Funcs{}
	for _, h := range []gostrict.Funcs{ui.Host(), canvas.Host(), three.Host(), big.Host()} {
		for k, v := range h {
			f[k] = v
		}
	}
	return f
}

// hlRun runs Run of the bundle in a window of an application of nago and returns its result.
func hlRun(t *testing.T, e gostrict.Engine, b *gostrict.Bundle, appID string) (string, error) {
	t.Helper()
	in, err := e.Load(context.Background(), b, hlHost())
	if err != nil {
		return "", err
	}
	defer func() { _ = in.Close() }()
	var once sync.Once
	var got string
	var callErr error
	app := nagotest.New(t, func(cfg *application.Configurator) {
		cfg.SetApplicationID(core.ApplicationID("de.worldiety.gostrict.headerlists." + appID))
		cfg.RootView("page", func(wnd core.Window) core.View {
			once.Do(func() {
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
				defer cancel()
				res, err := in.Call(ctx, "Run", ui.WindowOf(wnd).HostRef())
				if err != nil {
					callErr = err
					return
				}
				got, _ = res[0].(string)
			})
			return nui.Text("HL")
		})
	})
	app.Open(t, nil, "page")
	return got, callErr
}

// TestHeaderListsObserved runs the generated program in goja and in the vm and reports each deviation of a
// function from rows 1 to 3 of the table.
func TestHeaderListsObserved(t *testing.T) {
	prog, fs := hlEnumerated(t)
	p := hlBuild(prog, fs)
	if len(p.earlier) == 0 {
		t.Errorf("no enumerated function carries //gostrict:calls earlier (L-N11-9: ui.RadioStateGroup.Notify)")
	}
	var skipped []string
	for n, r := range p.skip {
		skipped = append(skipped, n+": "+r)
	}
	sort.Strings(skipped)
	for _, s := range skipped {
		t.Errorf("not observed: %s", s)
	}
	b, diags, err := gostrict.NewCache(1).Compile("llm.local/hl", gostrict.Module{Path: "llm.local/hl", FS: fstest.MapFS{"hl.go": {Data: []byte(p.src)}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(diags) > 0 {
		t.Fatalf("the generated program does not compile:\n%s", hlFirst(diags.Text(), 60))
	}
	for name, e := range map[string]gostrict.Engine{"goja": &gojaengine.Engine{}, "vm": &vm.Engine{}} {
		t.Run(name, func(t *testing.T) {
			got, err := hlRun(t, e, b, name)
			if err != nil {
				t.Fatalf("the program failed: %v", err)
			}
			if got != "" {
				t.Errorf("deviations of %d observed functions:\n%s", len(p.blocks), got)
			}
			t.Logf("%d functions observed, %d not, %d recorders registered through later or now", len(p.blocks), len(p.skip), p.regs)
		})
	}
}

func hlFirst(s string, n int) string {
	l := strings.Split(s, "\n")
	if len(l) > n {
		l = append(l[:n], fmt.Sprintf("… %d more lines", len(l)-n))
	}
	return strings.Join(l, "\n")
}

// TestHeaderListsDirectives is row 5 of the table: //gostrict:calls earlier at a function with a function parameter
// and now at one without are each a finding of the header; counter rows: earlier at a function without a function
// parameter and now at one with are without a finding, and ui.RadioStateGroup.Notify carries earlier, while a
// program, which calls it, has no finding.
func TestHeaderListsDirectives(t *testing.T) {
	for _, r := range []struct {
		name, decl string
		finding    bool
	}{
		{"earlier at a function with a function parameter", "//gostrict:calls earlier\nfunc F(f func()) {}\n", true},
		{"now at a function without a function parameter", "//gostrict:calls now\nfunc F() {}\n", true},
		{"counter row: earlier at a function without a function parameter", "//gostrict:calls earlier\nfunc F() {}\n", false},
		{"counter row: now at a function with a function parameter", "//gostrict:calls now\nfunc F(f func()) {}\n", false},
	} {
		t.Run(r.name, func(t *testing.T) {
			fsys := fstest.MapFS{
				"h/h.go": {Data: []byte("// Package h is an invented header.\n//\n//gostrict:header\npackage h\n\n// F is a function of the header.\n//\n//gostrict:host\n" + r.decl)},
				"app.go": {Data: []byte("// Package app uses the header.\npackage app\n\nimport \"llm.local/app/h\"\n\nfunc Run() int32 {\n\t_ = h.F\n\treturn 1\n}\n")},
			}
			prog, err := gostrict.LoadModules("llm.local/app", gostrict.Module{Path: "llm.local/app", FS: fsys, Trusted: true})
			if err != nil {
				t.Fatal(err)
			}
			var at []string
			for _, d := range gostrict.Check(prog) {
				if strings.HasSuffix(d.File, "h/h.go") && d.Code != gostrict.CodeTypeError {
					at = append(at, fmt.Sprintf("%s:%d %s %s", d.File, d.Line, d.Code, d.Msg))
				}
			}
			if r.finding && len(at) == 0 {
				t.Errorf("no finding of the header")
			}
			if !r.finding && len(at) > 0 {
				t.Errorf("findings of the header: %v", at)
			}
		})
	}
	t.Run("counter row: ui.RadioStateGroup.Notify carries earlier, a program, which calls it, has no finding", func(t *testing.T) {
		prog, fs := hlEnumerated(t)
		found := false
		for _, f := range fs {
			if f.Name == "go.wdy.de/nago/pkg/gostrict/lib/ui.RadioStateGroup.Notify" {
				found = true
				if f.Calls != "earlier" {
					t.Errorf("ui.RadioStateGroup.Notify carries %q, want earlier (L-N11-9)", f.Calls)
				}
			}
		}
		if !found {
			t.Fatalf("premise: the enumeration does not find ui.RadioStateGroup.Notify")
		}
		_ = prog
		src := "package s\n\nimport \"go.wdy.de/nago/pkg/gostrict/lib/ui\"\n\nfunc Run(wnd ui.Window) int32 {\n\tg := ui.AutoRadioStateGroup(wnd, \"g\", 2)\n\tg.Notify()\n\treturn g.SelectedIndex()\n}\n"
		p, err := gostrict.LoadModules("llm.local/s", gostrict.Module{Path: "llm.local/s", FS: fstest.MapFS{"s.go": {Data: []byte(src)}}})
		if err != nil {
			t.Fatal(err)
		}
		if ds := gostrict.Check(p); len(ds) > 0 {
			t.Errorf("the program, which calls Notify, has findings:\n%s", ds.Text())
		}
	})
}
