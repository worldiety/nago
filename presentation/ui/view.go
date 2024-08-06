package ui

import (
	"go.wdy.de/nago/presentation/core"
)

type DecoredView interface {
	core.View
	Padding(padding Padding) DecoredView
	Frame(frame Frame) DecoredView
	Border(border Border) DecoredView
	Visible(visible bool) DecoredView
	// AccessibilityLabel is used to help screen readers, see guidelines, when use them.
	// https://www.w3.org/WAI/tutorials/images/decision-tree/
	AccessibilityLabel(label string) DecoredView
}
