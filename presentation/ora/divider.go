package ora

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Divider struct {
	Type    ComponentType `json:"type" value:"d"`
	Frame   Frame         `json:"f,omitempty"`
	Border  Border        `json:"b,omitempty"`
	Padding Padding       `json:"p,omitempty"`
	component
}
