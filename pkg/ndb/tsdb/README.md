# tsdb — columnar time series storage engine

`tsdb` is a specialized storage engine for the `ndb` database, optimized for the
workload described in [`../timeseries/README.md`](../timeseries/README.md):
~90 % reads, ~10 % writes, with occasional huge burst inserts (billions of
measurements at once), data from 10-minute to 50 Hz sampling, oscillating
numeric values with small relative change, and occasional value updates and
deletes.

It registers itself with `ndb` under the engine kind `"tsdb"` and shares the
`ndb.DB`'s file-descriptor pool. Every access to a file of a column goes through
that pool, also creating, reading or writing a whole file, truncating, renaming,
removing and syncing; only listing, creating and removing the directories do
not. `ndb.FilePool.Stats` counts the files it opened and the bytes it read and
wrote.

## Data model

```
<engine-dir>/
  .engine                       # "tsdb" marker (written by ndb)
  .lock                         # single-process exclusive lock
  <bucket>/
    <column>/
      schema.json               # scheme + decimals (new file + rename)
      enum.dict                 # append-only enum dictionary (enum scheme)
      head.wal                  # out-of-order corrections log
      1700000000000_1700019999980.tsb   # finalized chunk: <minMillis>_<maxMillis>
      1700020000000_.tsb                 # pending (open) chunk, continued across restarts
```

A column holds no other file after a Sync, a Close, an open or a finished
compaction. A running compaction adds files ending in `.new` and `compact.log`
(see [Compaction](#compaction)).

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
* **Pending chunk** — each column has at most one open chunk `<minMillis>_.tsb`,
  which holds its newest points: sealed blocks, then the open block. The points
  of the open block, which a `Column.Sync` saved, are continuation records behind
  the sealed blocks; the others are only in memory. The append path seals the
  open block when it is full and the split policy finalizes the chunk under its
  `<minMillis>_<maxMillis>` name; from then on it is immutable like every
  finalized chunk. `Column.Sync`, `DB.Close` and a restart leave it open, so a
  process that starts once a day does not begin a new file per column and day.
  Only `Column.Flush` finalizes it early.
* **Chunk header** — `NAGO` `NTSB` and a version. Version 2 (every file this
  version writes) adds a random salt with a CRC and two slots, each with where the
  sealed blocks end, their points and their newest time, so that the open of a
  column reads only the end of its open chunk; a seal writes the older slot, so
  that a torn write leaves the other. The bodies of its blocks and continuation
  records are masked with a key stream of the salt and their offset, so that no
  value a caller writes appears in the file as written and neither the search for
  blocks nor the recovery can take its bytes for a frame. Version 1 files (blocks
  right behind the 9-byte header, unmasked) are read and continued as they are;
  their open chunk is read as a whole when a column opens, and they are never
  searched, only read from their first block. Every later version keeps the
  prefix of version 2 (magic, format, version, salt and the CRC over them), so
  that a reader tells a chunk of a later version from a corrupt file, see
  [Compatibility](#compatibility).

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
  the column's `Decimals` (killing float noise), then encoded in one of two ways,
  chosen per block (see below). Because values oscillate with small relative
  change, deltas are tiny. With the equidistant timestamp mode the value stream is
  the dominant cost, giving measured storage of **~1.5 bytes per point** for a
  power curve with one decimal (see below).
  * **Frame of reference of the deltas** (flag `0x10`, `forcodec.go`): the deltas
    of the values in sub-blocks of 128, each packed with the bit width of its
    range above its smallest delta, in groups of 64, so that a reader unpacks a
    value with a shift, a mask and two additions (`unpack_gen.go`, generated by
    `internal/unpackgen`), instead of decoding a varint. The positions of `NA`
    values stand in a list in front of the packed deltas, in which an `NA`
    repeats the value before it, so that a rare `NA` does not widen its
    sub-block to 64 bits.
  * **Delta + zig-zag varint**, as before. A numeric block, whose values are not
    all the same, is encoded both ways, the varints `s2`-compressed if that makes
    them smaller (as before), the frame of reference too, if it is larger than
    that, and holds the smaller one, the frame of reference on a tie; so no
    block, and no column, is larger than with varints alone. A status, which keeps
    one value almost always, stays varint with `s2`. A block is encoded twice, but
    without the byte-wise buffer of before, so appending is not slower (see the
    benchmarks).
  * **Constant values** — when every value in a block is identical (a signal
    parked at a setpoint, an idle status, a long unchanged run), the value
    stream is stored as a single varint and reconstructed by filling the slice.
    Combined with the equidistant timestamp mode a flat block becomes a fixed
    ~46-byte frame regardless of point count — near-zero bytes per point.
* **Enum** — dictionary ids, encoded like decimal values; the dictionary is
  append-only and ids are never reused. An all-same-id block uses the same
  constant encoding.
* Blocks (default 4096 points) are CRC-protected with an 8-byte sync marker for
  forward-scan recovery, and optionally `s2`-compressed.

## Write path — two lanes

The single most important design decision. Writes are split by monotonicity so
that burst ingest never accumulates in memory:

1. **Monotonic append fast path** (`ts > column max`, the burst/normal case).
   Points form the open block of the pending chunk (~4096 points). Those not yet
   saved are held in memory, encoded like a continuation record (a few bytes per
   point, at most 24). A full block is encoded and written to the pending chunk
   file; the split policy finalizes chunks. **No B-tree, no per-point WAL, no
   chunk rewrite.** Memory is bounded to the unsaved points of one block per
   column regardless of how many points are ingested — a 20-billion-point burst
   runs in **constant memory** — and a Sync, a sealed block or the open leave
   none of it.

2. **Out-of-order / overwrite / delete path** (`ts ≤ column max`, corrections
   and bulk rewrites). These go to a bounded in-memory head (B-tree) + `head.wal`
   and are merged on read (newest wins; tombstones mask). A put or a delete is
   saved when it returns: it is in the WAL. The head has a **hard
   cap** (`MaxHeadPoints`, default 1 M); exceeding it triggers a synchronous
   flush so head memory is strictly bounded even under adversarially
   out-of-order load. Compaction rewrites the affected chunks to physically
   apply the corrections and reclaim space.

### Sync

`Column.Sync` appends the points of the open block, which only the memory holds,
as one continuation record (marker, length, kind, count, the points, CRC; one
point takes at most 48 bytes besides its text) behind the open chunk and frees
their memory. It writes into one file, only appends and changes nothing else; with
nothing in memory it writes nothing. The points after it continue the open block.
When the block is full, its saved points from the records and those in memory are
sealed as one block, which takes the place of the records: first a copy record of
the saved points goes behind both, then the block is written over the records,
then the file is cut behind the block. So a block is the same with or without
Sync, and a crash at any byte of the seal leaves the saved points readable: as
the records, the copy or the block. The split policy is asked at the first point
after a Sync and after the open, otherwise at block boundaries; when it ends the
chunk, the open block is sealed there as a block of its own, as are `DB.Close` and
a compaction that rewrites the open chunk.

### Memory and idle columns

A column holds the unsaved points of its open block, its head, its dictionary and
its list of chunks; read buffers come from a pool shared by all columns. With
`Options.IdleAfter` set, the DB calls `DB.ReleaseIdle` every `IdleAfter`: it
saves every column, into which nothing was written for `IdleAfter` by the clock
`Options.Now`, like a Sync, and frees that memory. A read does not keep a column
from being idle. `IdleAfter` 0, the default, is off: no timer runs and
`ReleaseIdle` does nothing. `ColumnStats.Unsynced` tells the points only the
memory holds.

## Read API

* `ScanI64` / `ScanF64` / `ScanString` — columnar batch reader. Yields decoded
  blocks as parallel slices with (near) zero per-point cost; the path for
  billion-element scans and aggregations.
* `ScanF64` delivers `NA` as `math.NaN()` and every other value as
  `float64(v)/10^Decimals`, bit for bit like before; the divisor is determined
  once per read.
* `IterI64` / `IterF64` / `IterString` — `iter.Seq` of points; composes directly
  with `pkg/timeseries.M4` and `Series`. ~25 % slower than `Scan*` due to the
  per-point delivery, but ergonomic. Use `Scan*` for bulk work. **Limit:** an
  `iter.Seq` has no error, so `Iter*` end silently where `Scan*` returns an error,
  e.g. for data of a later version (see [Compatibility](#compatibility)); a caller,
  which must know, reads with `Scan*`.

Reads merge the finalized chunks, the sealed blocks of the pending chunk, its open
block (records and memory) and the head, in ascending time, each point once; an
entry of the head goes between the points, replaces the point of its time, or
removes it. A read skips chunks by their names and reads from a chunk only the
blocks that overlap its range and a little around them: a chunk, whose blocks take
more than 256 KiB and which begins before the range, is searched for the first of
them by a binary search over the file, a few small windows instead of every block
before it; a block it takes as a waypoint lies within the bounds of the chunk and
has its CRC. The scan stops at the first block after the range, so that a query
reads at most 2 MiB more than its blocks. It holds at most a window of 256 KiB and
a block in memory.

### Opening a column

Different columns open at the same time; the same column opens once, and every
call of `DB.Column` or `DB.LookupColumn` gets the same `*Column` until it is
dropped or the DB is closed. The open lists the directory, reads `schema.json`
and, of the open chunk, the header and what lies behind the end of its sealed
blocks that the newer slot names: one block and the records at most, if a seal
did not update the slot. The other chunks are taken from their names; their
headers, the dictionary and the head are read when a read or a write needs them.

## Compatibility

Old data stays: this version reads and continues the chunks, blocks,
continuation records, dictionaries and WALs of the versions before; it does not
rewrite them. Values written before read faster only once a compaction rewrites
their chunk. **A version before the frame-of-reference blocks reads new blocks
wrongly or not at all, without an error**: do not go back after an update.

From this version on, nago fails loudly on what a later version wrote. When the
magic and format of a chunk match and its version is larger than 2, and the CRC of
the prefix of its header matches or the file is shorter than a header of version 2;
when a block has its marker and CRC, but a flag this version does not know
(`0xE0`); when a continuation record has its marker and CRC, but a kind this
version does not know; when a record of the WAL has its CRC (over its payload), but
a kind this version does not know; or when `schema.json` has a version larger than
1, then every method with an error result, which meets it, fails with an error
(`tsdb: written by a newer version of tsdb`) and removes, cuts, overwrites or
compacts nothing of it:

* `Scan*`, `Count` and `DeleteRange` on a chunk or block of a later version;
* `DB.Column` and `DB.LookupColumn` on an open chunk (also one shorter than a
  header, which is not taken for a torn header and removed), a continuation
  record or a `schema.json` of a later version, `DB.SeriesColumns` on such a
  `schema.json`;
* every method, which loads the head (reads, writes, `Delete`, `DeleteRange`,
  `Flush`), on a record of the WAL of a later version; a torn last record of the
  WAL is still cut off;
* a compaction (`Column.Flush`, `DB.Close` and the background compaction of the
  head, which logs the error) on a chunk or block of a later version, which it
  would rewrite: it stops before it writes a chunk or a WAL. `Column.Flush`
  finalizes the open chunk before, as it always does. A put, which reaches
  `MaxHeadPoints` and compacts at once, succeeds, because its point is in the WAL;
  it logs the error of the compaction, and the next `Flush` or `Close` reports it.

Appending, `Sync` and putting a correction into the WAL do not read old blocks and
go on. The dictionary has no version or kind of its own; its entries are read as
before. A wrong marker or magic, or a CRC, which does not match, is corrupt data
and skipped as before. A skipped block, one with a wrong CRC or one, which does not
decode, is logged once per column and place with `slog.Warn` ("tsdb: damaged block
skipped", attributes `bucket`, `column`, `file` and `offset`), by a read, the open
of a column and a compaction, which logs it before it writes the chunk without it;
a later read of the same place logs nothing, until the DB is opened again. A chunk,
whose header is not one of a chunk, is logged at offset 0. When the open of a column
reads its open chunk after a crash, a frame, which runs over the end, is the torn
rest, and bytes before the records of the open block may be the remains of records
written over; neither is logged, but bytes between two of these records are. Every
later change of the continuation records, the WAL or the dictionary increases the
version of the chunk; new block encodings get a flag of their own.

## Benchmarks

### Reading after the frame-of-reference blocks

Measured on the Apple M1 Max of the benchmarks below on 2026-11-16 08:40 to 08:41
by throwaway programs, twice alternating with nago f4aca80c5 (the version before),
go1.27.3, `Compress: true`, 1-s raster, sums into a local variable, median of 5
after a warm-up run, load 6.7 to 14.6 (Defender, GoLand); both rounds within 2 %.

| 50 M points, power, 1 decimal | before    | now       |
|-------------------------------|-----------|-----------|
| `ScanI64`, counting only      | 254 M/s   | 897 M/s   |
| `ScanI64`, sum                | 236 M/s   | 694 M/s   |
| `ScanF64`, sum                | 153 M/s   | 298 M/s   |
| `PutI64`, appending           | 24.1 M/s  | 28.9 M/s  |
| bytes per point (files)       | 1.888     | 1.488     |

10 M points each, the kinds of data, bytes per point and `ScanI64` counting only
(`ScanString` for the status), before → now: power 1.888 → 1.488 B, 254 → 904 M/s;
power with 0.1 % `NA` 1.904 → 1.490 B; with 1 % `NA` 2.034 → 1.504 B, 239 → 867 M/s;
a counter 1.012 → 0.528 B, 432 → 1 249 M/s; a status (enum, 98.7 % one value) 0.0405
→ 0.0403 B, 62 → 63 M/s (the dictionary is the limit); noise of 64 bits 9.507 → 8.100 B,
72 → 309 M/s; a temperature with 2 decimals 1.160 → 1.052 B; a voltage with 3 decimals
1.978 → 1.704 B, 290 → 869 M/s.

A script in `engine/vm` of gostrict, which gets a callback per block of `ScanF64`
with the times as `list.List[int64]` and the values as `list.List[float64]` and
sums them, reads 61 M points/s instead of 35 M/s (`Writer.Int64s` and
`Writer.Float64s` instead of a value at a time); the loop of the interpreter is the
limit now.

The tables from here on were measured before the frame-of-reference blocks (with
varint values only) and are not measured again; reading is faster now, see above.
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

(`PutI64` measured again with the encoded open block, load 6: 33 ns/op and 17 B/op, d9eaf8db2 on the same machine 23.5 ns/op and 4 B/op.)

| Operation  | ns/op | throughput  | B/op | allocs/op |
|------------|-------|-------------|------|-----------|
| `PutI64`   | 33    | 30 M/s      | 17   | 0         |
| `ScanI64`  | —     | 348 M pts/s | ~1 buffer | 1    |
| `IterI64`  | —     | 217 M pts/s | ~1 buffer | 1    |
| `IterF64`  | —     | 130 M pts/s | ~1 buffer | 2    |

A read over a column with an unflushed pending chunk (data written but not yet
`Flush`-ed) performs within ~1 % of a flushed read (348 vs 348 M pts/s,
`BenchmarkScanI64Pending` vs `BenchmarkScanI64`): the read scans the sealed
blocks of the pending chunk in-place using their already-known end, so
reader-visible unflushed data adds no measurable cost.

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

### Column resolution (`DB.Column` / `LookupColumn`)

Opened columns are cached, so a stateless caller can re-resolve a column on every
operation instead of holding the handle. The cache is a `sync.Map`, so a hit is a
lock-free load with no reader cache-line contention; name/schema validation only
runs when a column is actually created. Resolving a live column
(`BenchmarkColumnAccessParallel`) therefore gets *faster* with more cores instead
of degrading:

| cores | RWMutex (before) | sync.Map (after) |
|-------|------------------|------------------|
| 1     | 54 ns/op         | 52 ns/op         |
| 4     | 93 ns/op         | 15 ns/op         |
| 10    | 123 ns/op        | 13 ns/op         |

Under a `RWMutex` the reader lock's shared state ping-pongs between cores, so
throughput *degraded* with concurrency; the lock-free cache removes that. (The
single remaining allocation per call is the `"bucket/column"` lookup key.)

## Essential findings

* **Constant-memory burst ingest is real.** Inserting 1 billion points grew the
  heap by only **3.6 MiB** and never required manual flushing or a special
  harness. This is the direct result of routing monotonic appends around the
  in-memory index straight into sealed chunk blocks. (Measured before the
  open block was held encoded; that holds less.)
* **Insert throughput is flat with data size** — 28.7 M/s at 1 M and 29.5 M/s at
  1 B points. There is no O(n²) write amplification, because monotonic appends
  never rewrite existing chunks (only out-of-order corrections do).
* **Writes are nearly allocation-free** on the append path (0 allocs/op rounded, ~33 ns/point): the open block
  in memory grows by an eighth at a time, 17 B/op, and is freed when it is sealed or saved.
* **Storage is ~1 byte per point** for oscillating decimal data at a constant
  sample rate (measured with varint values; the frame of reference takes the
  smaller of both per block): the equidistant timestamp mode stores the whole
  per-block time axis as `(start, step)`, so only the scaled-int value delta
  remains. Irregular or holed data falls back to delta-of-delta (~2 bytes/point).
* **Unchanged tracks are nearly free.** A block whose values are all identical is
  stored as one constant, so a flat + equidistant block is a fixed ~46-byte
  frame independent of point count: **~0.01 bytes/point** measured, ~180× smaller
  than oscillating data, and read ~2× faster (a fill loop, no varint decode).
* **Reads sustain ~300–900 M points/s** single-core via the columnar batch API
  for a long series (`ScanF64` with a sum to `ScanI64` counting, see above;
  ~130–260 M/s with varint values), with buffers from a pool shared by all
  columns. The ergonomic `Iter*` API is slower purely due to per-point delivery.
* **Concurrent columns scale.** Different columns share no column lock and use a
  per-path-sharded file pool, so reads/writes on disjoint columns run in
  parallel: writes reach ~4.2× and reads ~3.7× aggregate throughput before
  memory bandwidth (not locking) becomes the ceiling. All cross-column access is
  data-race free.

## Durability model (house rules, shared with `msgstore`)

No `fsync`, except in a compaction; atomic `pread`/`pwrite`; per-block and
per-record CRC + markers for forward-scan recovery; a new file + atomic rename
for the schema and the compaction, the dictionary is only appended to;
single-process `flock`; monotonic ids never reused. A crash of the process stops
the writes after any byte: what was written before it stands.

* The unsaved points of the open block live in memory and are lost on a crash of
  the process. After a `Column.Sync`, the next open finds every point written
  before it; a crash during a Sync costs only the points since the Sync before,
  because a Sync only appends. A correction or a deletion is saved when its put
  returns: it is in the WAL.
* What is written only to the page cache can be lost by a power failure, except
  what a compaction synced.
* `DB.Close` seals the open block of every column as a block of its own and
  folds the head into chunks; a head that reaches into the pending chunk rewrites
  that chunk, and the newest of the rewritten chunks stays open. The background
  compaction of the head works the same way, but leaves an open block, which it
  does not rewrite, open.

The next open continues the pending chunk, whether `DB.Close` or a crash left it:
it is taken up to its last intact block (marker, length and CRC) and the records
of its open block behind it. A record whose points are not newer than that block
belongs to it (the crash came after the block was written over its records); the
newest copy record replaces the records before it (the crash came while the block
was written); the records after it continue it, each newer than the one before. A
torn rest is cut off, and a file holding more than that gets its open block written
anew as one record, in the same order as a seal. Appends go on in the same file
until the split policy finalizes it. No intact point is lost, the torn rest does
not count, and no point is read twice.
Should a column hold more than one pending chunk with intact points, only the
newest is continued and the others are finalized, as is a pending chunk whose
points are not newer than every finalized chunk, so that reads keep their order;
a pending chunk without an intact point is removed.
Retention (`DropBefore`) removes a continued pending chunk like any other, only
as a whole and only when all its points, in memory or on disk, are too old.

### Compaction

A compaction (the background folding of the head, `Column.Flush` and `DB.Close`)
does not lose a point, which was in a chunk, saved or in the WAL when it began,
at any byte of a crash of the process, and after the next open each is read once:

1. It writes the new chunks, the WAL of the head entries it does not fold and
   `compact.log.new`, which lists the chunks it replaces and the files it renames,
   under names ending in `.new`, and syncs them and the directory.
2. It renames `compact.log.new` to `compact.log`: the commit. It syncs the
   directory.
3. It removes the replaced chunks and renames the new ones into place; what was
   written to the old WAL meanwhile goes behind the new one, which is renamed over
   it. Writers of corrections wait only for this step. It syncs the directory.
4. It removes `compact.log`.

Before the commit the next open removes the `.new` files and finds the old state;
after it the open finishes step 3. A power failure during a compaction does not
lose a point that was on the disk before it began. `DB.Close` prepares the
compactions of all columns side by side and runs each step for all of them at
once, so that it waits for the syncs once, not once per column.

A sync of many files (`ndb.FilePool.SyncAll`) syncs them side by side; on macOS,
where Go's `File.Sync` is `F_FULLFSYNC`, each one is an `fsync(2)` and one
`F_FULLFSYNC` at the end flushes the cache of the drive for all of them. The
price: a compaction needs the space of the chunks it rewrites twice and waits for
three such syncs.
