package ora

// An VStack aligns children elements in a vertical column.
// - the intrinsic component dimensions are the sum of all sizes of the contained children
// - the parent can define a custom width and height
// - if the container is larger than the contained views, it must center vertical or horizontal
// - the inner gap between components should be around 2dp
//
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type VStack struct {
	Type     ComponentType `json:"type" value:"vs"`
	Children []Component   `json:"c,omitempty"`
	// InnerGap is omitted, if empty
	Gap Length `json:"g,omitempty"`
	// Frame is omitted if empty
	Frame Frame `json:"f,omitempty"`
	// Alignment may be empty and omitted. Then Center (=c) must be applied.
	Alignment Alignment `json:"a,omitempty"`
	// BackgroundColor regular is always transparent
	BackgroundColor Color   `json:"bgc,omitempty"`
	Padding         Padding `json:"p,omitempty"`
	Border          Border  `json:"b,omitempty"`
	// see also https://www.w3.org/WAI/tutorials/images/decision-tree/
	AccessibilityLabel string `json:"al,omitempty"`
	Invisible          bool   `json:"iv,omitempty"`
	Font               Font   `json:"fn,omitempty"`
	Action             Ptr    `json:"t,omitempty"`

	HoveredBackgroundColor Color  `json:"hgc,omitempty"`
	PressedBackgroundColor Color  `json:"pgc,omitempty"`
	FocusedBackgroundColor Color  `json:"fbc,omitempty"`
	HoveredBorder          Border `json:"hb,omitempty"`
	PressedBorder          Border `json:"pb,omitempty"`
	FocusedBorder          Border `json:"fb,omitempty"`

	component
}
