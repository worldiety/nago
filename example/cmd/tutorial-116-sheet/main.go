// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"cmp"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"
	"sync"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	. "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/sheet"
	"go.wdy.de/nago/web/vuejs"
)

// position is a row of the example, e.g. of a planning table.
type position struct {
	Name     string
	Category string
	Amount   float64
	Price    float64
	Approved bool
	Note     string
	Checked  bool
}

var categories = []string{"Material", "Personal", "Reise", "Sonstiges"}

// model is kept by the application, the sheet only asks for the cells it shows.
type model struct {
	mutex     sync.RWMutex
	positions []position
}

func newModel(n int) *model {
	m := &model{positions: make([]position, n)}
	for i := range m.positions {
		m.positions[i] = position{
			Name:     fmt.Sprintf("Position %d", i+1),
			Category: categories[i%len(categories)],
			Amount:   float64(i%17 + 1),
			Price:    float64(i%97)*1.25 + 3,
			Approved: i%3 == 0,
			Note:     strings.Repeat("Ein längerer Hinweis, der umbricht. ", i%3),
		}
	}
	return m
}

const (
	colName = iota
	colCategory
	colAmount
	colPrice
	colTotal
	colApproved
	colNote
	colAction
	colCount
)

var header = []string{"Bezeichnung", "Kategorie", "Menge", "Preis", "Summe", "Freigabe", "Hinweis", "Aktion"}

var (
	headerStyle = sheet.Style{Borders: sheet.Borders{Bottom: sheet.Line{Width: 2, Style: sheet.LineSolid}}}
	totalStyle  = sheet.Style{Bold: true, BackgroundColor: "#fff4e0", Color: "#7a4b00"}
	noteStyle   = sheet.Style{Wrap: sheet.WrapWrap, Italic: true}
	nameStyle   = sheet.Style{Wrap: sheet.WrapEllipsis}
)

func money(v float64) string {
	return strings.ReplaceAll(strconv.FormatFloat(v, 'f', 2, 64), ".", ",") + " €"
}

// cell returns a cell of the position with the given index. changed renders the window again after a use case
// changed the model.
func (m *model) cell(index, col int, changed func()) sheet.Cell {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	p := m.positions[index]
	switch col {
	case colName:
		return sheet.Cell{Text: p.Name, Editor: sheet.EditorText, Style: nameStyle}
	case colCategory:
		return sheet.Cell{Text: p.Category, Editor: sheet.EditorSelect, Options: categories}
	case colAmount:
		return sheet.Cell{Text: strconv.FormatFloat(p.Amount, 'f', -1, 64), Value: strconv.FormatFloat(p.Amount, 'f', -1, 64), Type: sheet.ValueNumber, Editor: sheet.EditorNumber}
	case colPrice:
		return sheet.Cell{Text: money(p.Price), Value: strconv.FormatFloat(p.Price, 'f', 2, 64), Type: sheet.ValueNumber, Editor: sheet.EditorNumber}
	case colTotal:
		// a metric, which a use case calculates, e.g. after an edit
		return sheet.Cell{Text: money(p.Amount * p.Price), Value: strconv.FormatFloat(p.Amount*p.Price, 'f', 2, 64), Type: sheet.ValueNumber, Style: totalStyle}
	case colApproved:
		return sheet.Cell{Value: strconv.FormatBool(p.Approved), Type: sheet.ValueBool, Editor: sheet.EditorCheckbox}
	case colAction:
		// any view, e.g. a button, which invokes a use case
		if p.Checked {
			return sheet.Cell{Text: "Geprüft ✓"}
		}

		return sheet.Cell{Text: "Prüfen", View: HStack(Text("Prüfen").Color(I0)).Action(func() {
			m.mutex.Lock()
			m.positions[index].Checked = true
			m.mutex.Unlock()
			changed()
		})}
	default:
		return sheet.Cell{Text: p.Note, Editor: sheet.EditorText, Style: noteStyle}
	}
}

// text returns the text of a column, by which the positions are filtered and sorted.
func (p position) text(col int) string {
	switch col {
	case colName:
		return p.Name
	case colCategory:
		return p.Category
	case colNote:
		return p.Note
	case colApproved:
		return strconv.FormatBool(p.Approved)
	default:
		return ""
	}
}

// number returns the number of a column, by which the positions are sorted.
func (p position) number(col int) (float64, bool) {
	switch col {
	case colAmount:
		return p.Amount, true
	case colPrice:
		return p.Price, true
	case colTotal:
		return p.Amount * p.Price, true
	default:
		return 0, false
	}
}

// order stands for a use case, which filters and sorts the positions. It returns their indices in the order, in
// which the sheet shows them.
func (m *model) order(sorting sheet.Sorting, filters map[int]string) []int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var res []int
	for i, p := range m.positions {
		keep := true
		for col, text := range filters {
			value := p.text(col)
			if v, ok := p.number(col); ok {
				value = strconv.FormatFloat(v, 'f', -1, 64)
			}
			if !strings.Contains(strings.ToLower(value), strings.ToLower(text)) {
				keep = false
				break
			}
		}
		if keep {
			res = append(res, i)
		}
	}

	if sorting.Dir == sheet.SortNone {
		return res
	}

	slices.SortStableFunc(res, func(a, b int) int {
		pa, pb := m.positions[a], m.positions[b]
		var c int
		if va, ok := pa.number(sorting.Col); ok {
			vb, _ := pb.number(sorting.Col)
			c = cmp.Compare(va, vb)
		} else {
			c = strings.Compare(strings.ToLower(pa.text(sorting.Col)), strings.ToLower(pb.text(sorting.Col)))
		}
		if sorting.Dir == sheet.SortDesc {
			c = -c
		}
		return c
	})

	return res
}

// edit stands for any use case, which changes the state of the application.
func (m *model) edit(index int, e sheet.Edit) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	p := &m.positions[index]
	switch e.Col {
	case colName:
		if strings.TrimSpace(e.Value) == "" {
			return std.NewLocalizedError("Ungültig", "Die Bezeichnung darf nicht leer sein.")
		}
		p.Name = e.Value
	case colCategory:
		p.Category = e.Value
	case colAmount, colPrice:
		v, err := strconv.ParseFloat(strings.ReplaceAll(strings.TrimSuffix(strings.TrimSpace(e.Value), " €"), ",", "."), 64)
		if err != nil || v < 0 {
			return std.NewLocalizedError("Ungültig", "Bitte eine positive Zahl eingeben.")
		}
		if e.Col == colAmount {
			p.Amount = v
		} else {
			p.Price = v
		}
	case colApproved:
		p.Approved = e.Value == "true"
	case colNote:
		p.Note = e.Value
	}
	return nil
}

func main() {
	data := newModel(50_000)

	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		cfg.RootView(".", func(wnd core.Window) core.View {
			// a use case of an action within a cell changes the model, but no state, so a state tells nago to
			// render again
			version := core.AutoState[int](wnd)

			// sorting and filtering are use cases of the application as well, the sheet only shows their result
			sorting := core.StateOf[sheet.Sorting](wnd, "sorting")
			filters := core.StateOf[map[int]string](wnd, "filters")
			order := core.StateOf[[]int](wnd, "order").Init(func() []int { return data.order(sheet.Sorting{}, nil) })

			cell := func(row, col int) sheet.Cell {
				if row == 0 {
					return sheet.Cell{Text: header[col], Style: headerStyle}
				}
				return data.cell(order.Get()[row-1], col, func() { version.Set(version.Get() + 1) })
			}

			// the merge refers to rows, which only stay together, as long as the rows are not reordered
			var merges []sheet.Merge
			if sorting.Get().Dir == sheet.SortNone && len(filters.Get()) == 0 {
				merges = append(merges, sheet.Merge{Row: 1, Col: colNote, Rows: 3, Cols: 1})
			}

			return VStack(
				Text("Sheet mit 50.000 Zeilen").Font(Title),
				Text("Kopfzeile und erste Spalte sind festgepinnt. Pfeiltasten, Enter, Tab, Escape und F2 funktionieren wie in einer Tabellenkalkulation, und markierte Bereiche lassen sich nach Excel kopieren und aus Excel einfügen. Spaltenbreiten lassen sich am Rand der Kopfzeile ziehen, das Kontextmenü blendet Zeilen und Spalten aus, und die Kopfzeile sortiert und filtert."),
				sheet.Sheet("positions", len(order.Get())+1, colCount, cell).
					Frozen(1, 1).
					ColumnWidths(map[int]float64{colName: 180, colCategory: 120, colNote: 260}).
					RowHeights(map[int]float64{0: 32}).
					AllowPaste(true).
					Merges(merges...).
					OnEdit(func(e sheet.Edit) error { return data.edit(order.Get()[e.Row-1], e) }).
					Sorted(sorting.Get()).
					OnSort(func(s sheet.Sorting) {
						sorting.Set(s)
						order.Set(data.order(s, filters.Get()))
					}).
					SortableColumns(colName, colCategory, colAmount, colPrice, colTotal, colApproved).
					Filtered(filters.Get()).
					OnFilter(func(col int, text string) {
						f := maps.Clone(filters.Get())
						if f == nil {
							f = map[int]string{}
						}
						if text == "" {
							delete(f, col)
						} else {
							f[col] = text
						}
						filters.Set(f)
						order.Set(data.order(sorting.Get(), f))
					}).
					FilterableColumns(colName, colCategory, colAmount, colPrice, colNote).
					Frame(Frame{Width: Full, Height: "70dvh"}),
			).Gap(L16).
				Alignment(Leading).
				Padding(Padding{}.All(L24)).
				Frame(Frame{}.FullWidth())
		})
	}).Run()
}
