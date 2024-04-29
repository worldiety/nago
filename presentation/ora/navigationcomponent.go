package ora

type NavigationComponent struct {
	Ptr       Ptr                   `json:"id"`
	Type      ComponentType         `json:"type" value:"NavigationComponent"`
	Logo      Property[SVG]         `json:"logo"` // TODO replace with svg id
	Menu      Property[[]MenuEntry] `json:"menu"`
	Alignment Property[Alignment]   `json:"alignment"`
	component
}
