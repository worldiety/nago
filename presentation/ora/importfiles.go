package ora

// FileImportRequested asks the frontend to let the user pick some files.
// Depending on the actual backend configuration, this may cause
// a regular http multipart upload or some FFI calls providing data streams
// or accessor URIs.
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type FileImportRequested struct {
	Type             EventType `json:"type" value:"FileImportRequested"`
	ID               string    `json:"id"`
	ScopeID          string    `json:"scopeID"`
	Multiple         bool      `json:"multiple"`
	MaxBytes         int64     `json:"maxBytes"`
	AllowedMimeTypes []string  `json:"allowedMimeTypes"`
	event
}

func (f FileImportRequested) ReqID() RequestId {
	return 0
}
