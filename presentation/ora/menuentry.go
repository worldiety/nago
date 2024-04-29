package ora

type MenuEntry struct {
	Ptr   Ptr                   `json:"id"`
	Type  ComponentType         `json:"type" value:"MenuEntry"`
	Icon  Property[SVG]         `json:"icon"` // TODO replace with svg id
	Title Property[string]      `json:"title"`
	Menu  Property[[]MenuEntry] `json:"menu"`
	component
}
