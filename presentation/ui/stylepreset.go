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

// StylePreset allows to apply a build-in style to this component. This reduces over-the-wire boilerplate and
// also defines a stereotype, so that the applied component behavior may be indeed a bit different, because
// a native component may be used, e.g. for a native button. The order of appliance is first the preset and
// then customized properties on top.
type StylePreset uint

func (p StylePreset) ora() proto.StylePreset {
	return proto.StylePreset(p)
}

const (
	StyleButtonPrimary   StylePreset = StylePreset(proto.StyleButtonPrimary)
	StyleButtonSecondary StylePreset = StylePreset(proto.StyleButtonSecondary)
	StyleButtonTertiary  StylePreset = StylePreset(proto.StyleButtonTertiary)
)
