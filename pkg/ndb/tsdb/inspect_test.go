package tsdb

import "testing"

func TestColumnStats(t *testing.T) {
	db, err := Open(t.TempDir(), Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	c, err := db.Column("b", "c", Schema{Scheme: SchemeDecimal, Decimals: 2})
	if err != nil {
		t.Fatal(err)
	}

	// empty column
	if st := c.Stats(); st.HasData {
		t.Fatalf("empty column should have no data: %+v", st)
	}

	base := int64(1_700_000_000_000)
	for i := int64(0); i < 10000; i++ {
		if err := c.PutF64(base+i*20, float64(i%7)); err != nil {
			t.Fatal(err)
		}
	}
	if err := c.Flush(); err != nil {
		t.Fatal(err)
	}

	st := c.Stats()
	if !st.HasData {
		t.Fatal("expected data")
	}
	if st.Scheme != SchemeDecimal || st.Decimals != 2 {
		t.Fatalf("schema wrong: %+v", st)
	}
	if st.MinMillis != base {
		t.Fatalf("min = %d, want %d", st.MinMillis, base)
	}
	wantMax := base + 9999*20
	if st.MaxMillis != wantMax {
		t.Fatalf("max = %d, want %d", st.MaxMillis, wantMax)
	}
	if st.Chunks < 1 || st.Bytes <= 0 {
		t.Fatalf("chunks/bytes wrong: %+v", st)
	}
}

func TestColumnCount(t *testing.T) {
	db, err := Open(t.TempDir(), Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	c, _ := db.Column("b", "c", Schema{Scheme: SchemeDecimal, Decimals: 0})

	if n, err := c.Count(); err != nil || n != 0 {
		t.Fatalf("empty count = %d (err %v), want 0", n, err)
	}

	base := int64(1_700_000_000_000)
	for i := int64(0); i < 1234; i++ {
		if err := c.PutI64(base+i*20, i); err != nil {
			t.Fatal(err)
		}
	}
	c.Flush()
	// overwrite one point (must not double-count) and delete one (must drop)
	c.PutI64(base+100*20, 999)
	c.Delete(base + 200*20)
	c.Flush()

	n, err := c.Count()
	if err != nil {
		t.Fatal(err)
	}
	if n != 1233 {
		t.Fatalf("count = %d, want 1233 (1234 - 1 deleted)", n)
	}
}
