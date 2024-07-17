package ora

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
