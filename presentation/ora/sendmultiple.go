package ora

// SendMultipleRequested is an event for the frontend from the backend
// to send the according resources into the system environment.
// A Webbrowser may issue a regular download. A backend should not issue multiple downloads at once but instead
// pack multiple files into a zip file because the browser support for something like a multipart download
// is just broken today. An Android App may trigger the according Intent and opens a picker
// to select the receiving app.
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type SendMultipleRequested struct {
	Type      EventType  `json:"type" value:"SendMultipleRequested"`
	Resources []Resource `json:"resources"`

	event
}

// A Resource represents a blob with a name and a resource accessor URI.
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Resource struct {
	// Name must not be a path, just the human readable (and not unique) file name.
	Name string `json:"name"`

	// URI is likely an unreadable link to resolve the actual data. It may incorporate additional security tokens
	// and may have a limited lifetime and its scheme is undefined.
	URI URI `json:"uri"`

	// MimeType is optional and is a hint about the anticipated content.
	MimeType string `json:"mimeType"`
}

// URI is a Uniform Resource Identifier.
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type URI string
