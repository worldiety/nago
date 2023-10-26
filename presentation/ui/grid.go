package ui

import (
	"encoding/json"
	"go.wdy.de/nago/container/slice"
)

type GridCell struct {
	Start    int               // start column at
	End      int               // end column at
	Child    View              // first
	Children slice.Slice[View] // others
}

func (GridCell) isView() {}

func (v GridCell) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type  string `json:"type"`
		Start int    `json:"start"`
		End   int    `json:"end"`
		Views []View `json:"views"`
	}{
		Type:  "GridCell",
		Start: v.Start,
		End:   v.End,
		Views: joinViews(v.Child, v.Children),
	})
}

type Responsive struct {
	Default int
	Medium  int
	Large   int
}

// Grid shall be interpreted like the rules of the tailwind grid, see also https://tailwindcss.com/docs/grid-column.
type Grid struct {
	Columns int
	Gap     int
	Cells   slice.Slice[GridCell]
}

func (Grid) isView() {}

func (v Grid) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type    string     `json:"type"`
		Columns int        `json:"columns"`
		Gap     int        `json:"gap"`
		Cells   []GridCell `json:"cells"`
	}{
		Type:    "Grid",
		Columns: v.Columns,
		Gap:     v.Gap,
		Cells:   slice.UnsafeUnwrap(v.Cells),
	})
}

type IntLike interface {
	isInt()
}

type RInt ResponsiveValue[int]

func (RInt) isInt() {}

type Int int

func (Int) isInt() {}

// A ResponsiveValue has a default for all devices (or screen sizes) but has multiple
type ResponsiveValue[T any] struct {
	// Default is valid for any screen size. This applies always to mobile devices first.
	Default T
	// MD is for medium devices like 768 dip
	MD T
	// LG is for large devices like 1024 dip
	LG T
	// XL is for extra large device like 1280 dip
	XL T
	// XXL is for double extra large like 1536 dip.
	XXL T
}
