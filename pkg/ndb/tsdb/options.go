// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tsdb

import "go.wdy.de/nago/pkg/ndb"

// Options configures a tsdb engine instance.
type Options struct {
	// FilePool is the shared file-descriptor pool. When opened through ndb, the
	// DB injects its shared pool here; nil defaults to ndb.NewFilePool(1024).
	FilePool *ndb.FilePool

	// Split decides when the pending chunk of a column is sealed. nil defaults
	// to a German-quarter-or-64MiB policy (see defaultSplit).
	Split SplitFunc

	// BlockPoints is the target number of points per compressed block. nil/0
	// defaults to 4096. Larger blocks compress better but cost more to decode
	// for point lookups.
	BlockPoints int

	// Compress enables s2 compression of block bodies larger than 512 bytes.
	Compress bool

	// FlushBytes is the head-log payload volume (per column) that triggers a
	// flush of the head into a sealed chunk. 0 defaults to 8 MiB.
	FlushBytes int64

	// MaxHeadPoints is the hard cap on the number of out-of-order/overwrite
	// entries held in a column's in-memory head. When exceeded, the writing
	// goroutine flushes synchronously so head memory is strictly bounded
	// regardless of how out-of-order the workload is. 0 defaults to 1,000,000.
	// Monotonic appends never enter the head, so a normal in-order or bursty
	// ingest keeps the head empty and never hits this cap.
	MaxHeadPoints int

	// CompactTombstoneRatio triggers compaction of a column when the fraction
	// of head entries that are tombstones or overwrites exceeds this value.
	// 0 defaults to 0.25.
	CompactTombstoneRatio float64
}

const (
	defaultBlockPoints          = 4096
	defaultFlushBytes     int64 = 8 << 20
	defaultTombstoneRatio       = 0.25
	defaultMaxHeadPoints        = 1_000_000
)

func (o *Options) resolve() {
	if o.FilePool == nil {
		o.FilePool = ndb.NewFilePool(1024)
	}
	if o.Split == nil {
		o.Split = defaultSplit()
	}
	if o.BlockPoints <= 0 {
		o.BlockPoints = defaultBlockPoints
	}
	if o.FlushBytes <= 0 {
		o.FlushBytes = defaultFlushBytes
	}
	if o.MaxHeadPoints <= 0 {
		o.MaxHeadPoints = defaultMaxHeadPoints
	}
	if o.CompactTombstoneRatio <= 0 {
		o.CompactTombstoneRatio = defaultTombstoneRatio
	}
}
