package ora

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type WebView struct {
	Ptr   Ptr              `json:"id"`
	Type  ComponentType    `json:"type" value:"WebView"`
	Value Property[string] `json:"value"`
	component
}
