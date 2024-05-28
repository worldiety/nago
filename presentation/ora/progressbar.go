package ora

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type ProgressBar struct {
	Ptr            Ptr               `json:"id"`
	Type           ComponentType     `json:"type" value:"ProgressBar"`
	Max            Property[float64] `json:"max"`
	Value          Property[float64] `json:"value"`
	ShowPercentage Property[bool]    `json:"showPercentage"`
	component
}
