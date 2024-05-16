package ora

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type EventType string

const (
	EventsAggregatedT              EventType = "T"
	AcknowledgedT                  EventType = "A"
	SetPropertyValueRequestedT     EventType = "P"
	FunctionCallRequestedT         EventType = "F"
	PingT                          EventType = "Ping"
	NewComponentRequestedT         EventType = "NewComponentRequested"
	ComponentInvalidatedT          EventType = "ComponentInvalidated"
	ErrorOccurredT                 EventType = "ErrorOccurred"
	ComponentDestructionRequestedT EventType = "ComponentDestructionRequested"
	ScopeDestructionRequestedT     EventType = "ScopeDestructionRequested"
)

const (
	NewConfigurationRequestedT    EventType = "NewConfigurationRequested"
	ConfigurationDefinedT         EventType = "ConfigurationDefined"
	NavigationForwardToRequestedT EventType = "NavigationForwardToRequested"
	NavigationBackRequestedT      EventType = "NavigationBackRequested"
	NavigationResetRequestedT     EventType = "NavigationResetRequested"
	SessionAssignedT              EventType = "SessionAssigned"
	SendMultipleRequestedT        EventType = "SendMultipleRequested"
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

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type SetPropertyValueRequested struct {
	Type      EventType `json:"type" value:"P" description:"P stands for Set**P**ropertValue. It is expected, that we must process countless of these events."`
	Ptr       Ptr       `json:"p" description:"p denotes the remote pointer."`
	Value     string    `json:"v" description:"v denotes the serialized value to set the property to."`
	RequestId RequestId `json:"r"`
	event
}

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type FunctionCallRequested struct {
	Type      EventType `json:"type" value:"F" description:"F stands for **F**unctionCallRequested. It is expected, that we must process countless of these events."`
	Ptr       Ptr       `json:"p" description:"p denotes the remote pointer."`
	RequestId RequestId `json:"r"`
	event
}

type event struct {
}

func (e event) isEvent() {}
