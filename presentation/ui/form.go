// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
)

// TForm is a composite component (Form).
// It represents a form container that can group input fields and controls.
// A form can handle submission via an action callback, be styled with a frame,
// and optionally support autocomplete behavior.
type TForm struct {
	children     []core.View // child views (input fields, buttons, etc.)
	id           string      // unique identifier for the form
	action       func()      // callback invoked on form submission
	autocomplete bool        // enables/disables browser autocomplete
	frame        Frame       // layout frame for sizing and positioning
}

// Form creates a new form containing the given child views.
func Form(children ...core.View) TForm {
	return TForm{
		children: children,
	}
}

// ID assigns a unique identifier to the form.
func (c TForm) ID(id string) TForm {
	c.id = id
	return c
}

// Action sets the callback function to be executed when the form is submitted.
func (c TForm) Action(action func()) TForm {
	c.action = action
	return c
}

// Frame sets the layout frame of the form, including size and positioning.
func (c TForm) Frame(frame Frame) TForm {
	c.frame = frame
	return c
}

// Autocomplete enables or disables browser autocomplete for the form.
func (c TForm) Autocomplete(b bool) TForm {
	c.autocomplete = b
	return c
}

// Render builds and returns the protocol representation of the form,
// including its children, action callback, ID, autocomplete setting,
// and layout frame.
func (c TForm) Render(ctx core.RenderContext) core.RenderNode {
	return &proto.Form{
		Children:     renderComponents(ctx, c.children),
		Action:       ctx.MountCallback(c.action),
		Id:           proto.Str(c.id),
		Autocomplete: proto.Bool(c.autocomplete),
		Frame:        c.frame.ora(),
	}
}
