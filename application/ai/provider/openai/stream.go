// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package openai

import (
	"bufio"
	"bytes"
	"io"
	"strings"
)

// parseSSE reads a text/event-stream as emitted by the Chat Completions API and dispatches the data payload of
// each event. The terminating "[DONE]" sentinel ends the stream. Event names are ignored, because the API only
// uses unnamed events; comment/keep-alive lines (prefixed with ':') are skipped. Multi-line data fields are
// joined by a newline.
func parseSSE(r io.Reader, onData func(data []byte) error) error {
	sc := bufio.NewScanner(r)
	// model output can produce large single SSE frames, so use a generous buffer
	sc.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)

	var data []byte
	done := false

	dispatch := func() error {
		if len(data) == 0 {
			return nil
		}
		d := data
		data = nil
		if string(bytes.TrimSpace(d)) == "[DONE]" {
			done = true
			return nil
		}
		return onData(d)
	}

	for !done && sc.Scan() {
		line := sc.Text()
		switch {
		case line == "":
			if err := dispatch(); err != nil {
				return err
			}
		case strings.HasPrefix(line, ":"):
			// comment / keep-alive ping, ignore
		case strings.HasPrefix(line, "data:"):
			d := strings.TrimPrefix(line[len("data:"):], " ")
			if len(data) > 0 {
				data = append(data, '\n')
			}
			data = append(data, d...)
		}
	}

	if done {
		return nil
	}

	if err := sc.Err(); err != nil {
		return err
	}

	// flush a trailing event without terminating blank line
	return dispatch()
}
