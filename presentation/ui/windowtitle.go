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

// TWindowTitle is an utility component(Window Title).
// This component sets the browser or application window title which is also displayed in the browser tab.
type TWindowTitle struct {
	title string
}

func WindowTitle(title string) TWindowTitle {
	return TWindowTitle{title: title}
}

func (c TWindowTitle) Render(ctx core.RenderContext) core.RenderNode {
	return &proto.WindowTitle{
		Value: proto.Str(c.title),
	}
}

// H1 creates a level 1 heading (page title) and sets the window title accordingly.
func H1(title string) core.View {
	return Heading(1, title)
}

// H2 creates a level 2 heading with a slightly smaller bold font and a standard horizontal line.
func H2(title string) core.View {
	return Heading(2, title)
}

// Heading returns a default formatted heading text. Level 1 is page heading H1 and so forth. H1 levels also
// set automatically the window title.
func Heading(level int, title string) core.View {
	switch level {
	case 1:
		return VStack(
			WindowTitle(title),
			Text(title).Font(Font{
				Size:   "2rem",
				Weight: HeadlineAndTitleFontWeight,
			}),
			HLineWithColor(ColorAccent),
		).Alignment(Leading).Padding(Padding{Bottom: Length("2rem")})
	case 2:
		return VStack(
			Text(title).Font(Font{
				Size:   "1.2rem",
				Weight: HeadlineAndTitleFontWeight,
			}),
			HLine(),
		).Alignment(Leading).Padding(Padding{Bottom: Length("2rem")})
	case 3:
		return VStack(Text(title).Font(Title))
	default:
		return VStack(Text(title).Font(SubTitle))
	}
}
