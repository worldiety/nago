package ora

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Dropdown struct {
	Ptr             Ptr                      `json:"id"`
	Type            ComponentType            `json:"type" value:"Dropdown"`
	Items           Property[[]DropdownItem] `json:"items"`
	SelectedIndices Property[[]int64]        `json:"selectedIndices"`
	Multiselect     Property[bool]           `json:"multiselect"`
	Expanded        Property[bool]           `json:"expanded"`
	Disabled        Property[bool]           `json:"disabled"`
	Label           Property[string]         `json:"label"`
	Hint            Property[string]         `json:"hint"`
	Error           Property[string]         `json:"error"`
	OnClicked       Property[Ptr]            `json:"onClicked"`
	Searchable      Property[bool]           `json:"searchable"`
	Visible         Property[bool]           `json:"visible"`
	component
}
