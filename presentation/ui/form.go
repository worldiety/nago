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

type TForm struct {
	children     []core.View
	id           string
	action       func()
	autocomplete bool
	frame        Frame
}

func Form(children ...core.View) TForm {
	return TForm{
		children: children,
	}
}

func (c TForm) ID(id string) TForm {
	c.id = id
	return c
}

func (c TForm) Action(action func()) TForm {
	c.action = action
	return c
}

func (c TForm) Frame(frame Frame) TForm {
	c.frame = frame
	return c
}

func (c TForm) Autocomplete(b bool) TForm {
	c.autocomplete = b
	return c
}

func (c TForm) Render(ctx core.RenderContext) core.RenderNode {
	return &proto.Form{
		Children:     renderComponents(ctx, c.children),
		Action:       ctx.MountCallback(c.action),
		Id:           proto.Str(c.id),
		Autocomplete: proto.Bool(c.autocomplete),
		Frame:        c.frame.ora(),
	}
}
