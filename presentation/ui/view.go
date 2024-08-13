package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
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

// A Composable is a lazy factory of a view which is evaluated lately in the render cycle.
// This can be used, to more efficiently post-pone resource allocations in conditional renderings.
type Composable func() core.View

func (c Composable) Render(context core.RenderContext) ora.Component {
	return c().Render(context)
}
