package msgstore_test

import (
	"fmt"
	"math"
	"sync"
	"testing"

	"go.wdy.de/nago/pkg/ndb/msgstore"
)

// ---------- Write Benchmarks ----------

// BenchmarkAppend measures raw single-type append throughput at various payload
// sizes. Reports ns/op, B/op, allocs/op, MB/s and msg/s.
func BenchmarkAppend(b *testing.B) {
	for _, size := range []int{0, 64, 256, 1024, 4096, 16384} {
		b.Run(fmt.Sprintf("payload=%d", size), func(b *testing.B) {
			dir := b.TempDir()
			db, err := msgstore.Open(dir, msgstore.Options{
				Compress: msgstore.NoCompression,
			})
			if err != nil {
				b.Fatal(err)
			}
			defer db.Close()

			var traceID [16]byte
			payload := make([]byte, size)
			for i := range payload {
				payload[i] = byte(i)
			}

			b.SetBytes(int64(size))
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				if _, err := db.Append(1, traceID, payload); err != nil {
					b.Fatal(err)
				}
			}

			b.StopTimer()
			elapsed := b.Elapsed()
			if elapsed > 0 {
				msgPerSec := float64(b.N) / elapsed.Seconds()
				b.ReportMetric(msgPerSec, "msg/s")
			}
		})
	}
}

// BenchmarkAppendS2 measures append throughput with S2 compression enabled.
func BenchmarkAppendS2(b *testing.B) {
	for _, size := range []int{256, 1024, 4096, 16384} {
		b.Run(fmt.Sprintf("payload=%d", size), func(b *testing.B) {
			dir := b.TempDir()
			db, err := msgstore.Open(dir, msgstore.Options{
				Compress: msgstore.AlwaysS2,
			})
			if err != nil {
				b.Fatal(err)
			}
			defer db.Close()

			var traceID [16]byte
			payload := make([]byte, size)
			for i := range payload {
				payload[i] = byte(i % 64)
			}

			b.SetBytes(int64(size))
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				if _, err := db.Append(1, traceID, payload); err != nil {
					b.Fatal(err)
				}
			}

			b.StopTimer()
			elapsed := b.Elapsed()
			if elapsed > 0 {
				msgPerSec := float64(b.N) / elapsed.Seconds()
				b.ReportMetric(msgPerSec, "msg/s")
			}
		})
	}
}

// BenchmarkAppendMultiType measures append when events are spread across
// multiple event types (tests type-state map lookup overhead).
func BenchmarkAppendMultiType(b *testing.B) {
	for _, nTypes := range []int{1, 10, 100} {
		b.Run(fmt.Sprintf("types=%d", nTypes), func(b *testing.B) {
			dir := b.TempDir()
			db, err := msgstore.Open(dir, msgstore.Options{
				Compress: msgstore.NoCompression,
			})
			if err != nil {
				b.Fatal(err)
			}
			defer db.Close()

			var traceID [16]byte
			payload := []byte("benchmark-event-payload")

			b.SetBytes(int64(len(payload)))
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				typeID := msgstore.TypeID(i % nTypes)
				if _, err := db.Append(typeID, traceID, payload); err != nil {
					b.Fatal(err)
				}
			}

			b.StopTimer()
			elapsed := b.Elapsed()
			if elapsed > 0 {
				msgPerSec := float64(b.N) / elapsed.Seconds()
				b.ReportMetric(msgPerSec, "msg/s")
			}
		})
	}
}

// ---------- Read Benchmarks ----------

// BenchmarkReplay measures sequential read throughput at various payload sizes.
func BenchmarkReplay(b *testing.B) {
	for _, size := range []int{0, 64, 256, 1024, 4096} {
		b.Run(fmt.Sprintf("payload=%d", size), func(b *testing.B) {
			const count = 50_000

			dir := b.TempDir()
			db, err := msgstore.Open(dir, msgstore.Options{
				Compress: msgstore.NoCompression,
			})
			if err != nil {
				b.Fatal(err)
			}
			defer db.Close()

			var traceID [16]byte
			payload := make([]byte, size)
			for i := range payload {
				payload[i] = byte(i)
			}

			for i := 0; i < count; i++ {
				if _, err := db.Append(1, traceID, payload); err != nil {
					b.Fatal(err)
				}
			}

			b.SetBytes(int64(size) * count)
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				n := 0
				for _, msg := range db.Replay([]msgstore.TypeID{1}, 1, math.MaxUint64) {
					_ = msg
					n++
				}
				if n != count {
					b.Fatalf("expected %d messages, got %d", count, n)
				}
			}

			b.StopTimer()
			elapsed := b.Elapsed()
			if elapsed > 0 {
				totalMessages := float64(b.N) * float64(count)
				msgPerSec := totalMessages / elapsed.Seconds()
				b.ReportMetric(msgPerSec, "msg/s")
			}
		})
	}
}

// BenchmarkReplayMultiType measures replay throughput with k-way merge across
// multiple event types.
func BenchmarkReplayMultiType(b *testing.B) {
	for _, nTypes := range []int{1, 5, 20} {
		b.Run(fmt.Sprintf("types=%d", nTypes), func(b *testing.B) {
			const totalEvents = 50_000

			dir := b.TempDir()
			db, err := msgstore.Open(dir, msgstore.Options{
				Compress: msgstore.NoCompression,
			})
			if err != nil {
				b.Fatal(err)
			}
			defer db.Close()

			var traceID [16]byte
			payload := []byte("benchmark-event-payload")

			for i := 0; i < totalEvents; i++ {
				typeID := msgstore.TypeID(i % nTypes)
				if _, err := db.Append(typeID, traceID, payload); err != nil {
					b.Fatal(err)
				}
			}

			b.SetBytes(int64(len(payload)) * totalEvents)
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				n := 0
				for _, msg := range db.Replay(nil, 1, math.MaxUint64) {
					_ = msg
					n++
				}
				if n != totalEvents {
					b.Fatalf("expected %d messages, got %d", totalEvents, n)
				}
			}

			b.StopTimer()
			elapsed := b.Elapsed()
			if elapsed > 0 {
				totalMessages := float64(b.N) * float64(totalEvents)
				msgPerSec := totalMessages / elapsed.Seconds()
				b.ReportMetric(msgPerSec, "msg/s")
			}
		})
	}
}

// BenchmarkReplayS2 measures replay throughput with S2 compressed data.
func BenchmarkReplayS2(b *testing.B) {
	const count = 50_000
	const size = 1024

	dir := b.TempDir()
	db, err := msgstore.Open(dir, msgstore.Options{
		Compress: msgstore.AlwaysS2,
	})
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	var traceID [16]byte
	payload := make([]byte, size)
	for i := range payload {
		payload[i] = byte(i % 64)
	}

	for i := 0; i < count; i++ {
		if _, err := db.Append(1, traceID, payload); err != nil {
			b.Fatal(err)
		}
	}

	b.SetBytes(int64(size) * count)
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		n := 0
		for _, msg := range db.Replay([]msgstore.TypeID{1}, 1, math.MaxUint64) {
			_ = msg
			n++
		}
		if n != count {
			b.Fatalf("expected %d messages, got %d", count, n)
		}
	}

	b.StopTimer()
	elapsed := b.Elapsed()
	if elapsed > 0 {
		totalMessages := float64(b.N) * float64(count)
		msgPerSec := totalMessages / elapsed.Seconds()
		b.ReportMetric(msgPerSec, "msg/s")
	}
}

// ---------- Marshal / Unmarshal micro-benchmarks ----------

// BenchmarkMarshalMessage measures raw message serialization speed (no I/O).
func BenchmarkMarshalMessage(b *testing.B) {
	for _, size := range []int{0, 64, 1024, 4096} {
		b.Run(fmt.Sprintf("payload=%d", size), func(b *testing.B) {
			payload := make([]byte, size)
			msg := msgstore.Message{
				SequenceID:      42,
				Timestamp:       1234567890,
				Encoding:        msgstore.EncodingRaw,
				PayloadLen:      uint32(size),
				UncompressedLen: uint32(size),
				Payload:         payload,
			}

			var buf []byte
			b.SetBytes(int64(size))
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				buf = msg.MarshalInto(buf)
			}

			b.StopTimer()
			elapsed := b.Elapsed()
			if elapsed > 0 {
				msgPerSec := float64(b.N) / elapsed.Seconds()
				b.ReportMetric(msgPerSec, "msg/s")
			}
		})
	}
}

// BenchmarkUnmarshalMessage measures raw message deserialization speed (no I/O).
func BenchmarkUnmarshalMessage(b *testing.B) {
	for _, size := range []int{0, 64, 1024, 4096} {
		b.Run(fmt.Sprintf("payload=%d", size), func(b *testing.B) {
			payload := make([]byte, size)
			msg := msgstore.Message{
				SequenceID:      42,
				Timestamp:       1234567890,
				Encoding:        msgstore.EncodingRaw,
				PayloadLen:      uint32(size),
				UncompressedLen: uint32(size),
				Payload:         payload,
			}
			data := msg.MarshalBinary()

			b.SetBytes(int64(size))
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_, _, err := msgstore.UnmarshalMessageNoCopy(data, 16<<20)
				if err != nil {
					b.Fatal(err)
				}
			}

			b.StopTimer()
			elapsed := b.Elapsed()
			if elapsed > 0 {
				msgPerSec := float64(b.N) / elapsed.Seconds()
				b.ReportMetric(msgPerSec, "msg/s")
			}
		})
	}
}

// BenchmarkAppendConcurrent measures total write throughput when exactly N
// goroutines append to N different event types simultaneously (64-byte payload).
// Each goroutine writes exclusively to its own TypeID so the per-type mutex
// does not cause cross-type contention. The only shared serialization point
// is the time-index mutex (one 16-byte pwrite per message).
func BenchmarkAppendConcurrent(b *testing.B) {
	for _, nWriters := range []int{1, 2, 4, 8} {
		b.Run(fmt.Sprintf("writers=%d", nWriters), func(b *testing.B) {
			dir := b.TempDir()
			db, err := msgstore.Open(dir, msgstore.Options{
				Compress: msgstore.NoCompression,
			})
			if err != nil {
				b.Fatal(err)
			}
			defer db.Close()

			payload := make([]byte, 64)
			for i := range payload {
				payload[i] = byte(i)
			}

			// pre-create type states so lazy-init overhead doesn't skew results
			var traceID [16]byte
			for t := range nWriters {
				if _, err := db.Append(msgstore.TypeID(t+1), traceID, payload); err != nil {
					b.Fatal(err)
				}
			}

			// each goroutine gets b.N / nWriters messages
			perWriter := b.N / nWriters
			if perWriter == 0 {
				perWriter = 1
			}

			b.SetBytes(int64(len(payload)))
			b.ResetTimer()
			b.ReportAllocs()

			var wg sync.WaitGroup
			wg.Add(nWriters)
			for w := range nWriters {
				go func(typeID msgstore.TypeID) {
					defer wg.Done()
					var tid [16]byte
					for range perWriter {
						if _, err := db.Append(typeID, tid, payload); err != nil {
							b.Error(err)
							return
						}
					}
				}(msgstore.TypeID(w + 1))
			}
			wg.Wait()

			b.StopTimer()
			elapsed := b.Elapsed()
			if elapsed > 0 {
				totalMessages := float64(perWriter * nWriters)
				msgPerSec := totalMessages / elapsed.Seconds()
				b.ReportMetric(msgPerSec, "msg/s")
			}
		})
	}
}
