package ui

import "reflect"

type triggerType string

const (
	trOnClick triggerType = "onClick"
)

type eventSource struct {
	// TriggerType defines the kind of trigger mechanism of this event, e.g. an onClick or a timeout.
	TriggerType triggerType `json:"trigger"`
	// EventType determines the go event type, which is serialized and sent to the frontend and back for model
	// transformation. It has been registered at the page level.
	EventType string `json:"eventType"`
	// Data is the actual payload of the EventType.
	Data any `json:"data"`
}

// fqn return the full qualified name of v which must not be nil.
func fqn(v any) string {
	t := reflect.TypeOf(v)
	return t.PkgPath() + "." + t.Name()
}

func makeEvent(triggerType triggerType, v any) *eventSource {
	if v == nil {
		return nil
	}

	return &eventSource{
		TriggerType: triggerType,
		EventType:   fqn(v),
		Data:        v,
	}
}
