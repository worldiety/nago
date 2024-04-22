package ora

type Scaffold struct {
	Ptr         Ptr                 `json:"id"`
	Type        ComponentType       `json:"type" value:"Scaffold"`
	Title       Property[string]    `json:"title"`
	Body        Property[Component] `json:"body"`
	Breadcrumbs Property[[]Button]  `json:"breadcrumbs"`
	Menu        Property[[]Button]  `json:"menu"`
	TopbarLeft  Property[Component] `json:"topbarLeft"`
	TopbarMid   Property[Component] `json:"topbarMid"`
	TopbarRight Property[Component] `json:"topbarRight"`
	component
}
