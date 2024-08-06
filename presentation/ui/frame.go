package ui

import "go.wdy.de/nago/presentation/ora"

type Frame struct {
	MinWidth  Length
	MaxWidth  Length
	MinHeight Length
	MaxHeight Length
	Width     Length
	Height    Length
}

func (f Frame) ora() ora.Frame {
	return ora.Frame{
		MinWidth:  ora.Length(f.MinWidth),
		MaxWidth:  ora.Length(f.MaxWidth),
		MinHeight: ora.Length(f.MinHeight),
		MaxHeight: ora.Length(f.MaxHeight),
		Width:     ora.Length(f.Width),
		Height:    ora.Length(f.Height),
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
