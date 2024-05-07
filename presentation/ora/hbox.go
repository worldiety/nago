package ora

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type HBox struct {
	Ptr       Ptr                   `json:"id"`
	Type      ComponentType         `json:"type" value:"HBox"`
	Children  Property[[]Component] `json:"children"`
	Alignment Property[string]      `json:"alignment"` // TODO we need to define the UX and what is allowed, we have an unwanted html semantic here. must look fine in iOS/Android App
	component
}
