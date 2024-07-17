package ora

// The following Length sizes are common for the ORA design system and will automatically adjust to the root elements font size.
// It is similar to the effect of Androids SP unit, however its factor is by default at 16, because we just use the CSS semantics.
const (
	// L2 relates to about 2dp at default font scale.
	L2 Length = "0.125rem"
	// L4 relates to about 4dp at default font scale.
	L4 Length = "0.25rem"
	// L8 relates to about 8dp at default font scale.
	L8 Length = "0.5rem"
	//L12 relates to about 12dp at default font scale.
	L12 Length = "0.75rem"
	// L14 corresponds to 14dp at default font scale.
	L14 Length = "0.875rem"
	// L20 relates to about 20dp at default font scale.
	L20 Length = "1.25rem"
	//L40 relates to about 40dp at default font scale.
	L40 Length = "2.5rem"
	//L44 relates to about 44dp at default font scale.
	L44 Length = "2.75rem"
	//L160 relates to about 160dp at default font scale.
	L160 Length = "10rem"
	//L320 relates to about 320dp at default font scale.
	L320 Length = "20rem"
)

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Padding struct {
	Top    Length `json:"t,omitempty"`
	Left   Length `json:"l,omitempty"`
	Right  Length `json:"r,omitempty"`
	Bottom Length `json:"b,omitempty"`
}

func (p Padding) All(pad Length) Padding {
	p.Left = pad
	p.Right = pad
	p.Bottom = pad
	p.Top = pad
	return p
}

func (p Padding) Vertical(pad Length) Padding {
	p.Bottom = pad
	p.Top = pad
	return p
}

func (p Padding) Horizontal(pad Length) Padding {
	p.Left = pad
	p.Right = pad
	return p
}
