package ora

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Padding struct {
	Top    Length `json:"t,omitempty"`
	Left   Length `json:"l,omitempty"`
	Right  Length `json:"r,omitempty"`
	Bottom Length `json:"b,omitempty"`
}

func (p *Padding) All(pad Length) {
	p.Left = pad
	p.Right = pad
	p.Bottom = pad
	p.Top = pad
}
