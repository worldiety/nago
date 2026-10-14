// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package sheet

import (
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/proto"
	"golang.org/x/text/language"
)

// maxCopyCells bounds a range, which is copied on the server, e.g. a whole column of a large sheet.
const maxCopyCells = 500_000

// tsv returns the visible cells of the range as tab separated values, which a spreadsheet application pastes into
// cells. Numbers use the decimal separator of the locale, so that e.g. a German Excel parses them as numbers.
func (c TSheet) tsv(r Range, locale language.Tag) (string, error) {
	if r.Rows*r.Cols > maxCopyCells {
		return "", fmt.Errorf("the range has %d cells, but at most %d can be copied", r.Rows*r.Cols, maxCopyCells)
	}

	sep := decimalSeparator(locale)
	var sb strings.Builder
	first := true
	for row := r.Row; row < r.Row+r.Rows; row++ {
		if c.hiddenRows[row] {
			continue
		}

		if !first {
			sb.WriteString("\r\n")
		}
		first = false

		firstCol := true
		for col := r.Col; col < r.Col+r.Cols; col++ {
			if c.hiddenCols[col] {
				continue
			}

			if !firstCol {
				sb.WriteByte('\t')
			}
			firstCol = false

			sb.WriteString(tsvValue(c.cell(row, col), sep))
		}
	}

	return sb.String(), nil
}

func tsvValue(cell Cell, decimalSep string) string {
	v := cell.Value
	if v == "" {
		v = cell.Text
	}

	if cell.Type == ValueNumber && cell.Value != "" && decimalSep != "." {
		v = strings.Replace(v, ".", decimalSep, 1)
	}

	// a tab or a line break would move the following values into other cells
	if strings.ContainsAny(v, "\t\r\n\"") {
		v = `"` + strings.ReplaceAll(v, `"`, `""`) + `"`
	}

	return v
}

// commaLanguages write a decimal comma.
var commaLanguages = []string{"bg", "cs", "da", "de", "el", "es", "et", "fi", "fr", "hr", "hu", "id", "it", "lt", "lv", "nb", "nl", "nn", "no", "pl", "pt", "ro", "ru", "sk", "sl", "sr", "sv", "tr", "uk", "vi"}

func decimalSeparator(locale language.Tag) string {
	base, _ := locale.Base()
	if slices.Contains(commaLanguages, base.String()) {
		return ","
	}

	return "."
}

// errorText is shown at a cell, whose edit has been rejected.
func errorText(err error) string {
	var localized std.LocalizedError
	if errors.As(err, &localized) {
		if localized.Description() != "" {
			return localized.Description()
		}
		return localized.Title()
	}

	slog.Error("cannot apply the edit of a sheet cell", "err", err)
	return "Die Eingabe konnte nicht übernommen werden."
}

func sortInts(v proto.Ints) {
	slices.Sort(v)
}
