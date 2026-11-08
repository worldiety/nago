// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict_test

import (
	"go/ast"
	"go/types"
	"sort"
	"strings"
	"testing"
	"testing/fstest"

	"go.wdy.de/nago/pkg/gostrict"
)

// The enumeration of R-SCRIPT-GS-HEADER-LISTS of wokoda (L-N11-2, L-N11-9) over go/types: every exported function and
// method of a trusted header package, which the loader loads (here std/* and lib/* of nago),
//   - whose arguments, receiver, results or function parameters reach a list.List, list.Frozen or dict.Map (also
//     through a struct, an opt.Option, a pointer or a type argument; a type parameter can be one);
//   - with a handle as an argument or receiver, whose type has such a method;
//   - which has //gostrict:calls later, now or earlier, each method of a handle type, at which such a function stands
//     or which it takes, and each function with such a handle as an argument.
// A handle is a named struct of a header package with a field of type js.Ref.

// hlPackages are the trusted header packages of nago, which a program can import.
var hlPackages = []string{
	"go.wdy.de/nago/pkg/gostrict/std/big", "go.wdy.de/nago/pkg/gostrict/std/decimal", "go.wdy.de/nago/pkg/gostrict/std/dict",
	"go.wdy.de/nago/pkg/gostrict/std/fuel", "go.wdy.de/nago/pkg/gostrict/std/js", "go.wdy.de/nago/pkg/gostrict/std/list",
	"go.wdy.de/nago/pkg/gostrict/std/msg", "go.wdy.de/nago/pkg/gostrict/std/num", "go.wdy.de/nago/pkg/gostrict/std/opt",
	"go.wdy.de/nago/pkg/gostrict/std/text", "go.wdy.de/nago/pkg/gostrict/lib/asset", "go.wdy.de/nago/pkg/gostrict/lib/canvas",
	"go.wdy.de/nago/pkg/gostrict/lib/input", "go.wdy.de/nago/pkg/gostrict/lib/three", "go.wdy.de/nago/pkg/gostrict/lib/ui",
}

// hlFunc is an enumerated function or method.
type hlFunc struct {
	Name string // the full name, e.g. go.wdy.de/nago/pkg/gostrict/std/list.List.Delete
	Obj  *types.Func
	Recv *types.Named // the receiver or nil
	// Calls is the directive //gostrict:calls (now, later, earlier) or ""
	Calls string
	// Why tells, which criteria hold: l (lists), h (handle with such a method), c (callbacks)
	Why string
}

// hlLoad loads a program, which imports the packages, and returns it.
func hlLoad(t *testing.T, pkgs []string, modules ...gostrict.Module) *gostrict.Program {
	t.Helper()
	var imp strings.Builder
	for _, p := range pkgs {
		imp.WriteString("\t_ \"" + p + "\"\n")
	}
	src := "package s\n\nimport (\n" + imp.String() + ")\n"
	mods := append([]gostrict.Module{{Path: "llm.local/s", FS: fstest.MapFS{"s.go": {Data: []byte(src)}}}}, modules...)
	prog, err := gostrict.LoadModules("llm.local/s", mods...)
	if err != nil {
		t.Fatal(err)
	}
	return prog
}

// hlIsHandle tells, whether t is a handle of a header package.
func hlIsHandle(t types.Type) bool {
	n, ok := types.Unalias(t).(*types.Named)
	if !ok {
		return false
	}
	s, ok := n.Underlying().(*types.Struct)
	if !ok {
		return false
	}
	for i := range s.NumFields() {
		if strings.HasSuffix(s.Field(i).Type().String(), "/std/js.Ref") {
			return true
		}
	}
	return false
}

// hlReaches tells, whether a value of t may hold a list.List, list.Frozen or dict.Map.
func hlReaches(t types.Type, seen map[types.Type]bool) bool {
	if seen[t] {
		return false
	}
	seen[t] = true
	switch t := types.Unalias(t).(type) {
	case *types.Named:
		o := t.Origin().Obj()
		if o.Pkg() != nil {
			switch o.Pkg().Path() + "." + o.Name() {
			case "go.wdy.de/nago/pkg/gostrict/std/list.List", "go.wdy.de/nago/pkg/gostrict/std/list.Frozen", "go.wdy.de/nago/pkg/gostrict/std/dict.Map":
				return true
			}
		}
		for i := range t.TypeArgs().Len() {
			if hlReaches(t.TypeArgs().At(i), seen) {
				return true
			}
		}
		if hlIsHandle(t) {
			return false
		}
		return hlReaches(t.Underlying(), seen)
	case *types.Struct:
		for i := range t.NumFields() {
			if hlReaches(t.Field(i).Type(), seen) {
				return true
			}
		}
	case *types.Pointer:
		return hlReaches(t.Elem(), seen)
	case *types.Slice:
		return hlReaches(t.Elem(), seen)
	case *types.Signature:
		for _, tup := range []*types.Tuple{t.Params(), t.Results()} {
			for i := range tup.Len() {
				if hlReaches(tup.At(i).Type(), seen) {
					return true
				}
			}
		}
	case *types.TypeParam:
		return true
	}
	return false
}

// hlEnumerate enumerates the functions of the trusted header packages of the program, whose path has one of the
// prefixes, sorted by name.
func hlEnumerate(prog *gostrict.Program, prefixes ...string) []hlFunc {
	inScope := func(p string) bool {
		for _, x := range prefixes {
			if strings.HasPrefix(p, x) {
				return true
			}
		}
		return false
	}
	dirs := map[types.Object]string{}
	for _, p := range prog.Packages {
		if p.Kind == gostrict.KindUser {
			continue
		}
		for _, f := range p.Files {
			for _, d := range f.Decls {
				fd, ok := d.(*ast.FuncDecl)
				if !ok || fd.Doc == nil {
					continue
				}
				for _, c := range fd.Doc.List {
					if v, ok := strings.CutPrefix(c.Text, "//gostrict:calls "); ok {
						dirs[p.Info.Defs[fd.Name]] = strings.TrimSpace(v)
					}
				}
			}
		}
	}
	// the kinds of the handles come from all trusted header packages, e.g. ui.Window of nago for a header of wokoda;
	// the result holds the functions of the packages with the prefixes
	var all []hlFunc
	for _, p := range prog.Packages {
		if p.Kind == gostrict.KindUser {
			continue
		}
		sc := p.Types.Scope()
		for _, n := range sc.Names() {
			o := sc.Lookup(n)
			if !o.Exported() {
				continue
			}
			switch o := o.(type) {
			case *types.Func:
				all = append(all, hlFunc{Name: p.Path + "." + n, Obj: o, Calls: dirs[o]})
			case *types.TypeName:
				named, ok := o.Type().(*types.Named)
				if !ok {
					continue
				}
				for i := range named.NumMethods() {
					m := named.Method(i)
					// HostRef, HostFromRef, HostValues and the like convert values for a host in Go, a program can not
					// call them
					if m.Exported() && !strings.HasPrefix(m.Name(), "Host") {
						all = append(all, hlFunc{Name: p.Path + "." + n + "." + m.Name(), Obj: m, Recv: named, Calls: dirs[m]})
					}
				}
			}
		}
	}
	sig := func(f hlFunc) *types.Signature { return f.Obj.Type().(*types.Signature) }
	handles := func(f hlFunc) []string {
		var r []string
		if f.Recv != nil && hlIsHandle(f.Recv) {
			r = append(r, f.Recv.String())
		}
		for i := range sig(f).Params().Len() {
			if t := sig(f).Params().At(i).Type(); hlIsHandle(t) {
				r = append(r, types.TypeString(t, nil))
			}
		}
		return r
	}
	why := make([]string, len(all))
	listHandles, callHandles := map[string]bool{}, map[string]bool{}
	for i, f := range all {
		if hlReaches(f.Obj.Type(), map[types.Type]bool{}) || f.Recv != nil && hlReaches(f.Recv, map[types.Type]bool{}) {
			why[i] += "l"
			if f.Recv != nil && hlIsHandle(f.Recv) {
				listHandles[f.Recv.String()] = true
			}
		}
		if f.Calls != "" {
			why[i] += "c"
			for _, h := range handles(f) {
				callHandles[h] = true
			}
		}
	}
	var res []hlFunc
	for i, f := range all {
		w := why[i]
		for _, h := range handles(f) {
			if listHandles[h] && !strings.Contains(w, "h") {
				w += "h"
			}
			if callHandles[h] && !strings.Contains(w, "c") {
				w += "c"
			}
		}
		if w != "" && inScope(f.Name) {
			f.Why = w
			res = append(res, f)
		}
	}
	sort.Slice(res, func(i, j int) bool { return res[i].Name < res[j].Name })
	return res
}
