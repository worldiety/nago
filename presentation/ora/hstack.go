package ora

// An HStack aligns children elements in a horizontal row.
// - the intrinsic component dimensions are the sum of all sizes of the contained children
// - the parent can define a custom width and height
// - if the container is larger than the contained views, it must center vertical or horizontal
// - the inner gap between components should be around 2dp
//
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type HStack struct {
	Ptr      Ptr           `json:"id"`
	Type     ComponentType `json:"type" value:"hs"`
	Children []Component   `json:"children,omitempty"`
	// InnerGap is omitted, if empty
	InnerGap Length `json:"innerGap,omitempty"`
	// Frame is omitted if empty
	Frame Frame `json:"frame,omitempty"`
	// Alignment may be empty and omitted. Then Center (=c) must be applied.
	Alignment Alignment `json:"alignment,omitempty"`
	// BackgroundColor regular is always transparent
	BackgroundColor NamedColor `json:"backgroundColor,omitempty"`
	component
}
