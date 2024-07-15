package ora

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Text struct {
	Ptr   Ptr           `json:"id"`
	Type  ComponentType `json:"type" value:"Text"`
	Value string        `json:"value,omitempty"`
	// Color denotes the text color. Leave empty, for the context sensitiv default theme color.
	Color NamedColor `json:"color,omitempty"`

	// BackgroundColor denotes the color of the text background.  Leave empty, for the context sensitiv default theme color.
	BackgroundColor NamedColor `json:"backgroundColor,omitempty"`

	Size         string `json:"size,omitempty"` // TODO what is this size, which unit?
	OnClick      Ptr    `json:"onClick,omitempty"`
	OnHoverStart Ptr    `json:"onHoverStart,omitempty"`
	OnHoverEnd   Ptr    `json:"onHoverEnd,omitempty"`
	Invisible    bool   `json:"invisible,omitempty"`

	Padding Padding `json:"p,omitempty"`
	Frame   Frame   `json:"f,omitempty"`

	component
}

// Str is a much more simple text type, which will never have special formatting options.
// We introduced this, because our protocol encoding is so bloated by definition, that we must allocate
// all properties and even send them over wire. In larger tables this creates mind-blowing render trees with
// dozens of MiB in transfer size. Also, neither is the websocket compression working nor is it effective in practice.
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Str struct {
	Type  ComponentType `json:"type" value:"S"`
	Value string        `json:"v,omitempty"`

	component
}
