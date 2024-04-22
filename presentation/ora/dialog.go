package ora

type Dialog struct {
	Ptr     Ptr                 `json:"id"`
	Type    ComponentType       `json:"type" value:"Dialog"`
	Title   Property[string]    `json:"title"`
	Body    Property[Component] `json:"body"`
	Icon    Property[SVG]       `json:"icon"` // TODO replace me with reference
	Actions Property[[]Button]  `json:"actions"`
	component
}
