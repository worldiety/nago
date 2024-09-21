// Package sdb contains a naive and simple key-value database implementation.
// The NIH syndrom is kicking in, again:
//   - etcd/bbolt is slow as hell (100-200 tps on modern NVME storage) due to single file and massive fsync-based design. Turning fsync off, corrupts db entirely. I cannot reach per machine-scale with that (1000 instances per cloud server, hell not even a single hako instance)
//   - badger has a bad reputation, takes a huge amount of RAM and insanely amount of vlog storage (e.g. 2gib for a single entry), also looses unclosed data all the way. (1000 instances per cloud server becomes impossible, thats 2 tib for nothing)
//   - pebble, lmdb and others require cgo, which is unavailable and due to security concerns (e.g. random code execution at build time) unwanted at our hosting platform
//   - sqlite has massive concurrency scaling issues, a blocking user transaction will halt everything
//   - we don't need real transactions, however
//   - we need as less as possible locking and maximum read and write throughput
//   - we expect only a few hundred to thousand entries per bucket, but huge amount of read/writes due to our renderer design.
//
// So what crazy assumptions can we make?
//   - Appending a simple log and just reading into memory seems viable for a "long" time
//   - stupidly appending a log is insanely fast (like 500k tps writes). And at this point we have already beaten badger by 5x AND we have been persistent.
//   - writing an optimized index from time to time and truncating the log keeps performance up, just need a constant time fsync for that
//   - just keeping the index in memory is not a problem, e.g. holding 1 mio UUID with file offsets is what? 64mib?
package sdb
