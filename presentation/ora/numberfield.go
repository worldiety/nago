package ora

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type NumberField struct {
	Ptr            Ptr              `json:"id"`
	Type           ComponentType    `json:"type" value:"NumberField"`
	Label          Property[string] `json:"label"`
	Hint           Property[string] `json:"hint"`
	Error          Property[string] `json:"error"`
	Value          Property[string] `json:"value"`
	Placeholder    Property[string] `json:"placeholder"` // TODO that does not make any sense from UX, we have Label and Hint: remove me
	Disabled       Property[bool]   `json:"disabled"`
	Simple         Property[bool]   `json:"simple"` // TODO what is that? Better use a documented enum?
	OnValueChanged Property[Ptr]    `json:"onValueChanged"`
	component
}
