// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package program

import (
	"encoding/json"
	"errors"
	"math"
	"testing"
	"testing/fstest"

	"go.wdy.de/nago/pkg/gostrict"
)

const shapes = `package app

import (
	"go.wdy.de/nago/pkg/gostrict/std/dict"
	"go.wdy.de/nago/pkg/gostrict/std/list"
)

type Shape interface{ isShape() }

type Circle struct{ R float64 }

type Rect struct{ W, H float64 }

func (Circle) isShape() {}
func (Rect) isShape()   {}

type Drawing struct {
	Shapes list.List[Shape]
	Tags   dict.Map[string, int32]
	Big    Shape
}

type State struct{ N int32 }

func Check(d Drawing, f float64) (Drawing, float64, error) { return d, f, nil }
`

// TestCodec converts values of all kinds into JSON in the format of the browser and back.
func TestCodec(t *testing.T) {
	b, diags, err := Compile("llm.local/app", gostrict.Module{Path: "llm.local/app", FS: fstest.MapFS{"app.go": {Data: []byte(shapes)}}})
	if err != nil || len(diags) > 0 {
		t.Fatal(err, diags)
	}
	f := b.Exports["Check"]

	drawing := []any{
		[]any{gostrict.Variant{Tag: "llm.local/app.Circle", Fields: []any{1.5}}, gostrict.Variant{Tag: "llm.local/app.Rect", Fields: []any{2.0, math.Inf(1)}}},
		[]any{[]any{"a", int32(-3)}, []any{"b", int32(4)}},
		gostrict.Variant{Tag: "llm.local/app.Circle", Fields: []any{math.NaN()}},
	}
	c := codec{b: b}
	src, err := c.encode([]any{drawing, math.Copysign(0, -1), errors.New("boom")}, append(f.Params, f.Results[2]))
	if err != nil {
		t.Fatal(err)
	}

	// the format, which the reviver of the browser reads
	var tree []any
	if err := json.Unmarshal([]byte(src), &tree); err != nil {
		t.Fatal(err)
	}
	want := `[{"Big":{"$":"llm.local/app.Circle","R":{"$f":"NaN"}},"Shapes":[{"$":"llm.local/app.Circle","R":1.5},{"$":"llm.local/app.Rect","H":{"$f":"Inf"},"W":2}],"Tags":{"$m":[["a",-3],["b",4]]}},-0,{"$e":"boom"}]`
	if src != want {
		t.Fatalf("unexpected JSON\n got: %s\nwant: %s", src, want)
	}

	// the browser sends -0 as a marker, because JSON.stringify writes 0
	back, err := c.decode(`[`+src[1:len(src)-len(`,-0,{"$e":"boom"}]`)]+`,{"$f":"-0"},{"$e":"boom"}]`, append(f.Params, f.Results[2]))
	if err != nil {
		t.Fatal(err)
	}
	if !gostrict.SameValues(back[:2], []any{drawing, math.Copysign(0, -1)}) || back[2].(error).Error() != "boom" {
		t.Fatalf("unexpected values %#v", back)
	}

	// a tag, which is no variant of the union, is rejected
	if _, err := c.decode(`[{"Shapes": [{"$": "llm.local/app.State", "N": 1}], "Tags": {"$m": []}, "Big": {"$": "llm.local/app.Circle", "R": 1}}, 1]`, f.Params); err == nil {
		t.Fatal("a wrong variant was accepted")
	}
}

// TestCodecBig converts a big.Int and a decimal.Dec, which holds one, into JSON and back and rejects other digits.
func TestCodecBig(t *testing.T) {
	src := `package app

import (
	"go.wdy.de/nago/pkg/gostrict/std/big"
	"go.wdy.de/nago/pkg/gostrict/std/decimal"
)

type Line struct {
	Amount decimal.Dec
	Count  big.Int
}

func Total(l Line) Line { return l }
`
	b, diags, err := Compile("llm.local/app", gostrict.Module{Path: "llm.local/app", FS: fstest.MapFS{"app.go": {Data: []byte(src)}}})
	if err != nil || len(diags) > 0 {
		t.Fatal(err, diags)
	}
	f := b.Exports["Total"]
	c := codec{b: b}

	line := []any{[]any{"-123456789012345678901234567890", int32(2)}, "7"}
	js, err := c.encode([]any{line}, f.Params)
	if err != nil {
		t.Fatal(err)
	}
	want := `[{"Amount":{"M":{"$b":"-123456789012345678901234567890"},"S":2},"Count":{"$b":"7"}}]`
	if js != want {
		t.Fatalf("unexpected JSON\n got: %s\nwant: %s", js, want)
	}
	back, err := c.decode(js, f.Params)
	if err != nil {
		t.Fatal(err)
	}
	if !gostrict.SameValues(back, []any{line}) {
		t.Fatalf("unexpected values %#v", back)
	}

	for _, bad := range []string{`{"$b":"1.5"}`, `{"$b":""}`, `{"$b":7}`, `7`, `{"$b":"1e3"}`, `{"$b":"-+5"}`, `{"$b":"+"}`} {
		if _, err := c.decode(`[{"Amount":{"M":{"$b":"1"},"S":0},"Count":`+bad+`}]`, f.Params); err == nil {
			t.Errorf("%s was accepted as big.Int", bad)
		}
	}
}

// TestCompileFS derives the module path from the import of the package of messages.
func TestCompileFS(t *testing.T) {
	fsys := fstest.MapFS{
		"scene/scene.go": {Data: []byte(`package scene

import (
	"example.com/app/scene/messages"
	"go.wdy.de/nago/pkg/gostrict/std/msg"
)

type State struct{ N int32 }

func Init() *State { return &State{} }

func OnPing(s *State, p messages.Ping) msg.Out[messages.Ping] {
	out := msg.New[messages.Ping]()
	out.Send(p)
	return out
}
`)},
		"scene/messages/messages.go": {Data: []byte("package messages\n\ntype Ping struct{ N int32 }\n")},
	}
	b, diags, err := CompileFS(fsys, "scene")
	if err != nil || len(diags) > 0 {
		t.Fatal(err, diags)
	}
	if b.Handlers["example.com/app/scene/messages.Ping"] != "OnPing" {
		t.Fatalf("unexpected handlers %v", b.Handlers)
	}
}
