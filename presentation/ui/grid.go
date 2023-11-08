package ui

import (
	"encoding/json"
	"go.wdy.de/nago/container/slice"
)

type GridCell struct {
	RowSpan  int
	ColSpan  int
	Child    View              // first
	Children slice.Slice[View] // others
}

func (GridCell) isView() {}

func (v GridCell) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type    string `json:"type"`
		RowSpan int    `json:"rowSpan"`
		ColSpan int    `json:"colSpan"`
		Views   []View `json:"views"`
	}{
		Type:    "GridCell",
		RowSpan: v.RowSpan,
		ColSpan: v.ColSpan,
		Views:   joinViews(v.Child, v.Children),
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
	Rows    int
	Gap     Length
	Padding Length
	Cells   slice.Slice[GridCell]
}

func (Grid) isView() {}

func (v Grid) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type    string     `json:"type"`
		Columns int        `json:"columns"`
		Rows    int        `json:"rows"`
		Gap     Length     `json:"gap"`
		Padding Length     `json:"padding"`
		Cells   []GridCell `json:"cells"`
	}{
		Type:    "Grid",
		Columns: v.Columns,
		Rows:    v.Rows,
		Gap:     v.Gap,
		Padding: v.Padding,
		Cells:   slice.UnsafeUnwrap(v.Cells),
	})
}

type IntLike interface {
	isInt()
}

type RInt ResponsiveValue[int]

func (RInt) isInt() {}

type Int2 int

func (Int2) isInt() {}

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
