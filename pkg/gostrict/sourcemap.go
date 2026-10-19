// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict

import (
	"encoding/json"
	"go/token"
	"strconv"
	"strings"
)

// The emitter marks each line of JavaScript, which a statement of Go produces, with the position of the
// statement: a NUL, the index of the position and another NUL. The translated code never contains NUL itself,
// because string literals escape it. After the bundle is assembled, the marks are removed and become the source
// map, which maps each line to the line of Go, so that a browser shows and debugs the Go source.

const markChar = '\x00'

// mark returns the mark of the position, which a line starts with.
func (e *emitter) mark(pos token.Pos) string {
	if !pos.IsValid() {
		return ""
	}
	e.positions = append(e.positions, pos)
	return string(markChar) + strconv.Itoa(len(e.positions)-1) + string(markChar)
}

// sourceName returns the name of a file of the source, which a source map and the position of an error show, e.g.
// example.com/app/main.go.
func (e *emitter) sourceName(filename string) string {
	if name := e.prog.importNames[filename]; name != "" {
		return name
	}
	return filename
}

// errorPos returns the index of the position, where the program creates an error, which the runtime keeps in the
// error, so that an engine finds the position in [Bundle.Positions].
func (e *emitter) errorPos(pos token.Pos) string {
	e.errPositions = append(e.errPositions, pos)
	return strconv.Itoa(len(e.errPositions) - 1)
}

// errorPositions returns the positions of errorPos.
func (e *emitter) errorPositions() []Position {
	res := make([]Position, len(e.errPositions))
	for i, pos := range e.errPositions {
		p := e.prog.Fset.Position(pos)
		res[i] = Position{File: e.sourceName(p.Filename), Line: p.Line}
	}
	return res
}

type sourceMap struct {
	Version        int      `json:"version"`
	File           string   `json:"file,omitempty"`
	Sources        []string `json:"sources"`
	SourcesContent []string `json:"sourcesContent"`
	Names          []string `json:"names"`
	Mappings       string   `json:"mappings"`
}

// stripMarks removes the marks from the code and returns the code and its source map in the format of version 3.
func (e *emitter) stripMarks(code string) (string, []byte) {
	lines := strings.Split(code, "\n")
	sm := sourceMap{Version: 3, Names: []string{}}
	sourceIdx := map[string]int{}
	var mappings strings.Builder
	// the fields of a segment are relative to the former segment, except for the column of the line
	prevSource, prevLine, prevCol := 0, 0, 0

	for i, l := range lines {
		if i > 0 {
			mappings.WriteByte(';')
		}
		start := strings.IndexByte(l, markChar)
		if start < 0 {
			continue
		}
		var pos token.Pos
		var sb strings.Builder
		rest := l
		first := true
		for {
			s := strings.IndexByte(rest, markChar)
			if s < 0 {
				sb.WriteString(rest)
				break
			}
			end := strings.IndexByte(rest[s+1:], markChar)
			if end < 0 {
				sb.WriteString(rest)
				break
			}
			if first {
				if n, err := strconv.Atoi(rest[s+1 : s+1+end]); err == nil && n < len(e.positions) {
					pos = e.positions[n]
				}
				first = false
			}
			sb.WriteString(rest[:s])
			rest = rest[s+1+end+1:]
		}
		lines[i] = sb.String()

		p := e.prog.Fset.Position(pos)
		if !p.IsValid() {
			continue
		}
		idx, ok := sourceIdx[p.Filename]
		if !ok {
			idx = len(sm.Sources)
			sourceIdx[p.Filename] = idx
			sm.Sources = append(sm.Sources, e.sourceName(p.Filename))
			sm.SourcesContent = append(sm.SourcesContent, string(e.prog.sources[p.Filename]))
		}
		// the generated column is the first character after the indentation
		col := len(lines[i]) - len(strings.TrimLeft(lines[i], "\t"))
		mappings.WriteString(vlq(col))
		mappings.WriteString(vlq(idx - prevSource))
		mappings.WriteString(vlq(p.Line - 1 - prevLine))
		mappings.WriteString(vlq(p.Column - 1 - prevCol))
		prevSource, prevLine, prevCol = idx, p.Line-1, p.Column-1
	}

	sm.Mappings = mappings.String()
	if sm.Sources == nil {
		sm.Sources = []string{}
		sm.SourcesContent = []string{}
	}
	buf, _ := json.Marshal(sm)
	return strings.Join(lines, "\n"), buf
}

const base64Chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"

// vlq encodes a number as a base 64 variable length quantity of a source map.
func vlq(n int) string {
	v := n << 1
	if n < 0 {
		v = (-n << 1) | 1
	}
	var sb strings.Builder
	for {
		digit := v & 31
		v >>= 5
		if v > 0 {
			digit |= 32
		}
		sb.WriteByte(base64Chars[digit])
		if v == 0 {
			return sb.String()
		}
	}
}
