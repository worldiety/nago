package ora

type StepInfo struct {
	Ptr     Ptr              `json:"id"`
	Type    ComponentType    `json:"type" value:"StepInfo"`
	Number  Property[string] `json:"number"`
	Caption Property[string] `json:"caption"`
	Details Property[string] `json:"details"`
	component
}
