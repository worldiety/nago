package ui

import "go.wdy.de/nago/presentation/proto"

type Shadow struct {
	Color  Color
	Radius Length
	X      Length
	Y      Length
}

func (s Shadow) ora() proto.Shadow {
	return proto.Shadow{
		Color:  proto.Color(s.Color),
		Radius: proto.Length(s.Radius),
		X:      proto.Length(s.X),
		Y:      proto.Length(s.Y),
	}
}
