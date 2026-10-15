// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package shapes uses sealed interfaces as sum types, an enum, maps and unions as messages.
package shapes

import (
	"math"

	"go.wdy.de/nago/pkg/gostrict/std/dict"
	"go.wdy.de/nago/pkg/gostrict/std/list"
	"go.wdy.de/nago/pkg/gostrict/std/msg"
	"go.wdy.de/nago/pkg/gostrict/std/text"
)

// Shape is a sum type: its unexported method seals it, so its variants are the types below.
type Shape interface{ isShape() }

type Circle struct{ R float64 }

type Rect struct{ W, H float64 }

type Point struct{ X, Y float64 }

type Poly struct{ Points list.List[Point] }

func (Circle) isShape() {}
func (Rect) isShape()   {}
func (Poly) isShape()   {}

type Color int32

const (
	Red Color = iota
	Green
	Blue
)

func ColorName(c Color) string {
	switch c {
	case Red:
		return "red"
	case Green:
		return "green"
	case Blue:
		return "blue"
	}
	return "unknown"
}

// Area handles each variant, which the compiler checks.
func Area(s Shape) float64 {
	switch s := s.(type) {
	case Circle:
		return math.Pi * s.R * s.R
	case Rect:
		return s.W * s.H
	case Poly:
		sum := 0.0
		n := s.Points.Len()
		for i, p := range s.Points.All() {
			q, _ := s.Points.Get((i + 1) % max(n, 1))
			sum += p.X*q.Y - q.X*p.Y
		}
		return math.Abs(sum) / 2
	}
	return 0
}

func Scale(s Shape, f float64) Shape {
	switch s := s.(type) {
	case Circle:
		return Circle{R: s.R * f}
	case Rect:
		return Rect{W: s.W * f, H: s.H * f}
	case Poly:
		points := list.New[Point]()
		for p := range s.Points.Values() {
			points.Append(Point{X: p.X * f, Y: p.Y * f})
		}
		return Poly{Points: points}
	}
	return s
}

type Drawing struct {
	Name   string
	Shapes list.List[Shape]
	Tags   dict.Map[string, Color]
}

// Summary returns the total area and the tags with their colors.
func Summary(d Drawing) (float64, string) {
	total := 0.0
	for s := range d.Shapes.Values() {
		total += Area(s)
	}
	tags := ""
	for name, c := range d.Tags.All() {
		tags += name + "=" + ColorName(c) + " "
	}
	return total, d.Name + ": " + text.FromInt(d.Shapes.Len()) + " shapes, " + tags
}

// Largest returns the shape with the largest area, or the default shape.
func Largest(d Drawing, fallback Shape) Shape {
	best := fallback
	for s := range d.Shapes.Values() {
		if Area(s) > Area(best) {
			best = s
		}
	}
	return best
}

// Canvas is the state, which receives events.
type Canvas struct {
	Shapes list.List[Shape]
	Total  float64
}

func NewCanvas() *Canvas {
	return &Canvas{Shapes: list.New[Shape]()}
}

type Event interface{ isEvent() }

type Added struct{ Shape Shape }

type Cleared struct{}

func (Added) isEvent()   {}
func (Cleared) isEvent() {}

type Total struct {
	Area  float64
	Count int32
}

// OnEvent handles all variants of Event, which the backend sends.
func OnEvent(c *Canvas, e Event) msg.Out[Total] {
	switch e := e.(type) {
	case Added:
		c.Shapes.Append(e.Shape)
		c.Total += Area(e.Shape)
	case Cleared:
		c.Shapes = list.New[Shape]()
		c.Total = 0
	}
	out := msg.New[Total]()
	out.Send(Total{Area: c.Total, Count: c.Shapes.Len()})
	return out
}
