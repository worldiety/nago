package ora

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Text struct {
	Ptr   Ptr           `json:"id"`
	Type  ComponentType `json:"type" value:"Text"`
	Value string        `json:"value,omitempty"`
	// Color denotes the text color. Leave empty, for the context sensitiv default theme color.
	Color Color `json:"color,omitempty"`

	// BackgroundColor denotes the color of the text background.  Leave empty, for the context sensitiv default theme color.
	BackgroundColor Color `json:"backgroundColor,omitempty"`

	OnClick      Ptr  `json:"onClick,omitempty"`
	OnHoverStart Ptr  `json:"onHoverStart,omitempty"`
	OnHoverEnd   Ptr  `json:"onHoverEnd,omitempty"`
	Invisible    bool `json:"invisible,omitempty"`

	Padding Padding `json:"p,omitempty"`
	Frame   Frame   `json:"f,omitempty"`

	// see also https://www.w3.org/WAI/tutorials/images/decision-tree/ but makes probably no sense.
	AccessibilityLabel string `json:"al,omitempty"`

	Font Font `json:"fn,omitempty"`
	component
}
