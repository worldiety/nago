// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tsdb

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"go.wdy.de/nago/pkg/ndb"
)

// chunk file header (9 bytes), matching the house style of msgstore:
//
//	| Magic        4 | 0x4E41474F "NAGO"
//	| Format Magic 4 | 0x4E545342 "NTSB" Nago Time Series Block
//	| Version      1 | 0x01
const (
	chunkMagic     uint32 = 0x4E41474F // "NAGO"
	chunkFormat    uint32 = 0x4E545342 // "NTSB"
	chunkVersion   byte   = 1
	chunkHeaderLen        = 9
	chunkExt              = ".tsb"
)

// chunkInfo describes a chunk file on disk. A finalized chunk is named
// "<minMillis>_<maxMillis>.tsb"; a pending (open) chunk is "<minMillis>_.tsb".
// Filenames encode the contained time range so range queries skip whole files
// and retention drops whole files, without opening them — the timezone-free
// analog of msgstore's seq-range filenames.
type chunkInfo struct {
	path      string
	minMillis int64
	maxMillis int64 // for pending chunks this is the last known max (may grow)
	pending   bool
	sizeBytes int64 // total file size; cached for finalized (immutable) chunks, 0 = unknown
}

func pendingChunkName(minMillis int64) string {
	return strconv.FormatInt(minMillis, 10) + "_" + chunkExt
}

func finalChunkName(minMillis, maxMillis int64) string {
	return strconv.FormatInt(minMillis, 10) + "_" + strconv.FormatInt(maxMillis, 10) + chunkExt
}

// parseChunkName parses a chunk filename into its time bounds. ok is false for
// files that are not chunk files.
func parseChunkName(name string) (info chunkInfo, ok bool) {
	if !strings.HasSuffix(name, chunkExt) {
		return chunkInfo{}, false
	}
	base := strings.TrimSuffix(name, chunkExt)
	us := strings.IndexByte(base, '_')
	if us < 0 {
		return chunkInfo{}, false
	}
	minStr, maxStr := base[:us], base[us+1:]
	minV, err := strconv.ParseInt(minStr, 10, 64)
	if err != nil {
		return chunkInfo{}, false
	}
	if maxStr == "" {
		return chunkInfo{minMillis: minV, maxMillis: minV, pending: true}, true
	}
	maxV, err := strconv.ParseInt(maxStr, 10, 64)
	if err != nil {
		return chunkInfo{}, false
	}
	return chunkInfo{minMillis: minV, maxMillis: maxV}, true
}

// listChunks returns all chunk files in dir sorted ascending by minMillis. At
// most one pending chunk is expected; if several exist (crash mid-finalize),
// all are returned and the caller reconciles.
func listChunks(dir string) ([]chunkInfo, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("tsdb: list chunks: %w", err)
	}
	var out []chunkInfo
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, ok := parseChunkName(e.Name())
		if !ok {
			continue
		}
		info.path = filepath.Join(dir, e.Name())
		if fi, err := e.Info(); err == nil {
			info.sizeBytes = fi.Size()
		}
		out = append(out, info)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].minMillis != out[j].minMillis {
			return out[i].minMillis < out[j].minMillis
		}
		return out[i].maxMillis < out[j].maxMillis
	})
	return out, nil
}

// writeChunkHeader writes the 9-byte header at offset 0 via the pool.
func writeChunkHeader(pool *ndb.FilePool, path string) error {
	var hdr [chunkHeaderLen]byte
	binary.BigEndian.PutUint32(hdr[0:], chunkMagic)
	binary.BigEndian.PutUint32(hdr[4:], chunkFormat)
	hdr[8] = chunkVersion
	if _, err := pool.WriteAt(path, hdr[:], 0); err != nil {
		return fmt.Errorf("tsdb: write chunk header: %w", err)
	}
	return nil
}

// validateChunkHeader reads and checks the header of an existing chunk file.
func validateChunkHeader(pool *ndb.FilePool, path string) error {
	var hdr [chunkHeaderLen]byte
	if _, err := pool.ReadAt(path, hdr[:], 0); err != nil {
		return fmt.Errorf("tsdb: read chunk header: %w", err)
	}
	if binary.BigEndian.Uint32(hdr[0:]) != chunkMagic ||
		binary.BigEndian.Uint32(hdr[4:]) != chunkFormat ||
		hdr[8] != chunkVersion {
		return errCorruptChunk
	}
	return nil
}

// scanChunkBlocks yields every valid block header in the chunk file at path,
// forward-scanning past corruption. It reads the whole file into memory once;
// chunk files are bounded by the split policy (default 64 MiB). The callback
// receives the raw frame slice and its parsed header; returning false stops.
func scanChunkBlocks(pool *ndb.FilePool, path string, fn func(frame []byte, h blockHeader) bool) error {
	var scratch []byte
	return scanChunkBlocksBuf(pool, path, 0, &scratch, fn)
}

// scanChunkBlocksBuf is scanChunkBlocks with a caller-owned reusable read
// buffer. The buffer is grown as needed and retained across calls, so a full
// read over many chunks allocates the file buffer at most once (grown to the
// largest chunk) rather than once per chunk. The frame passed to fn aliases
// *bufp and is only valid during the callback.
//
// sizeHint is the known file size; when > 0 (finalized, immutable chunks) it is
// used directly and the per-scan Fstat syscall is skipped. Pass 0 for pending
// chunks whose size may still be growing.
func scanChunkBlocksBuf(pool *ndb.FilePool, path string, sizeHint int64, bufp *[]byte, fn func(frame []byte, h blockHeader) bool) error {
	size := sizeHint
	if size <= 0 {
		fi, err := pool.Stat(path)
		if err != nil {
			return fmt.Errorf("tsdb: stat chunk: %w", err)
		}
		size = fi.Size()
	}
	if size <= chunkHeaderLen {
		return nil
	}
	need := int(size - chunkHeaderLen)
	if cap(*bufp) < need {
		*bufp = make([]byte, need)
	}
	buf := (*bufp)[:need]
	if _, err := pool.ReadAt(path, buf, chunkHeaderLen); err != nil {
		return fmt.Errorf("tsdb: read chunk body: %w", err)
	}

	off := 0
	for off < len(buf) {
		h, perr := parseBlockHeader(buf[off:])
		if perr != nil {
			// forward-scan to the next sync marker
			next := findNextSync(buf, off+1)
			if next < 0 {
				break
			}
			off = next
			continue
		}
		if !fn(buf[off:off+h.frameLen], h) {
			return nil
		}
		off += h.frameLen
	}
	return nil
}

// findNextSync returns the index >= from of the next sync marker in buf, or -1.
func findNextSync(buf []byte, from int) int {
	if from < 0 {
		from = 0
	}
	for i := from; i+8 <= len(buf); i++ {
		if binary.LittleEndian.Uint64(buf[i:]) == syncMarker {
			return i
		}
	}
	return -1
}

// finalizeOrphanPending finalizes a pending chunk left over from a prior crash:
// it scans the intact blocks to find the true min/max timestamps (skipping any
// torn tail block via the CRC/sync-marker scan), then renames the file to its
// final <min>_<max>.tsb name. ok is false if the pending chunk contains no
// readable block (it is then skipped and left on disk).
func finalizeOrphanPending(pool *ndb.FilePool, ci chunkInfo) (chunkInfo, bool) {
	var minTs, maxTs int64
	var have bool
	var lastGoodEnd int64 = chunkHeaderLen
	_ = scanChunkBlocks(pool, ci.path, func(frame []byte, h blockHeader) bool {
		if !have || h.minMillis < minTs {
			minTs = h.minMillis
		}
		if !have || h.maxMillis > maxTs {
			maxTs = h.maxMillis
		}
		have = true
		lastGoodEnd += int64(h.frameLen)
		return true
	})
	if !have {
		return chunkInfo{}, false
	}
	finalPath := filepath.Join(filepath.Dir(ci.path), finalChunkName(minTs, maxTs))
	pool.Evict(ci.path)
	if err := os.Rename(ci.path, finalPath); err != nil {
		return chunkInfo{}, false
	}
	return chunkInfo{path: finalPath, minMillis: minTs, maxMillis: maxTs, sizeBytes: lastGoodEnd}, true
}
