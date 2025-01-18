package ora

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
	// is used. See also https://developer.mozilla.org/de/docs/Web/CSS/position#absolute.
	PositionAbsolute
	// PositionFixed removes the element from the layout and places it at a fixed position according to the viewport
	// independent of the scroll position. See also https://developer.mozilla.org/de/docs/Web/CSS/position#absolute.
	PositionFixed
	// PositionSticky is here for completion, and it is unclear which rules to follow on mobile clients.
	// See also https://developer.mozilla.org/de/docs/Web/CSS/position#absolute.
	PositionSticky
)

type Position struct {
	Kind   PositionType `json:"k,omitempty"`
	Left   Length       `json:"l,omitempty"`
	Top    Length       `json:"t,omitempty"`
	Right  Length       `json:"r,omitempty"`
	Bottom Length       `json:"b,omitempty"`
}
