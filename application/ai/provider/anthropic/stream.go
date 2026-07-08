// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package anthropic

import (
	"bufio"
	"io"
	"strings"
)

// CreateMessageStream performs a streaming Messages API call. onEvent is invoked synchronously for every
// server-sent event (SSE) with the raw event name and the associated JSON data payload.
func (c *Client) CreateMessageStream(req apiRequest, onEvent func(event string, data []byte) error) error {
	req.Stream = true

	var cbErr error
	r := c.newReq().
		URL("messages").
		Assert2xx(true).
		BodyJSON(req).
		Header("Accept", "text/event-stream").
		ToCloser(func(rc io.ReadCloser) {
			defer func() { _ = rc.Close() }()
			cbErr = parseSSE(rc, onEvent)
		})

	if requestUsesFileSource(req) {
		// See CreateMessage: file-id sources require the Files API beta header on the Messages request.
		r = r.Header("anthropic-beta", filesAPIBeta)
	}

	err := r.Post()

	if err != nil {
		return mapErr(err)
	}

	return cbErr
}

// parseSSE reads a text/event-stream and dispatches each event. It supports multi-line data fields and
// ignores comment/ping lines (prefixed with ':').
func parseSSE(r io.Reader, onEvent func(event string, data []byte) error) error {
	sc := bufio.NewScanner(r)
	// model output can produce large single SSE frames, so use a generous buffer
	sc.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)

	var event string
	var data []byte

	dispatch := func() error {
		if event == "" && len(data) == 0 {
			return nil
		}
		err := onEvent(event, data)
		event = ""
		data = nil
		return err
	}

	for sc.Scan() {
		line := sc.Text()
		switch {
		case line == "":
			if err := dispatch(); err != nil {
				return err
			}
		case strings.HasPrefix(line, ":"):
			// comment / keep-alive ping, ignore
		case strings.HasPrefix(line, "event:"):
			event = strings.TrimSpace(line[len("event:"):])
		case strings.HasPrefix(line, "data:"):
			d := strings.TrimPrefix(line[len("data:"):], " ")
			if len(data) > 0 {
				data = append(data, '\n')
			}
			data = append(data, d...)
		}
	}

	if err := sc.Err(); err != nil {
		return err
	}

	// flush a trailing event without terminating blank line
	return dispatch()
}
