package ora

// why is this so stupid? Because it is more or less impossible (because so ineffective) to parse
// adjacent encoded types in typescript
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Length string

// Weight is between 0-1 and can be understood as 1 = 100%, however implementations must normalize the total
// of all weights and recalculate the effective percentage.
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Weight float64

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Frame struct {
	// MinWidth is omitted if empty
	MinWidth Length `json:"wi,omitempty"`
	// MaxWidth is omitted if empty
	MaxWidth Length `json:"wx,omitempty"`
	// MinHeight is omitted if empty
	MinHeight Length `json:"hi,omitempty"`
	// MaxHeight is omitted if empty
	MaxHeight Length `json:"hx,omitempty"`
	// Width is omitted if empty
	Width Length `json:"w,omitempty"`
	// Height is omitted if empty
	Height Length `json:"h,omitempty"`
}
