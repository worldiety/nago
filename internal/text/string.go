// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package text

import (
	"regexp"
	"strings"
)

var safeRegex = regexp.MustCompile(`[^a-z0-9_-]+`)

func SafeName(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "_")
	s = safeRegex.ReplaceAllString(s, "")
	return s
}
