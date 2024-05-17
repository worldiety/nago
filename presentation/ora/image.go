package ora

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Image struct {
	Ptr     Ptr              `json:"id"`
	Type    ComponentType    `json:"type" value:"Image"`
	URI     Property[URI]    `json:"uri"`
	Caption Property[string] `json:"caption"`
	component
}
