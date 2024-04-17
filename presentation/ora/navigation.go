package ora

// NavigationForwardToRequested is an Event triggered by the backend which requests a forward navigation action within the frontend.
// A frontend must put the new component to create by the factory on top of the current component within the scope.
// The frontend is free keep multiple components alive at the same time, however it must ensure that the UX is sane.
type NavigationForwardToRequested struct {
	Type    EventType          `json:"type" value:"NavigationForwardToRequested"`
	Factory ComponentFactoryId `json:"factory"` // Factory tells which component factory shall be called to invalidate the frontend.
	Values  map[string]string  `json:"values"`
	event
}

// NavigationResetRequested removes the entire history in the scope and pushes the target on top.
type NavigationResetRequested struct {
	Type    EventType          `json:"type" value:"NavigationResetRequested"`
	Factory ComponentFactoryId `json:"factory"` // Factory tells which component factory shall be called to invalidate the frontend.
	Values  map[string]string  `json:"values"`
	event
}

// NavigationBackRequested steps back causing a likely destruction of the most top component.
// The frontend may decide to ignore that, if the stack would be empty/undefined otherwise.
type NavigationBackRequested struct {
	Type EventType `json:"type" value:"NavigationBackRequested"`
	event
}
