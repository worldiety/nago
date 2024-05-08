package ora

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type MenuEntry struct {
	Ptr        Ptr                   `json:"id"`
	Type       ComponentType         `json:"type" value:"MenuEntry"`
	Icon       Property[SVG]         `json:"icon"`       // TODO replace with svg id
	IconActive Property[SVG]         `json:"iconActive"` // TODO replace with svg id
	Title      Property[string]      `json:"title"`
	Action     Property[Ptr]         `json:"action"`
	Menu       Property[[]MenuEntry] `json:"menu"`
	Badge      Property[string]      `json:"badge"`
	Expanded   Property[bool]        `json:"expanded"`
	OnFocus    Property[Ptr]         `json:"onFocus"`
	component
}
