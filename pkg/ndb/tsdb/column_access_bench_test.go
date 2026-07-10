// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tsdb

import (
	"fmt"
	"testing"
)

// BenchmarkColumnAccessSerial measures the cost of resolving an already-open
// column from a single goroutine (the cache-hit path).
func BenchmarkColumnAccessSerial(b *testing.B) {
	db, _ := Open(b.TempDir(), Options{})
	defer db.Close()
	schema := Schema{Scheme: SchemeDecimal, Decimals: 2}
	if _, err := db.Column("bucket", "col0", schema); err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := db.Column("bucket", "col0", schema); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkColumnAccessParallel simulates stateless callers that re-resolve the
// column on every operation from many goroutines concurrently. It stresses the
// column-cache lookup path for reader contention.
func BenchmarkColumnAccessParallel(b *testing.B) {
	db, _ := Open(b.TempDir(), Options{})
	defer db.Close()
	schema := Schema{Scheme: SchemeDecimal, Decimals: 2}
	const cols = 16
	names := make([]string, cols)
	for i := 0; i < cols; i++ {
		names[i] = fmt.Sprintf("col%d", i)
		if _, err := db.Column("bucket", names[i], schema); err != nil {
			b.Fatal(err)
		}
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			name := names[i%cols]
			i++
			if _, err := db.Column("bucket", name, schema); err != nil {
				b.Fatal(err)
			}
		}
	})
}
