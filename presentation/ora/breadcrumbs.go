package ora

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Breadcrumbs struct {
	Ptr               Ptr                `json:"id"`
	Type              ComponentType      `json:"type" value:"Breadcrumbs"`
	Items             Property[[]string] `json:"items"`
	SelectedItemIndex Property[int64]    `json:"selectedItemIndex"`
	Icon              Property[SVG]      `json:"icon"`
	component
}
