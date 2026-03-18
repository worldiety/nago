// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package switcher

import (
	"slices"

	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
	"go.wdy.de/nago/presentation/ui"
)

type SwitcherLayout int

const (
	SwitcherLayoutAuto SwitcherLayout = iota
	SwitcherLayoutVertical
	SwitcherLayoutHorizontal
)

// TSwitcher is a content component.
// You can add multiple content pages to it, that you can switch between.
// It is responsive and can switch between horizontal and vertical orientation.
type TSwitcher struct {
	pages            []TSwitcherPage
	id               string
	layout           SwitcherLayout
	frame            ui.Frame
	contentNoPadding bool
	dynamicHeight    bool
	value            string
	inputValue       *core.State[string]
}

// Switcher is a responsive variant which decides between HSwitcher and VSwitcher.
func Switcher(pages []TSwitcherPage, state *core.State[string]) TSwitcher {
	checkPages(pages)

	value := pages[0].id
	if state != nil && len(state.Get()) > 0 {
		value = state.Get()
	}

	return TSwitcher{
		pages:      pages,
		layout:     SwitcherLayoutAuto,
		value:      value,
		inputValue: state,
	}
}

// HSwitcher is a fixed horizontal variant of Switcher
func HSwitcher(pages []TSwitcherPage, state *core.State[string]) TSwitcher {
	checkPages(pages)

	value := pages[0].id
	if state != nil && len(state.Get()) > 0 {
		value = state.Get()
	}

	c := TSwitcher{
		pages:      pages,
		layout:     SwitcherLayoutHorizontal,
		value:      value,
		inputValue: state,
	}

	return c
}

// VSwitcher is a fixed vertical variant of Switcher
func VSwitcher(pages []TSwitcherPage, state *core.State[string]) TSwitcher {
	checkPages(pages)

	value := pages[0].id
	if state != nil && len(state.Get()) > 0 {
		value = state.Get()
	}

	c := TSwitcher{
		pages:      pages,
		layout:     SwitcherLayoutVertical,
		value:      value,
		inputValue: state,
	}

	return c
}

// checkPages checks for valid pages configuration (at least one page, no duplicate IDs, etc.)
func checkPages(pages []TSwitcherPage) {
	if len(pages) == 0 {
		panic("switcher must have at least one page")
	}

	temp := make([]string, 0)
	for _, page := range pages {
		if len(page.id) == 0 {
			panic("every switcher page must have an ID")
		}
		if slices.Contains(temp, page.id) {
			panic("switcher pages must have unique IDs")
		}
		temp = append(temp, page.id)
	}
}

// ID assigns a unique identifier to the switcher
func (c TSwitcher) ID(id string) TSwitcher {
	c.id = id
	return c
}

// Append adds more pages to the switcher
func (c TSwitcher) Append(pages ...TSwitcherPage) TSwitcher {
	c.pages = append(c.pages, pages...)
	return c
}

// Frame sets the switcher's frame
func (c TSwitcher) Frame(frame ui.Frame) TSwitcher {
	c.frame = frame
	return c
}

// FullWidth sets the switcher's frame to full width
func (c TSwitcher) FullWidth() TSwitcher {
	c.frame.Width = ui.Full
	return c
}

// ContentNoPadding sets whether the content part should use the prestyled padding
func (c TSwitcher) ContentNoPadding() TSwitcher {
	c.contentNoPadding = true
	return c
}

// DynamicHeight sets the switcher to dynamically change its height by the active page
func (c TSwitcher) DynamicHeight() TSwitcher {
	c.dynamicHeight = true
	return c
}

// InputValue binds the switcher to an external string state,
// allowing it to be controlled from outside the component.
func (c TSwitcher) InputValue(input *core.State[string]) TSwitcher {
	c.inputValue = input
	return c
}

// With applies a transformation function to the switcher itself and returns the result.
// Useful for chaining configuration in a functional style.
func (c TSwitcher) With(fn func(switcher TSwitcher) TSwitcher) TSwitcher {
	return fn(c)
}

// Layout sets the switcher's layout
func (c TSwitcher) Layout(layout SwitcherLayout) TSwitcher {
	c.layout = layout
	return c
}

func (c TSwitcher) Render(ctx core.RenderContext) core.RenderNode {
	wnd := ctx.Window()

	layout := c.layout
	if layout == SwitcherLayoutAuto {
		if wnd.Info().SizeClass < core.SizeClassMedium {
			layout = SwitcherLayoutVertical
		} else {
			layout = SwitcherLayoutHorizontal
		}
	}

	var orientation proto.Orientation
	if layout == SwitcherLayoutVertical {
		orientation = proto.Vertical
	} else {
		orientation = proto.Horizontal
	}

	pages := make(proto.SwitcherPages, 0)
	for _, page := range c.pages {
		myUri := page.lightUri
		if page.lightUri != "" && page.darkUri != "" {
			if ctx.Window().Info().ColorScheme == core.Dark {
				myUri = page.darkUri
			} else {
				myUri = page.lightUri
			}
		}

		pages = append(pages, proto.SwitcherPage{
			Id:      proto.Str(page.id),
			Title:   proto.Str(page.title),
			Toggle:  ui.HStack(ui.Image().Embed(page.icon).Frame(ui.Frame{}.Size(ui.L32, ui.L32))).Render(ctx),
			Content: page.content.Render(ctx),
			Img:     proto.URI(myUri),
		})
	}

	return &proto.Switcher{
		Id:    proto.Str(c.id),
		Pages: pages,
		Frame: proto.Frame{
			MinWidth:  proto.Length(c.frame.MinWidth),
			MaxWidth:  proto.Length(c.frame.MaxWidth),
			MinHeight: proto.Length(c.frame.MinHeight),
			MaxHeight: proto.Length(c.frame.MaxHeight),
			Width:     proto.Length(c.frame.Width),
			Height:    proto.Length(c.frame.Height),
		},
		DynamicHeight: proto.Bool(c.dynamicHeight),
		Orientation:   orientation,
		Value:         proto.Str(c.value),
		InputValue:    c.inputValue.Ptr(),
	}
}
