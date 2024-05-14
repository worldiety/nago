package ora

import (
	"crypto/rand"
	"encoding/base64"
)

// A ScopeID has at least a 32 byte entropy and must be generated using a secure random source.
// It must be treated as a secret at the frontend (e.g. no exposing into URLs), because
// it allows the hijacking of connections and allocated components.
// These components may likely contain already authorized credentials, thus leaking the ScopeID
// also means leaking the access rights.
//
// If you know, that you are done, destroy the scope to release all associated backend resources.
// Keep the lifetime of the scope small to trade resume comfort and security and resource usage.
//
// Note that allocations of components inside a Scope are unrelated and must be managed explicitly.
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type ScopeID string

// NewScopeID returns a secure URL safe base64 encoded random 32 byte entropy.
func NewScopeID() ScopeID {
	var tmp [32]byte
	if _, err := rand.Read(tmp[:]); err != nil {
		panic(err)
	}

	return ScopeID(base64.URLEncoding.EncodeToString(tmp[:]))
}

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type ScopeDestructionRequested struct {
	Type      EventType `json:"type" value:"ScopeDestructionRequested"`
	RequestId RequestId `json:"r" description:"Request ID."`
	event
}
