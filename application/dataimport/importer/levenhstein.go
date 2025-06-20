// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package importer

func LevenshteinDistance(a, b string) int {
	la := len(a)
	lb := len(b)

	dist := make([][]int, la+1)
	for i := range dist {
		dist[i] = make([]int, lb+1)
	}

	for i := 0; i <= la; i++ {
		dist[i][0] = i
	}
	for j := 0; j <= lb; j++ {
		dist[0][j] = j
	}

	for i := 1; i <= la; i++ {
		for j := 1; j <= lb; j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			dist[i][j] = min(
				dist[i-1][j]+1,      // delete
				dist[i][j-1]+1,      // insert
				dist[i-1][j-1]+cost, // replace
			)
		}
	}

	return dist[la][lb]
}
