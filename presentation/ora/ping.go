package ora

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Ping struct {
	Type EventType `json:"type" value:"Ping"`
	event
}
