package ora

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type StepInfo struct {
	Ptr     Ptr              `json:"id"`
	Type    ComponentType    `json:"type" value:"StepInfo"`
	Number  Property[string] `json:"number"`
	Caption Property[string] `json:"caption"`
	Details Property[string] `json:"details"`
	component
}
