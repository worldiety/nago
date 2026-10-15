// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package input declares the input events, which a nago component like threejs or canvas2 passes to its program.
// They are the input events of nago itself, see core.InputEvent, but go to the program in the frontend without a
// round trip to the server. A program receives them with a handler, like a message of its backend:
//
//	func OnInput(g *Game, e input.Event) {
//		switch e.Type {
//		case input.PointerDown:
//			...
//		}
//	}
//
// The component delivers the events between two frames. It passes all pointer down, up and cancel events and key
// events, but only the last pointer move of each pointer and the sum of the wheel events since the last frame.
//
// The fields, which an event sets:
//
//	PointerDown, PointerMove, PointerUp, PointerCancel  X, Y, PointerID, Buttons, Pressure, Shift, Ctrl, Alt, Meta
//	MouseWheel                                           X, Y, DeltaX, DeltaY, DeltaZ in pixels (DeltaY > 0 scrolls down), modifiers
//	KeyDown, KeyUp                                       Code (e.g. KeyA, Digit1, ArrowLeft, Space, Delete), Key, Repeat, modifiers
//
// On macOS, the command key sets Meta, so an undo checks e.Ctrl || e.Meta. The component focuses its canvas, when
// the user touches it, so that the keys arrive.
package input

// Type is the kind of an event. Its values are those of core.InputEventType.
type Type uint32

const (
	PointerUp     Type = 1
	PointerDown   Type = 2
	PointerMove   Type = 3
	PointerCancel Type = 4
	KeyDown       Type = 5
	KeyUp         Type = 6
	Invalidate    Type = 7
	MouseDown     Type = 8
	MouseEnter    Type = 9
	MouseLeave    Type = 10
	MouseOut      Type = 11
	MouseMove     Type = 12
	MouseOver     Type = 13
	MouseUp       Type = 14
	MouseWheel    Type = 15
	PointerLeave  Type = 16
	PointerOut    Type = 17
	PointerOver   Type = 18
	Blur          Type = 19
)

// Event is an input event of the component. Its fields are those of core.InputEvent.
type Event struct {
	Type Type
	// X and Y are the position in CSS pixels relative to the component.
	X, Y float64
	// Code is the physical key of a key event, e.g. KeyA or ArrowLeft.
	Code string
	// PointerID distinguishes the pointers of a multi touch.
	PointerID int32
	// Buttons is the bit mask of the pressed buttons: 1 primary, 2 secondary, 4 auxiliary.
	Buttons uint32
	// Pressure of the pointer from 0 to 1.
	Pressure float64
	Shift    bool
	Ctrl     bool
	Alt      bool
	Meta     bool
	// DeltaX, DeltaY and DeltaZ are the amounts of a wheel event in pixels.
	DeltaX, DeltaY, DeltaZ float64
	// Key is the value of the key, which respects the keyboard layout, e.g. a or Enter.
	Key string
	// Repeat tells, whether the key is held down.
	Repeat bool
}
