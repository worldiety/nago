package ora

type Ping struct {
	Type EventType `json:"type" value:"Ping"`
	event
}
