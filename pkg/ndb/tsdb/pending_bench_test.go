package tsdb

import "testing"

// benchColumnNoFlush fills a column but does NOT flush, so most data sits in the
// pending chunk file (sealed blocks not yet finalized) plus the in-memory buffer.
func benchColumnNoFlush(b *testing.B, n int64) *Column {
	b.Helper()
	db, err := Open(b.TempDir(), Options{Compress: false})
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() { db.Close() })
	c, _ := db.Column("bench", "sig", Schema{Scheme: SchemeDecimal, Decimals: 2})
	base := int64(1_700_000_000_000)
	for i := int64(0); i < n; i++ {
		c.PutI64(base+i*20, 2000+(i%13))
	}
	return c // no Flush: pending chunk present
}

func BenchmarkScanI64Flushed(b *testing.B) {
	const n = 1_000_000
	c := benchColumn(b, n) // flushed -> finalized chunks only
	b.ResetTimer()
	var sink int64
	for i := 0; i < b.N; i++ {
		_ = c.ScanI64(minInt64, maxInt64, func(ts, vals []int64) bool {
			for _, v := range vals {
				sink += v
			}
			return true
		})
	}
	b.StopTimer()
	b.ReportMetric(float64(n*int64(b.N))/b.Elapsed().Seconds(), "points/s")
	_ = sink
}

func BenchmarkScanI64Pending(b *testing.B) {
	const n = 1_000_000
	c := benchColumnNoFlush(b, n) // unflushed -> pending chunk scanned each read
	b.ResetTimer()
	var sink int64
	for i := 0; i < b.N; i++ {
		_ = c.ScanI64(minInt64, maxInt64, func(ts, vals []int64) bool {
			for _, v := range vals {
				sink += v
			}
			return true
		})
	}
	b.StopTimer()
	b.ReportMetric(float64(n*int64(b.N))/b.Elapsed().Seconds(), "points/s")
	_ = sink
}
