package ui

import (
	"fmt"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

// Border adds the defined border and dimension to the component. Note, that a border will change the dimension.
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Border struct {
	TopLeftRadius     Length
	TopRightRadius    Length
	BottomLeftRadius  Length
	BottomRightRadius Length

	LeftWidth   Length
	TopWidth    Length
	RightWidth  Length
	BottomWidth Length

	LeftColor   Color
	TopColor    Color
	RightColor  Color
	BottomColor Color

	BoxShadow Shadow `json:"s,omitempty"`
}

func (b Border) ora() ora.Border {
	return ora.Border{
		TopLeftRadius:     ora.Length(b.TopLeftRadius),
		TopRightRadius:    ora.Length(b.TopRightRadius),
		BottomLeftRadius:  ora.Length(b.BottomLeftRadius),
		BottomRightRadius: ora.Length(b.BottomRightRadius),
		LeftWidth:         ora.Length(b.LeftWidth),
		TopWidth:          ora.Length(b.TopWidth),
		RightWidth:        ora.Length(b.RightWidth),
		BottomWidth:       ora.Length(b.BottomWidth),
		LeftColor:         ora.Color(b.LeftColor),
		TopColor:          ora.Color(b.TopColor),
		RightColor:        ora.Color(b.RightColor),
		BottomColor:       ora.Color(b.BottomColor),
		BoxShadow:         b.BoxShadow.ora(),
	}
}

func (b Border) Radius(radius Length) Border {
	b.TopLeftRadius = radius
	b.TopRightRadius = radius
	b.BottomLeftRadius = radius
	b.BottomRightRadius = radius
	return b
}

func (b Border) TopRadius(radius Length) Border {
	b.TopLeftRadius = radius
	b.TopRightRadius = radius
	return b
}

func (b Border) Circle() Border {
	return b.Radius("999999dp")
}

func (b Border) Width(width Length) Border {
	b.LeftWidth = width
	b.TopWidth = width
	b.RightWidth = width
	b.BottomWidth = width
	return b
}

func (b Border) Color(c Color) Border {
	b.LeftColor = c
	b.TopColor = c
	b.BottomColor = c
	b.RightColor = c
	return b
}

func (b Border) Shadow(radius Length) Border {
	b.BoxShadow.Radius = radius
	b.BoxShadow.Color = "#00000054"
	b.BoxShadow.X = ""
	b.BoxShadow.Y = ""
	return b
}

// Elevate by DP
func (b Border) Elevate(elevation core.DP) Border {
	rem := float64(elevation) / 16
	b.BoxShadow.Radius = Length(fmt.Sprintf("%.2frem", rem*3))
	b.BoxShadow.Color = "#00000030"
	b.BoxShadow.X = ""
	b.BoxShadow.Y = Length(fmt.Sprintf("%.2frem", rem))
	return b
}
