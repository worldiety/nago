package ora

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type FileField struct {
	Ptr       Ptr              `json:"id"`
	Type      ComponentType    `json:"type" value:"FileField"`
	Label     Property[string] `json:"label"`
	HintLeft  Property[string] `json:"hintLeft"`
	HintRight Property[string] `json:"hintRight"`
	Error     Property[string] `json:"error"`
	Disabled  Property[bool]   `json:"disabled"`
	Filter    Property[string] `json:"filter"`
	Multiple  Property[bool]   `json:"multiple"`
	MaxBytes  Property[int64]  `json:"maxBytes"`
	Visible   Property[bool]   `json:"visible"`
	component
}
