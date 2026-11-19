// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package tdb contains a naive and simple key-value database implementation (Torbens Database).
// The NIH syndrom is kicking in, again:
//   - etcd/bbolt is slow as hell (100-200 tps on modern NVME storage) due to single file and massive fsync-based design. Turning fsync off, corrupts db entirely. I cannot reach per machine-scale with that (1000 instances per cloud server, hell not even a single hako instance)
//   - badger has a bad reputation, takes a huge amount of RAM and insanely amount of vlog storage (e.g. 2gib for a single entry), also looses unclosed data all the way. (1000 instances per cloud server becomes impossible, thats 2 tib for nothing)
//   - lmdb and others require cgo, which is unavailable and due to security concerns (e.g. random code execution at build time) unwanted at our hosting platform
//   - pebble has a cgo-free target, however it is still slow (10.000 inserts/sec at best on my machine)
//   - sqlite has massive concurrency scaling issues, a blocking user transaction will halt everything, non-cgo-version is an untrusted big-ball-of-mud mess
//   - we don't need real transactions, however
//   - we need as less as possible locking and maximum read and write throughput
//   - we expect only a few hundred to thousand entries per bucket, but huge amount of read/writes due to our renderer design.
//
// So what crazy assumptions can we make?
//   - Appending a simple log and just reading into memory seems viable for a "long" time
//   - stupidly appending a log is insanely fast (like 500k tps writes). And at this point we have already beaten badger by 5x AND we have been persistent.
//   - writing an optimized index from time to time and truncating the log keeps performance up, just need a constant time fsync for that
//   - just keeping the index in memory is not a problem, e.g. holding 1 mio UUID with file offsets is what? 64mib?
//     It is about 70 bytes per entry with a key of 36 bytes: an entry of 32 bytes in nearly full B-tree nodes and
//     the key in a large block, see buildTree and keyArena.
//   - pwrite and pread is atomic within a single process and safe enough on process crashes.
//   - pread requests are cached by the OS anyway and syscall costs are acceptable
//   - Machine crashes do not occur in practice on virtual cloud servers. For a bare metal restore, we have backups anyway.
//
// The files of a directory:
//   - data.tdb holds the entries of the last compaction, all with the transaction of that compaction.
//   - tdb.wal appends every change since. A load skips its records up to the transaction of data.tdb.
//   - tdb.lock is held by the process, which has opened the directory. It also locks data.tdb and tdb.wal, as former
//     versions did, which do not know tdb.lock.
//
// A compaction runs in the background, when the garbage of the files reaches the live data, like GOGC=100, see
// [Options]. The garbage are the records of overwritten and removed entries and the removals themselves. It writes a
// new data.tdb from a snapshot of the index and replaces the former one, which is consistent with the former log. It
// copies the log behind the snapshot into a new log and replaces the former log with it, holding the write lock only
// for the last part. Readers and iterators keep the former files open, until they are done. A crash in between leaves
// a consistent state or one, which the next Open completes.
//
// A crash may leave an incomplete record at the end of the log, which the next Open removes. It never removes valid
// records behind it: former versions have appended behind such a record, which Open then skips, and anything else
// fails the Open as corruption.
//
// When a few entries change often, they stay in the log and leave most of the garbage there. Then, a compaction
// rewrites the log alone with its live records and removals, and keeps data.tdb with the entries, which rarely change.
// Thus, it does not rewrite the whole database each time.
package tdb
