package ora

// A Box aligns children elements in absolute within its bounds.
// - there is no intrinsic component dimension, so you have to set it by hand
// - z-order is defined as defined children order, thus later children are put on top of others
// - it is undefined behavior, to define multiple children with the same alignment. So this must not be rendered.
//
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Box struct {
	Type     ComponentType      `json:"type" value:"bx"`
	Children []AlignedComponent `json:"c,omitempty"`
	// Frame is omitted if empty
	Frame Frame `json:"frame,omitempty"`

	// BackgroundColor regular is always transparent
	BackgroundColor Color   `json:"bgc,omitempty"`
	Padding         Padding `json:"p,omitempty"`
	Border          Border  `json:"b,omitempty"`
	component
}

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type AlignedComponent struct {
	Component Component `json:"c,omitempty"`

	// Alignment may be empty and omitted. Then Center (=c) must be applied.
	Alignment Alignment `json:"a,omitempty"`
}
