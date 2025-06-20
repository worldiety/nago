// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package json

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"github.com/worldiety/jsonptr"
	"go.wdy.de/nago/application/dataimport/parser"
	"go.wdy.de/nago/pkg/xiter"
	icons "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"io"
	"iter"
	"log/slog"
	"unicode"
)

const ID parser.ID = "nago.data.parser.json"

type jsonParser struct {
}

func NewParser() parser.Parser {
	return jsonParser{}
}

func (p jsonParser) Identity() parser.ID {
	return ID
}

func (p jsonParser) Configuration() parser.Configuration {
	return parser.Configuration{
		Image:       icons.FileCode,
		Name:        "JSON",
		Description: "Der JSON Importer unterstützt den Import eines einzelnen JSON Objektes oder eines ganzen JSON-Arrays von JSON-Objekten. Die Textkodierung muss als UTF-8 vorliegen. Ebenfalls unterstützt wird JSONL.",
		FromUpload: parser.FromUpload{
			Enabled:       true,
			MimeTypes:     []string{"application/json"},
			MaxUploadSize: 1024 * 1024 * 1024, // 1GiB is easily reached when importing time series data
		},
	}
}

func (p jsonParser) Parse(ctx context.Context, reader io.Reader, opts parser.Options) iter.Seq2[*jsonptr.Obj, error] {

	// Wrap with buffered reader for peek
	buffered := bufio.NewReader(reader)

	// Peek to detect delimiter from first line
	peekBytes, err := buffered.Peek(4096)
	if err != nil && err != io.EOF {
		return xiter.WithError[*jsonptr.Obj](err)
	}

	if maybeArray(peekBytes) {
		return parseArray(buffered)
	}

	if maybeJSONL(peekBytes) {
		return parseJSONL(buffered)
	} else {
		slog.Info("json import parser did not detect JSONL. If this is unexpected, check if first line is longer than 4KiB, otherwise the first object is truncated and we cannot detect it.")
		return parseObj(buffered)
	}

}

func parseArray(r io.Reader) iter.Seq2[*jsonptr.Obj, error] {
	return func(yield func(*jsonptr.Obj, error) bool) {
		var tmp []jsonptr.Obj
		dec := json.NewDecoder(r)
		if err := dec.Decode(&tmp); err != nil {
			yield(&jsonptr.Obj{}, err)
			return
		}

		for _, obj := range tmp {
			if !yield(&obj, nil) {
				return
			}
		}
	}
}

func parseJSONL(r io.Reader) iter.Seq2[*jsonptr.Obj, error] {
	return func(yield func(*jsonptr.Obj, error) bool) {
		scanner := bufio.NewScanner(r)

		for scanner.Scan() {
			line := scanner.Bytes()
			var obj *jsonptr.Obj

			if err := json.Unmarshal(line, &obj); err != nil {
				if !yield(obj, err) {
					return
				}
			}

			if !yield(obj, nil) {
				return
			}
		}

		if err := scanner.Err(); err != nil {
			if !yield(&jsonptr.Obj{}, err) {
				return
			}
		}
	}
}

func parseObj(r io.Reader) iter.Seq2[*jsonptr.Obj, error] {
	return func(yield func(*jsonptr.Obj, error) bool) {
		var tmp *jsonptr.Obj
		dec := json.NewDecoder(r)
		if err := dec.Decode(&tmp); err != nil {
			yield(tmp, err)
			return
		}

		yield(tmp, nil)
	}
}

func maybeArray(buf []byte) bool {
	for _, b := range buf {
		// inspect without producing any garbage
		if unicode.IsSpace(rune(b)) {
			continue
		}

		return b == '['
	}

	return false
}

func maybeJSONL(buf []byte) bool {
	for _, line := range bytes.Split(buf, []byte("\n")) {
		return json.Valid(line) // check if first line is a valid json something
	}

	return false
}
