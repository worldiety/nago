package ora

// Checkbox represents an user interface element which spans a visible area to click or tap from the user.
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Checkbox struct {
	Ptr      Ptr            `json:"id"`
	Type     ComponentType  `json:"type" value:"Checkbox"`
	Selected Property[bool] `json:"selected"`
	Clicked  Property[Ptr]  `json:"clicked"`
	Disabled Property[bool] `json:"disabled"`
	component
}
