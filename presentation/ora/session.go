package ora

// SessionAssigned must not be used by browser clients directly.
// A http channel implementation must issue this by itself due to security concerns like http-only cookies.
// Native client (mobile or desktop) should use this event instead.
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type SessionAssigned struct {
	Type      EventType `json:"type" value:"SessionAssigned"`
	SessionID string    `json:"sessionID"`
	event
}
