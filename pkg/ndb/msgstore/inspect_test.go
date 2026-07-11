package msgstore

import (
	"testing"

	"go.wdy.de/nago/pkg/ndb"
)

func TestTypesAndTypeStat(t *testing.T) {
	db, err := Open(t.TempDir(), Options{Compress: NoCompression, ShouldSplit: SplitByCount(3)})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var tr ndb.TraceID
	for i := 0; i < 10; i++ {
		if _, err := db.Append("orders", tr, []byte("x")); err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 2; i++ {
		if _, err := db.Append("users", tr, []byte("y")); err != nil {
			t.Fatal(err)
		}
	}

	types, err := db.Types()
	if err != nil {
		t.Fatal(err)
	}
	if len(types) != 2 || types[0] != "orders" || types[1] != "users" {
		t.Fatalf("types = %v", types)
	}

	os, err := db.TypeStat("orders")
	if err != nil {
		t.Fatal(err)
	}
	if os.Type != "orders" || os.Segments < 1 || os.Bytes <= 0 {
		t.Fatalf("orders stat = %+v", os)
	}
	// seqs are global and start at 0; 10 appends -> seqs 0..9, then a pending
	// segment named by the next min seq (10). MinSeq must be 0, and a pending
	// segment must be reported.
	if os.MinSeq != 0 {
		t.Fatalf("orders MinSeq want 0: %+v", os)
	}
	if !os.HasPending {
		t.Fatalf("orders should have a pending segment: %+v", os)
	}
	if os.MaxSeq < 9 {
		t.Fatalf("orders MaxSeq want >= 9: %+v", os)
	}

	empty, err := db.TypeStat("missing")
	if err != nil {
		t.Fatal(err)
	}
	if empty.Segments != 0 || empty.Bytes != 0 {
		t.Fatalf("missing type should be zero: %+v", empty)
	}
}

func TestCountType(t *testing.T) {
	db, err := Open(t.TempDir(), Options{Compress: NoCompression, ShouldSplit: SplitByCount(3)})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var tr ndb.TraceID

	// missing type -> 0
	if n, err := db.CountType("missing"); err != nil || n != 0 {
		t.Fatalf("missing count = %d (err %v), want 0", n, err)
	}

	seqs := make([]ndb.Seq, 0, 10)
	for i := 0; i < 10; i++ {
		s, err := db.Append("orders", tr, []byte("x"))
		if err != nil {
			t.Fatal(err)
		}
		seqs = append(seqs, s)
	}

	if n, err := db.CountType("orders"); err != nil || n != 10 {
		t.Fatalf("count = %d (err %v), want 10", n, err)
	}

	// tombstone one -> excluded from the count
	if err := db.DeleteSeq("orders", seqs[4]); err != nil {
		t.Fatal(err)
	}
	if n, err := db.CountType("orders"); err != nil || n != 9 {
		t.Fatalf("count after delete = %d (err %v), want 9", n, err)
	}
}
