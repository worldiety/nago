// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build vmstats

package vmcode

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

// statsOn counts the executed instructions and their pairs, e.g. for the choice of superinstructions:
// go test -tags vmstats ...
const statsOn = true

var (
	statsMu sync.Mutex
	ops     = map[op]int{}
	pairs   = map[[2]op]int{}
)

func countOp(prev, cur op) {
	statsMu.Lock()
	ops[cur]++
	pairs[[2]op{prev, cur}]++
	statsMu.Unlock()
}

// Stats returns the most frequent instructions and pairs and resets them.
func Stats() string {
	statsMu.Lock()
	defer statsMu.Unlock()
	var sb strings.Builder
	type kv struct {
		k string
		v int
	}
	var list []kv
	total := 0
	for o, n := range ops {
		list = append(list, kv{opNames[o], n})
		total += n
	}
	sort.Slice(list, func(i, j int) bool { return list[i].v > list[j].v })
	fmt.Fprintf(&sb, "%d instructions\n", total)
	for i, e := range list {
		if i < 25 {
			fmt.Fprintf(&sb, "  %-12s %5.1f%%\n", e.k, float64(e.v)*100/float64(total))
		}
	}
	list = list[:0]
	for p, n := range pairs {
		list = append(list, kv{opNames[p[0]] + " " + opNames[p[1]], n})
	}
	sort.Slice(list, func(i, j int) bool { return list[i].v > list[j].v })
	for i, e := range list {
		if i < 25 {
			fmt.Fprintf(&sb, "  %-24s %5.1f%%\n", e.k, float64(e.v)*100/float64(total))
		}
	}
	clear(ops)
	clear(pairs)
	return sb.String()
}
