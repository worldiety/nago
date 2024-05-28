package ora

// deprecated: use flexcontainer
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type VBox struct {
	Ptr      Ptr                   `json:"id"`
	Type     ComponentType         `json:"type" value:"VBox"`
	Children Property[[]Component] `json:"children"`
	component
}
