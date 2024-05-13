package ora

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Page struct {
	Ptr    Ptr                   `json:"id"`
	Type   ComponentType         `json:"type" value:"Page"`
	Body   Property[Component]   `json:"body"`
	Modals Property[[]Component] `json:"modals"`
	component
}
