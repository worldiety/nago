// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package mail

import (
	"encoding/base64"
	"io"
	"strings"
)

func protect(str string) string {
	str = strings.ReplaceAll(str, "\n\r", " ")
	str = strings.ReplaceAll(str, "\r", " ")
	str = strings.ReplaceAll(str, "\n", " ")
	return str
}

func split1000(str string) []string {
	if len(str) < 1000 {
		return []string{str}
	}

	var lines []string
	for _, line := range strings.Split(str, "\n") {
		if len(line) < 1000 {
			lines = append(lines, line)
		} else {
			var tmp strings.Builder

			for _, r := range line {
				tmp.WriteRune(r)
				if tmp.Len() > 1000 {
					lines = append(lines, tmp.String())
					tmp.Reset()
				}
			}

			if tmp.Len() > 0 {
				lines = append(lines, tmp.String())
			}
		}
	}

	return lines
}

func b64(data []byte) []string {
	str := base64.StdEncoding.EncodeToString(data)
	stride := 72
	lines := make([]string, (len(str)/stride)+1)[:]
	for i := 0; i < len(str); i += stride {
		end := i + stride
		if end > len(str) {
			end = len(str)
		}
		lines = append(lines, str[i:end])
	}
	return lines
}

func writeText(str string, writer io.Writer) error {
	for _, line := range split1000(str) {
		_, err := writer.Write([]byte(line))
		if err != nil {
			return err
		}
		_, err = writer.Write([]byte("\r\n"))
		if err != nil {
			return err
		}
	}
	return nil
}
