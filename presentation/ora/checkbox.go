package ora

// Checkbox represents a user interface element which spans a visible area to click or tap from the user.
// Use it for controls, which do not cause an immediate effect. See also [Toggle].
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Checkbox struct {
	Type ComponentType `json:"type" value:"c"`
	// Value is the initial checked value.
	Value bool `json:"v,omitempty"`
	// InputValue is where updated value of the checked states are written.
	InputValue Ptr  `json:"i,omitempty"`
	Disabled   bool `json:"d,omitempty"`
	Invisible  bool `json:"iv,omitempty"`
	component
}
