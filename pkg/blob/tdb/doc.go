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
//   - pwrite and pread is atomic within a single process and safe enough on process crashes.
//   - pread requests are cached by the OS anyway and syscall costs are acceptable
//   - Machine crashes do not occur in practice on virtual cloud servers. For a bare metal restore, we have backups anyway.
package tdb
