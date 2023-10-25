package ui

import (
	"encoding/json"
	"go.wdy.de/nago/container/slice"
)

type Table struct {
	Rows slice.Slice[TableRow]
}

func (t Table) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type string                `json:"type"`
		Rows slice.Slice[TableRow] `json:"rows"`
	}{
		Type: "Table",
		Rows: t.Rows,
	})
}

func (t Table) isView() {
}

type TableRow struct {
	Columns slice.Slice[TableCell]
}

func (t TableRow) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type    string                 `json:"type"`
		Columns slice.Slice[TableCell] `json:"columns"`
	}{
		Type:    "TableRow",
		Columns: t.Columns,
	})
}

type TableCell struct {
	Child    View
	Children slice.Slice[View]
}

func (t TableCell) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type  string `json:"type"`
		Views []View `json:"views"`
	}{
		Type:  "TableCell",
		Views: joinViews(t.Child, t.Children),
	})
}
