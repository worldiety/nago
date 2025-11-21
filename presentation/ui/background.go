// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui

import (
	"strings"

	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
)

type Background struct {
	URL   core.URI
	Fit   ObjectFit
	lGrad []Color
}

func (b Background) LinearGradient(colors ...Color) Background {
	b.lGrad = colors
	return b
}

func (b *Background) proto() *proto.Background {
	if b == nil {
		return nil
	}

	bg := &proto.Background{}

	switch b.Fit {
	case FitFill:
		bg.Size = "100% 100%"
	case FitContain:
		bg.Size = "contain"
		bg.PositionX = 50
		bg.PositionY = 50
		bg.Repeat = "no-repeat"
	case FitCover:
		bg.Size = "cover"
	case FitNone:
		bg.Size = "contain"
		bg.Repeat = "repeat"
	}

	if len(b.lGrad) > 0 {
		var tmp strings.Builder
		for i, color := range b.lGrad {
			if color.isAbsolute() {
				tmp.WriteString(string(color))
			} else {
				tmp.WriteString("var(--")
				tmp.WriteString(string(color))
				tmp.WriteString(")")
			}

			if i < len(b.lGrad)-1 {
				tmp.WriteString(", ")
			}
		}
		bg.Image = append(bg.Image, proto.Str("linear-gradient("+tmp.String()+")"))
	}

	// implementation note: order is important and must be stacked accordingly
	if b.URL != "" {
		bg.Image = append(bg.Image, "url("+proto.Str(b.URL)+")")
	}

	return bg
}
