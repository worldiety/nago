package ora

type Slider struct {
	Ptr              Ptr               `json:"id"`
	Type             ComponentType     `json:"type" value:"Slider"`
	Disabled         Property[bool]    `json:"disabled"`
	Label            Property[string]  `json:"label"`
	Hint             Property[string]  `json:"hint"`
	Error            Property[string]  `json:"error"`
	StartValue       Property[float64] `json:"startValue"`
	EndValue         Property[float64] `json:"endValue"`
	Min              Property[float64] `json:"min"`
	Max              Property[float64] `json:"max"`
	Stepsize         Property[float64] `json:"stepsize"`
	StartInitialized Property[bool]    `json:"startInitialized"`
	EndInitialized   Property[bool]    `json:"endInitialized"`
	OnChanged        Property[Ptr]     `json:"onChanged"`
	component
}
