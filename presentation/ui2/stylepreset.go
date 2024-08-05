package ui

import (
	"go.wdy.de/nago/presentation/ora"
)

// StylePreset allows to apply a build-in style to this component. This reduces over-the-wire boilerplate and
// also defines a stereotype, so that the applied component behavior may be indeed a bit different, because
// a native component may be used, e.g. for a native button. The order of appliance is first the preset and
// then customized properties on top.
type StylePreset string

func (p StylePreset) ora() ora.StylePreset {
	return ora.StylePreset(p)
}

const (
	StyleButtonPrimary   StylePreset = "p"
	StyleButtonSecondary StylePreset = "s"
	StyleButtonTertiary  StylePreset = "t"
)
