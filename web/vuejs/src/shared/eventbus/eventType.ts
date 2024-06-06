// TODO what is this? Should we autogenerate that based on the sum type discriminators?
export enum EventType {
	INVALIDATED = 'ComponentInvalidated',
	INVALIDATION_REQUESTED = 'ComponentInvalidationRequested',
	NAVIGATE_BACK_REQUESTED = 'NavigationBackRequested',
	NAVIGATE_RELOAD_REQUESTED = 'NavigationReloadRequested',
	SEND_MULTIPLE_REQUESTED = 'SendMultipleRequested',
	NAVIGATE_FORWARD_REQUESTED = 'NavigationForwardToRequested',
	NAVIGATION_RESET_REQUESTED = 'NavigationResetRequested',
	ACKNOWLEDGED = 'A',
	TRANSACTION = 'T',
	NEW_COMPONENT_REQUESTED = 'NewComponentRequested',
	ERROR_OCCURRED = 'ErrorOccurred',
	DESTROY_COMPONENT_REQUESTED = 'ComponentDestructionRequested',
	SCOPE_DESTRUCTION_REQUESTED = 'ScopeDestructionRequested',
	CONFIGURATION_REQUESTED = 'ConfigurationRequested',
	CONFIGURATION_DEFINED = 'ConfigurationDefined',
	SET_PROPERTY_REQUESTED = 'P',
	FUNCTION_CALL_REQUESTED = 'F',
	SESSION_ASSIGNED = 'SessionAssigned',
	PING = 'Ping',
	WindowInfoChanged = 'WindowInfoChanged'
}
