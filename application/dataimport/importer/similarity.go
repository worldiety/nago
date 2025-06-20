// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package importer

// Similarity returns a value between 0.0 (no similarity) and 1.0 (identical).
func Similarity(a, b string) float64 {
	if a == b {
		return 1.0
	}

	distance := LevenshteinDistance(a, b)
	maxLen := max(len(a), len(b))
	return 1.0 - float64(distance)/float64(maxLen)
}
