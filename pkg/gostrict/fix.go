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
	"slices"
	"strconv"
	"strings"
)

// Edit is a change of the source code, which fixes a diagnostic mechanically, e.g. int into int32. It replaces the
// text from Line:Col up to EndLine:EndCol, which is excluded, by New. Columns count bytes from 1, like those of
// the diagnostic. An insertion has the same start and end.
type Edit struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Col     int    `json:"col"`
	EndLine int    `json:"endLine"`
	EndCol  int    `json:"endCol"`
	New     string `json:"new"`
}

// Edits returns the edits of all diagnostics of the file, see [ApplyEdits].
func (d Diagnostics) Edits(file string) []Edit {
	var res []Edit
	for _, diag := range d {
		for _, e := range diag.Edits {
			if e.File == file {
				res = append(res, e)
			}
		}
	}
	return res
}

// ApplyEdits applies the edits to the source code of their file. Identical edits are applied once, e.g. the
// import, which several fixes need, and an edit, which overlaps a former one, is skipped, so that a fix never
// breaks the code. Run the check again afterwards, because a fix may reveal further problems.
func ApplyEdits(src []byte, edits []Edit) ([]byte, error) {
	lineStart := []int{0}
	for i, b := range src {
		if b == '\n' {
			lineStart = append(lineStart, i+1)
		}
	}
	offset := func(line, col int) (int, error) {
		if line < 1 || line > len(lineStart) {
			return 0, fmt.Errorf("line %d is out of range", line)
		}
		o := lineStart[line-1] + col - 1
		if col < 1 || o > len(src) {
			return 0, fmt.Errorf("column %d of line %d is out of range", col, line)
		}
		return o, nil
	}

	type span struct {
		from, to int
		text     string
	}
	var spans []span
	seen := map[span]bool{}
	for _, e := range edits {
		from, err := offset(e.Line, e.Col)
		if err != nil {
			return nil, err
		}
		to, err := offset(e.EndLine, e.EndCol)
		if err != nil {
			return nil, err
		}
		s := span{from, to, e.New}
		if to < from || seen[s] {
			continue
		}
		seen[s] = true
		spans = append(spans, s)
	}
	slices.SortStableFunc(spans, func(a, b span) int { return a.from - b.from })

	var sb strings.Builder
	pos := 0
	for _, s := range spans {
		if s.from < pos {
			// overlaps the former edit
			continue
		}
		sb.Write(src[pos:s.from])
		sb.WriteString(s.text)
		pos = s.to
	}
	sb.Write(src[pos:])
	return []byte(sb.String()), nil
}

// fix associates an edit with the diagnostic of the code at the line of at, if it is reported.
func (d *diagnoser) fix(code Code, at, from, to token.Pos, text string) {
	a, f, t := d.fset.Position(at), d.fset.Position(from), d.fset.Position(to)
	if d.edits == nil {
		d.edits = map[string][]Edit{}
	}
	key := fmt.Sprintf("%s:%d:%s", a.Filename, a.Line, code)
	d.edits[key] = append(d.edits[key], Edit{File: f.Filename, Line: f.Line, Col: f.Column, EndLine: t.Line, EndCol: t.Column, New: text})
}

// attachEdits adds the edits to their diagnostics.
func (d *diagnoser) attachEdits() {
	for i, diag := range d.diags {
		if edits, ok := d.edits[fmt.Sprintf("%s:%d:%s", diag.File, diag.Line, diag.Code)]; ok {
			d.diags[i].Edits = edits
		}
	}
}

// collectFixes finds the mechanical fixes of the file, which become edits of the diagnostics on their lines.
func (c *checker) collectFixes(f *ast.File) {
	intType := types.Universe.Lookup("int")
	importText := func(at token.Pos) {
		for _, imp := range f.Imports {
			if strings.Trim(imp.Path.Value, `"`) == textPath {
				return
			}
		}
		c.d.fix(CodeString, at, f.Name.End(), f.Name.End(), "\n\nimport "+strconv.Quote(textPath))
	}

	ast.Inspect(f, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.Ident:
			// the type int, e.g. of a parameter, a variable or a conversion
			if c.info.Uses[n] == intType {
				c.d.fix(CodeNumber, n.Pos(), n.Pos(), n.End(), "int32")
			}
		case *ast.AssignStmt:
			// i := 0 makes an int of an untyped constant
			if n.Tok != token.DEFINE || len(n.Lhs) != len(n.Rhs) {
				return true
			}
			for i, l := range n.Lhs {
				id, ok := l.(*ast.Ident)
				if !ok || c.info.Defs[id] == nil || c.info.Defs[id].Type() != intType.Type() {
					continue
				}
				if lit, ok := ast.Unparen(n.Rhs[i]).(*ast.BasicLit); ok && lit.Kind == token.INT {
					c.d.fix(CodeNumber, id.Pos(), lit.Pos(), lit.Pos(), "int32(")
					c.d.fix(CodeNumber, id.Pos(), lit.End(), lit.End(), ")")
				}
			}
		case *ast.RangeStmt:
			// range over an untyped constant ranges over int
			if tv, ok := c.info.Types[n.X]; ok && tv.Value != nil && tv.Value.Kind() == constant.Int && types.Identical(tv.Type, intType.Type()) && n.Key != nil {
				c.d.fix(CodeNumber, n.Key.Pos(), n.X.Pos(), n.X.Pos(), "int32(")
				c.d.fix(CodeNumber, n.Key.Pos(), n.X.End(), n.X.End(), ")")
			}
		case *ast.CallExpr:
			id, ok := ast.Unparen(n.Fun).(*ast.Ident)
			if !ok || len(n.Args) != 1 {
				return true
			}
			if b, ok := c.info.Uses[id].(*types.Builtin); !ok || b.Name() != "len" {
				return true
			}
			switch t := c.info.TypeOf(n.Args[0]); {
			case isStringType(t):
				c.d.fix(CodeString, n.Pos(), id.Pos(), id.End(), "text.RuneCount")
				importText(n.Pos())
			case isOpaque(t) && !isOption(t):
				c.d.fix(CodeBuiltin, n.Pos(), n.Pos(), n.Args[0].Pos(), "")
				c.d.fix(CodeBuiltin, n.Pos(), n.Args[0].End(), n.End(), ".Len()")
			}
		}
		return true
	})
}

// optionalFix records the fix of x, _ := g() or x, _ = g() (R-SCRIPT-GS-OPTIONAL) with the operator of the original
// and returns the new statement: x := l.GetOr(i, <zero value>) for Get of a list or a dict, x := opt.OrZero(g()) or
// opt.OrZeroErr(g()) for a function, which returns the zero value, if it fails, and whose result contains no nil. The
// fix returns the same value as before. There is none for a function value, a method of an interface or of a type
// parameter, Invoke, three or more results and a zero value, which the checker can not write.
func (c *checker) optionalFix(as *ast.AssignStmt, p pairing, g *getCall) (string, bool) {
	if as == nil || len(as.Lhs) != 2 || p.tuple.Len() != 2 || isBlank(as.Lhs[0]) || !isBlank(as.Lhs[1]) {
		return "", false
	}
	f := c.fileOf(as)
	if f == nil {
		return "", false
	}
	q := c.newQualifier(f)
	var rhs string
	switch {
	case g != nil:
		zero, ok := c.zeroText(g.elem, q, map[types.Type]bool{})
		if !ok {
			return "", false
		}
		rhs = c.text(g.recv) + ".GetOr(" + c.text(g.arg) + ", " + zero + ")"
	default:
		if nilable(p.tuple.At(0).Type()) || !c.zeroOnFailure(p) {
			return "", false
		}
		name := "OrZero"
		if p.errIdx >= 0 {
			name = "OrZeroErr"
		}
		rhs = q.pkg(optPath, "opt") + "." + name + "(" + c.text(p.call) + ")"
	}
	if q.failed {
		return "", false
	}
	stmt := c.text(as.Lhs[0]) + " " + as.Tok.String() + " " + rhs
	c.d.fix(CodeGuard, as.Pos(), as.Pos(), as.End(), stmt)
	for _, path := range q.added {
		c.d.fix(CodeGuard, as.Pos(), f.Name.End(), f.Name.End(), "\n\nimport "+strconv.Quote(path))
	}
	return stmt, true
}

// text returns the source code of the node.
func (c *checker) text(n ast.Node) string {
	from, to := c.d.fset.Position(n.Pos()), c.d.fset.Position(n.End())
	src, ok := c.d.sources[from.Filename]
	if !ok || from.Offset < 0 || to.Offset > len(src) || from.Offset > to.Offset {
		return types.ExprString(n.(ast.Expr))
	}
	return string(src[from.Offset:to.Offset])
}

// zeroOnFailure tells, whether the function, which the call calls by its name, returns the zero value of each other
// result, if its last bool is false or its last error is not nil: a function of a header package with the directive
// //gostrict:zero, or a function of the program, each of whose returns with a last result, which is not the literal
// true or nil, returns the zero value as a literal.
func (c *checker) zeroOnFailure(p pairing) bool {
	fn := c.namedCallee(p.call)
	if fn == nil {
		return false
	}
	src, ok := c.decls[fn.Origin()]
	if !ok {
		return false
	}
	if src.pkg.Kind != KindUser {
		return src.zero
	}
	if src.decl.Body == nil {
		return false
	}
	info := src.pkg.Info
	zero := true
	ast.Inspect(src.decl.Body, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.FuncLit:
			return false
		case *ast.ReturnStmt:
			if len(n.Results) != p.tuple.Len() {
				zero = false
				return false
			}
			last := n.Results[len(n.Results)-1]
			if (p.okIdx >= 0 && isTrueConst(info, last)) || (p.errIdx >= 0 && isNilIn(info, last)) {
				return false
			}
			for j := range p.values {
				if !isZeroLiteral(info, n.Results[j]) {
					zero = false
				}
			}
			return false
		}
		return zero
	})
	return zero
}

func isNilIn(info *types.Info, e ast.Expr) bool {
	tv, ok := info.Types[e]
	return ok && tv.IsNil()
}

// isZeroLiteral tells, whether e writes a zero value: a constant 0, "" or false, an empty literal of a struct or
// opt.None.
func isZeroLiteral(info *types.Info, e ast.Expr) bool {
	e = ast.Unparen(e)
	if tv, ok := info.Types[e]; ok && tv.Value != nil {
		switch tv.Value.Kind() {
		case constant.Bool:
			return !constant.BoolVal(tv.Value)
		case constant.String:
			return constant.StringVal(tv.Value) == ""
		case constant.Int, constant.Float:
			return constant.Sign(tv.Value) == 0
		}
		return false
	}
	switch e := e.(type) {
	case *ast.CompositeLit:
		_, isStruct := info.TypeOf(e).Underlying().(*types.Struct)
		return isStruct && len(e.Elts) == 0
	case *ast.CallExpr:
		fun := ast.Unparen(e.Fun)
		if ix, ok := fun.(*ast.IndexExpr); ok {
			fun = ix.X
		}
		se, ok := fun.(*ast.SelectorExpr)
		if !ok || len(e.Args) != 0 {
			return false
		}
		fn, ok := info.Uses[se.Sel].(*types.Func)
		return ok && fn.Pkg() != nil && fn.Pkg().Path() == optPath && fn.Name() == "None"
	}
	return false
}

// qualifier names the packages of a file in the text of a fix and collects the imports of std, which it needs.
type qualifier struct {
	c     *checker
	names map[string]string
	added []string
	// failed tells, that a type of a package, which the file does not import, is needed
	failed bool
}

func (c *checker) newQualifier(f *ast.File) *qualifier {
	q := &qualifier{c: c, names: map[string]string{}}
	for _, imp := range f.Imports {
		if pn := c.info.PkgNameOf(imp); pn != nil {
			q.names[pn.Imported().Path()] = pn.Name()
		}
	}
	return q
}

// pkg returns the name of the package of std in the file and imports it, if the file does not.
func (q *qualifier) pkg(path, name string) string {
	if n, ok := q.names[path]; ok {
		return n
	}
	q.names[path] = name
	q.added = append(q.added, path)
	return name
}

// typ returns the type as the file names it.
func (q *qualifier) typ(t types.Type) string {
	return types.TypeString(t, func(p *types.Package) string {
		if p == q.c.pkg.Types {
			return ""
		}
		if n, ok := q.names[p.Path()]; ok {
			return n
		}
		switch p.Path() {
		case listPath, dictPath, optPath, bigPath, decimalPath:
			return q.pkg(p.Path(), p.Name())
		}
		q.failed = true
		return p.Name()
	})
}

// zeroText writes the zero value of t as gostrict accepts it: a constant, an empty literal, and for each field, which
// contains nil, list.New[T](), list.Freeze(list.New[T]()), dict.New[K, V](), big.Zero(), decimal.Zero() or opt.None,
// recursively. It fails for a pointer, a function, an interface, a handle, an outbox, a type parameter and a field,
// which can not be set outside its package.
func (c *checker) zeroText(t types.Type, q *qualifier, seen map[types.Type]bool) (string, bool) {
	if seen[t] {
		return "", false
	}
	seen[t] = true
	defer delete(seen, t)
	t = types.Unalias(t)
	if n, ok := t.(*types.Named); ok && n.Obj().Pkg() != nil {
		args := n.TypeArgs()
		switch p, name := n.Obj().Pkg().Path(), n.Obj().Name(); {
		case p == optPath && name == "Option":
			return q.pkg(optPath, "opt") + ".None[" + q.typ(args.At(0)) + "]()", true
		case p == listPath && name == "List":
			return q.pkg(listPath, "list") + ".New[" + q.typ(args.At(0)) + "]()", true
		case p == listPath && name == "Frozen":
			l := q.pkg(listPath, "list")
			return l + ".Freeze(" + l + ".New[" + q.typ(args.At(0)) + "]())", true
		case p == dictPath && name == "Map":
			return q.pkg(dictPath, "dict") + ".New[" + q.typ(args.At(0)) + ", " + q.typ(args.At(1)) + "]()", true
		case p == bigPath && name == "Int":
			return q.pkg(bigPath, "big") + ".Zero()", true
		case p == decimalPath && name == "Dec":
			return q.pkg(decimalPath, "decimal") + ".Zero()", true
		case p == msgPath:
			return "", false
		}
	}
	if isHandle(t) {
		return "", false
	}
	if _, ok := t.(*types.TypeParam); ok {
		return "", false
	}
	switch u := t.Underlying().(type) {
	case *types.Basic:
		switch {
		case u.Info()&types.IsBoolean != 0:
			return "false", true
		case u.Info()&types.IsString != 0:
			return `""`, true
		case u.Info()&types.IsNumeric != 0:
			return "0", true
		}
	case *types.Struct:
		if _, named := t.(*types.Named); !named {
			return "", false
		}
		var parts []string
		for f := range u.Fields() {
			if !nilable(f.Type()) {
				continue
			}
			if !f.Exported() && f.Pkg() != c.pkg.Types {
				return "", false
			}
			z, ok := c.zeroText(f.Type(), q, seen)
			if !ok {
				return "", false
			}
			parts = append(parts, f.Name()+": "+z)
		}
		return q.typ(t) + "{" + strings.Join(parts, ", ") + "}", true
	}
	return "", false
}
