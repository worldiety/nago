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
