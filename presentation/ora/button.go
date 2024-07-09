package ora

// Button represents an user interface element which spans a visible area to click or tap from the user.
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Button struct {
	Ptr      Ptr                  `json:"id"`
	Type     ComponentType        `json:"type" value:"Button"`
	Caption  Property[string]     `json:"caption" description:"Caption of the button"`
	PreIcon  Property[SVG]        `json:"preIcon"`  // TODO replace with svg id
	PostIcon Property[SVG]        `json:"postIcon"` // TODO replace with svg id
	Color    Property[NamedColor] `json:"color"`
	Disabled Property[bool]       `json:"disabled"`
	Action   Property[Ptr]        `json:"action"`
	Visible  Property[bool]       `json:"visible"`
	// Frame is omitted, if empty
	Frame Frame `json:"frame,omitempty"`
	component
}
