package ora

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type DatePicker struct {
	Ptr                Ptr              `json:"id"`
	Type               ComponentType    `json:"type" value:"DatePicker"`
	Disabled           Property[bool]   `json:"disabled"`
	Label              Property[string] `json:"label"`
	Hint               Property[string] `json:"hint"`
	Error              Property[string] `json:"error"`
	Expanded           Property[bool]   `json:"expanded"`
	RangeMode          Property[bool]   `json:"rangeMode"`
	StartDateSelected  Property[bool]   `json:"startDateSelected"`
	SelectedStartDay   Property[int64]  `json:"selectedStartDay"`
	SelectedStartMonth Property[int64]  `json:"selectedStartMonth"`
	SelectedStartYear  Property[int64]  `json:"selectedStartYear"`
	EndDateSelected    Property[bool]   `json:"endDateSelected"`
	SelectedEndDay     Property[int64]  `json:"selectedEndDay"`
	SelectedEndMonth   Property[int64]  `json:"selectedEndMonth"`
	SelectedEndYear    Property[int64]  `json:"selectedEndYear"`
	OnClicked          Property[Ptr]    `json:"onClicked"`
	OnSelectionChanged Property[Ptr]    `json:"onSelectionChanged"`
	Visible            Property[bool]   `json:"visible"`
	component
}
