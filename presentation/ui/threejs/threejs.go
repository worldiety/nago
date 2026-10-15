// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package threejs renders 3D scenes with three.js, which a program of the gostrict subset builds and animates in
// the frontend, see package go.wdy.de/nago/pkg/gostrict/lib/three.
package threejs

import (
	"context"

	"go.wdy.de/nago/pkg/gostrict"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/internal"
	"go.wdy.de/nago/presentation/ui/program"
)

// TThreeJS is a basic component (ThreeJS).
// It runs a program of the gostrict subset in the frontend, which builds a three.js scene with
// go.wdy.de/nago/pkg/gostrict/lib/three and shows it with three.Show. The component renders the scene in each
// animation frame after it called the function of [TThreeJS.Animate], so an animation runs at the frame rate of
// the browser without the server. The program exchanges messages with its backend, see [TThreeJS.OnMessage].
type TThreeJS struct {
	id        string
	spec      program.Spec
	frame     ui.Frame
	invisible bool
	antialias bool
}

// ThreeJS creates a component, which runs the program of the bundle, e.g. one of [program.MustCompile]. The id must
// be unique within the window, because messages and input events refer to it.
func ThreeJS(id string, b *gostrict.Bundle) TThreeJS {
	return TThreeJS{id: id, spec: program.Spec{Bundle: b}}
}

// Init sets the exported function, which the component calls first, and its arguments as neutral values, see
// [gostrict.ValueKind]. If it returns a state, the component passes the state to the other functions, which take
// one.
func (c TThreeJS) Init(name string, args ...any) TThreeJS {
	c.spec.Init = name
	c.spec.Args = args
	return c
}

// Animate sets the exported function, which the component calls in each animation frame with the time in seconds,
// e.g. func Frame(s *State, t float64).
func (c TThreeJS) Animate(name string) TThreeJS {
	c.spec.Frame = name
	return c
}

// Resize sets the exported function, which the component calls with the aspect ratio, with the width and the
// height or with the width, the height and the device pixel ratio of its area in CSS pixels, before the first
// frame and whenever the size changes.
func (c TThreeJS) Resize(name string) TThreeJS {
	c.spec.Resize = name
	return c
}

// Inputs sets the input events, which the component passes to the handler of input.Event of the program, see
// package go.wdy.de/nago/pkg/gostrict/lib/input. If the program has such a handler and no inputs are set, it
// receives the pointer, wheel and key events.
func (c TThreeJS) Inputs(types ...core.InputEventType) TThreeJS {
	c.spec.Inputs = types
	return c
}

// OnMessage sets the function, which handles the messages of the program, usually the Dispatch of a
// [gostrict.Router] with the typed functions of the backend. The backend sends messages to the program through
// [program.ChannelOf] with the id of the component.
func (c TThreeJS) OnMessage(fn func(ctx context.Context, e gostrict.Envelope) error) TThreeJS {
	c.spec.Receive = fn
	return c
}

// Antialias smooths the edges of the scene at some cost of performance.
func (c TThreeJS) Antialias(antialias bool) TThreeJS {
	c.antialias = antialias
	return c
}

// Visible hides the component, if false.
func (c TThreeJS) Visible(visible bool) TThreeJS {
	c.invisible = !visible
	return c
}

// Frame sets the layout of the component. It should have a size, e.g. a fixed height, because the program draws
// into the area, which the component gets.
func (c TThreeJS) Frame(frame ui.Frame) TThreeJS {
	c.frame = frame
	return c
}

func (c TThreeJS) Render(ctx core.RenderContext) core.RenderNode {
	if c.invisible {
		return nil
	}

	spec := c.spec
	if len(spec.Inputs) == 0 {
		spec.Inputs = defaultInputs
	}
	return &proto.ThreeJS{
		Id:        proto.Str(c.id),
		Frame:     internal.FrameToOra(c.frame),
		Program:   spec.Render(ctx.Window(), c.id),
		Antialias: proto.Bool(c.antialias),
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
