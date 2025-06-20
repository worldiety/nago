// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package pdf

import (
	"bytes"
	"context"
	"github.com/worldiety/jsonptr"
	"go.wdy.de/nago/application/dataimport/parser"
	icons "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"golang.org/x/text/encoding/unicode"
	"io"
	"iter"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
)

const ID parser.ID = "nago.data.parser.pdf"

type pdfParser struct {
}

func NewParser() parser.Parser {
	return pdfParser{}
}

func (p pdfParser) Identity() parser.ID {
	return ID
}

func (p pdfParser) Configuration() parser.Configuration {
	return parser.Configuration{
		Image:       icons.FilePdf,
		Name:        "PDF AcroForms",
		Description: "Der PDF AcroForms Importer unterstützt unverschlüsselte und unkomprimierte Formularfelder in UTF-16 Kodierung. Die komplette PDF-Datei wird immer als ein einzelnes Formularobjekt behandelt. Tabellen oder sich wiederholende Formulare werden nicht unterstützt.",
		FromUpload: parser.FromUpload{
			Enabled:       true,
			MimeTypes:     []string{"application/pdf"},
			MaxUploadSize: 1024 * 1024 * 128, // 128MiB is already large for a PDF form, everything else is probably a broken PDF
		},
	}
}

func (p pdfParser) Parse(ctx context.Context, reader io.Reader, opts parser.Options) iter.Seq2[*jsonptr.Obj, error] {
	return func(yield func(*jsonptr.Obj, error) bool) {
		buf, err := io.ReadAll(reader)
		if err != nil {
			if !yield(&jsonptr.Obj{}, err) {
				return
			}
		}

		if !yield(parsePDF(buf), nil) {
			return
		}
	}
}

func parsePDF(buf []byte) *jsonptr.Obj {
	fieldRegex := regexp.MustCompile(`/T\s*\(([^)]+)\)[\s\S]+?/V\s*\(([^)]+)\)`)

	matches := fieldRegex.FindAllSubmatch(buf, -1)

	res := &jsonptr.Obj{}
	for _, m := range matches {
		key := strings.TrimSpace(string(m[1]))
		val, err := decodeValue(string(m[2]))
		if err != nil {
			slog.Error("decode pdf field value error", "err", err.Error())
		}

		res.Put(key, jsonptr.String(val))
	}

	return res
}

// decodeValue attempts simple Unicode/ASCII sanitization
func decodeValue(val string) (string, error) {
	// If no \000 is included, return string directly
	if !strings.Contains(val, `\000`) {
		return val, nil
	}

	// Extract the UTF-16BE characters as a byte sequence
	var buf bytes.Buffer
	r := regexp.MustCompile(`\\([0-7]{3})|.`) // \000-style or single char
	matches := r.FindAllString(val, -1)

	for _, part := range matches {
		if strings.HasPrefix(part, `\`) {
			// Octal-Escape wie \000
			n, err := strconv.ParseInt(part[1:], 8, 16)
			if err == nil {
				buf.WriteByte(byte(n))
			}
		} else {
			//  (z. B. ˛ oder ﬂ)
			buf.WriteByte(part[0])
		}
	}

	// UTF-16BE decode
	decoder := unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM).NewDecoder()
	utf8val, err := decoder.Bytes(buf.Bytes())
	if err != nil {
		return "", err
	}

	return strings.TrimPrefix(string(utf8val), "\uFEFF"), nil
}
