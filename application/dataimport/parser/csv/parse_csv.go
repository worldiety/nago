// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package csv

import (
	"bufio"
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"github.com/worldiety/jsonptr"
	"go.wdy.de/nago/application/dataimport/parser"
	icons "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
	"io"
	"iter"
	"strings"
)

const ID parser.ID = "nago.data.parser.csv"

type csvParser struct {
}

func NewParser() parser.Parser {
	return csvParser{}
}

func (p csvParser) Identity() parser.ID {
	return ID
}

func (p csvParser) Configuration() parser.Configuration {
	return parser.Configuration{
		Image:       icons.FileCsv,
		Name:        "CSV (Comma Separated Values)",
		Description: "Der CSV Importer unterstützt UTF-8, UTF-16 LE und UTF-16 BE Kodierungen. Die erste Zeile muss die Schlüsselnamen enthalten. Es erfolgt eine automatische Erkennung ob Punkt oder Komma als Trennzeichen verwendet wird.",
		FromUpload: parser.FromUpload{
			Enabled:       true,
			MimeTypes:     []string{"text/csv"},
			MaxUploadSize: 1024 * 1024 * 1024, // 1GiB is easily reached when importing time series data
		},
	}
}

func (p csvParser) Parse(ctx context.Context, reader io.Reader, opts parser.Options) iter.Seq2[*jsonptr.Obj, error] {
	return func(yield func(*jsonptr.Obj, error) bool) {
		buf, err := io.ReadAll(reader)
		if err != nil {
			if !yield(&jsonptr.Obj{}, err) {
				return
			}
		}

		objs, err := parseCSV(buf)
		if err != nil {
			yield(&jsonptr.Obj{}, fmt.Errorf("unable to parse csv: %w", err))
			return
		}

		for _, obj := range objs {
			if !yield(&obj, nil) {
				return
			}
		}
	}
}

func parseCSV(buf []byte) ([]jsonptr.Obj, error) {
	// Decode according to BOM
	decodedReader, err := detectEncoding(buf)
	if err != nil {
		return nil, err
	}

	// Wrap with buffered reader for peek
	buffered := bufio.NewReader(decodedReader)

	// Peek to detect delimiter from first line
	peekBytes, err := buffered.Peek(4096)
	if err != nil && err != io.EOF {
		return nil, err
	}
	firstLine := string(peekBytes)
	firstLine = strings.Split(firstLine, "\n")[0]
	delimiter := detectDelimiter(firstLine)

	// Setup CSV reader
	csvReader := csv.NewReader(buffered)
	csvReader.Comma = delimiter
	csvReader.FieldsPerRecord = -1 // Allow variable field count

	// Read all records
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) < 2 {
		// nothing to extract, at least 2 rows required: first row is the keys and the second is a lonely data row
		return nil, nil
	}

	keyNames := records[0]

	var res []jsonptr.Obj
	for _, record := range records[1:] {
		obj := jsonptr.Obj{}
		for fidx, field := range record {
			if fidx < len(keyNames) {
				key := keyNames[fidx]
				obj.Put(key, jsonptr.String(field))
			}
		}

		res = append(res, obj)
	}

	return res, nil
}

// detectEncoding returns an io.Reader that decodes the byte data
// according to the detected BOM (UTF-8, UTF-16 LE/BE)
func detectEncoding(data []byte) (io.Reader, error) {
	switch {
	case bytes.HasPrefix(data, []byte{0xEF, 0xBB, 0xBF}): // UTF-8 BOM
		return bytes.NewReader(data[3:]), nil
	case bytes.HasPrefix(data, []byte{0xFF, 0xFE}): // UTF-16 LE BOM
		return transform.NewReader(bytes.NewReader(data), unicode.UTF16(unicode.LittleEndian, unicode.ExpectBOM).NewDecoder()), nil
	case bytes.HasPrefix(data, []byte{0xFE, 0xFF}): // UTF-16 BE BOM
		return transform.NewReader(bytes.NewReader(data), unicode.UTF16(unicode.BigEndian, unicode.ExpectBOM).NewDecoder()), nil
	default:
		// Assume UTF-8 without BOM
		return bytes.NewReader(data), nil
	}
}

// detectDelimiter tries to detect the delimiter (comma or semicolon)
// by counting occurrences in the first line
func detectDelimiter(line string) rune {
	commaCount := strings.Count(line, ",")
	semicolonCount := strings.Count(line, ";")
	if semicolonCount > commaCount {
		return ';'
	}
	return ','
}
