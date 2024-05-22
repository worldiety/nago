package ora

// Radiobutton represents an user interface element which spans a visible area to click or tap from the user.
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Radiobutton struct {
	Ptr       Ptr            `json:"id"`
	Type      ComponentType  `json:"type" value:"Radiobutton"`
	Selected  Property[bool] `json:"selected"`
	OnClicked Property[Ptr]  `json:"onClicked"`
	Disabled  Property[bool] `json:"disabled"`
	component
}
