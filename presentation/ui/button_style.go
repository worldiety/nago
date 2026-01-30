// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui

import (
	"go.wdy.de/nago/presentation/proto"
)

// deprecated: use ButtonStylePreset
type StylePreset = ButtonStyle

// ButtonStyle allows to apply a build-in style to this component. This reduces over-the-wire boilerplate and
// also defines a stereotype, so that the applied component behavior may be indeed a bit different, because
// a native component may be used, e.g. for a native button. The order of appliance is first the preset and
// then customized properties on top.
type ButtonStyle uint

func (p ButtonStyle) ora() proto.StylePreset {
	return proto.StylePreset(p)
}

func (p ButtonStyle) String() string {
	switch p {
	case ButtonStylePrimary:
		return "primary"
	case ButtonStyleSecondary:
		return "secondary"
	case ButtonStyleTertiary:
		return "tertiary"
	default:
		return "unknown"
	}
}

const (
	// deprecated: use ButtonStylePrimary
	StyleButtonPrimary = ButtonStylePrimary
	// deprecated: use ButtonStyleSecondary
	StyleButtonSecondary = ButtonStyleSecondary
	// deprecated: use StyleButtonTertiary
	StyleButtonTertiary = ButtonStyleTertiary
)

const (
	ButtonStylePrimary   = ButtonStyle(proto.StyleButtonPrimary)
	ButtonStyleSecondary = ButtonStyle(proto.StyleButtonSecondary)
	ButtonStyleTertiary  = ButtonStyle(proto.StyleButtonTertiary)
)

func ButtonStyles() []ButtonStyle {
	return []ButtonStyle{
		ButtonStylePrimary,
		ButtonStyleSecondary,
		ButtonStyleTertiary,
	}
}
