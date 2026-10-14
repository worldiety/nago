// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"fmt"
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
	colCount
)

var header = []string{"Bezeichnung", "Kategorie", "Menge", "Preis", "Summe", "Freigabe", "Hinweis"}

var (
	headerStyle = sheet.Style{Borders: sheet.Borders{Bottom: sheet.Line{Width: 2, Style: sheet.LineSolid}}}
	totalStyle  = sheet.Style{Bold: true, BackgroundColor: "#fff4e0", Color: "#7a4b00"}
	noteStyle   = sheet.Style{Wrap: sheet.WrapWrap, Italic: true}
	nameStyle   = sheet.Style{Wrap: sheet.WrapEllipsis}
)

func money(v float64) string {
	return strings.ReplaceAll(strconv.FormatFloat(v, 'f', 2, 64), ".", ",") + " €"
}

func (m *model) cell(row, col int) sheet.Cell {
	if row == 0 {
		return sheet.Cell{Text: header[col], Style: headerStyle}
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()
	p := m.positions[row-1]
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
	default:
		return sheet.Cell{Text: p.Note, Editor: sheet.EditorText, Style: noteStyle}
	}
}

// edit stands for any use case, which changes the state of the application.
func (m *model) edit(e sheet.Edit) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	p := &m.positions[e.Row-1]
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
			return VStack(
				Text("Sheet mit 50.000 Zeilen").Font(Title),
				Text("Kopfzeile und erste Spalte sind festgepinnt. Pfeiltasten, Enter, Tab, Escape und F2 funktionieren wie in einer Tabellenkalkulation, und markierte Bereiche lassen sich nach Excel kopieren."),
				sheet.Sheet("positions", len(data.positions)+1, colCount, data.cell).
					Frozen(1, 1).
					ColumnWidths(map[int]float64{colName: 180, colCategory: 120, colNote: 260}).
					Merges(sheet.Merge{Row: 1, Col: colNote, Rows: 3, Cols: 1}).
					OnEdit(data.edit).
					Frame(Frame{Width: Full, Height: "70dvh"}),
			).Gap(L16).
				Alignment(Leading).
				Padding(Padding{}.All(L24)).
				Frame(Frame{}.FullWidth())
		})
	}).Run()
}
