package ora

type Button struct {
	Ptr      Ptr              `json:"id"`
	Type     ComponentType    `json:"type" value:"Button"`
	Caption  Property[string] `json:"caption" description:"Caption of the button"`
	PreIcon  Property[SVG]    `json:"preIcon"`  // TODO replace with svg id
	PostIcon Property[SVG]    `json:"postIcon"` // TODO replace with svg id
	Color    Property[Intent] `json:"color"`
	Disabled Property[bool]   `json:"disabled"`
	Action   Property[Ptr]    `json:"action"`
	component
	_ struct{} `description:"A Button is the only button"`
}
