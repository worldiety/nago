package ora

// A ScrollView can either be horizontal or vertical.
//
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type ScrollView struct {
	Type            ComponentType  `json:"type" value:"V"`
	Content         Component      `json:"c,omitempty"`
	Axis            ScrollViewAxis `json:"a,omitempty"`
	Invisible       bool           `json:"iv,omitempty"`
	Border          Border         `json:"b,omitempty"`
	Frame           Frame          `json:"f,omitempty"`
	Padding         Padding        `json:"p,omitempty"`
	BackgroundColor Color          `json:"bgc,omitempty"`
	component
}

// ScrollViewAxis is by default vertical (v). Alternatively is horizontal (h).
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type ScrollViewAxis string
