# tsdb — columnar time series storage engine

`tsdb` is a specialized storage engine for the `ndb` database, optimized for the
workload described in [`../timeseries/README.md`](../timeseries/README.md):
~90 % reads, ~10 % writes, with occasional huge burst inserts (billions of
measurements at once), data from 10-minute to 50 Hz sampling, oscillating
numeric values with small relative change, and occasional value updates and
deletes.

It registers itself with `ndb` under the engine kind `"tsdb"` and shares the
`ndb.DB`'s file-descriptor pool.

## Data model

```
<engine-dir>/
  .engine                       # "tsdb" marker (written by ndb)
  .lock                         # single-process exclusive lock
  <bucket>/
    <column>/
      schema.json               # scheme + decimals (atomic write)
      enum.dict                 # append-only enum dictionary (enum scheme)
      head.wal                  # out-of-order corrections log
      1700000000000_1700019999980.tsb   # finalized chunk: <minMillis>_<maxMillis>
      1700020000000_.tsb                 # pending (open) chunk
```

* **Bucket / column** — the only grouping. Each column is a single typed signal
  with its own **independent time axis** (no shared timebase). Fan-out by
  `bucket/column` gives a free filter: a query for a few columns does zero I/O
  for the rest.
* **Column schemes** — `decimal` (float64 rounded to a scaled int64 by the
  column's `Decimals`), `enum` (runtime-extensible string dictionary), `string`
  (free variable string).
* **Chunk files** are immutable, named by the raw epoch-millisecond range they
  contain, so range queries skip whole files and retention drops whole files
  without opening them. The naming is **timezone-free**; any calendar
  interpretation lives only inside the pluggable `SplitFunc` partitioning policy
  (default: German-quarter *or* 64 MiB, whichever first).

## Encoding

* **Timestamps** — two modes, chosen per block:
  * **Equidistant** (default when a block has a single constant step, i.e. no
    holes): the whole timestamp stream is stored as just `(start, step)` and
    reconstructed by the closed form `ts[i] = start + i*step`. This drops the
    per-point timestamp byte entirely (~4 KB → ~10 bytes for a 4096-point
    block) and makes decoding a branch-free arithmetic loop. A single hole
    downgrades only the one block containing it to delta-of-delta.
  * **Delta-of-delta + zig-zag varint** (fallback for holes / irregular
    sampling). Even here an equidistant run compresses to ~1 byte per point.
* **Decimal values** — the noisy `float64` is rounded to a scaled `int64` using
  the column's `Decimals` (killing float noise), then delta + zig-zag varint.
  Because values oscillate with small relative change, deltas are tiny (1–2
  bytes). With the equidistant timestamp mode the value stream is the dominant
  cost, giving measured storage of **~1 byte per point** (see below).
  * **Constant values** — when every value in a block is identical (a signal
    parked at a setpoint, an idle status, a long unchanged run), the value
    stream is stored as a single varint and reconstructed by filling the slice.
    Combined with the equidistant timestamp mode a flat block becomes a fixed
    ~46-byte frame regardless of point count — near-zero bytes per point.
* **Enum** — dictionary ids, delta-coded; the dictionary is append-only and ids
  are never reused. An all-same-id block uses the same constant encoding.
* Blocks (default 4096 points) are CRC-protected with an 8-byte sync marker for
  forward-scan recovery, and optionally `s2`-compressed.

## Write path — two lanes

The single most important design decision. Writes are split by monotonicity so
that burst ingest never accumulates in memory:

1. **Monotonic append fast path** (`ts > column max`, the burst/normal case).
   Points accumulate in one in-memory block (~4096 points); a full block is
   encoded and appended straight to the pending chunk file; the split policy
   finalizes chunks. **No B-tree, no per-point WAL, no chunk rewrite.** Memory is
   bounded to a single block per column regardless of how many points are
   ingested — a 20-billion-point burst runs in **constant memory**.

2. **Out-of-order / overwrite / delete path** (`ts ≤ column max`, corrections
   and bulk rewrites). These go to a bounded in-memory head (B-tree) + `head.wal`
   and are merged on read (newest wins; tombstones mask). The head has a **hard
   cap** (`MaxHeadPoints`, default 1 M); exceeding it triggers a synchronous
   flush so head memory is strictly bounded even under adversarially
   out-of-order load. Compaction rewrites the affected chunks to physically
   apply the corrections and reclaim space.

Reads merge sealed chunks + the unsealed append tail + head overrides. The
common post-flush read (empty head) takes an allocation-free fast path that
streams decoded blocks directly to the caller.

## Read API

* `ScanI64` / `ScanF64` / `ScanString` — columnar batch reader. Yields decoded
  blocks as parallel slices with (near) zero per-point cost; the path for
  billion-element scans and aggregations.
* `IterI64` / `IterF64` / `IterString` — `iter.Seq` of points; composes directly
  with `pkg/timeseries.M4` and `Series`. ~25 % slower than `Scan*` due to the
  per-point delivery, but ergonomic. Use `Scan*` for bulk work.

## Benchmarks

Measured on an Apple M1 Max (macOS, arm64, 10 cores), decimal column, 2 decimal
places, 50 Hz equidistant timestamps, oscillating values, no compression.

### End-to-end scale (single monotonic series)

Reproduce with:

```
go test -run '^TestScale$' -v ./pkg/ndb/tsdb/                                  # 1M
TSDB_SCALES=1000000,1000000000 go test -run '^TestScale$' -v -timeout 30m \
    ./pkg/ndb/tsdb/                                                            # 1M + 1B
```

| points | insert  | insert pts/s | read   | read pts/s | on-disk  | bytes/pt | peak heap growth |
|--------|---------|--------------|--------|------------|----------|----------|------------------|
| 1 M    | 26.7 ms | 37.4 M/s     | 6.0 ms | 165 M/s    | 987 KiB  | 1.01     | 0.0 MiB          |
| 1 B    | 25.3 s  | 39.5 M/s     | 3.5 s  | 286 M/s    | 964 MiB  | 1.01     | 3.5 MiB          |

(Oscillating 50 Hz values. The equidistant-timestamp mode stores the per-block
time axis as `(start, step)`, so only the value delta remains → ~1 byte/point.)

### Constant / unchanged tracks (`TSDB_FLAT=1`)

When values do not change (a setpoint, an idle signal), the constant-value block
mode makes storage and speed essentially independent of point count:

| points | insert  | insert pts/s | read   | read pts/s | on-disk  | bytes/pt |
|--------|---------|--------------|--------|------------|----------|----------|
| 1 M    | 16.8 ms | 59.7 M/s     | 4.2 ms | 239 M/s    | 11.1 KiB | 0.01     |
| 1 B    | 15.7 s  | 63.8 M/s     | 1.94 s | 514 M/s    | 10.7 MiB | 0.01     |

A billion unchanged samples occupy **10.7 MiB** (≈180× smaller than the
oscillating case) and read at ~510 M points/s.

### Micro-benchmarks (`go test -bench . -benchmem`)

| Operation  | ns/op | throughput  | B/op | allocs/op |
|------------|-------|-------------|------|-----------|
| `PutI64`   | 25    | 40 M/s      | 4    | 0         |
| `ScanI64`  | —     | 348 M pts/s | ~1 buffer | 1    |
| `IterI64`  | —     | 217 M pts/s | ~1 buffer | 1    |
| `IterF64`  | —     | 130 M pts/s | ~1 buffer | 2    |

A read over a column with an unflushed pending chunk (data written but not yet
`Flush`-ed) performs within ~1 % of a flushed read (348 vs 348 M pts/s,
`BenchmarkScanI64Pending` vs `BenchmarkScanI64`): the read merges the pending
chunk in-place using its already-known size, so reader-visible unflushed data
adds no measurable cost.

### Concurrent columns (read/write scaling)

Each worker operates on its own column, on its own goroutine. Columns are fully
independent: a read holds only the per-column lock, a write only the per-column
lock plus the sharded [`ndb.FilePool`](../filepool.go), so disjoint columns run
in parallel with no shared serialization point. Aggregate throughput therefore
rises with the number of concurrent columns until it saturates memory bandwidth.
Reproduce with:

```
go test -run '^TestConcurrentColumnsScaleAndCorrect$' -v ./pkg/ndb/tsdb/
go test -run '^$' -bench BenchmarkParallelReadColumns ./pkg/ndb/tsdb/
```

Write (each worker appends 2 M points to its own column):

| workers | total pts/s | speedup |
|---------|-------------|---------|
| 1       | 36 M/s      | 1.00×   |
| 2       | 72 M/s      | 1.99×   |
| 4       | 125 M/s     | 3.45×   |
| 8       | 153 M/s     | 4.21×   |

Read (each worker scans its own 1 M-point column):

| workers | total pts/s | speedup |
|---------|-------------|---------|
| 1       | 259 M/s     | 1.00×   |
| 2       | 500 M/s     | 1.93×   |
| 4       | 956 M/s     | 3.69×   |
| 8       | 773 M/s     | 2.98×   |

Reads scale near-linearly to 4 workers (≈960 M points/s aggregate) and then flatten
as the shared memory bandwidth — not any lock — becomes the ceiling. A mutex
profile of 8 concurrent readers shows the FilePool contention reduced ~60× by the
per-path shard sharding (single-mutex pool: 189 ms; sharded pool: 3 ms), so the
remaining limit is hardware, not the engine. Runs are data-race clean under
`go test -race`.

## Essential findings

* **Constant-memory burst ingest is real.** Inserting 1 billion points grew the
  heap by only **3.6 MiB** and never required manual flushing or a special
  harness. This is the direct result of routing monotonic appends around the
  in-memory index straight into sealed chunk blocks.
* **Insert throughput is flat with data size** — 28.7 M/s at 1 M and 29.5 M/s at
  1 B points. There is no O(n²) write amplification, because monotonic appends
  never rewrite existing chunks (only out-of-order corrections do).
* **Writes are allocation-free** on the append path (0 allocs/op, ~33 ns/point).
* **Storage is ~1 byte per point** for oscillating decimal data at a constant
  sample rate: the equidistant timestamp mode stores the whole per-block time
  axis as `(start, step)`, so only the scaled-int value delta remains. Irregular
  or holed data falls back to delta-of-delta (~2 bytes/point).
* **Unchanged tracks are nearly free.** A block whose values are all identical is
  stored as one constant, so a flat + equidistant block is a fixed ~46-byte
  frame independent of point count: **~0.01 bytes/point** measured, ~180× smaller
  than oscillating data, and read ~2× faster (a fill loop, no varint decode).
* **Reads sustain ~130–260 M points/s** single-core via the columnar batch API,
  with a constant, one-time-per-call allocation (a reused block buffer). The
  ergonomic `Iter*` API is ~25 % slower purely due to per-point delivery.
* **Concurrent columns scale.** Different columns share no column lock and use a
  per-path-sharded file pool, so reads/writes on disjoint columns run in
  parallel: writes reach ~4.2× and reads ~3.7× aggregate throughput before
  memory bandwidth (not locking) becomes the ceiling. All cross-column access is
  data-race free.

## Durability model (house rules, shared with `msgstore`)

No `fsync`; atomic `pread`/`pwrite`; per-block CRC + sync marker for forward-scan
recovery; temp-file + atomic-rename for schema, dictionary, and compaction;
single-process `flock`; monotonic ids never reused. The not-yet-sealed append
tail block lives in memory and is lost on a hard crash; a pending chunk left over
from a crash is finalized on the next open, recovering all intact blocks and
dropping any torn tail block.
