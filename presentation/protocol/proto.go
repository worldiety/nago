package protocol

type EventType string

const (
	EventsAggregatedT          EventType = "T"
	AcknowledgedT              EventType = "A"
	SetPropertyValueRequestedT EventType = "P"
	FunctionCallRequestedT     EventType = "F"
	NewComponentRequestedT     EventType = "NewComponentRequested"
	ComponentInvalidatedT      EventType = "ComponentInvalidated"
)

const (
	NewConfigurationRequestedT EventType = "NewConfigurationRequested"
	ConfigurationDefinedT      EventType = "ConfigurationDefined"
)

// Event is a sum type of
//
//	EventsAggregated |
//	Acknowledged |
//	SetPropertyValueRequested |
//	CallFunctionRequested |
//	NewComponentRequested |
//	ComponentInvalidated
type Event interface {
	isEvent()
}

type SetPropertyValueRequested struct {
	Type  EventType `json:"type" value:"P" description:"P stands for Set**P**ropertValue. It is expected, that we must process countless of these events."`
	Ptr   Ptr       `json:"p" description:"p denotes the remote pointer."`
	Value any       `json:"v" description:"v denotes the value to set the property to."`
	event
}

type FunctionCallRequested struct {
	Type EventType `json:"type" value:"F" description:"F stands for **F**unctionCallRequested. It is expected, that we must process countless of these events."`
	Ptr  Ptr       `json:"p" description:"p denotes the remote pointer."`
	event
}

type event struct {
}

func (e event) isEvent() {}
