// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict

import (
	"encoding/json"
	"fmt"
	"go/token"
	"slices"
	"strings"
)

// Code identifies a kind of diagnostic. The codes are stable, so a tool or a model may refer to them.
type Code string

const (
	CodeTypeError       Code = "GS000" // the code is not valid Go
	CodeImport          Code = "GS001" // the package can not be imported
	CodeStdFunc         Code = "GS002" // the function of an allowed package is not part of the subset
	CodeSlice           Code = "GS003" // slices
	CodeMap             Code = "GS004" // maps
	CodeArray           Code = "GS005" // arrays
	CodeNumber          Code = "GS006" // numeric types other than int, int32, uint32 and float64
	CodeBits            Code = "GS007" // bit operations on int
	CodeString          Code = "GS008" // byte based string operations
	CodeAddress         Code = "GS009" // taking the address of a variable or field
	CodePointer         Code = "GS010" // pointers to anything but structs
	CodeConcurrency     Code = "GS011" // goroutines and channels
	CodeControl         Code = "GS012" // defer, recover, goto and fallthrough
	CodeTypeSwitch      Code = "GS013" // type switches and assertions
	CodeEmbedding       Code = "GS014" // embedded fields
	CodeVariadic        Code = "GS015" // declaring variadic functions
	CodeInit            Code = "GS016" // init functions
	CodeAny             Code = "GS017" // any as the type of a value
	CodeInterfaceEqual  Code = "GS018" // comparing interface values
	CodeBuiltin         Code = "GS019" // builtin functions
	CodeInvalidUTF8     Code = "GS020" // string literals, which are not valid UTF-8
	CodeUnsupported     Code = "GS021" // anything else, which is not part of the subset
	CodeRange           Code = "GS022" // ranging over something else than int, an iterator or a list
	CodeRangeFuncBranch Code = "GS023" // no longer reported: labeled break or continue out of a range over an iterator
	CodeUnsafeInt       Code = "GS024" // integer constants beyond the safe range of JavaScript
	CodeLocalType       Code = "GS025" // type declarations inside functions
	CodeOpaqueLiteral   Code = "GS026" // composite literals of list.List or dict.Map
	CodeImportForm      Code = "GS027" // dot and blank imports
	CodeHeader          Code = "GS028" // functions of header packages without a JavaScript binding or with invalid directives
	CodePanic           Code = "GS029" // panic
	CodeDivision        Code = "GS030" // integer division by a divisor, which may be zero
	CodeFloatConversion Code = "GS031" // converting a float64 into an integer
	CodeNil             Code = "GS032" // sources of nil
	CodeGuard           Code = "GS033" // values, which may be nil, without a check
	CodeBridge          Code = "GS034" // declarations of header packages, which can not cross the boundary
	CodeBinding         Code = "GS035" // invalid JavaScript bindings
	CodeGlobal          Code = "GS036" // package variables, which change or hold mutable values
	CodeHandler         Code = "GS037" // several handlers of the same message
	CodeExhaustive      Code = "GS038" // switches, which miss variants or constants
	CodeUntrusted       Code = "GS039" // header packages in modules, which are not trusted
	CodeTermination     Code = "GS040" // loops and recursion, which may not end, in the terminating profile
	CodeBigEqual        Code = "GS041" // == on values, which contain a big.Int
	CodeEntry           Code = "GS042" // an exported function of a root of a whole program, which is no entry point
	CodeConstArg        Code = "GS043" // arguments of parameters with //gostrict:const, which are not constants
	CodeConstValue      Code = "GS044" // functions with //gostrict:const parameters used as values
	CodeInternal        Code = "GS099" // the compiler can not translate a valid construct
)

// Diagnostic is a problem in the source code. It tells what is wrong, why the subset forbids it and how to
// rewrite the code, so that a developer or a language model can fix it without further documentation.
type Diagnostic struct {
	Code Code   `json:"code"`
	File string `json:"file"`
	Line int    `json:"line"`
	Col  int    `json:"col"`
	// EndCol is the column after the marked range, if it ends on the same line.
	EndCol int    `json:"endCol,omitempty"`
	Msg    string `json:"msg"`
	Why    string `json:"why,omitempty"`
	Fix    string `json:"fix,omitempty"`
	// Source is the line of source code.
	Source string `json:"source,omitempty"`
	// Edits fix the problem mechanically, if they are known, see [ApplyEdits] and the command gostrict fix.
	Edits []Edit `json:"edits,omitempty"`
}

// Diagnostics is a list of problems sorted by their position.
type Diagnostics []Diagnostic

func (d Diagnostics) Error() string {
	return d.Text()
}

// Text formats the diagnostics for humans and models: the position, the line of source code with a marker and
// the reason and the fix.
func (d Diagnostics) Text() string {
	var sb strings.Builder
	for _, diag := range d {
		fmt.Fprintf(&sb, "%s:%d:%d: %s %s\n", diag.File, diag.Line, diag.Col, diag.Code, diag.Msg)
		if src := diag.Source; src != "" {
			fmt.Fprintf(&sb, "    %s\n", strings.ReplaceAll(src, "\t", "    "))
			fmt.Fprintf(&sb, "    %s\n", marker(src, diag.Col, diag.EndCol))
		}
		if diag.Why != "" {
			fmt.Fprintf(&sb, "  why: %s\n", diag.Why)
		}
		if diag.Fix != "" {
			fmt.Fprintf(&sb, "  fix: %s\n", strings.ReplaceAll(diag.Fix, "\n", "\n       "))
		}
		if len(diag.Edits) > 0 {
			sb.WriteString("  gostrict fix applies the fix\n")
		}
	}
	return sb.String()
}

// JSON returns the diagnostics as a JSON array.
func (d Diagnostics) JSON() []byte {
	if d == nil {
		d = Diagnostics{}
	}
	buf, _ := json.MarshalIndent(d, "", "  ")
	return buf
}

// marker returns the line, which marks the columns [col, endCol) of src below it. Columns count bytes and a tab
// is shown as 4 spaces.
func marker(src string, col, endCol int) string {
	var sb strings.Builder
	for i := 0; i < col-1 && i < len(src); i++ {
		if src[i] == '\t' {
			sb.WriteString("    ")
		} else if src[i]&0xC0 != 0x80 {
			sb.WriteByte(' ')
		}
	}
	n := endCol - col
	if n < 1 {
		n = 1
	}
	sb.WriteString(strings.Repeat("^", n))
	return sb.String()
}

type diagnoser struct {
	fset    *token.FileSet
	sources map[string][]byte
	diags   Diagnostics
	seen    map[string]bool
	// edits are the mechanical fixes by file, line and code of their diagnostic
	edits map[string][]Edit
}

// line returns the line of the loaded source file.
func (d *diagnoser) line(file string, line int) string {
	src, ok := d.sources[file]
	if !ok {
		return ""
	}
	lines := strings.Split(string(src), "\n")
	if line < 1 || line > len(lines) {
		return ""
	}
	return strings.TrimRight(lines[line-1], "\r")
}

func (d *diagnoser) report(pos, end token.Pos, code Code, msg, why, fix string) {
	p := d.fset.Position(pos)
	// one diagnostic per kind and line is enough to fix the line, more would only repeat it
	key := fmt.Sprintf("%s:%d:%s", p.Filename, p.Line, code)
	if d.seen == nil {
		d.seen = map[string]bool{}
	}
	if d.seen[key] {
		return
	}
	d.seen[key] = true

	diag := Diagnostic{Code: code, File: p.Filename, Line: p.Line, Col: p.Column, Msg: msg, Why: why, Fix: fix, Source: d.line(p.Filename, p.Line)}
	if end.IsValid() {
		if e := d.fset.Position(end); e.Line == p.Line && e.Column > p.Column {
			diag.EndCol = e.Column
		}
	}
	d.diags = append(d.diags, diag)
}

func (d *diagnoser) sorted() Diagnostics {
	slices.SortStableFunc(d.diags, func(a, b Diagnostic) int {
		if c := strings.Compare(a.File, b.File); c != 0 {
			return c
		}
		if a.Line != b.Line {
			return a.Line - b.Line
		}
		return a.Col - b.Col
	})
	return d.diags
}
