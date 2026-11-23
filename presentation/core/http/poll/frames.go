// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package poll

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// WriteFrames writes the messages in the body format of both directions: the number of the first message as
// uvarint, followed by each message as uvarint length and bytes. A message is what a websocket message holds, thus
// an encoded nago event. A body without messages consists of the number of the next message only.
func WriteFrames(w io.Writer, first uint64, msgs [][]byte) error {
	buf := binary.AppendUvarint(nil, first)
	if _, err := w.Write(buf); err != nil {
		return err
	}

	for _, msg := range msgs {
		buf = binary.AppendUvarint(buf[:0], uint64(len(msg)))
		if _, err := w.Write(buf); err != nil {
			return err
		}

		if _, err := w.Write(msg); err != nil {
			return err
		}
	}

	return nil
}

// ReadFrames reads a body of [WriteFrames]. A message must not exceed maxMsg bytes.
func ReadFrames(r io.Reader, maxMsg int) (first uint64, msgs [][]byte, err error) {
	br := bufio.NewReader(r)
	first, err = binary.ReadUvarint(br)
	if err != nil {
		return 0, nil, fmt.Errorf("cannot read the number of the first message: %w", err)
	}

	for {
		size, err := binary.ReadUvarint(br)
		if errors.Is(err, io.EOF) {
			return first, msgs, nil
		}

		if err != nil {
			return 0, nil, fmt.Errorf("cannot read the length of message %d: %w", len(msgs), err)
		}

		if size > uint64(maxMsg) {
			return 0, nil, fmt.Errorf("message %d is too large: %d > %d", len(msgs), size, maxMsg)
		}

		msg := make([]byte, size)
		if _, err := io.ReadFull(br, msg); err != nil {
			return 0, nil, fmt.Errorf("cannot read message %d: %w", len(msgs), err)
		}

		msgs = append(msgs, msg)
	}
}
