package ora

type Scaffold struct {
	Ptr                 Ptr                           `json:"id"`
	Type                ComponentType                 `json:"type" value:"Scaffold"`
	Body                Property[Component]           `json:"body"`
	NavigationComponent Property[NavigationComponent] `json:"navigationComponent"`
	component
}
