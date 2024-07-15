package ora

import "time"

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type TextField struct {
	Type          ComponentType    `json:"type" value:"TextField"`
	Label         string           `json:"label"`
	Hint          string           `json:"hint"`
	Help          string           `json:"help"`
	Error         string           `json:"error"`
	Text          Property[string] `json:"value"`
	Placeholder   string           `json:"placeholder"` // TODO that does not make any sense from UX, we have Label and Hint: remove me
	Disabled      bool             `json:"disabled"`
	Simple        bool             `json:"simple"` // TODO what is that? Better use a documented enum?
	OnTextChanged Ptr              `json:"onTextChanged"`
	// OnDebouncedTextChanged is called, after no changes within the DebounceTime have been seen.
	// Note that the frontend is allowed to suppress any property updates, until the debouncer kicks in.
	// This is by intention, to ensure that the backend does not re-render anyway due to always dirty views.
	OnDebouncedTextChanged Ptr           `json:"onDebouncedTextChanged"`
	DebounceTime           time.Duration `json:"debounceTime"`
	Invisible              bool          `json:"visible"`
	Frame                  Frame         `json:"frame,omitempty"`
	component
}
