package ui

import "go.wdy.de/nago/presentation/ora"

type PositionType int

const (
	// PositionDefault is the default and any explicit position value have no effect.
	// See also https://developer.mozilla.org/de/docs/Web/CSS/position#static.
	PositionDefault PositionType = iota
	// PositionOffset is like PositionDefault but moves the element by applying the given position values after
	// layouting. See also https://developer.mozilla.org/de/docs/Web/CSS/position#relative.
	PositionOffset
	// PositionAbsolute removes the element from the layout and places it using the given values in an absolute way
	// within any of its parent layouted as PositionOffset. If no parent with PositionOffset is found, the viewport
	// is used. See also https://developer.mozilla.org/de/docs/Web/CSS/position#absolute. The position and
	// size will not be accounted within its parent, thus you need to provide the parent size either explicitly
	// or explicitly set from the parents parent.
	PositionAbsolute
	// PositionFixed removes the element from the layout and places it at a fixed position according to the viewport
	// independent of the scroll position. See also https://developer.mozilla.org/de/docs/Web/CSS/position#absolute.
	PositionFixed
	// PositionSticky is here for completion, and it is unclear which rules to follow on mobile clients.
	// See also https://developer.mozilla.org/de/docs/Web/CSS/position#absolute.
	PositionSticky
)

type Position struct {
	Type PositionType
	// Left defines the absolute length from within the left border of the (anchor) parent.
	Left Length
	// Top defines the absolute length from within the top border of the (anchor) parent.
	Top Length
	// Right defines the absolute length seen from the right border of the (anchor) parent.
	// Note, that you must not define the right but can instead set an explicit width.
	Right Length
	// Bottom defines the absolute length seen from the bottom border of the (anchor) parent.
	// Note, that you must not define the bottom but can instead set an explicit height.
	Bottom Length
}

func (p Position) ora() ora.Position {
	return ora.Position{
		Left:   p.Left.ora(),
		Top:    p.Top.ora(),
		Right:  p.Right.ora(),
		Bottom: p.Bottom.ora(),
		Kind:   ora.PositionType(p.Type),
	}
}
