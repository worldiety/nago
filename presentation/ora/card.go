package ora

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Card struct {
	Ptr      Ptr                   `json:"id"`
	Type     ComponentType         `json:"type" value:"Card"`
	Children Property[[]Component] `json:"children"`
	Action   Property[Ptr]         `json:"action"`
	Visible  Property[bool]        `json:"visible"`
	component
}
