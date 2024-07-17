package ora

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Text struct {
	Type  ComponentType `json:"type" value:"T"`
	Value string        `json:"v,omitempty"`
	// Color denotes the text color. Leave empty, for the context sensitiv default theme color.
	Color Color `json:"c,omitempty"`

	// BackgroundColor denotes the color of the text background.  Leave empty, for the context sensitiv default theme color.
	BackgroundColor Color `json:"bgc,omitempty"`

	OnClick      Ptr    `json:"onClick,omitempty"`
	OnHoverStart Ptr    `json:"onHoverStart,omitempty"`
	OnHoverEnd   Ptr    `json:"onHoverEnd,omitempty"`
	Invisible    bool   `json:"i,omitempty"`
	Border       Border `json:"b,omitempty"`

	Padding Padding `json:"p,omitempty"`
	Frame   Frame   `json:"f,omitempty"`

	// see also https://www.w3.org/WAI/tutorials/images/decision-tree/ but makes probably no sense.
	AccessibilityLabel string `json:"al,omitempty"`

	Font Font `json:"o,omitempty"`
	component
}
