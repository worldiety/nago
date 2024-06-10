package ora

import "time"

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type TextField struct {
	Ptr           Ptr              `json:"id"`
	Type          ComponentType    `json:"type" value:"TextField"`
	Label         Property[string] `json:"label"`
	Hint          Property[string] `json:"hint"`
	Help          Property[string] `json:"help"`
	Error         Property[string] `json:"error"`
	Value         Property[string] `json:"value"`
	Placeholder   Property[string] `json:"placeholder"` // TODO that does not make any sense from UX, we have Label and Hint: remove me
	Disabled      Property[bool]   `json:"disabled"`
	Simple        Property[bool]   `json:"simple"` // TODO what is that? Better use a documented enum?
	OnTextChanged Property[Ptr]    `json:"onTextChanged"`
	// OnDebouncedTextChanged is called, after no changes within the DebounceTime have been seen.
	// Note that the frontend is allowed to suppress any property updates, until the debouncer kicks in.
	// This is by intention, to ensure that the backend does not re-render anyway due to always dirty views.
	OnDebouncedTextChanged Property[Ptr]           `json:"onDebouncedTextChanged"`
	DebounceTime           Property[time.Duration] `json:"debounceTime"`
	Visible                Property[bool]          `json:"visible"`
	component
}
