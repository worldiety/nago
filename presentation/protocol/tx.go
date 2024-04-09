package protocol

type RequestId int64

type EventsAggregated struct {
	Type      EventType `json:"type" value:"T" description:"The magic type constant for a Transaction."`
	Events    []Event   `json:"events" description:"The aggregated events to apply in-order at once."`
	RequestId RequestId `json:"r"`
	event
}

type Acknowledged struct {
	Type      EventType `json:"type" value:"A" description:"The magic type constant."`
	RequestId RequestId `json:"r" description:"The request identifier, which is sent back."`
	event
}
