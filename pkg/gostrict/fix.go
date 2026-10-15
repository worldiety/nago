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
