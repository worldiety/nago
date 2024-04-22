package ora

type DropdownItem struct {
	Ptr       Ptr              `json:"id"`
	Type      ComponentType    `json:"type" value:"DropdownItem"`
	Content   Property[string] `json:"content"`
	OnClicked Property[Ptr]    `json:"onClicked"`
	component
}
