package ora

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Text struct {
	Ptr          Ptr              `json:"id"`
	Type         ComponentType    `json:"type" value:"Text"`
	Value        Property[string] `json:"value"`
	Color        Property[string] `json:"color"` // TODO how to mix color and intent? A customer may want a special color
	Size         Property[string] `json:"size"`  // TODO what is this size, which unit?
	OnClick      Property[Ptr]    `json:"onClick"`
	OnHoverStart Property[Ptr]    `json:"onHoverStart"`
	OnHoverEnd   Property[Ptr]    `json:"onHoverEnd"`
	Visible      Property[bool]   `json:"visible"`

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
