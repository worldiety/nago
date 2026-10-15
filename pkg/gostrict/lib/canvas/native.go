// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build !gostrict

package canvas

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"

	"go.wdy.de/nago/pkg/gostrict/std/js"
)

// context is the native context, which records the commands.
type context struct {
	width, height float64
	font          string
}

// object is a native path, gradient, pattern or image, which is named after its kind.
type object struct {
	name string
}

var (
	mu       sync.Mutex
	commands []string
	// size is the size of the canvas of a native context in CSS pixels, see SetSize.
	size = [2]float64{300, 150}
	ids  int
)

// SetSize sets the size of the canvas of the contexts, which Get returns from now on, e.g. for a test.
func SetSize(width, height float64) {
	mu.Lock()
	defer mu.Unlock()
	size = [2]float64{width, height}
}

// Commands returns the commands, which the programs have drawn natively since the last [Reset], e.g. to compare
// them with those of another engine in a test.
func Commands() []string {
	mu.Lock()
	defer mu.Unlock()
	return append([]string(nil), commands...)
}

// Reset forgets the recorded commands.
func Reset() {
	mu.Lock()
	defer mu.Unlock()
	commands = nil
	ids = 0
}

func newContext() js.Ref {
	mu.Lock()
	defer mu.Unlock()
	return js.NewRef(&context{width: size[0], height: size[1], font: "10px sans-serif"})
}

func ctxOf(r js.Ref) *context {
	if c, ok := r.Value().(*context); ok {
		return c
	}
	return &context{}
}

func newObject(kind string) js.Ref {
	mu.Lock()
	defer mu.Unlock()
	ids++
	return js.NewRef(&object{name: kind + "#" + strconv.Itoa(ids)})
}

// ctxName names a handle in the recorded commands.
func ctxName(r js.Ref) string {
	switch v := r.Value().(type) {
	case *object:
		return v.name
	case *context:
		return "ctx"
	}
	return "?"
}

func refOf[T Context | Path | Gradient | Pattern | Image](h T) js.Ref {
	switch h := any(h).(type) {
	case Context:
		return h.ref
	case Path:
		return h.ref
	case Gradient:
		return h.ref
	case Pattern:
		return h.ref
	case Image:
		return h.ref
	}
	panic("unreachable")
}

// dashes parses the pattern of SetLineDash like its binding.
func dashes(pattern string) []float64 {
	res := []float64{}
	for _, f := range strings.Fields(pattern) {
		if v, err := strconv.ParseFloat(f, 64); err == nil && v >= 0 && !math.IsInf(v, 0) {
			res = append(res, v)
		}
	}
	return res
}

// clearAll records, what the binding of Clear does.
func clearAll(r js.Ref) {
	c := ctxOf(r)
	rec(r, "save")
	rec(r, "setTransform", 1.0, 0.0, 0.0, 1.0, 0.0, 0.0)
	rec(r, "clearRect", 0.0, 0.0, c.width, c.height)
	rec(r, "restore")
}

func hex(rgba uint32) string {
	return fmt.Sprintf("#%08x", rgba)
}

// rec records a command of the context or path.
func rec(target js.Ref, name string, args ...any) {
	if c, ok := target.Value().(*context); ok && name == "font" && len(args) == 1 {
		c.font = fmt.Sprint(args[0])
	}

	var sb strings.Builder
	sb.WriteString(ctxName(target))
	sb.WriteByte('.')
	sb.WriteString(name)
	sb.WriteByte('(')
	for i, a := range args {
		if i > 0 {
			sb.WriteString(", ")
		}
		switch a := a.(type) {
		case float64:
			sb.WriteString(strconv.FormatFloat(a, 'g', 6, 64))
		case string:
			sb.WriteString(strconv.Quote(a))
		default:
			fmt.Fprint(&sb, a)
		}
	}
	sb.WriteByte(')')

	mu.Lock()
	defer mu.Unlock()
	commands = append(commands, sb.String())
}

// measure approximates the metrics of a text: each rune is 0.6 of the size of the font.
func measure(c *context, text string) (float64, float64, float64) {
	px := 10.0
	for _, part := range strings.Fields(c.font) {
		if v, ok := strings.CutSuffix(part, "px"); ok {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				px = f
			}
		}
	}
	return float64(utf8.RuneCountInString(text)) * px * 0.6, px * 0.8, px * 0.2
}
