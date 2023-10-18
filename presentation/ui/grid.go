package ui

import (
	"encoding/json"
	"go.wdy.de/nago/container/slice"
)

type GridCell struct {
	Start    int               // start column at
	Span     int               // span number of columns
	End      int               // end column at
	Child    View              // first
	Children slice.Slice[View] // others
}

func (GridCell) isView() {}

func (v GridCell) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type  string `json:"type"`
		Start int    `json:"start"`
		Span  int    `json:"span"`
		End   int    `json:"end"`
		Views []View `json:"views"`
	}{
		Type:  "GridCell",
		Start: v.Start,
		Span:  v.Span,
		End:   v.End,
		Views: joinViews(v.Child, v.Children),
	})
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
		Columns int        `json:"start"`
		Gap     int        `json:"span"`
		Cells   []GridCell `json:"cells"`
	}{
		Type:    "GridCell",
		Columns: v.Columns,
		Gap:     v.Gap,
		Cells:   slice.UnsafeUnwrap(v.Cells),
	})
}
