package ora

import "fmt"

// why is this so stupid? Because it is more or less impossible (because so ineffective) to parse
// adjacent encoded types in typescript
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Length string

var Auto = Relative(0)

var Full = Relative(1)

// ViewportHeight is a magic value which sets the intrinsic size of an Element to be the smallest available viewport
// height. This is useful, if you have to center a component vertically on screen. Note, that scrollbars may
// or truncated views may appear, if contained view is larger than the view ports height.
const ViewportHeight = "100svh"

func Absolute(v DP) Length {
	return Length(fmt.Sprintf("%vdp", v))
}

// Relative sizes must be interpreted according to the parents intrinsic size. E.g. setting 1 to height or width
// will not cause a visible change, as long as the parent has wrap content semantics. Thus, the parent
// must either have its own intrinsic size or its parent must force a specific size for it also.
func Relative(v Weight) Length {
	return Length(fmt.Sprintf("%v%%", v*100))
}

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

func (f Frame) Size(w, h Length) Frame {
	f.Height = h
	f.Width = w
	return f
}
