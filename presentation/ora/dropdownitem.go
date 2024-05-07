package ora

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type DropdownItem struct {
	Ptr       Ptr              `json:"id"`
	Type      ComponentType    `json:"type" value:"DropdownItem"`
	Content   Property[string] `json:"content"`
	OnClicked Property[Ptr]    `json:"onClicked"`
	component
}
