// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package canvas2 draws on a 2D canvas with a program of the gostrict subset, which runs in the frontend, see package
// go.wdy.de/nago/pkg/gostrict/lib/canvas. Unlike package canvas, which streams each command from the server, the
// program draws each frame itself, so it reacts to the input of the user within the same frame.
package canvas2

import (
	"context"

	"go.wdy.de/nago/pkg/gostrict"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/internal"
	"go.wdy.de/nago/presentation/ui/program"
)

// TCanvas is a basic component (Canvas2).
// It runs a program of the gostrict subset in the frontend, which draws with go.wdy.de/nago/pkg/gostrict/lib/canvas
// into the context, which canvas.Show hands to the host. The component scales the canvas to the device pixels and
// calls the function of [TCanvas.Animate] in each animation frame. The program exchanges messages with its
// backend, see [TCanvas.OnMessage].
type TCanvas struct {
	id        string
	spec      program.Spec
	frame     ui.Frame
	invisible bool
}

// Canvas creates a component, which runs the program of the bundle, e.g. one of [program.MustCompile]. The id must
// be unique within the window, because messages and input events refer to it.
func Canvas(id string, b *gostrict.Bundle) TCanvas {
	return TCanvas{id: id, spec: program.Spec{Bundle: b}}
}

// Init sets the exported function, which the component calls first, and its arguments as neutral values, see
// [gostrict.ValueKind]. If it returns a state, the component passes the state to the other functions, which take
// one.
func (c TCanvas) Init(name string, args ...any) TCanvas {
	c.spec.Init = name
	c.spec.Args = args
	return c
}

// Animate sets the exported function, which the component calls in each animation frame with the time in seconds,
// e.g. func Frame(s *State, t float64).
func (c TCanvas) Animate(name string) TCanvas {
	c.spec.Frame = name
	return c
}

// Resize sets the exported function, which the component calls with the aspect ratio, with the width and the
// height or with the width, the height and the device pixel ratio of its area in CSS pixels, before the first
// frame and whenever the size changes.
func (c TCanvas) Resize(name string) TCanvas {
	c.spec.Resize = name
	return c
}

// Inputs sets the input events, which the component passes to the handler of input.Event of the program, see
// package go.wdy.de/nago/pkg/gostrict/lib/input. If the program has such a handler and no inputs are set, it
// receives the pointer, wheel and key events.
func (c TCanvas) Inputs(types ...core.InputEventType) TCanvas {
	c.spec.Inputs = types
	return c
}

// OnMessage sets the function, which handles the messages of the program, usually the Dispatch of a
// [gostrict.Router] with the typed functions of the backend. The backend sends messages to the program through
// [program.ChannelOf] with the id of the component.
func (c TCanvas) OnMessage(fn func(ctx context.Context, e gostrict.Envelope) error) TCanvas {
	c.spec.Receive = fn
	return c
}

// Visible hides the component, if false.
func (c TCanvas) Visible(visible bool) TCanvas {
	c.invisible = !visible
	return c
}

// Frame sets the layout of the component. It should have a size, e.g. a fixed height, because the program draws
// into the area, which the component gets.
func (c TCanvas) Frame(frame ui.Frame) TCanvas {
	c.frame = frame
	return c
}

func (c TCanvas) Render(ctx core.RenderContext) core.RenderNode {
	if c.invisible {
		return nil
	}

	spec := c.spec
	if len(spec.Inputs) == 0 {
		spec.Inputs = defaultInputs
	}
	return &proto.Canvas2{
		Id:      proto.Str(c.id),
		Frame:   internal.FrameToOra(c.frame),
		Program: spec.Render(ctx.Window(), c.id),
	}
}

// defaultInputs are passed to a handler of input.Event, if the component sets none.
var defaultInputs = []core.InputEventType{
	core.InputEventPointerDown,
	core.InputEventPointerMove,
	core.InputEventPointerUp,
	core.InputEventPointerCancel,
	core.InputEventMouseWheel,
	core.InputEventKeyDown,
	core.InputEventKeyUp,
}
