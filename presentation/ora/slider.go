package ora

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Slider struct {
	Ptr              Ptr               `json:"id"`
	Type             ComponentType     `json:"type" value:"Slider"`
	Disabled         Property[bool]    `json:"disabled"`
	Label            Property[string]  `json:"label"`
	Hint             Property[string]  `json:"hint"`
	Error            Property[string]  `json:"error"`
	RangeMode        Property[bool]    `json:"rangeMode"`
	StartValue       Property[float64] `json:"startValue"`
	EndValue         Property[float64] `json:"endValue"`
	Min              Property[float64] `json:"min"`
	Max              Property[float64] `json:"max"`
	Stepsize         Property[float64] `json:"stepsize"`
	StartInitialized Property[bool]    `json:"startInitialized"`
	EndInitialized   Property[bool]    `json:"endInitialized"`
	ShowLabel        Property[bool]    `json:"showLabel"`
	LabelSuffix      Property[string]  `json:"labelSuffix"`
	ShowTickMarks    Property[bool]    `json:"showTickMarks"`
	OnChanged        Property[Ptr]     `json:"onChanged"`
	Visible          Property[bool]    `json:"visible"`
	component
}
