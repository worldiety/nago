// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package ssr provides a server-side renderer for the nago protocol tree.
// It converts a proto.Component tree into a pkg/dom HTML structure with
// inline CSS styles – analogous to the TypeScript helpers in
// web/vuejs/src/components/shared/.
package ssr

import (
	"fmt"
	"strings"

	"go.wdy.de/nago/presentation/proto"
)

// LengthCSS converts a proto.Length into a CSS length string.
// Port of cssLengthValue() from length.ts:
//   - replaces "dp" with "px"
//   - numeric values are returned as-is
//   - "calc(…)" expressions are returned as-is
//   - "auto" is returned as-is
//   - named tokens become var(--token)
func LengthCSS(l proto.Length) string {
	s := string(l)
	if s == "" {
		return ""
	}
	s = strings.ReplaceAll(s, "dp", "px")
	if s[0] == '-' || (s[0] >= '0' && s[0] <= '9') {
		return s
	}
	if strings.HasPrefix(s, "calc") {
		return s
	}
	if s == "auto" {
		return s
	}
	return fmt.Sprintf("var(--%s)", s)
}

// LengthCSSOrZero is like LengthCSS but returns "0px" for empty lengths.
// Port of cssLengthValue0Px() from length.ts.
func LengthCSSOrZero(l proto.Length) string {
	if l == "" {
		return "0px"
	}
	return LengthCSS(l)
}

// FrameCSS converts a proto.Frame into CSS declarations.
// Port of frameCSS() from frame.ts.
func FrameCSS(f proto.Frame) []string {
	var s []string
	if f.Width != "" {
		s = append(s, "width:"+LengthCSS(f.Width))
	}
	if f.MinWidth != "" {
		s = append(s, "min-width:"+LengthCSS(f.MinWidth))
	}
	if f.MaxWidth != "" {
		s = append(s, "max-width:"+LengthCSS(f.MaxWidth))
	}
	if f.Height != "" {
		s = append(s, "height:"+LengthCSS(f.Height))
	}
	if f.MinHeight != "" {
		s = append(s, "min-height:"+LengthCSS(f.MinHeight))
	}
	if f.MaxHeight != "" {
		s = append(s, "max-height:"+LengthCSS(f.MaxHeight))
	}
	return s
}

// PaddingCSS converts a proto.Padding into CSS declarations.
// Negative values become margin-* (same behaviour as paddingCSS() in padding.ts).
func PaddingCSS(p proto.Padding) []string {
	var s []string
	addPad := func(side string, l proto.Length) {
		if l == "" {
			return
		}
		v := LengthCSS(l)
		if strings.HasPrefix(string(l), "-") {
			s = append(s, fmt.Sprintf("margin-%s:%s", side, v))
		} else {
			s = append(s, fmt.Sprintf("padding-%s:%s", side, v))
		}
	}
	addPad("top", p.Top)
	addPad("bottom", p.Bottom)
	addPad("left", p.Left)
	addPad("right", p.Right)
	return s
}

// BorderCSS converts a proto.Border into CSS declarations.
// Port of borderCSS() from border.ts (including box-shadow).
func BorderCSS(b proto.Border) []string {
	var s []string

	// Only emit border-style when there is at least one border width set,
	// to avoid "border-style: solid" on elements with no visible border.
	hasBorderWidth := b.TopWidth != "" || b.BottomWidth != "" ||
		b.LeftWidth != "" || b.RightWidth != ""

	if hasBorderWidth {
		switch b.BorderStyle {
		case proto.StyleDotted:
			s = append(s, "border-style:dotted")
		case proto.StyleDashed:
			s = append(s, "border-style:dashed")
		default:
			s = append(s, "border-style:solid")
		}
	}

	// radii
	if b.TopLeftRadius != "" {
		s = append(s, "border-top-left-radius:"+LengthCSS(b.TopLeftRadius))
	}
	if b.TopRightRadius != "" {
		s = append(s, "border-top-right-radius:"+LengthCSS(b.TopRightRadius))
	}
	if b.BottomLeftRadius != "" {
		s = append(s, "border-bottom-left-radius:"+LengthCSS(b.BottomLeftRadius))
	}
	if b.BottomRightRadius != "" {
		s = append(s, "border-bottom-right-radius:"+LengthCSS(b.BottomRightRadius))
	}

	// widths
	if b.TopWidth != "" {
		s = append(s, "border-top-width:"+LengthCSS(b.TopWidth))
	}
	if b.BottomWidth != "" {
		s = append(s, "border-bottom-width:"+LengthCSS(b.BottomWidth))
	}
	if b.LeftWidth != "" {
		s = append(s, "border-left-width:"+LengthCSS(b.LeftWidth))
	}
	if b.RightWidth != "" {
		s = append(s, "border-right-width:"+LengthCSS(b.RightWidth))
	}

	// colors
	if b.TopColor != "" {
		s = append(s, "border-top-color:"+string(b.TopColor))
	}
	if b.BottomColor != "" {
		s = append(s, "border-bottom-color:"+string(b.BottomColor))
	}
	if b.LeftColor != "" {
		s = append(s, "border-left-color:"+string(b.LeftColor))
	}
	if b.RightColor != "" {
		s = append(s, "border-right-color:"+string(b.RightColor))
	}

	// box-shadow
	if !b.BoxShadow.IsZero() {
		radius := b.BoxShadow.Radius
		if radius == "" {
			radius = "10px"
		}
		color := b.BoxShadow.Color
		if color == "" {
			color = "#00000020"
		}
		x := b.BoxShadow.X
		if x == "" {
			x = "0px"
		}
		y := b.BoxShadow.Y
		if y == "" {
			y = "0px"
		}
		s = append(s, fmt.Sprintf("box-shadow:%s %s %s 0 %s",
			LengthCSSOrZero(x), LengthCSSOrZero(y), LengthCSS(radius), string(color)))
	}

	return s
}

// AlignmentMainAxisCSS returns the CSS justify-content value for the main axis.
func AlignmentMainAxisCSS(a proto.Alignment) string {
	switch a {
	case proto.Leading, proto.TopLeading, proto.BottomLeading:
		return "flex-start"
	case proto.Trailing, proto.TopTrailing, proto.BottomTrailing:
		return "flex-end"
	case proto.Stretch:
		return "stretch"
	default:
		return "center"
	}
}

// AlignmentCrossAxisCSS returns the CSS align-items value for the cross axis.
func AlignmentCrossAxisCSS(a proto.Alignment) string {
	switch a {
	case proto.Top, proto.TopLeading, proto.TopTrailing:
		return "flex-start"
	case proto.Bottom, proto.BottomLeading, proto.BottomTrailing:
		return "flex-end"
	case proto.Stretch:
		return "stretch"
	default:
		return "center"
	}
}

// JoinCSS joins CSS declarations with semicolons.
func JoinCSS(parts ...[]string) string {
	var all []string
	for _, p := range parts {
		all = append(all, p...)
	}
	return strings.Join(all, ";")
}

// TableAlignmentCSS returns vertical-align and text-align for a cell alignment.
// Default (zero) for body cells is Leading; for header cells pass isHeader=true (→ Center).
func TableAlignmentCSS(a proto.Alignment, isHeader bool) []string {
	// if zero value, apply the spec default
	if a == 0 {
		if isHeader {
			a = proto.Center
		} else {
			a = proto.Leading
		}
	}
	switch a {
	case proto.Leading:
		return []string{"vertical-align:middle", "text-align:start"}
	case proto.Trailing:
		return []string{"vertical-align:middle", "text-align:end"}
	case proto.Center:
		return []string{"vertical-align:middle", "text-align:center"}
	case proto.TopLeading:
		return []string{"vertical-align:top", "text-align:start"}
	case proto.BottomLeading:
		return []string{"vertical-align:bottom", "text-align:start"}
	case proto.TopTrailing:
		return []string{"vertical-align:top", "text-align:end"}
	case proto.Top:
		return []string{"vertical-align:top", "text-align:center"}
	case proto.BottomTrailing:
		return []string{"vertical-align:bottom", "text-align:end"}
	case proto.Bottom:
		return []string{"vertical-align:bottom", "text-align:center"}
	default:
		return []string{"vertical-align:middle", "text-align:start"}
	}
}
