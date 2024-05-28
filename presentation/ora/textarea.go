package ora

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type TextArea struct {
	Ptr           Ptr              `json:"id"`
	Type          ComponentType    `json:"type" value:"TextArea"`
	Label         Property[string] `json:"label"`
	Hint          Property[string] `json:"hint"`
	Error         Property[string] `json:"error"`
	Value         Property[string] `json:"value"`
	Rows          Property[int64]  `json:"rows"`
	Disabled      Property[bool]   `json:"disabled"`
	OnTextChanged Property[Ptr]    `json:"onTextChanged"`
	Visible       Property[bool]   `json:"visible"`
	component
}
