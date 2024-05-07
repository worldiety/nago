package ora

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Toggle struct {
	Ptr              Ptr              `json:"id"`
	Type             ComponentType    `json:"type" value:"Toggle"`
	Label            Property[string] `json:"label"`
	Checked          Property[bool]   `json:"checked"`
	Disabled         Property[bool]   `json:"disabled"`
	OnCheckedChanged Property[Ptr]    `json:"onCheckedChanged"`
	component
}
