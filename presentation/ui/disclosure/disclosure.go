// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package disclosure provides a section, which opens and closes in the browser without a round trip to the server.
//
// In contrast to the accordion, whose open state lives in a server state, a disclosure keeps its open state in the
// browser (a native details element). A click neither sends an event nor causes a rendering, so it fits for sections
// whose content is rendered anyway and only shown or hidden, e.g. the settings of a detail page. Use the accordion
// instead, if the content shall only be rendered while the section is open.
package disclosure

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
	"go.wdy.de/nago/presentation/ui"
)

// TDisclosure is a section with a summary, which the user opens and closes in the browser to show or hide its
// content. The content is always rendered and transmitted.
type TDisclosure struct {
	summary    core.View
	content    core.View
	frame      ui.Frame
	open       bool
	inputValue *core.State[bool]
	separator  bool
	id         string
}

// Disclosure creates a closed section with the given summary, which is always visible, and the content, which is
// shown while the section is open. Like the accordion, it shows a separator below by default.
func Disclosure(summary, content core.View) TDisclosure {
	return TDisclosure{
		summary:   summary,
		content:   content,
		separator: true,
	}
}

// Open sets the open state, which the browser applies when the disclosure appears, e.g. for a deep link to a
// section. Afterward, the browser keeps the state the user has chosen across renderings. Only if a rendering changes
// the open state compared to the previous rendering, the browser applies it again. Thus, a constant value never
// overrides what the user did. See also [TDisclosure.InputValue], which takes precedence.
func (t TDisclosure) Open(open bool) TDisclosure {
	t.open = open
	return t
}

// InputValue binds the open state to the given state. The disclosure appears with its value and applies a change
// made by the application, like [TDisclosure.Open]. The browser reports each change of the user to the state in the
// background: opening and closing does not wait for the server, but each report updates the state and therefore
// causes a rendering. Without a state, opening and closing costs no server event at all.
func (t TDisclosure) InputValue(state *core.State[bool]) TDisclosure {
	t.inputValue = state
	return t
}

// Frame sets the frame of the disclosure.
func (t TDisclosure) Frame(frame ui.Frame) TDisclosure {
	t.frame = frame
	return t
}

// FullWidth sets the width of the frame to full.
func (t TDisclosure) FullWidth() TDisclosure {
	t.frame.Width = ui.Full
	return t
}

// HideSeparator hides the separator below the disclosure.
func (t TDisclosure) HideSeparator() TDisclosure {
	t.separator = false
	return t
}

// ID sets an identifier, which must be either empty or unique within the view tree. The frontend uses it as the id
// of the details element, e.g. to scroll to a section of a deep link.
func (t TDisclosure) ID(id string) TDisclosure {
	t.id = id
	return t
}

func (t TDisclosure) Render(ctx core.RenderContext) core.RenderNode {
	open := t.open
	if t.inputValue != nil {
		open = t.inputValue.Get()
	}

	var summary, content proto.Component
	if t.summary != nil {
		summary = t.summary.Render(ctx)
	}

	if t.content != nil {
		content = t.content.Render(ctx)
	}

	return &proto.Disclosure{
		Summary: summary,
		Content: content,
		Frame: proto.Frame{
			MinWidth:  proto.Length(t.frame.MinWidth),
			MaxWidth:  proto.Length(t.frame.MaxWidth),
			MinHeight: proto.Length(t.frame.MinHeight),
			MaxHeight: proto.Length(t.frame.MaxHeight),
			Width:     proto.Length(t.frame.Width),
			Height:    proto.Length(t.frame.Height),
			FlexProperties: proto.FlexProperties{
				Grow:          proto.Bool(t.frame.FlexGrow),
				PreventShrink: proto.Bool(t.frame.FlexPreventShrink),
			},
		},
		InputValue: t.inputValue.Ptr(),
		Value:      proto.Bool(open),
		Separator:  proto.Bool(t.separator),
		Id:         proto.Str(t.id),
	}
}
