// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License
//
//gostrict:header
package ui

import (
	"go.wdy.de/nago/pkg/gostrict/std/js"
	nui "go.wdy.de/nago/presentation/ui"
)

// TWindowTitle is a utility component, which sets the title of the browser or application window, see WindowTitle.
type TWindowTitle struct{ ref js.Ref }

func (TWindowTitle) isView() {}

// WindowTitle returns a component, which sets the title of the window to title.
//
//gostrict:host
func WindowTitle(title string) TWindowTitle {
	return TWindowTitle{ref: wrap(nui.WindowTitle(title))}
}

// Title sets the title of the window.
//
//gostrict:host
func (c TWindowTitle) Title(title string) TWindowTitle {
	return TWindowTitle{ref: wrap(get[nui.TWindowTitle](c.ref).Title(title))}
}

// H1 returns a level 1 heading (page title), which also sets the title of the window.
//
//gostrict:host
func H1(title string) TStack {
	return TStack{ref: wrap(nui.H1(title))}
}

// H2 returns a level 2 heading with a slightly smaller bold font and a standard horizontal line.
//
//gostrict:host
func H2(title string) TStack {
	return TStack{ref: wrap(nui.H2(title))}
}

// H3 returns a level 3 heading.
//
//gostrict:host
func H3(title string) TStack {
	return TStack{ref: wrap(nui.H3(title))}
}

// H4 returns a level 4 heading.
//
//gostrict:host
func H4(title string) TStack {
	return TStack{ref: wrap(nui.H4(title))}
}

// H5 returns a level 5 heading.
//
//gostrict:host
func H5(title string) TStack {
	return TStack{ref: wrap(nui.H5(title))}
}

// H6 returns a level 6 heading.
//
//gostrict:host
func H6(title string) TStack {
	return TStack{ref: wrap(nui.H6(title))}
}

// Heading returns a heading of the level, where 1 is the page heading H1 and so forth. A level 1 heading also sets the
// title of the window.
//
//gostrict:host
func Heading(level int32, title string) TStack {
	return TStack{ref: wrap(nui.Heading(int(level), title))}
}

// ContentHeading is like Heading, but never sets the title of the window. It is meant for headings within content,
// e.g. of a rendered document.
//
//gostrict:host
func ContentHeading(level int32, title string) TStack {
	return TStack{ref: wrap(nui.ContentHeading(int(level), title))}
}
