package ora

type Chip struct {
	Ptr     Ptr              `json:"id"`
	Type    ComponentType    `json:"type" value:"Chip"`
	Caption Property[string] `json:"caption"`
	Action  Property[Ptr]    `json:"action"`
	OnClose Property[Ptr]    `json:"onClose"`
	Color   Property[string] `json:"color"` // TODO this must respect themes like dark,light, colorblindnesses and intent
	component
}
