package ora

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type BreadcrumbItem struct {
	Ptr    Ptr              `json:"id"`
	Type   ComponentType    `json:"type" value:"BreadcrumbItem"`
	Label  Property[string] `json:"label"`
	Action Property[Ptr]    `json:"action"`
	component
}
