package ora

// Toggle is just a kind of checkbox without a label. However, a toggle shall be used for immediate activation
// functions. In contrast to that, use a checkbox for form things without an immediate effect.
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Toggle struct {
	Type ComponentType `json:"type" value:"t"`
	// Value is the initial checked value.
	Value bool `json:"v,omitempty"`
	// InputValue is where updated value of the checked states are written.
	InputValue Ptr  `json:"i,omitempty"`
	Disabled   bool `json:"d,omitempty"`
	Invisible  bool `json:"iv,omitempty"`
	component
}
