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
	"strconv"

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
				if decl.Tok != token.TYPE {
					continue
				}
				for _, spec := range decl.Specs {
					ts := spec.(*ast.TypeSpec)
					obj, ok := p.Info.Defs[ts.Name].(*types.TypeName)
					if !ts.Name.IsExported() || !ok || isHandle(obj.Type()) || isBoundaryConstraint(obj.Type()) {
						continue
					}
					d.report(ts.Name.Pos(), ts.Name.End(), CodeBridge, "header packages export only handles and constraints", "a handle is a struct with a single unexported field of type js.Ref, a constraint a union of types, which cross the boundary", "declare type T struct{ ref js.Ref } or type Kind interface{ A | B }")
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

	const why = "only bool, int32, uint32, float64, string, big.Int, handles, records without handles and lists and maps of them cross the boundary, error only as a result, and functions of such values as parameters"
	if decl.Recv != nil {
		d.report(decl.Name.Pos(), decl.Name.End(), CodeBridge, "header packages can not declare methods", why, "declare a function, which takes the handle as its first parameter")
		return
	}

	instances, ok := instantiations(fn)
	if !ok {
		d.report(decl.Type.Pos(), decl.Type.End(), CodeBridge, "the type parameters of "+fn.Name()+" must be unions of types, which cross the boundary", why, "constrain them like [T int32 | float64] or [T Mesh | Group]")
		return
	}
	for _, targs := range instances {
		sig := fn.Signature()
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

	if msg := checkBinding(js, fn.Signature().Params().Len()); msg != "" {
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
