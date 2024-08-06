package ui

import "go.wdy.de/nago/presentation/ora"

type Shadow struct {
	Color  Color
	Radius Length
	X      Length
	Y      Length
}

func (s Shadow) ora() ora.Shadow {
	return ora.Shadow{
		Color:  ora.Color(s.Color),
		Radius: ora.Length(s.Radius),
		X:      ora.Length(s.X),
		Y:      ora.Length(s.Y),
	}
}
