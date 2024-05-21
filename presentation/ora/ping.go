package ora

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Ping struct {
	Type EventType `json:"type" value:"Ping"`
	event
}

func (e Ping) ReqID() RequestId {
	return 0 // this is by definition nothing to answer
}
