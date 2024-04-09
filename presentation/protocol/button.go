package protocol

type Button struct {
	Ptr      Ptr              `json:"id"`
	Type     ComponentType    `json:"type" value:"Button"`
	Caption  Property[string] `json:"caption" description:"Caption of the button"`
	PreIcon  Property[RIDSVG] `json:"preIcon"`
	PostIcon Property[RIDSVG] `json:"postIcon"`
	Color    Property[Intent] `json:"color"`
	Disabled Property[bool]   `json:"disabled"`
	Action   Property[Ptr]    `json:"action"`
	component
	_ struct{} `description:"A Button is the only button"`
}
