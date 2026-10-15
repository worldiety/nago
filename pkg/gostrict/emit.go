// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"slices"
	"strings"
)

//go:embed runtime.js
var runtimeJS string

// Bundle is a translated program. Its script evaluates to a function, which takes an object with the host
// functions by their full name and returns an object with the exported functions of the main package. Its
// property $externs holds the functions of header packages by their full name, as the host provided them or
// as their JavaScript binding, and $new the constructors of the records, which take the fields in their order.
type Bundle struct {
	JS string
	// Hash identifies the script, e.g. for a cache of an engine or as an immutable URL for a browser.
	Hash string
	// Exports are the exported functions of the main package, whose values can cross the boundary of an engine.
	Exports map[string]Func
	// Records are the records of the exports by their full name.
	Records map[string]Record
	// Unions are the sealed interfaces of the exports by their full name.
	Unions map[string]Union
	// Handlers maps the full name of a message to the exported function, which handles it, see [Session].
	Handlers map[string]string
	// Externs are the functions of header packages, which the program calls, e.g. "example.com/host.Mix".
	Externs []Extern
}

// Extern is a function of a header package.
type Extern struct {
	Func
	// JS is the JavaScript binding, a function expression.
	JS string
}

// Compile loads, checks and translates the package in dir. If the package is not in the subset, it returns the
// diagnostics and no bundle.
func Compile(dir string) (*Bundle, Diagnostics, error) {
	prog, err := Load(dir)
	if err != nil {
		return nil, nil, err
	}

	if diags := Check(prog); len(diags) > 0 {
		return nil, diags, nil
	}

	b, diags := Emit(prog)
	return b, diags, nil
}

// Emit translates a checked program into JavaScript. Every function is translated once for each combination of
// type arguments, it is called with, so the translated code knows the concrete type of every value.
func Emit(prog *Program) (b *Bundle, diags Diagnostics) {
	e := &emitter{
		prog:    prog,
		decls:   funcDecls(prog),
		tctx:    types.NewContext(),
		names:   map[string]string{},
		used:    map[string]bool{},
		prefix:  map[*types.Package]string{},
		globals: map[*types.Var]string{},
		diag:    &diagnoser{fset: prog.Fset, sources: prog.sources},
	}

	defer func() {
		if r := recover(); r != nil {
			err, ok := r.(emitError)
			if !ok {
				panic(r)
			}
			e.diag.report(err.pos, token.NoPos, CodeInternal, err.msg, "the compiler can not translate this construct yet", "rewrite it with simpler constructs")
			b, diags = nil, e.diag.sorted()
		}
	}()

	return e.emit()
}

type emitError struct {
	pos token.Pos
	msg string
}

func fail(pos token.Pos, format string, args ...any) {
	panic(emitError{pos: pos, msg: fmt.Sprintf(format, args...)})
}

type emitter struct {
	prog    *Program
	decls   map[*types.Func]funcSrc
	tctx    *types.Context
	out     strings.Builder
	externs strings.Builder
	names   map[string]string
	used    map[string]bool
	prefix  map[*types.Package]string
	globals map[*types.Var]string
	queue   []func()
	extern  []Extern
	diag    *diagnoser
}

func (e *emitter) emit() (*Bundle, Diagnostics) {
	var vars, init strings.Builder
	for _, p := range e.prog.Packages {
		if p.Kind != KindUser {
			continue
		}

		c := newCtx(e, p, nil)
		for _, f := range p.Files {
			for _, d := range f.Decls {
				g, ok := d.(*ast.GenDecl)
				if !ok || g.Tok != token.VAR {
					continue
				}
				for _, spec := range g.Specs {
					for _, n := range spec.(*ast.ValueSpec).Names {
						if v, ok := p.Info.Defs[n].(*types.Var); ok && n.Name != "_" {
							fmt.Fprintf(&vars, "let %s = %s;\n", e.global(v), c.zero(v.Type()))
						}
					}
				}
			}
		}

		for _, ini := range p.Info.InitOrder {
			c.initializer(ini)
		}
		init.WriteString(c.b.String())
	}

	exports := map[string]Func{}
	bnd := newBoundary()
	var ret strings.Builder
	main := e.prog.Main
	for _, name := range main.Types.Scope().Names() {
		fn, ok := main.Types.Scope().Lookup(name).(*types.Func)
		if !ok || !fn.Exported() {
			continue
		}
		sig, ok := bnd.exportSignature(name, fn.Signature())
		if !ok {
			continue
		}
		exports[name] = sig
		fmt.Fprintf(&ret, "\t%s: %s,\n", name, e.funcName(fn, nil, nil))
	}

	// the classes of the records and the methods of the variants must exist, before the program is written
	for _, name := range sortedNames(bnd.records) {
		e.structName(bnd.structs[name])
	}
	for _, name := range sortedNames(bnd.unions) {
		for _, v := range bnd.unions[name].Variants {
			e.methodTable(bnd.structs[v])
		}
	}

	for len(e.queue) > 0 {
		next := e.queue[0]
		e.queue = e.queue[1:]
		next()
	}

	var sb strings.Builder
	sb.WriteString("(function ($host) {\n\"use strict\";\n")
	sb.WriteString(usedRuntime(e.externs.String(), e.out.String(), vars.String(), init.String()))
	sb.WriteString("\n// externs\n")
	sb.WriteString(e.externs.String())
	sb.WriteString("\n// program\n")
	sb.WriteString(e.out.String())
	sb.WriteString("\n// package variables\n")
	sb.WriteString(vars.String())
	sb.WriteString(init.String())
	sb.WriteString("\nreturn {\n")
	sb.WriteString(ret.String())
	sb.WriteString("\t$new: {\n")
	for _, name := range sortedNames(bnd.records) {
		cls := e.structName(bnd.structs[name])
		fmt.Fprintf(&sb, "\t\t%s: (...f) => new %s(...f),\n", jsString(name), cls)
	}
	sb.WriteString("\t},\n")
	sb.WriteString("\t$unions: {\n")
	for _, name := range sortedNames(bnd.unions) {
		fmt.Fprintf(&sb, "\t\t%s: {\n", jsString(name))
		for _, v := range bnd.unions[name].Variants {
			mt := e.methodTable(bnd.structs[v])
			fmt.Fprintf(&sb, "\t\t\t%s: [%s, (...f) => $box(new %s(...f), %s)],\n", jsString(v), mt, e.structName(bnd.structs[v]), mt)
		}
		sb.WriteString("\t\t},\n")
	}
	sb.WriteString("\t},\n")
	sb.WriteString("\t$externs: {\n")
	for _, x := range e.extern {
		fmt.Fprintf(&sb, "\t\t%s: %s,\n", jsString(x.Name), e.names["X:"+x.Name])
	}
	sb.WriteString("\t},\n};\n})")

	slices.SortFunc(e.extern, func(a, b Extern) int { return strings.Compare(a.Name, b.Name) })
	sum := sha256.Sum256([]byte(sb.String()))
	return &Bundle{JS: sb.String(), Hash: hex.EncodeToString(sum[:16]), Exports: exports, Records: bnd.records, Unions: bnd.unions, Handlers: handlers(exports, bnd.unions), Externs: e.extern}, nil
}

// ident returns a unique identifier of the bundle based on base.
func (e *emitter) ident(base string) string {
	base = sanitize(base)
	name := base
	for i := 2; e.used[name]; i++ {
		name = fmt.Sprintf("%s$%d", base, i)
	}
	e.used[name] = true
	return name
}

func sanitize(s string) string {
	var sb strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '_', r == '$':
			sb.WriteRune(r)
		case r == ' ', r == ']':
		default:
			sb.WriteByte('$')
		}
	}
	return sb.String()
}

func (e *emitter) pkgPrefix(p *types.Package) string {
	if pre, ok := e.prefix[p]; ok {
		return pre
	}
	pre := e.ident(p.Name())
	e.prefix[p] = pre
	return pre
}

func (e *emitter) global(v *types.Var) string {
	if n, ok := e.globals[v]; ok {
		return n
	}
	n := e.ident(e.pkgPrefix(v.Pkg()) + "$" + v.Name())
	e.globals[v] = n
	return n
}

func typeKey(t types.Type) string {
	return types.TypeString(t, func(p *types.Package) string { return p.Path() })
}

func typeBase(t types.Type) string {
	return types.TypeString(t, func(p *types.Package) string { return p.Name() })
}

func typeListBase(ts []types.Type) string {
	var parts []string
	for _, t := range ts {
		parts = append(parts, typeBase(t))
	}
	return strings.Join(parts, "$")
}

func typeListKey(ts []types.Type) string {
	var parts []string
	for _, t := range ts {
		parts = append(parts, typeKey(t))
	}
	return strings.Join(parts, ",")
}

func typeArgs(l *types.TypeList) []types.Type {
	if l == nil {
		return nil
	}
	return slices.Collect(l.Types())
}

// funcName requests the translation of the function or method for the type arguments and returns its name.
func (e *emitter) funcName(origin *types.Func, recvArgs, targs []types.Type) string {
	origin = origin.Origin()
	key := fmt.Sprintf("F:%p[%s][%s]", origin, typeListKey(recvArgs), typeListKey(targs))
	if n, ok := e.names[key]; ok {
		return n
	}

	src, ok := e.decls[origin]
	if !ok {
		fail(origin.Pos(), "%s has no declaration", origin.FullName())
	}

	base := e.pkgPrefix(origin.Pkg()) + "$"
	if recv := origin.Signature().Recv(); recv != nil {
		base += recvTypeName(recv.Type()) + "$"
		if len(recvArgs) > 0 {
			base += "$" + typeListBase(recvArgs) + "$"
		}
	}
	base += origin.Name()
	if len(targs) > 0 {
		base += "$$" + typeListBase(targs)
	}

	name := e.ident(base)
	e.names[key] = name
	e.queue = append(e.queue, func() { e.emitFunc(name, src, recvArgs, targs) })
	return name
}

func recvTypeName(t types.Type) string {
	if p, ok := t.(*types.Pointer); ok {
		t = p.Elem()
	}
	if n, ok := types.Unalias(t).(*types.Named); ok {
		return n.Obj().Name()
	}
	return "recv"
}

// externName returns the name of the binding of a function of a header package.
func (e *emitter) externName(fn *types.Func, targs []types.Type, src funcSrc, pos token.Pos) string {
	full := externKey(fn, targs)
	key := "X:" + full
	if n, ok := e.names[key]; ok {
		return n
	}
	if src.js == "" {
		fail(pos, "%s has no //gostrict:js binding", fn.FullName())
	}
	fsig := fn.Signature()
	if len(targs) > 0 {
		inst, err := types.Instantiate(e.tctx, fn.Origin().Type(), targs, true)
		if err != nil {
			fail(pos, "cannot instantiate %s: %v", fn.FullName(), err)
		}
		fsig = inst.(*types.Signature)
	}
	sig, ok := signature(full, fsig)
	if !ok {
		fail(pos, "%s has values, which can not cross the boundary", fn.FullName())
	}

	base := e.pkgPrefix(fn.Pkg()) + "$" + fn.Name()
	if len(targs) > 0 {
		base += "$$" + typeListBase(targs)
	}
	name := e.ident(base)
	e.names[key] = name
	e.extern = append(e.extern, Extern{Func: sig, JS: src.js})
	fmt.Fprintf(&e.externs, "const %s = $host[%s] !== void 0 ? $host[%s] : (%s\n);\n", name, jsString(full), jsString(full), src.js)
	return name
}

// isStruct tells, whether values of t are structs with value semantics. list.List and dict.Map are structs in Go,
// but have reference semantics.
func isStruct(t types.Type) bool {
	if isOpaque(t) || isHandle(t) {
		return false
	}
	_, ok := t.Underlying().(*types.Struct)
	return ok
}

func isInterface(t types.Type) bool {
	if _, ok := t.(*types.TypeParam); ok {
		return false
	}
	_, ok := t.Underlying().(*types.Interface)
	return ok
}

func isPointer(t types.Type) bool {
	_, ok := t.Underlying().(*types.Pointer)
	return ok
}

// structName requests the class and helpers of the struct type and returns the name of the class. The helpers
// are named after the class: $z returns the zero value, $c copies, $s copies into an existing value and $e
// compares.
func (e *emitter) structName(t types.Type) string {
	n, ok := types.Unalias(t).(*types.Named)
	if !ok {
		fail(token.NoPos, "anonymous struct type %s", t)
	}

	key := "S:" + typeKey(n)
	if name, ok := e.names[key]; ok {
		return name
	}

	base := e.pkgPrefix(n.Obj().Pkg()) + "$" + n.Obj().Name()
	if n.TypeArgs().Len() > 0 {
		base += "$$" + typeListBase(typeArgs(n.TypeArgs()))
	}
	name := e.ident(base)
	e.names[key] = name

	st := n.Underlying().(*types.Struct)
	var fields, zeros, clones, sets, eqs []string
	for i := 0; i < st.NumFields(); i++ {
		f := st.Field(i)
		fn := fieldName(f, i)
		ft := f.Type()
		fields = append(fields, fn)
		zeros = append(zeros, e.zero(ft))
		clones = append(clones, e.clone("x."+fn, ft))
		if isStruct(ft) {
			sets = append(sets, fmt.Sprintf("%s$s(d.%s, x.%s);", e.structName(ft), fn, fn))
		} else {
			sets = append(sets, fmt.Sprintf("d.%s = x.%s;", fn, fn))
		}
		eqs = append(eqs, e.eq("x."+fn, "y."+fn, ft))
	}

	eq := "true"
	if len(eqs) > 0 {
		eq = strings.Join(eqs, " && ")
	}

	var ctor strings.Builder
	for _, f := range fields {
		fmt.Fprintf(&ctor, "this.%s = %s; ", f, f)
	}

	fmt.Fprintf(&e.out, "class %s { constructor(%s) { %s} }\n", name, strings.Join(fields, ", "), ctor.String())
	fmt.Fprintf(&e.out, "function %s$z() { return new %s(%s); }\n", name, name, strings.Join(zeros, ", "))
	fmt.Fprintf(&e.out, "function %s$c(x) { return new %s(%s); }\n", name, name, strings.Join(clones, ", "))
	fmt.Fprintf(&e.out, "function %s$s(d, x) { %s return d; }\n", name, strings.Join(sets, " "))
	fmt.Fprintf(&e.out, "function %s$e(x, y) { return %s; }\n", name, eq)
	return name
}

func fieldName(f *types.Var, i int) string {
	if f.Name() == "_" {
		return fmt.Sprintf("$_%d", i)
	}
	if f.Name() == "__proto__" {
		return "$__proto__"
	}
	return f.Name()
}

// zero returns an expression, which creates the zero value of t.
func (e *emitter) zero(t types.Type) string {
	if isStruct(t) {
		return e.structName(t) + "$z()"
	}
	switch u := t.Underlying().(type) {
	case *types.Basic:
		switch {
		case u.Info()&types.IsString != 0:
			return `""`
		case u.Info()&types.IsBoolean != 0:
			return "false"
		case u.Info()&types.IsNumeric != 0:
			return "0"
		}
	}
	return "null"
}

// zeroFn returns a function, which creates the zero value of t.
func (e *emitter) zeroFn(t types.Type) string {
	if isStruct(t) {
		return e.structName(t) + "$z"
	}
	return "(() => " + e.zero(t) + ")"
}

// clone returns an expression, which copies the value of x, if t has value semantics.
func (e *emitter) clone(x string, t types.Type) string {
	if isStruct(t) {
		return e.structName(t) + "$c(" + x + ")"
	}
	return x
}

func (e *emitter) cloneFn(t types.Type) string {
	if isStruct(t) {
		return e.structName(t) + "$c"
	}
	return "$id"
}

// eq returns an expression, which compares a and b of type t like == of Go.
func (e *emitter) eq(a, b string, t types.Type) string {
	if isStruct(t) {
		return e.structName(t) + "$e(" + a + ", " + b + ")"
	}
	return "(" + a + " === " + b + ")"
}

// methodTable requests the methods of t, which an interface value of a t calls, and returns its name. A value
// receiver gets a copy, so a method can not change the value of the interface.
func (e *emitter) methodTable(t types.Type) string {
	key := "MT:" + typeKey(t)
	if n, ok := e.names[key]; ok {
		return n
	}

	name := e.ident(e.pkgPrefixOf(t) + "$" + sanitize(typeBase(t)) + "$mt")
	e.names[key] = name

	named, _ := types.Unalias(t).(*types.Named)
	if p, ok := t.(*types.Pointer); ok {
		named, _ = types.Unalias(p.Elem()).(*types.Named)
	}

	var entries []string
	mset := types.NewMethodSet(t)
	for sel := range mset.Methods() {
		fn := sel.Obj().(*types.Func)
		if fn.Signature().TypeParams().Len() > 0 || named == nil {
			continue
		}
		fname := e.funcName(fn.Origin(), typeArgs(named.TypeArgs()), nil)
		_, ptrRecv := fn.Origin().Signature().Recv().Type().(*types.Pointer)
		if !ptrRecv && isStruct(named) {
			fname = fmt.Sprintf("((r, ...a) => %s(%s, ...a))", fname, e.clone("r", named))
		}
		entries = append(entries, fmt.Sprintf("%s: %s", fn.Name(), fname))
	}

	fmt.Fprintf(&e.out, "const %s = { %s };\n", name, strings.Join(entries, ", "))
	return name
}

func (e *emitter) pkgPrefixOf(t types.Type) string {
	if p, ok := t.(*types.Pointer); ok {
		t = p.Elem()
	}
	if n, ok := types.Unalias(t).(*types.Named); ok && n.Obj().Pkg() != nil {
		return e.pkgPrefix(n.Obj().Pkg())
	}
	return "$builtin"
}

func sortedNames[V any](m map[string]V) []string {
	var names []string
	for n := range m {
		names = append(names, n)
	}
	slices.Sort(names)
	return names
}
