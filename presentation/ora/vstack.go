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
	BackgroundColor NamedColor `json:"bgc,omitempty"`
	Padding         Padding    `json:"p,omitempty"`
	component
}
