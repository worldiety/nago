package ora

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Dialog struct {
	Ptr     Ptr                   `json:"id"`
	Type    ComponentType         `json:"type" value:"Dialog"`
	Title   Property[string]      `json:"title"`
	Body    Property[Component]   `json:"body"`
	Footer  Property[Component]   `json:"footer"`
	Icon    Property[SVG]         `json:"icon"` // TODO replace me with reference
	Visible Property[bool]        `json:"visible"`
	Size    Property[ElementSize] `json:"size"`
	component
}
