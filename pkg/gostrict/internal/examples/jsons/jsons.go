// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package binary reads and writes a binary format like a recorder of an industrial plant.
// Package jsons reads JSON documents into the types of a program, so that every engine can be compared.
package jsons

import (
	"go.wdy.de/nago/pkg/gostrict/std/big"
	"go.wdy.de/nago/pkg/gostrict/std/bin"
	"go.wdy.de/nago/pkg/gostrict/std/dict"
	"go.wdy.de/nago/pkg/gostrict/std/json"
	"go.wdy.de/nago/pkg/gostrict/std/list"
	"go.wdy.de/nago/pkg/gostrict/std/opt"
	"go.wdy.de/nago/pkg/gostrict/std/text"
)

// Point is a point of a shape.
type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Tree refers to itself through a list.
type Tree struct {
	Name     string          `json:"name"`
	Children list.List[Tree] `json:"children"`
}

// Doc has every type, which json.Decode reads.
type Doc struct {
	Name   string `json:"name"`
	ID     int64  `json:"id"`
	Max    uint64 `json:"max"`
	Count  int32
	Mask   uint32                  `json:"mask"`
	Ratio  float64                 `json:"ratio"`
	Ok     bool                    `json:"ok"`
	Tags   list.List[string]       `json:"tags"`
	Points list.List[Point]        `json:"points"`
	Frozen list.Frozen[int32]      `json:"frozen"`
	Index  dict.Map[string, int64] `json:"index"`
	Note   opt.Option[string]      `json:"note"`
	Shape  opt.Option[Point]       `json:"shape"`
	Data   bin.Bytes               `json:"data"`
	Big    big.Int                 `json:"big"`
	Tree   Tree                    `json:"tree"`
	Skip   string                  `json:"-"`
	hidden list.List[int32]
}

func decode(src string) (Doc, error) {
	return json.Decode[Doc](bin.FromText(src))
}

// Scalars returns the numbers, strings and booleans.
func Scalars(src string) (string, int64, uint64, int32, uint32, float64, bool, string, error) {
	d, err := decode(src)
	if err != nil {
		return "", 0, 0, 0, 0, 0, false, "", err
	}
	return d.Name, d.ID, d.Max, d.Count, d.Mask, d.Ratio, d.Ok, d.Skip, nil
}

// Containers returns the lists, the map and the options.
func Containers(src string) (list.List[string], list.List[Point], int32, dict.Map[string, int64], string, string, int32, error) {
	d, err := decode(src)
	if err != nil {
		return list.New[string](), list.New[Point](), 0, dict.New[string, int64](), "", "", 0, err
	}
	var frozen int32
	for _, v := range d.Frozen.All() {
		frozen += v
	}
	note := "none"
	if n, ok := d.Note.Get(); ok {
		note = "some " + n
	}
	shape := "none"
	if p, ok := d.Shape.Get(); ok {
		shape = text.FromFloat(p.X, 1) + "," + text.FromFloat(p.Y, 1)
	}
	return d.Tags, d.Points, frozen, d.Index, note, shape, d.hidden.Len(), nil
}

// Values returns the bytes and the big integer.
func Values(src string) (bin.Bytes, big.Int, error) {
	d, err := decode(src)
	if err != nil {
		return bin.Empty(), big.FromInt64(0), err
	}
	return d.Data, d.Big, nil
}

// Nodes counts the nodes of the tree and joins their names in depth-first order.
func Nodes(src string) (int32, string, error) {
	d, err := decode(src)
	if err != nil {
		return 0, "", err
	}
	names := list.New[string]()
	walk(d.Tree, names)
	return names.Len(), text.Join(names, "/"), nil
}

func walk(t Tree, names list.List[string]) {
	names.Append(t.Name)
	for _, c := range t.Children.All() {
		walk(c, names)
	}
}

// Text reads a string as the whole document.
func Text(src string) (string, error) {
	return json.Decode[string](bin.FromText(src))
}

// Raw reads a string from bytes, which need not be valid UTF-8.
func Raw(hex string) (string, error) {
	data, ok := bin.FromHex(hex)
	if !ok {
		return "", nil
	}
	return json.Decode[string](data)
}

// Numbers reads a list of int64 as the whole document.
func Numbers(src string) (list.List[int64], error) {
	return json.Decode[list.List[int64]](bin.FromText(src))
}

// Floats reads a list of float64 as the whole document.
func Floats(src string) (list.List[float64], error) {
	return json.Decode[list.List[float64]](bin.FromText(src))
}

// Small reads a list of int32 and of uint32 and returns the messages of their errors.
func Small(a, b string) (list.List[int32], list.List[uint32], string) {
	msg := ""
	x := list.New[int32]()
	if v, err := json.Decode[list.List[int32]](bin.FromText(a)); err == nil {
		x = v
	} else {
		msg = err.Error()
	}
	y := list.New[uint32]()
	if v, err := json.Decode[list.List[uint32]](bin.FromText(b)); err == nil {
		y = v
	} else {
		msg += "|" + err.Error()
	}
	return x, y, msg
}

// Sum reads the points and returns the sum of their coordinates, e.g. to measure the decoding.
func Sum(src string) (float64, int32, error) {
	points, err := json.Decode[list.List[Point]](bin.FromText(src))
	if err != nil {
		return 0, 0, err
	}
	var s float64
	for _, p := range points.All() {
		s += p.X + p.Y
	}
	return s, points.Len(), nil
}

// Payload is written by the program.
type Payload struct {
	Text   string                 `json:"text"`
	Value  float64                `json:"value"`
	Values list.List[float64]     `json:"values"`
	Small  int32                  `json:"small,omitempty"`
	Note   opt.Option[string]     `json:"note,omitempty"`
	Point  opt.Option[Point]      `json:"point"`
	Empty  list.List[string]      `json:"empty,omitempty"`
	Names  dict.Map[string, bool] `json:"names"`
}

// Encoded reads the document and writes it again.
func Encoded(src string) (string, error) {
	d, err := decode(src)
	if err != nil {
		return "", err
	}
	b, err := json.Encode(d)
	if err != nil {
		return "", err
	}
	return b.Text(), nil
}

// Written writes a payload, which the program builds, with numbers derived from f.
func Written(f float64, s string) (string, error) {
	names := dict.New[string, bool]()
	names.Set(s, true)
	names.Set("z", false)
	names.Set(s, false)
	p := Payload{
		Text:   s + text.FromCodePoint(0xD800) + text.FromCodePoint(0x1F600),
		Value:  f,
		Values: list.Of(f, -f, f*1e30, f/1e30, f*3.3e-9, 0.1, 1e21, 1e-7, 123456789.125),
		Note:   opt.None[string](),
		Point:  opt.Some(Point{X: f, Y: 2}),
		Empty:  list.New[string](),
		Names:  names,
	}
	b, err := json.Encode(p)
	if err != nil {
		return "", err
	}
	return b.Text(), nil
}

// Points writes n points and returns the size of the document, e.g. to measure the encoding.
func Points(n int32) (int64, error) {
	points := list.New[Point]()
	for i := range n {
		points.Append(Point{X: float64(i) + 0.25, Y: -float64(i) - 0.5})
	}
	b, err := json.Encode(points)
	if err != nil {
		return 0, err
	}
	return b.Len(), nil
}
