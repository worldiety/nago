package ui

import "go.wdy.de/nago/presentation/proto"

type Frame struct {
	MinWidth  Length
	MaxWidth  Length
	MinHeight Length
	MaxHeight Length
	Width     Length
	Height    Length
}

func (f Frame) ora() proto.Frame {
	return proto.Frame{
		MinWidth:  proto.Length(f.MinWidth),
		MaxWidth:  proto.Length(f.MaxWidth),
		MinHeight: proto.Length(f.MinHeight),
		MaxHeight: proto.Length(f.MaxHeight),
		Width:     proto.Length(f.Width),
		Height:    proto.Length(f.Height),
	}
}

func (f Frame) Size(w, h Length) Frame {
	f.Height = h
	f.Width = w
	return f
}

func (f Frame) MatchScreen() Frame {
	f.Height = ViewportHeight
	f.Width = Full
	return f
}

func (f Frame) FullWidth() Frame {
	f.Width = "100%"
	return f
}

func (f Frame) FullHeight() Frame {
	f.Height = "100%"
	return f
}
