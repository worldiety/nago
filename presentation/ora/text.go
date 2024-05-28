package ora

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Text struct {
	Ptr          Ptr              `json:"id"`
	Type         ComponentType    `json:"type" value:"Text"`
	Value        Property[string] `json:"value"`
	Color        Property[string] `json:"color"`     // TODO how to mix color and intent? A customer may want a special color
	ColorDark    Property[string] `json:"colorDark"` // TODO how to mix color and intent? What about the other profiles?
	Size         Property[string] `json:"size"`      // TODO what is this size, which unit?
	OnClick      Property[Ptr]    `json:"onClick"`
	OnHoverStart Property[Ptr]    `json:"onHoverStart"`
	OnHoverEnd   Property[Ptr]    `json:"onHoverEnd"`
	Visible      Property[bool]   `json:"visible"`

	component
}
