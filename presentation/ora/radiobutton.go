package ora

// Radiobutton represents a user interface element which spans a visible area to click or tap from the user.
// Usually a radiobutton belongs to a group, where only a single element can be picked. Thus, it is quite similar
// to a Spinner/Select/Combobox.
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Radiobutton struct {
	Type ComponentType `json:"type" value:"R"`
	// Value is the initial checked value.
	Value bool `json:"v,omitempty"`
	// InputValue is where updated value of the checked states are written.
	InputValue Ptr  `json:"i,omitempty"`
	Disabled   bool `json:"d,omitempty"`
	Invisible  bool `json:"iv,omitempty"`
	component
}
