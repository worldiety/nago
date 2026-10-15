// Package proto declares the messages between the paint program and its backend.
package proto

import "go.wdy.de/nago/pkg/gostrict/std/list"

// Point is a point of a stroke in CSS pixels.
type Point struct {
	X, Y float64
}

// Stroke is a finished freehand stroke, which the program sends to the backend.
type Stroke struct {
	Points list.List[Point]
	// Color is 0xRRGGBBAA.
	Color uint32
	Width float64
}

// Clear tells the program to remove all strokes.
type Clear struct {
	// Reason is an optional explanation, e.g. for a log.
	Reason string
}
