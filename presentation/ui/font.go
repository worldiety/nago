package ui

import "go.wdy.de/nago/presentation/ora"

type Font struct {
	// Name of the font or family name as fallback. Extra fallback declarations are unspecified and must be comma
	// separated.
	Name string

	// Size of the font
	Size Length

	Style FontStyle

	Weight FontWeight
}

func (f Font) ora() ora.Font {
	return ora.Font{
		Name:   f.Name,
		Size:   ora.Length(f.Size),
		Style:  ora.FontStyle(f.Style),
		Weight: ora.FontWeight(f.Weight),
	}
}

var (
	Title = Font{
		Size:   "1.5rem",
		Weight: BoldFontWeight,
	}
)

type FontStyle string

const (
	ItalicFontStyle FontStyle = "i"
	NormalFontStyle           = "n"
)

type FontWeight int

const (
	NormalFontWeight FontWeight = 400
	BoldFontWeight   FontWeight = 700
)
