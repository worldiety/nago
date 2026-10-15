// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"path"
	"slices"
	"strings"
	"unicode/utf8"
)

// Check checks the packages of the program, which are not trusted, against the subset. It returns the
// diagnostics of [Program.Diagnostics] as well, so an empty result means, that the program can be translated.
func Check(prog *Program) Diagnostics {
	d := &diagnoser{fset: prog.Fset, sources: prog.sources}
	d.diags = append(d.diags, prog.Diagnostics...)

	decls := funcDecls(prog)
	for _, p := range prog.Packages {
		if p.Kind == KindHeader && !strings.HasPrefix(p.Path, stdPrefix) {
			checkHeader(p, d, decls)
		}
		if p.Kind != KindUser {
			continue
		}
		c := &checker{
			pkg:     p,
			prog:    prog,
			info:    p.Info,
			d:       d,
			decls:   decls,
			assigns: collectAssignments(p.Info, p.Files),
			pairs:   map[types.Object]types.Object{},
		}
		for _, f := range p.Files {
			c.checkImports(f)
			c.checkDefs(f)
			c.checkGlobals(f)
			c.collectPairs(f)
			ast.Inspect(f, c.visit)
			c.collectFixes(f)
			if prog.terminating(p) {
				c.checkLoops(f)
			}
		}
	}

	if prog.Main.Kind == KindUser {
		checkHandlers(prog.Main, d)
	}
	checkRecursion(prog, d, decls)
	d.attachEdits()
	return dropFollowUps(d.sorted())
}

// dropFollowUps removes the reports of a forbidden type on a line, whose expression is rejected anyway, e.g. the
// int of n := len(s), whose fix replaces the expression.
func dropFollowUps(diags Diagnostics) Diagnostics {
	causes := map[string]bool{}
	for _, d := range diags {
		switch d.Code {
		case CodeString, CodeBuiltin, CodeSlice, CodeStdFunc:
			causes[fmt.Sprintf("%s:%d", d.File, d.Line)] = true
		}
	}
	var res Diagnostics
	for _, d := range diags {
		if d.Code == CodeNumber && causes[fmt.Sprintf("%s:%d", d.File, d.Line)] {
			continue
		}
		res = append(res, d)
	}
	return res
}

type checker struct {
	pkg     *Package
	prog    *Program
	info    *types.Info
	d       *diagnoser
	decls   map[*types.Func]funcSrc
	stack   []ast.Node
	assigns map[types.Object][]assignment
	// pairs maps the value of v, ok := ... to ok, which must be checked, before v is used.
	pairs map[types.Object]types.Object
}

func (c *checker) report(n ast.Node, code Code, msg, why, fix string) {
	c.d.report(n.Pos(), n.End(), code, msg, why, fix)
}

func (c *checker) checkImports(f *ast.File) {
	for _, imp := range f.Imports {
		p := strings.Trim(imp.Path.Value, `"`)
		if imp.Name != nil && (imp.Name.Name == "." || imp.Name.Name == "_") {
			c.report(imp, CodeImportForm, "dot and blank imports are not allowed", "every name must be qualified, so that a reader and the compiler see where it comes from", `import "`+p+`" and use `+path.Base(p)+".Name")
		}

		if p == jsPath {
			c.report(imp, CodeImport, "package js is only for header packages", "a handle refers to an object of the host, which only a header package creates", "call the functions of a header package, which return handles")
			continue
		}
		if !isStdPath(p) {
			continue
		}
		if _, ok := stdAllowed[p]; ok {
			continue
		}

		why := "the package is not part of the subset, which allows only math, strings, strconv, errors and iter of the standard library"
		fix := `use the packages "go.wdy.de/nago/pkg/gostrict/std/list", ".../dict" and ".../text"`
		if h, ok := importHints[p]; ok {
			why, fix = h.why, h.fix
		}
		c.report(imp, CodeImport, "package "+p+" is not allowed", why, fix)
	}
}

// checkDefs checks the types of all declared variables, fields, parameters and types.
func (c *checker) checkDefs(f *ast.File) {
	// parameters and results are checked with their function type
	params := map[*ast.Ident]bool{}
	ast.Inspect(f, func(n ast.Node) bool {
		if ft, ok := n.(*ast.FuncType); ok {
			for _, list := range []*ast.FieldList{ft.Params, ft.Results} {
				if list != nil {
					for _, field := range list.List {
						for _, name := range field.Names {
							params[name] = true
						}
					}
				}
			}
		}
		return true
	})

	ast.Inspect(f, func(n ast.Node) bool {
		id, ok := n.(*ast.Ident)
		if !ok || id.Name == "_" || params[id] {
			return true
		}
		obj := c.info.Defs[id]
		if obj == nil {
			return true
		}

		switch obj := obj.(type) {
		case *types.Var:
			c.checkType(id, obj.Type())
		case *types.Const:
			c.checkType(id, obj.Type())
		case *types.TypeName:
			if obj.IsAlias() {
				c.checkType(id, obj.Type())
				return true
			}
			if _, ok := obj.Type().(*types.TypeParam); ok {
				return true
			}
			switch u := obj.Type().Underlying().(type) {
			case *types.Struct:
				// the fields are checked as variables
			case *types.Interface:
				// methods are checked as functions
			default:
				c.checkType(id, u)
			}
		}
		return true
	})

	for id, inst := range c.info.Instances {
		if c.fileOf(id) != f {
			continue
		}
		for t := range inst.TypeArgs.Types() {
			c.checkType(id, t)
		}
	}
}

func (c *checker) fileOf(n ast.Node) *ast.File {
	for _, f := range c.pkg.Files {
		if f.FileStart <= n.Pos() && n.Pos() <= f.FileEnd {
			return f
		}
	}
	return nil
}

func (c *checker) visit(n ast.Node) bool {
	if n == nil {
		c.stack = c.stack[:len(c.stack)-1]
		return false
	}

	if e, ok := n.(ast.Expr); ok {
		if tv, ok := c.info.Types[e]; ok && tv.Value != nil {
			// a constant is translated as its value, so its operations behave the same anyway
			c.checkConstant(e, tv)
			return false
		}
	}

	c.stack = append(c.stack, n)
	c.visitNode(n)
	return true
}

func (c *checker) visitNode(n ast.Node) {
	switch n := n.(type) {
	case *ast.GoStmt:
		c.report(n, CodeConcurrency, "go statements are not allowed", "the subset runs single threaded, like JavaScript", "call the function directly")
	case *ast.SelectStmt:
		c.report(n, CodeConcurrency, "select is not allowed", "the subset has no channels", "call functions directly")
	case *ast.SendStmt:
		c.report(n, CodeConcurrency, "channels are not allowed", "the subset has no channels", "use a list.List as a queue")
	case *ast.DeferStmt:
		c.report(n, CodeControl, "defer is not allowed", "the subset has no deferred calls", "call the function before each return")
	case *ast.BranchStmt:
		c.checkBranch(n)
	case *ast.TypeSwitchStmt:
		c.checkTypeSwitch(n)
	case *ast.SwitchStmt:
		c.checkEnumSwitch(n)
	case *ast.TypeAssertExpr:
		if n.Type != nil {
			c.report(n, CodeTypeSwitch, "type assertions are not allowed", "an assertion fails at runtime, if the value has another type", "use a type switch over a sealed interface, which the compiler checks, or declare a method in the interface")
		}
	case *ast.StructType:
		for _, f := range n.Fields.List {
			if len(f.Names) == 0 {
				c.report(f, CodeEmbedding, "embedded fields are not allowed", "promoted fields and methods are not part of the subset", "declare a named field, e.g. Base Base, and call its methods explicitly")
			}
		}
	case *ast.FuncDecl:
		if n.Recv == nil && n.Name.Name == "init" {
			c.report(n.Name, CodeInit, "init functions are not allowed", "the order of initialization is part of the program and must be visible", "call a setup function explicitly, or initialize the package variable with a function call")
		}
	case *ast.FuncType:
		c.checkFuncType(n)
		c.checkNamedResults(n)
	case *ast.Ident:
		c.checkIdent(n)
	case *ast.ValueSpec:
		c.checkValueSpec(n)
	case *ast.IncDecStmt:
		c.checkGlobalWrite(n.X)
	case *ast.AssignStmt:
		if n.Tok != token.DEFINE {
			for _, l := range n.Lhs {
				c.checkGlobalWrite(l)
			}
		}
		switch n.Tok {
		case token.QUO_ASSIGN, token.REM_ASSIGN:
			c.checkDivision(n, n.Lhs[0], n.Rhs[0])
		case token.SHL_ASSIGN, token.SHR_ASSIGN:
			c.checkShift(n, n.Rhs[0])
		}
	case *ast.DeclStmt:
		if g, ok := n.Decl.(*ast.GenDecl); ok && g.Tok == token.TYPE {
			c.report(n, CodeLocalType, "types must be declared at package level", "types declared inside functions are not supported", "move the type declaration out of the function")
		}
	case *ast.UnaryExpr:
		c.checkUnary(n)
	case *ast.BinaryExpr:
		c.checkBinary(n)
	case *ast.IndexExpr:
		if c.isString(n.X) {
			c.report(n, CodeString, "indexing a string is not allowed", "Go indexes the bytes of UTF-8, JavaScript the units of UTF-16", "use c, ok := text.Chars(s).Get(i) for the i-th character or text.CodePoints(s).Get(i) for its code point")
		}
	case *ast.SliceExpr:
		if c.isString(n.X) {
			c.report(n, CodeString, "slicing a string is not allowed", "Go slices the bytes of UTF-8, JavaScript the units of UTF-16", "split the string with text.Split or text.Chars and join the parts with text.Join")
		} else {
			c.report(n, CodeSlice, "slice expressions are not allowed", "the subset replaces slices by list.List", "copy the elements into a new list.List")
		}
	case *ast.CallExpr:
		c.checkCall(n)
		c.checkErrorCall(n)
		c.checkMake(n)
		c.checkGlobalMethod(n)
	case *ast.SelectorExpr:
		c.checkSelector(n)
	case *ast.RangeStmt:
		c.checkRange(n)
	case *ast.CompositeLit:
		c.checkType(n, c.info.TypeOf(n))
		c.checkMissingFields(n)
		if t := c.info.TypeOf(n); isOption(t) {
			c.report(n, CodeOpaqueLiteral, "composite literals of opt.Option are not allowed", "an option is created by its functions", "use opt.Some(v) or opt.None[T]()")
		} else if isOpaque(t) {
			c.report(n, CodeOpaqueLiteral, "composite literals of list.List and dict.Map are not allowed", "the zero value can not be appended to", "use list.New[T](), list.Of(a, b) or dict.New[K, V]()")
		}
	case *ast.ChanType:
		c.report(n, CodeConcurrency, "channels are not allowed", "the subset has no channels", "use a list.List as a queue")
	case *ast.Ellipsis:
		if n.Elt == nil {
			c.report(n, CodeSlice, "spreading a slice is not allowed", "the subset replaces slices by list.List", "pass the elements one by one or pass the list.List")
		}
	}
}

func (c *checker) checkConstant(e ast.Expr, tv types.TypeAndValue) {
	if !c.stdIntArg(e) {
		c.checkType(e, tv.Type)
	}
	switch tv.Value.Kind() {
	case constant.String:
		if !utf8.ValidString(constant.StringVal(tv.Value)) {
			c.report(e, CodeInvalidUTF8, "the string is not valid UTF-8", "JavaScript strings can not hold arbitrary bytes", "use only valid characters, e.g. \\u00e4 instead of \\xc3")
		}
	case constant.Int:
		if b, ok := tv.Type.Underlying().(*types.Basic); ok && b.Kind() == types.Int {
			if v, exact := constant.Int64Val(tv.Value); !exact || v > maxSafeInt || v < -maxSafeInt {
				c.report(e, CodeUnsafeInt, "the integer constant exceeds the safe range of JavaScript", "an int is a float64 in JavaScript and exact only up to ±2^53-1", "use a smaller value or a float64")
			}
		}
	}
}

const maxSafeInt = 1<<53 - 1

// stdIntArg tells, whether the constant is an argument of an int parameter of an allowed function of the standard
// library, e.g. the sign of math.IsInf(x, 0), which the runtime implements, so the int never becomes a variable.
func (c *checker) stdIntArg(e ast.Expr) bool {
	if len(c.stack) == 0 {
		return false
	}
	call, ok := c.stack[len(c.stack)-1].(*ast.CallExpr)
	if !ok || !slices.Contains(call.Args, e) || c.isForbiddenStdCall(call) {
		return false
	}
	se, ok := ast.Unparen(call.Fun).(*ast.SelectorExpr)
	if !ok {
		return false
	}
	fn, ok := c.info.Uses[se.Sel].(*types.Func)
	return ok && fn.Pkg() != nil && isStdPath(fn.Pkg().Path())
}

func (c *checker) checkFuncType(n *ast.FuncType) {
	if n.Params != nil {
		for i, f := range n.Params.List {
			if _, ok := f.Type.(*ast.Ellipsis); ok && i == len(n.Params.List)-1 {
				c.report(f, CodeVariadic, "variadic parameters are not allowed", "a variadic parameter is a slice, which the subset replaces by list.List", "take a list.List[T] parameter and call it with list.Of(a, b, c)")
				continue
			}
			c.checkType(f.Type, c.info.TypeOf(f.Type))
		}
	}
	if n.Results != nil {
		for _, f := range n.Results.List {
			c.checkType(f.Type, c.info.TypeOf(f.Type))
		}
	}
}

func (c *checker) checkBranch(n *ast.BranchStmt) {
	switch n.Tok {
	case token.GOTO:
		c.report(n, CodeControl, "goto is not allowed", "the subset has only structured control flow", "use a loop with break or continue, or a function")
		return
	case token.FALLTHROUGH:
		c.report(n, CodeControl, "fallthrough is not allowed", "the cases of a switch are translated into an if chain", "list several values in one case, e.g. case 1, 2:")
		return
	}
	// a labeled break or continue, which leaves a range over an iterator, is stored and performed after the call
	// of the iterator, see fnCtx.branchTo
}

func (c *checker) checkUnary(n *ast.UnaryExpr) {
	switch n.Op {
	case token.ARROW:
		c.report(n, CodeConcurrency, "channels are not allowed", "the subset has no channels", "use a list.List as a queue")
	case token.XOR:
		if c.hasIntTerm(c.info.TypeOf(n.X)) {
			c.report(n, CodeBits, "bit operations on int are not allowed", "JavaScript computes bit operations with 32 bits, Go with 64", "convert to uint32 or int32, e.g. ^uint32(x)")
		}
	case token.AND:
		if _, ok := ast.Unparen(n.X).(*ast.CompositeLit); !ok {
			c.report(n, CodeAddress, "taking the address of a variable or field is not allowed", "only structs created with &T{...} have an address in the subset", "create the struct with &T{...} and pass the pointer, or return the new value")
		}
	}
}

func (c *checker) checkBinary(n *ast.BinaryExpr) {
	switch n.Op {
	case token.SHL, token.SHR:
		c.checkShift(n, n.Y)
	case token.QUO, token.REM:
		c.checkDivision(n, n.X, n.Y)
	case token.EQL, token.NEQ:
		if c.isNil(n.X) || c.isNil(n.Y) {
			return
		}
		t := c.info.TypeOf(n.X)
		if isInterface(c.info.TypeOf(n.Y)) {
			t = c.info.TypeOf(n.Y)
		}
		if containsBig(t, map[types.Type]bool{}) {
			fix := "compare with big.Cmp(a, b) == 0, and compare the other fields of a struct one by one"
			if named, ok := types.Unalias(t).(*types.Named); ok && named.Obj().Pkg() != nil && named.Obj().Pkg().Path() == decimalPath {
				fix = "compare with decimal.Cmp(a, b) == 0, which also treats 1.0 and 1.00 as equal"
			}
			c.report(n, CodeBigEqual, "== is not allowed for "+typeBase(t)+", which holds a big.Int", "a big.Int is a BigInt in JavaScript, but a pointer in Go, so == would compare differently", fix)
			return
		}
		if containsInterface(t) && !unionComparable(t, map[types.Type]bool{}) {
			c.report(n, CodeInterfaceEqual, "comparing interface values is not allowed", "the subset compares interfaces only with nil and sealed interfaces, whose variants are all comparable, like the union types of golang/go#76920", "compare a distinct value, e.g. the result of a method, or compare with nil")
		}
	}
}

func (c *checker) checkCall(n *ast.CallExpr) {
	if tv, ok := c.info.Types[n.Fun]; ok && tv.IsType() {
		c.checkConversion(n, tv.Type)
		return
	}

	if tv, ok := c.info.Types[n]; ok && !c.isForbiddenStdCall(n) {
		c.checkType(n, tv.Type)
	}

	id, ok := ast.Unparen(n.Fun).(*ast.Ident)
	if !ok {
		return
	}
	b, ok := c.info.Uses[id].(*types.Builtin)
	if !ok {
		return
	}

	argType := func(i int) types.Type {
		if i < len(n.Args) {
			return c.info.TypeOf(n.Args[i])
		}
		return nil
	}

	switch b.Name() {
	case "panic":
		c.report(n, CodePanic, "panic is not allowed", "a checked program never fails at runtime, so the subset has no panics", "return an error, e.g. return errors.New(\"invalid input\"), and check it at the caller")
	case "min", "max":
		for _, a := range n.Args {
			if c.isString(a) {
				c.report(n, CodeBuiltin, b.Name()+" of strings is not allowed", "strings compare by bytes in Go and by units of UTF-16 in JavaScript, the subset compares them only with operators", "compare with < or strings.Compare")
				break
			}
		}
	case "len":
		switch {
		case c.isString(n.Args[0]):
			c.report(n, CodeString, "len of a string is not allowed", "Go counts the bytes of UTF-8, JavaScript the units of UTF-16, e.g. for \"ä\"", "use text.RuneCount(s) for the number of characters or text.ByteLen(s) for the bytes of UTF-8")
		case isOpaque(argType(0)):
			c.report(n, CodeBuiltin, "len of a list.List or dict.Map is not allowed", "list.List and dict.Map are types of the library", "use l.Len()")
		default:
			c.report(n, CodeBuiltin, "len is not allowed", "the subset has no slices, maps or arrays", "use the Len method of list.List or dict.Map")
		}
	case "append":
		c.report(n, CodeSlice, "append is not allowed", "the subset replaces slices by list.List", "use l.Append(v) on a list.List[T]")
	case "make":
		switch argType(0).Underlying().(type) {
		case *types.Map:
			c.report(n, CodeMap, "make of a map is not allowed", "the subset replaces maps by dict.Map", "use dict.New[K, V]()")
		case *types.Chan:
			c.report(n, CodeConcurrency, "channels are not allowed", "the subset has no channels", "use a list.List as a queue")
		default:
			c.report(n, CodeSlice, "make of a slice is not allowed", "the subset replaces slices by list.List", "use list.New[T]() or list.Make[T](n)")
		}
	case "cap", "copy":
		c.report(n, CodeSlice, b.Name()+" is not allowed", "the subset replaces slices by list.List", "use the methods of list.List, e.g. l.Clone()")
	case "delete":
		c.report(n, CodeMap, "delete is not allowed", "the subset replaces maps by dict.Map", "use m.Delete(k)")
	case "new":
		c.report(n, CodeBuiltin, "new is not allowed", "only structs have an address in the subset", "use &T{} for a struct")
	case "close":
		c.report(n, CodeConcurrency, "channels are not allowed", "the subset has no channels", "use a list.List as a queue")
	case "recover":
		c.report(n, CodeControl, "recover is not allowed", "the subset has no deferred calls", "return an error instead of panicking")
	case "complex", "real", "imag":
		c.report(n, CodeNumber, "complex numbers are not allowed", "JavaScript has no complex numbers", "use a struct with two float64 fields")
	case "print", "println":
		c.report(n, CodeBuiltin, b.Name()+" is not allowed", "the output differs between the targets", "return the values and check them in a test")
	default:
		c.report(n, CodeBuiltin, b.Name()+" is not allowed", "the builtin is not part of the subset", "")
	}
}

func (c *checker) checkConversion(n *ast.CallExpr, to types.Type) {
	c.checkType(n, to)
	if len(n.Args) != 1 {
		return
	}
	from := c.info.TypeOf(n.Args[0])
	if isFloat(from) && isInteger(to) {
		c.report(n, CodeFloatConversion, "converting a float64 into an integer is not allowed", "Go leaves the result undefined for NaN and values out of range", "use num.Trunc32(f) or num.Round32(f) from \"go.wdy.de/nago/pkg/gostrict/std/num\", which are defined for every value")
	}
	if isStringType(to) && !isStringType(from) {
		c.report(n, CodeString, "converting a number to a string is not allowed", "string(n) returns the character of a code point, which is easily confused with formatting the number", "use strconv.Itoa(n) to format a number or text.FromCodePoint(r) for a character")
	}
}

func (c *checker) checkSelector(n *ast.SelectorExpr) {
	if sel := c.info.Selections[n]; sel != nil {
		if sel.Kind() == types.MethodExpr {
			c.report(n, CodeUnsupported, "method expressions are not allowed", "they are not part of the subset", "use a function literal, e.g. func(v T) int { return v.M() }")
		}
		return
	}

	id, ok := n.X.(*ast.Ident)
	if !ok {
		return
	}
	pkgName, ok := c.info.Uses[id].(*types.PkgName)
	if !ok {
		return
	}

	imported := pkgName.Imported().Path()
	obj := c.info.Uses[n.Sel]
	if _, ok := obj.(*types.Const); ok {
		return
	}

	if funcs, ok := stdAllowed[imported]; ok {
		if funcs[n.Sel.Name] {
			return
		}
		why := "the function is not part of the subset, because it would behave differently in JavaScript"
		fix := "see the allowed functions of " + imported + ": " + strings.Join(sortedKeys(funcs), ", ")
		if h, ok := stdFuncHints[imported+"."+n.Sel.Name]; ok {
			why, fix = h.why, h.fix
		}
		c.report(n, CodeStdFunc, imported+"."+n.Sel.Name+" is not allowed", why, fix)
		return
	}

	fn, ok := obj.(*types.Func)
	if !ok || isStdPath(imported) || strings.HasPrefix(imported, stdPrefix) {
		return
	}
	if src, ok := c.decls[fn]; ok && src.pkg.Kind == KindHeader && src.js == "" {
		c.report(n, CodeHeader, imported+"."+n.Sel.Name+" has no JavaScript binding", "a function of a header package needs a //gostrict:js directive", "add //gostrict:js <expression> to the documentation of the function")
	}
}

func (c *checker) checkRange(n *ast.RangeStmt) {
	t := c.info.TypeOf(n.X)
	if t == nil {
		return
	}

	switch u := t.Underlying().(type) {
	case *types.Basic:
		if u.Info()&types.IsString != 0 {
			c.report(n.X, CodeString, "ranging over a string is not allowed", "Go ranges over the bytes of UTF-8, JavaScript over the units of UTF-16", "use for i, c := range text.Chars(s).All() or text.CodePoints(s).All()")
		}
		return
	case *types.Signature:
		return
	}

	if isOpaque(t) {
		c.report(n.X, CodeRange, "ranging over a list.List or dict.Map is not allowed", "list.List and dict.Map are types of the library", "range over l.All() or l.Values()")
		return
	}
	c.report(n.X, CodeRange, "ranging over "+t.String()+" is not allowed", "the subset ranges over int, iterators and list.List", "range over l.All() of a list.List")
}

// checkType reports the first part of t, which is not allowed.
func (c *checker) checkType(n ast.Node, t types.Type) {
	if t == nil {
		return
	}
	if p := typeProblem(t); p != nil {
		c.report(n, p.code, p.msg, p.why, p.fix)
	}
}

type problem struct {
	code          Code
	msg, why, fix string
}

func typeProblem(t types.Type) *problem {
	switch t := types.Unalias(t).(type) {
	case *types.Basic:
		switch t.Kind() {
		case types.Bool, types.Int32, types.Uint32, types.Float64, types.String,
			types.UntypedBool, types.UntypedInt, types.UntypedRune, types.UntypedFloat, types.UntypedString, types.UntypedNil:
			return nil
		case types.UnsafePointer:
			return &problem{CodeUnsupported, "unsafe.Pointer is not allowed", "it is not part of the subset", ""}
		case types.Uint8:
			return &problem{CodeNumber, "byte is not allowed", "the subset has no bytes, because strings are not byte based in JavaScript", "use int32 for numbers and the functions of text for strings"}
		case types.Int:
			return &problem{CodeNumber, "int is not allowed", "int has 64 bits in Go but JavaScript computes with float64, so an overflow would differ", "use int32, which wraps around in Go and JavaScript alike, or float64; write int32(0) or var n int32 for a constant, and range over int32(n)"}
		case types.Int64, types.Uint64, types.Uint, types.Uintptr:
			return &problem{CodeNumber, t.Name() + " is not allowed", "JavaScript represents numbers as float64, which holds only 53 bits of an integer", "use int32 or uint32 for integers and float64 for anything larger"}
		default:
			return &problem{CodeNumber, t.Name() + " is not allowed", "the subset allows only int32, uint32 and float64 as numbers", "use int32 for integers, float64 for decimals and uint32 for bit operations"}
		}
	case *types.Pointer:
		elem := types.Unalias(t.Elem())
		if isOpaque(elem) {
			return &problem{CodePointer, "pointers to list.List or dict.Map are not allowed", "list.List and dict.Map refer to their elements anyway", "use the list.List or dict.Map directly"}
		}
		if n, ok := elem.(*types.Named); ok {
			if _, ok := n.Underlying().(*types.Struct); ok {
				return typeProblem(n)
			}
		}
		return &problem{CodePointer, "pointers to " + elem.String() + " are not allowed", "only structs have an address in the subset", "wrap the value in a struct and use a pointer to it, or return the new value"}
	case *types.Named:
		if t.TypeArgs() != nil {
			for a := range t.TypeArgs().Types() {
				if p := typeProblem(a); p != nil {
					return p
				}
			}
		}
		return nil
	case *types.TypeParam:
		return nil
	case *types.Slice:
		return &problem{CodeSlice, "slices are not allowed: " + t.String(), "slices alias and grow differently in Go and JavaScript", "use list.List[" + t.Elem().String() + "] from \"go.wdy.de/nago/pkg/gostrict/std/list\", e.g. list.Of(a, b) or list.New[T]()"}
	case *types.Array:
		return &problem{CodeArray, "arrays are not allowed: " + t.String(), "arrays are not part of the subset", "use a struct with fields, e.g. type Vec3 struct{ X, Y, Z float64 }, or a list.List"}
	case *types.Map:
		return &problem{CodeMap, "maps are not allowed: " + t.String(), "the iteration order of maps is random in Go and differs in JavaScript", "use dict.Map[" + t.Key().String() + ", " + t.Elem().String() + "] from \"go.wdy.de/nago/pkg/gostrict/std/dict\", e.g. dict.New[K, V]()"}
	case *types.Chan:
		return &problem{CodeConcurrency, "channels are not allowed", "the subset has no channels", "use a list.List as a queue"}
	case *types.Struct:
		return &problem{CodeUnsupported, "anonymous struct types are not allowed", "the subset needs a name for each struct type", "declare a named struct type at package level"}
	case *types.Interface:
		if t.NumMethods() == 0 && t.IsMethodSet() {
			return &problem{CodeAny, "any is not allowed as the type of a value", "the subset has no type assertions, so a value of type any could not be used", "use a type parameter, e.g. func F[T any](v T), or an interface with methods"}
		}
		return nil
	case *types.Signature:
		if p := tupleProblem(t.Params(), t.Variadic()); p != nil {
			return p
		}
		return tupleProblem(t.Results(), false)
	case *types.Tuple:
		return tupleProblem(t, false)
	}
	return nil
}

func tupleProblem(t *types.Tuple, variadic bool) *problem {
	for i := 0; i < t.Len(); i++ {
		if variadic && i == t.Len()-1 {
			continue
		}
		if p := typeProblem(t.At(i).Type()); p != nil {
			return p
		}
	}
	return nil
}

func (c *checker) isString(e ast.Expr) bool {
	t := c.info.TypeOf(e)
	if t == nil {
		return false
	}
	if tp, ok := t.(*types.TypeParam); ok {
		return termsHave(tp, func(b *types.Basic) bool { return b.Info()&types.IsString != 0 })
	}
	return isStringType(t)
}

func (c *checker) isNil(e ast.Expr) bool {
	tv, ok := c.info.Types[e]
	return ok && tv.IsNil()
}

// hasIntTerm tells, whether t is int or a type parameter, which may be int.
func (c *checker) hasIntTerm(t types.Type) bool {
	if t == nil {
		return false
	}
	isInt := func(b *types.Basic) bool { return b.Kind() == types.Int }
	if tp, ok := t.(*types.TypeParam); ok {
		return termsHave(tp, isInt)
	}
	b, ok := t.Underlying().(*types.Basic)
	return ok && isInt(b)
}

func termsHave(tp *types.TypeParam, pred func(*types.Basic) bool) bool {
	iface, ok := tp.Constraint().Underlying().(*types.Interface)
	if !ok {
		return false
	}
	for i := 0; i < iface.NumEmbeddeds(); i++ {
		if u, ok := iface.EmbeddedType(i).Underlying().(*types.Union); ok {
			for j := 0; j < u.Len(); j++ {
				if b, ok := u.Term(j).Type().Underlying().(*types.Basic); ok && pred(b) {
					return true
				}
			}
		} else if b, ok := iface.EmbeddedType(i).Underlying().(*types.Basic); ok && pred(b) {
			return true
		}
	}
	return false
}

func isStringType(t types.Type) bool {
	if t == nil {
		return false
	}
	b, ok := t.Underlying().(*types.Basic)
	return ok && b.Info()&types.IsString != 0
}

// isOpaque tells, whether t is a list.List, dict.Map, msg.Out or opt.Option, which the runtime represents
// natively.
func isOpaque(t types.Type) bool {
	if t == nil {
		return false
	}
	n, ok := types.Unalias(t).(*types.Named)
	if !ok || n.Obj().Pkg() == nil {
		return false
	}
	p := n.Obj().Pkg().Path()
	return (p == listPath && (n.Obj().Name() == "List" || n.Obj().Name() == "Frozen")) || (p == dictPath && n.Obj().Name() == "Map") || (p == msgPath && n.Obj().Name() == "Out") || isOption(t)
}

// isOption tells, whether t is an opt.Option, whose zero value is valid: it holds nothing. In JavaScript, it is
// null or an object with the value, which never changes.
func isOption(t types.Type) bool {
	n, ok := types.Unalias(t).(*types.Named)
	return ok && n.Obj().Pkg() != nil && n.Obj().Pkg().Path() == optPath && n.Obj().Name() == "Option"
}

func containsInterface(t types.Type) bool {
	return containsInterfaceSeen(t, map[types.Type]bool{})
}

func containsInterfaceSeen(t types.Type, seen map[types.Type]bool) bool {
	if t == nil || seen[t] {
		return false
	}
	seen[t] = true
	switch u := t.Underlying().(type) {
	case *types.Interface:
		_, isTP := t.(*types.TypeParam)
		return !isTP
	case *types.Struct:
		if isOpaque(t) {
			return false
		}
		for f := range u.Fields() {
			if containsInterfaceSeen(f.Type(), seen) {
				return true
			}
		}
	}
	return false
}

func sortedKeys(m map[string]bool) []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

// isForbiddenStdCall tells, whether the call calls a function of the standard library, which is reported anyway.
func (c *checker) isForbiddenStdCall(call *ast.CallExpr) bool {
	se, ok := ast.Unparen(call.Fun).(*ast.SelectorExpr)
	if !ok {
		return false
	}
	id, ok := se.X.(*ast.Ident)
	if !ok {
		return false
	}
	pkgName, ok := c.info.Uses[id].(*types.PkgName)
	if !ok {
		return false
	}
	funcs, ok := stdAllowed[pkgName.Imported().Path()]
	return !ok || !funcs[se.Sel.Name]
}

// containsBig tells, whether a value of t holds a big.Int, which == compares differently in Go and JavaScript.
func containsBig(t types.Type, seen map[types.Type]bool) bool {
	if t == nil || seen[t] {
		return false
	}
	seen[t] = true
	if isBig(t) {
		return true
	}
	if st, ok := t.Underlying().(*types.Struct); ok && !isOpaque(t) {
		for f := range st.Fields() {
			if containsBig(f.Type(), seen) {
				return true
			}
		}
	}
	return false
}
