package ora

// A Modal can be declared at any place in the composed view tree. However, these dialogs are teleported into
// the modal space in tree declaration order. A Modal is layouted above all other regular content and will
// catch focus and disable controls of the views behind. Its bounds are at most the maximum possible screen size.
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Modal struct {
	Type    ComponentType `json:"type" value:"M"`
	Content Component     `json:"b,omitempty"`
	// OnDismissRequest is called, if the user wants to dismiss the dialog, e.g. by clicking outside or pressing escape.
	// You can then decide to disable you dialog, or not.
	OnDismissRequest Ptr `json:"odr,omitempty"`
	component
}
