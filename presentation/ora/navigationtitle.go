package ora

// WindowTitle is an invisible component which teleports its Value into the current active window navigation title.
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type WindowTitle struct {
	Type  ComponentType `json:"type" value:"W"`
	Value string        `json:"v,omitempty"`
	component
}
