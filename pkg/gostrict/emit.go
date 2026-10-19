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
	"maps"
	"slices"
	"strings"

	"go.wdy.de/nago/pkg/gostrict/internal/vmcode"
)

//go:embed runtime.js
var runtimeJS string

// Bundle is a translated program. Its script evaluates to a function, which takes an object with the host
// functions by their full name and returns an object with the exported functions of the main package. Its
// property $externs holds the functions of header packages by their full name, as the host provided them or
// as their JavaScript binding, and $new the constructors of the records, which take the fields in their order.
type Bundle struct {
	JS string
	// SourceMap maps the lines of JS to the lines of the Go source in the format of version 3, including the
	// sources, so that a browser shows and debugs them.
	SourceMap []byte
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
	// States maps the full name of each state of the exports to the hash of its structure, which is empty, if a new
	// version of the program can not take it over, see [Bundle.CanTakeOver].
	States map[string]string
	// Positions are the positions in the Go source, where the program creates errors. An error of the program
	// refers to one by its index, which an engine resolves with [Bundle.ProgramError].
	Positions []Position
	// vm is the program of the interpreter of package engine/vm, or vmErr, why it could not be compiled
	vm    *vmcode.Program
	vmErr error
}

// Extern is a function of a header package.
type Extern struct {
	Func
	// JS is the JavaScript binding, a function expression, or empty, if only the host implements the function, see
	// the directive //gostrict:host. Then an engine needs the function of the host.
	JS string
}

// MissingHost returns the first function of a header package, which has no JavaScript binding and which the host
// does not implement, or false.
func (b *Bundle) MissingHost(host Funcs) (string, bool) {
	for _, x := range b.Externs {
		if _, ok := host[x.Name]; !ok && x.JS == "" {
			return x.Name, true
		}
	}
	return "", false
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
	return EmitEntries(prog)
}

// EmitEntries translates a checked whole program, see [LoadRoots], for the given entry points by their full name,
// e.g. the use case, which a host runs, or for all entry points, if there are none. The bundle contains only the
// packages and functions, which they reach, so it is small and loads fast, while the program is checked once.
func EmitEntries(prog *Program, entries ...string) (b *Bundle, diags Diagnostics) {
	e := &emitter{
		prog:    prog,
		decls:   funcDecls(prog),
		tctx:    types.NewContext(),
		names:   map[string]string{},
		used:    map[string]bool{},
		prefix:  map[*types.Package]string{},
		globals: map[*types.Var]string{},
		diag:    &diagnoser{fset: prog.Fset, sources: prog.sources},
		mut:     analyzeMutation(prog),
		entries: entryPoints(prog),
		// the initializers of the package variables may call functions of header packages already
		bnd: newBoundary(),
	}
	if len(entries) > 0 {
		e.targets = map[string]bool{}
		for _, name := range entries {
			if !slices.Contains(slices.Collect(maps.Values(e.entries)), name) {
				return nil, Diagnostics{{Code: CodeEntry, Msg: name + " is no entry point of the program", Why: "a bundle exports only entry points, the exported functions of the roots of a whole program"}}
			}
			e.targets[name] = true
		}
	}

	defer func() {
		if r := recover(); r != nil {
			err, ok := r.(emitError)
			if !ok {
				// the programs come from users or language models, so a defect of the compiler must not end the host
				err = emitError{pos: token.NoPos, msg: fmt.Sprintf("the compiler failed: %v", r)}
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
	// bnd collects the records of the exports and of the functions of header packages
	bnd *boundary
	// readonlyCache keeps the parameters of functions, which they only read, see readonly
	readonlyCache map[*types.Func]readonlyInfo
	// positions are the positions of the marks of the source map, see mark
	positions []token.Pos
	// errPositions are the positions, where the program creates errors, see errorPos
	errPositions []token.Pos
	// mut tells, which struct types the program never changes, so that their values need no copy
	mut *mutation
	// hvars are the bindings of the variables of header packages, see headerVars
	hvars map[*types.Var]string
	// entries are the entry points of a whole program by their full name, see LoadRoots
	entries map[*types.Func]string
	// targets are the entry points, which the bundle exports, or nil for all
	targets map[string]bool
	// dynamic are the names of the methods, which the program calls through an interface, see dynamicMethods
	dynamic map[string]bool
}

// entryPoints returns the exported functions of the roots of a whole program.
func entryPoints(prog *Program) map[*types.Func]string {
	res := map[*types.Func]string{}
	for _, p := range prog.Roots {
		for _, name := range p.Types.Scope().Names() {
			if fn, ok := p.Types.Scope().Lookup(name).(*types.Func); ok && fn.Exported() {
				res[fn] = p.Path + "." + name
			}
		}
	}
	return res
}

// reached returns the packages, which the bundle needs: all, or those, which the packages of the targets import.
func (e *emitter) reached() map[*Package]bool {
	res := map[*Package]bool{}
	var visit func(p *types.Package)
	visit = func(p *types.Package) {
		pkg := e.prog.Package(p.Path())
		if pkg == nil || res[pkg] {
			return
		}
		res[pkg] = true
		for _, imp := range p.Imports() {
			visit(imp)
		}
	}
	for fn, name := range e.entries {
		if e.targets == nil || e.targets[name] {
			visit(fn.Pkg())
		}
	}
	if len(e.entries) == 0 {
		for _, p := range e.prog.Packages {
			res[p] = true
		}
	}
	return res
}

// hook returns the function, which calls the entry point fn of a whole program within the enter and exit of the
// aspect, see [Aspect]. The host and the other packages call the entry point only through it.
func (e *emitter) hook(fn *types.Func) string {
	key := "H:" + e.entries[fn]
	if n, ok := e.names[key]; ok {
		return n
	}
	inner := e.funcName(fn, nil, nil)
	name := e.ident(inner + "$h")
	e.names[key] = name

	sig := fn.Signature()
	params := make([]string, sig.Params().Len())
	for i := range params {
		params[i] = fmt.Sprintf("a%d", i)
	}
	// the error of the entry point tells the aspect, whether it failed
	errResult := "null"
	if n := sig.Results().Len(); n > 0 && isError(sig.Results().At(n-1).Type()) {
		errResult = "$r"
		if n > 1 {
			errResult = fmt.Sprintf("$r[%d]", n-1)
		}
	}
	full := jsString(e.entries[fn])
	fmt.Fprintf(&e.out, "function %s(%s) {\n\t$aspect_enter(%s);\n\tlet $r, $done = false;\n\ttry {\n\t\t$r = %s(%s);\n\t\t$done = true;\n\t} finally {\n\t\t$aspect_exit(%s, $done ? %s : $errors_New(\"aborted\"));\n\t}\n\treturn $r;\n}\n",
		name, strings.Join(params, ", "), full, inner, strings.Join(params, ", "), full, errResult)
	return name
}

func (e *emitter) emit() (*Bundle, Diagnostics) {
	var vars, init strings.Builder
	reached := e.reached()
	for _, p := range e.prog.Packages {
		if p.Kind != KindUser || !reached[p] {
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
	exportFuncs := map[string]*types.Func{}
	bnd := e.bnd
	var ret strings.Builder
	if len(e.entries) > 0 {
		// the host calls the entry points of a whole program by their full name within the aspect
		e.extern = append(e.extern, Extern{Func: Func{Name: EnterFunc, Params: []Type{String}}}, Extern{Func: Func{Name: ExitFunc, Params: []Type{String, Error}}})
		fmt.Fprintf(&e.externs, "const $aspect_enter = $host[%s];\nconst $aspect_exit = $host[%s];\n", jsString(EnterFunc), jsString(ExitFunc))
		e.names["X:"+EnterFunc], e.names["X:"+ExitFunc] = "$aspect_enter", "$aspect_exit"
		for _, fn := range sortedFuncs(e.entries) {
			full := e.entries[fn]
			if e.targets != nil && !e.targets[full] {
				continue
			}
			sig, ok := bnd.exportSignature(full, fn.Signature())
			if !ok {
				continue
			}
			exports[full] = sig
			exportFuncs[full] = fn
			fmt.Fprintf(&ret, "\t%s: %s,\n", jsString(full), e.hook(fn))
		}
	}
	main := e.prog.Main
	for _, name := range main.Types.Scope().Names() {
		if len(e.entries) > 0 {
			break
		}
		fn, ok := main.Types.Scope().Lookup(name).(*types.Func)
		if !ok || !fn.Exported() {
			continue
		}
		sig, ok := bnd.exportSignature(name, fn.Signature())
		if !ok {
			continue
		}
		exports[name] = sig
		exportFuncs[name] = fn
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
		if len(e.queue) == 0 {
			// the records of the functions of header packages are found with their calls
			for _, name := range sortedNames(bnd.records) {
				e.structName(bnd.structs[name])
			}
		}
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
		fmt.Fprintf(&sb, "\t\t%s: %s,\n", jsString(name), cls)
	}
	sb.WriteString("\t},\n")
	sb.WriteString("\t$unions: {\n")
	for _, name := range sortedNames(bnd.unions) {
		fmt.Fprintf(&sb, "\t\t%s: {\n", jsString(name))
		for _, v := range bnd.unions[name].Variants {
			mt := e.methodTable(bnd.structs[v])
			if bnd.unions[name].Handles {
				fmt.Fprintf(&sb, "\t\t\t%s: [%s, (h) => $box(h, %s)],\n", jsString(v), mt, mt)
				continue
			}
			fmt.Fprintf(&sb, "\t\t\t%s: [%s, (...f) => $box(%s(...f), %s)],\n", jsString(v), mt, e.structName(bnd.structs[v]), mt)
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
	for name, rec := range bnd.records {
		rec.Immutable = e.frozen(bnd.structs[name], bnd.records, map[string]bool{})
		bnd.records[name] = rec
	}
	js, sourceMap := e.stripMarks(sb.String())
	sum := sha256.Sum256([]byte(js))
	states := map[string]string{}
	for _, f := range exports {
		for _, t := range append(slices.Clone(f.Params), f.Results...) {
			if t.Kind == ValueState {
				states[t.Name] = stateHash(bnd.structs[t.Name])
			}
		}
	}
	hs := handlers(exports, bnd.unions)
	if len(e.entries) > 0 {
		// the entry points of a whole program are called by name and handle no messages of a browser
		hs = map[string]string{}
	}
	b := &Bundle{JS: js, SourceMap: sourceMap, Hash: hex.EncodeToString(sum[:16]), Exports: exports, Records: bnd.records, Unions: bnd.unions, Handlers: hs, Externs: e.extern, States: states, Positions: e.errorPositions()}
	b.vm, b.vmErr = e.emitVM(b, exportFuncs)
	return b, nil
}

// frozen tells, whether the values of the record t never change: its type is immutable and its fields hold only
// basic values and such records, but no list or map, which the program may change through any copy.
func (e *emitter) frozen(t types.Type, records map[string]Record, seen map[string]bool) bool {
	name := typeKey(t)
	if seen[name] {
		return true
	}
	seen[name] = true
	if !e.mut.immutable(t) {
		return false
	}
	for _, f := range records[name].Fields {
		switch f.Type.Kind {
		case ValueRecord:
			ft, ok := e.bnd.structs[f.Type.Name]
			if !ok || !e.frozen(ft, records, seen) {
				return false
			}
		case ValueBool, ValueInt32, ValueUint32, ValueFloat64, ValueString, ValueBig:
		default:
			return false
		}
	}
	return true
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
	if js, ok := e.headerVars()[v]; ok {
		// a variable of a header package is the value of its binding
		fmt.Fprintf(&e.externs, "const %s = (%s\n);\n", n, js)
	}
	return n
}

// headerVars returns the bindings of the exported variables of the header packages.
func (e *emitter) headerVars() map[*types.Var]string {
	if e.hvars != nil {
		return e.hvars
	}
	e.hvars = map[*types.Var]string{}
	for _, p := range e.prog.Packages {
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
						if v, ok := p.Info.Defs[n].(*types.Var); ok {
							e.hvars[v] = varBinding(g, vs)
						}
					}
				}
			}
		}
	}
	return e.hvars
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
	if src.js == "" && !src.host {
		fail(pos, "%s has no //gostrict:js binding", fn.FullName())
	}
	fsig := externSignature(fn)
	if len(targs) > 0 {
		inst, err := types.Instantiate(e.tctx, fn.Origin().Type(), targs, true)
		if err != nil {
			fail(pos, "cannot instantiate %s: %v", fn.FullName(), err)
		}
		fsig = inst.(*types.Signature)
	}
	sig, ok := e.bnd.headerSignature(full, fsig)
	if !ok {
		fail(pos, "%s has values, which can not cross the boundary", fn.FullName())
	}

	base := e.pkgPrefix(fn.Pkg()) + "$" + fn.Name()
	if recv := fn.Signature().Recv(); recv != nil {
		base = e.pkgPrefix(fn.Pkg()) + "$" + recvTypeName(recv.Type()) + "$" + fn.Name()
	}
	if len(targs) > 0 {
		base += "$$" + typeListBase(targs)
	}
	name := e.ident(base)
	e.names[key] = name
	e.extern = append(e.extern, Extern{Func: sig, JS: src.js})
	if src.host {
		// an engine refuses a host, which does not implement it
		fmt.Fprintf(&e.externs, "const %s = $host[%s];\n", name, jsString(full))
		return name
	}
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
		if isStruct(ft) && !e.mut.immutable(ft) {
			sets = append(sets, fmt.Sprintf("%s$s(d.%s, x.%s);", e.structName(ft), fn, fn))
		} else {
			// a value of an immutable type may be shared, so it is replaced and not changed
			sets = append(sets, fmt.Sprintf("d.%s = x.%s;", fn, fn))
		}
		eqs = append(eqs, e.eq("x."+fn, "y."+fn, ft))
	}

	eq := "true"
	if len(eqs) > 0 {
		eq = strings.Join(eqs, " && ")
	}

	// the record is an object, which the function of its name creates from the fields in their order
	params := make([]string, len(fields))
	for i := range params {
		params[i] = fmt.Sprintf("a%d", i)
	}
	fmt.Fprintf(&e.out, "function %s(%s) { return %s; }\n", name, strings.Join(params, ", "), objectLit(st, params))
	if e.mut.immutable(n) {
		// the program never changes a value of the type, so all zero values are the same
		fmt.Fprintf(&e.out, "let %s$Z;\nfunction %s$z() { return %s$Z || (%s$Z = %s); }\n", name, name, name, name, objectLit(st, zeros))
	} else {
		fmt.Fprintf(&e.out, "function %s$z() { return %s; }\n", name, objectLit(st, zeros))
	}
	fmt.Fprintf(&e.out, "function %s$c(x) { return %s; }\n", name, objectLit(st, clones))
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
	if isStruct(t) && !e.mut.immutable(t) {
		return e.structName(t) + "$c(" + x + ")"
	}
	return x
}

func (e *emitter) cloneFn(t types.Type) string {
	if isStruct(t) && !e.mut.immutable(t) {
		return e.structName(t) + "$c"
	}
	return "$id"
}

// eq returns an expression, which compares a and b of type t like == of Go.
func (e *emitter) eq(a, b string, t types.Type) string {
	if isStruct(t) {
		return e.structName(t) + "$e(" + a + ", " + b + ")"
	}
	if isInterface(t) && isSealed(t) {
		return e.unionEq(t) + "(" + a + ", " + b + ")"
	}
	return "(" + a + " === " + b + ")"
}

// unionEq requests the function, which compares two values of a sealed interface: they are equal, if they have the
// same variant and equal values of it.
func (e *emitter) unionEq(t types.Type) string {
	key := "UE:" + typeKey(t)
	if n, ok := e.names[key]; ok {
		return n
	}
	name := e.ident(e.pkgPrefixOf(t) + "$" + sanitize(typeBase(t)) + "$ueq")
	e.names[key] = name

	variants, _ := sealedVariants(t)
	var sb strings.Builder
	fmt.Fprintf(&sb, "function %s(a, b) {\n\tif (a.m !== b.m) return false;\n", name)
	for _, v := range variants {
		if isStruct(v) || (isInterface(v) && isSealed(v)) {
			fmt.Fprintf(&sb, "\tif (a.m === %s) return %s;\n", e.methodTable(v), e.eq("a.v", "b.v", v))
		}
	}
	sb.WriteString("\treturn a.v === b.v;\n}\n")
	e.out.WriteString(sb.String())
	return name
}

// dynamicMethods returns the names of the methods, which the program may call through an interface, and Error,
// which an engine calls to convert an error. A method table needs no other method, e.g. the builders of the views
// of lib/ui, which the program calls only statically, so that a bundle binds only the functions of the host, which
// it calls.
func (e *emitter) dynamicMethods() map[string]bool {
	if e.dynamic != nil {
		return e.dynamic
	}
	e.dynamic = map[string]bool{"Error": true}
	for _, p := range e.prog.Packages {
		if p.Kind != KindUser {
			continue
		}
		for _, sel := range p.Info.Selections {
			fn, ok := sel.Obj().(*types.Func)
			if !ok || sel.Kind() == types.FieldVal {
				continue
			}
			if types.IsInterface(sel.Recv()) || types.IsInterface(fn.Signature().Recv().Type()) {
				e.dynamic[fn.Name()] = true
			}
		}
	}
	return e.dynamic
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
	if named != nil && isHandle(named) {
		// the table of a handle names its type, e.g. a variant of the View of lib/ui, which a host receives
		entries = append(entries, "$tag: "+jsString(typeKey(named)))
	}
	mset := types.NewMethodSet(t)
	dynamic := e.dynamicMethods()
	for sel := range mset.Methods() {
		fn := sel.Obj().(*types.Func)
		if fn.Signature().TypeParams().Len() > 0 || named == nil || !dynamic[fn.Name()] {
			continue
		}
		if src, ok := e.decls[fn.Origin()]; ok && src.pkg.Kind == KindHeader {
			if fn.Exported() {
				entries = append(entries, fmt.Sprintf("%s: %s", fn.Name(), e.externName(fn.Origin(), nil, src, fn.Pos())))
			}
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

// sortedFuncs returns the functions of the map in the order of their names, so that a bundle is deterministic.
func sortedFuncs(m map[*types.Func]string) []*types.Func {
	res := slices.Collect(maps.Keys(m))
	slices.SortFunc(res, func(a, b *types.Func) int { return strings.Compare(m[a], m[b]) })
	return res
}
