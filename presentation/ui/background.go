// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui

import (
	"slices"
	"strings"

	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
)

type Background struct {
	fit         ObjectFit
	effectStack []proto.Str
}

func (b Background) Fit(fit ObjectFit) Background {
	b.fit = fit
	return b
}

// AppendURI appends the given uri on the background effect stack. Beware of the effect order,
// which is defined as each append comes on top of the prior
// layer. This is the inverse order of CSS but technically more logical.
func (b Background) AppendURI(uri core.URI) Background {
	b.effectStack = append(b.effectStack, "url("+proto.Str(uri)+")")
	return b
}

// AppendLinearGradient appends the defined gradient equally distributed between all given colors on the background
// effect stack. Beware of the effect order, which is defined as each append comes on top of the prior
// layer. This is the inverse order of CSS but technically more logical.
func (b Background) AppendLinearGradient(colors ...Color) Background {
	var tmp strings.Builder
	for i, color := range colors {
		if color.isAbsolute() {
			tmp.WriteString(string(color))
		} else {
			tmp.WriteString("var(--")
			tmp.WriteString(string(color))
			tmp.WriteString(")")
		}

		if i < len(colors)-1 {
			tmp.WriteString(", ")
		}
	}

	b.effectStack = append(b.effectStack, proto.Str("linear-gradient("+tmp.String()+")"))
	return b
}

func (b *Background) proto() *proto.Background {
	if b == nil {
		return nil
	}

	bg := &proto.Background{}

	switch b.fit {
	case FitFill:
		bg.Size = "100% 100%"
	case FitContain:
		bg.Size = "contain"
		bg.PositionX = 50
		bg.PositionY = 50
		bg.Repeat = "no-repeat"
	case FitCover:
		bg.Size = "cover"
		bg.PositionX = 50
		bg.PositionY = 50
	case FitNone:
		bg.Size = "contain"
		bg.Repeat = "repeat"
	}

	slices.Reverse(b.effectStack)
	bg.Image = b.effectStack

	return bg
}
