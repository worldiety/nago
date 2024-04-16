package ora

type Stepper struct {
	Ptr           Ptr                  `json:"id"`
	Type          ComponentType        `json:"type" value:"Stepper"`
	Steps         Property[[]StepInfo] `json:"steps"`
	SelectedIndex Property[int64]      `json:"selectedIndex"`
	component
}
