package ora

type NavigationReloadRequested struct {
	Type EventType `json:"type" value:"NavigationReloadRequested"`
	event
}

func (e NavigationReloadRequested) ReqID() RequestId {
	return 0 // TODO the role request-response role is inversed here, should the frontend respond with ack?
}

// NavigationForwardToRequested is an Event triggered by the backend which requests a forward navigation action within the frontend.
// A frontend must put the new component to create by the factory on top of the current component within the scope.
// The frontend is free keep multiple components alive at the same time, however it must ensure that the UX is sane.
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type NavigationForwardToRequested struct {
	Type    EventType          `json:"type" value:"NavigationForwardToRequested"`
	Factory ComponentFactoryId `json:"factory"` // Factory tells which component factory shall be called to invalidate the frontend.
	Values  map[string]string  `json:"values"`
	event
}

func (e NavigationForwardToRequested) ReqID() RequestId {
	return 0 // TODO the role request-response role is inversed here, should the frontend respond with ack?
}

// NavigationResetRequested removes the entire history in the scope and pushes the target on top.
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type NavigationResetRequested struct {
	Type    EventType          `json:"type" value:"NavigationResetRequested"`
	Factory ComponentFactoryId `json:"factory"` // Factory tells which component factory shall be called to invalidate the frontend.
	Values  map[string]string  `json:"values"`
	event
}

func (e NavigationResetRequested) ReqID() RequestId {
	return 0 // TODO the role request-response role is inversed here, should the frontend respond with ack?
}

// NavigationBackRequested steps back causing a likely destruction of the most top component.
// The frontend may deora.Ptre to ignore that, if the stack would be empty/undefined otherwise.
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type NavigationBackRequested struct {
	Type EventType `json:"type" value:"NavigationBackRequested"`
	event
}

func (e NavigationBackRequested) ReqID() RequestId {
	return 0 // TODO the role request-response role is inversed here, should the frontend respond with ack?
}

type ThemeRequested struct {
	Type  EventType `json:"type" value:"ThemeRequested"`
	Theme string    `json:"theme"`
	event
}

func (e ThemeRequested) ReqID() RequestId {
	return 0 // TODO the role request-response role is inversed here, should the frontend respond with ack?
}
