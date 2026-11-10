// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict

import (
	"go/ast"
	"go/token"
	"go/types"
	"slices"
	"strconv"
	"strings"

	jsast "github.com/dop251/goja/ast"
	jsparser "github.com/dop251/goja/parser"
)

// checkHeader checks the boundary of a header package: the bodies of its functions are trusted, but their
// signatures and bindings are checked, so that a mistake shows up at compile time and not in a browser.
func checkHeader(p *Package, d *diagnoser, decls map[*types.Func]funcSrc) {
	for _, f := range p.Files {
		for _, decl := range f.Decls {
			switch decl := decl.(type) {
			case *ast.GenDecl:
				switch decl.Tok {
				case token.VAR:
					checkHeaderVars(p, d, decl)
					continue
				case token.TYPE:
				default:
					continue
				}
				for _, spec := range decl.Specs {
					ts := spec.(*ast.TypeSpec)
					obj, ok := p.Info.Defs[ts.Name].(*types.TypeName)
					if !ts.Name.IsExported() || !ok || headerType(obj.Type()) {
						continue
					}
					d.report(ts.Name.Pos(), ts.Name.End(), CodeBridge, "header packages export only handles, sealed interfaces of handles, records, enums and constraints", "a handle is a struct with a single unexported field of type js.Ref, a record a struct of exported fields, which cross the boundary, an enum a named bool, int32, uint32, float64 or string, a constraint a union of types, which cross the boundary", "declare type T struct{ ref js.Ref }, type View interface{ isView() } or type Kind interface{ A | B }")
				}
			case *ast.FuncDecl:
				checkHeaderFunc(p, d, decls, decl)
			}
		}
	}
}

func checkHeaderFunc(p *Package, d *diagnoser, decls map[*types.Func]funcSrc, decl *ast.FuncDecl) {
	fn, ok := p.Info.Defs[decl.Name].(*types.Func)
	if !ok {
		return
	}
	checkConstDirective(d, decls[fn], fn)
	checkCallsDirective(d, decls[fn], fn)
	checkZeroDirective(d, decls[fn], fn)

	const why = "only bool, int32, uint32, int64, uint64, float64, string, big.Int, bin.Bytes, handles, records without handles and lists and maps of them cross the boundary, error only as a result, and functions of such values as parameters"
	if decl.Recv != nil {
		recv := fn.Signature().Recv().Type()
		if _, ptr := recv.(*types.Pointer); ptr || types.Unalias(recv).(*types.Named).TypeParams().Len() > 0 {
			d.report(decl.Name.Pos(), decl.Name.End(), CodeBridge, "a method of a header package needs a receiver, which is a value of a non-generic type", why, "declare func (t T) "+fn.Name()+"(...)")
			return
		}
		if !fn.Exported() {
			// an unexported method seals an interface of handles, the program can not call it
			if sig := fn.Signature(); sig.Params().Len() > 0 || sig.Results().Len() > 0 {
				d.report(decl.Name.Pos(), decl.Name.End(), CodeBridge, "an unexported method of a header package only seals an interface", "the program can not call it, so it has neither parameters nor results", "declare func (T) "+fn.Name()+"() {}")
			}
			return
		}
	}

	instances, ok := instantiations(fn)
	if !ok {
		d.report(decl.Type.Pos(), decl.Type.End(), CodeBridge, "the type parameters of "+fn.Name()+" must be unions of types, which cross the boundary", why, "constrain them like [T int32 | float64] or [T Mesh | Group]")
		return
	}
	for _, targs := range instances {
		sig := externSignature(fn)
		if len(targs) > 0 {
			inst, err := types.Instantiate(nil, fn.Type(), targs, true)
			if err != nil {
				d.report(decl.Type.Pos(), decl.Type.End(), CodeBridge, err.Error(), why, "")
				return
			}
			sig = inst.(*types.Signature)
		}
		if _, ok := signature(fn.Name(), sig); !ok {
			d.report(decl.Type.Pos(), decl.Type.End(), CodeBridge, "the signature of "+fn.Name()+" can not cross the boundary: "+types.TypeString(sig, types.RelativeTo(fn.Pkg())), why, "use only these types, e.g. pass a struct as its fields")
			return
		}
	}

	js := decls[fn].js
	if decls[fn].host {
		if js != "" {
			d.report(decl.Name.Pos(), decl.Name.End(), CodeHeader, fn.Name()+" has a JavaScript binding and "+hostDirective, hostDirective+" declares, that only the host implements the function", "remove one of the directives")
		}
		return
	}
	if js == "" {
		d.report(decl.Name.Pos(), decl.Name.End(), CodeHeader, fn.Name()+" has no JavaScript binding", "a function of a header package needs a //gostrict:js directive, or "+hostDirective+", if only a host in Go implements it", "add //gostrict:js (a, b) => ... to the documentation of the function")
		return
	}

	if msg := checkBinding(js, externSignature(fn).Params().Len()); msg != "" {
		pos := decl.Name.Pos()
		if decl.Doc != nil {
			for _, c := range decl.Doc.List {
				if len(c.Text) >= len(jsDirective) && c.Text[:len(jsDirective)] == jsDirective {
					pos = c.Pos()
				}
			}
		}
		d.report(pos, token.NoPos, CodeBinding, "the binding of "+fn.Name()+" is invalid: "+msg, "the binding is a JavaScript function expression, which the program calls with the parameters", "write e.g. //gostrict:js (a, b) => a * b")
	}
}

// checkConstDirective checks, that the constDirective of a function names parameters, which take constants.
func checkConstDirective(d *diagnoser, src funcSrc, fn *types.Func) {
	if src.constPos == token.NoPos {
		return
	}
	report := func(msg string) {
		d.report(src.constPos, token.NoPos, CodeHeader, msg, constDirective+" names the parameters of a function, whose arguments the host reads from the source at compile time, so they must be constants of a basic type", "write e.g. "+constDirective+" name for func Activity(name string)")
	}
	sig := fn.Signature()
	if sig.Recv() != nil {
		report(constDirective + " is only allowed on functions, because a method may also be called through an interface")
		return
	}
	if len(src.consts) == 0 {
		report(constDirective + " names no parameter")
		return
	}
	for _, name := range src.consts {
		i := paramIndex(sig, name)
		switch {
		case i < 0:
			report(fn.Name() + " has no parameter " + name)
		case sig.Variadic() && i == sig.Params().Len()-1:
			report("the variadic parameter " + name + " of " + fn.Name() + " takes no constant")
		case !constType(sig.Params().At(i).Type()):
			report("the parameter " + name + " of " + fn.Name() + " has the type " + types.TypeString(sig.Params().At(i).Type(), types.RelativeTo(fn.Pkg())) + ", which has no constants")
		}
	}
}

// checkCallsDirective checks the callsDirective of a function of a header package: now, later or concurrent on a
// function, which takes a function, earlier on one, which takes none, and concurrent not on a method, because a method
// may also be called through an interface, which bypasses the check of its arguments.
func checkCallsDirective(d *diagnoser, src funcSrc, fn *types.Func) {
	if src.callsPos == token.NoPos {
		return
	}
	report := func(msg string) {
		d.report(src.callsPos, token.NoPos, CodeHeader, msg, callsDirective+" tells the checks, whether the function calls its function arguments during the call, only after it returned or on another goroutine, or whether it calls functions, which it got earlier", "write "+callsDirective+" now, later or concurrent on a function, which takes a function, and "+callsDirective+" earlier on one, which takes none, but calls functions, which another call got")
	}
	takes := len(funcParams(fn.Signature())) > 0
	switch {
	case src.calls != CallsNow && src.calls != CallsLater && src.calls != CallsEarlier && src.calls != CallsConcurrent:
		report(callsDirective + " takes now, later, earlier or concurrent, not " + strconv.Quote(string(src.calls)))
	case src.calls == CallsEarlier && takes:
		report(fn.Name() + " takes a function, so it calls it now or later, not earlier")
	case src.calls != CallsEarlier && !takes:
		report(fn.Name() + " takes no function, so " + callsDirective + " " + string(src.calls) + " has nothing to tell")
	case src.calls == CallsConcurrent && fn.Signature().Recv() != nil:
		report(callsDirective + " concurrent is only allowed on functions, because a method may also be called through an interface")
	}
}

// checkZeroDirective checks the zeroDirective of a function of a header package: its results pair their values with a
// last bool or error, see [resultPairing].
func checkZeroDirective(d *diagnoser, src funcSrc, fn *types.Func) {
	if !src.zeroPos.IsValid() {
		return
	}
	if _, _, _, ok := resultPairing(fn.Signature().Results()); !ok {
		d.report(src.zeroPos, token.NoPos, CodeHeader, fn.Name()+" returns no value besides a last bool or error, so "+zeroDirective+" has nothing to tell", zeroDirective+" tells the fix of GS033, that the function returns zero values, if its last bool is false or its last error is not nil", "remove "+zeroDirective+" or declare it on a function like func F(s string) (int32, bool)")
	}
}

// HeaderCallback is an exported function or method of a header package, which takes a function, see
// [HeaderCallbacks].
type HeaderCallback struct {
	// Name is the full name, e.g. "go.wdy.de/nago/pkg/gostrict/lib/ui.TBox.WithFrame" or
	// "go.wdy.de/nago/pkg/gostrict/std/list.SortFunc".
	Name string
	// Calls is the time of the directive //gostrict:calls or "" without one.
	Calls CallTime
}

// HeaderCallbacks returns the exported functions and methods of the header packages of the program, which take a
// function, with the time, at which they call it, sorted by name. The check of recursion follows a call of such a
// function to its function arguments, unless its directive is //gostrict:calls later or concurrent, e.g. the action of a
// button.
// A function without the directive counts as now, so a test of a host checks, that each of its header functions
// has one.
func HeaderCallbacks(prog *Program) []HeaderCallback {
	var res []HeaderCallback
	for fn, src := range funcDecls(prog) {
		if src.pkg.Kind == KindUser || !fn.Exported() || len(funcParams(fn.Signature())) == 0 {
			continue
		}
		if recv := fn.Signature().Recv(); recv != nil {
			if n, ok := types.Unalias(derefType(recv.Type())).(*types.Named); !ok || !n.Obj().Exported() {
				continue
			}
		}
		res = append(res, HeaderCallback{Name: src.pkg.Path + "." + funcName(fn), Calls: src.calls})
	}
	slices.SortFunc(res, func(a, b HeaderCallback) int { return strings.Compare(a.Name, b.Name) })
	return res
}

// funcName returns the name of the function or Type.Method for a method.
func funcName(fn *types.Func) string {
	recv := fn.Signature().Recv()
	if recv == nil {
		return fn.Name()
	}
	if n, ok := types.Unalias(derefType(recv.Type())).(*types.Named); ok {
		return n.Obj().Name() + "." + fn.Name()
	}
	return fn.Name()
}

func derefType(t types.Type) types.Type {
	if p, ok := t.(*types.Pointer); ok {
		return p.Elem()
	}
	return t
}

// funcParams returns the indices of the parameters of the signature, whose types are functions or, for a variadic
// parameter, slices of functions.
func funcParams(sig *types.Signature) []int {
	var res []int
	for i := range sig.Params().Len() {
		t := sig.Params().At(i).Type()
		if s, ok := t.Underlying().(*types.Slice); ok && sig.Variadic() && i == sig.Params().Len()-1 {
			t = s.Elem()
		}
		if coreSignature(t) != nil {
			res = append(res, i)
		}
	}
	return res
}

// paramIndex returns the index of the parameter of the signature with the name or -1.
func paramIndex(sig *types.Signature, name string) int {
	for i := range sig.Params().Len() {
		if sig.Params().At(i).Name() == name {
			return i
		}
	}
	return -1
}

// constType tells, whether constants of the type exist: bool, numbers and strings and the enums of them, but not
// type parameters.
func constType(t types.Type) bool {
	if _, ok := t.(*types.TypeParam); ok {
		return false
	}
	_, ok := t.Underlying().(*types.Basic)
	return ok
}

// checkBinding parses the binding and returns a problem or "".
func checkBinding(js string, params int) string {
	// the line break ends a comment at the end of the binding
	prog, err := jsparser.ParseFile(nil, "binding.js", "("+js+"\n)", 0)
	if err != nil {
		return err.Error()
	}
	if len(prog.Body) != 1 {
		return "it must be a single expression"
	}
	stmt, ok := prog.Body[0].(*jsast.ExpressionStatement)
	if !ok {
		return "it must be an expression"
	}

	var list *jsast.ParameterList
	switch f := stmt.Expression.(type) {
	case *jsast.ArrowFunctionLiteral:
		if f.Async {
			return "it must not be async, the program calls it synchronously"
		}
		list = f.ParameterList
	case *jsast.FunctionLiteral:
		if f.Async || f.Generator {
			return "it must not be async or a generator, the program calls it synchronously"
		}
		list = f.ParameterList
	default:
		return "it must be a function expression like (a, b) => a * b"
	}

	if list.Rest != nil {
		return "it must not have rest parameters"
	}
	if len(list.List) != params {
		return "it takes " + strconv.Itoa(len(list.List)) + " parameters, but the function has " + strconv.Itoa(params)
	}
	return ""
}

// maxInstantiations limits the combinations of the type arguments of a header function, each of which needs its
// own native adapter.
const maxInstantiations = 64

// instantiations returns all combinations of the terms of the type parameters of a header function, or false, if
// a type parameter is not a union of types, which cross the boundary.
func instantiations(fn *types.Func) ([][]types.Type, bool) {
	tparams := fn.Signature().TypeParams()
	res := [][]types.Type{nil}
	for i := 0; i < tparams.Len(); i++ {
		terms, ok := boundaryTerms(tparams.At(i))
		if !ok {
			return nil, false
		}
		var next [][]types.Type
		for _, prefix := range res {
			for _, t := range terms {
				next = append(next, append(append([]types.Type{}, prefix...), t))
			}
		}
		res = next
		if len(res) > maxInstantiations {
			return nil, false
		}
	}
	return res, true
}

// isBoundaryConstraint tells, whether t is an interface, which only unites types, which cross the boundary.
func isBoundaryConstraint(t types.Type) bool {
	iface, ok := t.Underlying().(*types.Interface)
	if !ok {
		return false
	}
	_, ok = boundaryTerms(types.NewTypeParam(types.NewTypeName(0, nil, "T", nil), iface))
	return ok
}

// headerType tells, whether a header package may export the type t.
func headerType(t types.Type) bool {
	if isHandle(t) || isBoundaryConstraint(t) || isHandleUnion(t) {
		return true
	}
	if k := valueKind(t); k >= ValueBool && k <= ValueString {
		// an enum
		return true
	}
	bt, ok := newBoundary().typeOf(t)
	return ok && bt.Kind == ValueRecord
}

// checkHeaderVars checks, that the exported variables of a header package have a binding, which is the value of
// the variable in JavaScript, e.g. a preset of a font, and a type, which crosses the boundary.
func checkHeaderVars(p *Package, d *diagnoser, decl *ast.GenDecl) {
	for _, spec := range decl.Specs {
		vs := spec.(*ast.ValueSpec)
		for _, n := range vs.Names {
			v, ok := p.Info.Defs[n].(*types.Var)
			if !ok || !n.IsExported() {
				continue
			}
			if vs.Type == nil {
				d.report(n.Pos(), n.End(), CodeHeader, n.Name+" needs an explicit type", "the program reads the binding of the variable, so its value in Go is not checked", "declare var "+n.Name+" T = ...")
			}
			if varBinding(decl, vs) == "" {
				d.report(n.Pos(), n.End(), CodeHeader, n.Name+" has no JavaScript binding", "a variable of a header package is a value, which the program reads in JavaScript as well", "add //gostrict:js followed by the value as an expression to the documentation of the variable")
			}
			if _, ok := newBoundary().headerType(v.Type(), false); !ok {
				d.report(n.Pos(), n.End(), CodeBridge, "the variable "+n.Name+" can not cross the boundary", "a variable of a header package holds a basic value, an enum or a record", "declare a function, which returns the value")
			}
		}
	}
}

// externSignature returns the signature of a function of a header package as the program calls it: a method
// takes its receiver as its first parameter.
func externSignature(fn *types.Func) *types.Signature {
	sig := fn.Signature()
	recv := sig.Recv()
	if recv == nil {
		return sig
	}
	vars := []*types.Var{types.NewParam(recv.Pos(), recv.Pkg(), "recv", recv.Type())}
	for v := range sig.Params().Variables() {
		vars = append(vars, v)
	}
	return types.NewSignatureType(nil, nil, nil, types.NewTuple(vars...), sig.Results(), sig.Variadic())
}

// HeaderVar is an exported variable of a header package, whose binding is its value in JavaScript.
type HeaderVar struct {
	// Name is the full name, e.g. "go.wdy.de/nago/pkg/gostrict/lib/ui.Full".
	Name string
	JS   string
	Type Type
}

// HeaderVars returns the variables of the header packages of the program and the records of their types, so that
// a test compares their bindings with their values in Go, see the function HostVars of a generated file.
func HeaderVars(prog *Program) ([]HeaderVar, map[string]Record) {
	b := newBoundary()
	var res []HeaderVar
	for _, p := range prog.Packages {
		if p.Kind != KindHeader {
			continue
		}
		for _, f := range p.Files {
			for _, d := range f.Decls {
				g, ok := d.(*ast.GenDecl)
				if !ok || g.Tok != token.VAR {
					continue
				}
				for _, spec := range g.Specs {
					vs := spec.(*ast.ValueSpec)
					for _, n := range vs.Names {
						v, ok := p.Info.Defs[n].(*types.Var)
						if !ok || !n.IsExported() {
							continue
						}
						t, ok := b.headerType(v.Type(), false)
						if ok {
							res = append(res, HeaderVar{Name: p.Path + "." + n.Name, JS: varBinding(g, vs), Type: t})
						}
					}
				}
			}
		}
	}
	return res, b.records
}
