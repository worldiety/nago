package ui

import "go.wdy.de/nago/presentation/ora"

type Padding struct {
	Top    Length
	Left   Length
	Right  Length
	Bottom Length
}

func (p Padding) ora() ora.Padding {
	return ora.Padding{
		Top:    ora.Length(p.Top),
		Left:   ora.Length(p.Left),
		Right:  ora.Length(p.Right),
		Bottom: ora.Length(p.Bottom),
	}
}

func (p Padding) All(pad Length) Padding {
	p.Left = pad
	p.Right = pad
	p.Bottom = pad
	p.Top = pad
	return p
}

// Vertical means Y axis, so top and bottom are set to the padding value.
func (p Padding) Vertical(pad Length) Padding {
	p.Bottom = pad
	p.Top = pad
	return p
}

// Horizontal means X axis, so left and right are set to the padding value.
func (p Padding) Horizontal(pad Length) Padding {
	p.Left = pad
	p.Right = pad
	return p
}
