package ora

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Image struct {
	Type ComponentType `json:"type" value:"I"`
	URI  URI           `json:"u,omitempty"`
	// see also https://www.w3.org/WAI/tutorials/images/decision-tree/
	AccessibilityLabel string  `json:"al,omitempty"`
	Invisible          bool    `json:"iv,omitempty"`
	Border             Border  `json:"b,omitempty"`
	Frame              Frame   `json:"f,omitempty"`
	Padding            Padding `json:"p,omitempty"`
	SVG                SVG     `json:"s,omitempty"`
	CachedSVG          Ptr     `json:"v,omitempty"`
	FillColor          Color   `json:"c,omitempty"`
	StrokeColor        Color   `json:"k,omitempty"`
	component
}
