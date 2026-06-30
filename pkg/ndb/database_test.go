package ndb_test

import (
	"math"
	"slices"
	"testing"

	"github.com/worldiety/option"
	"go.wdy.de/nago/pkg/ndb"
	"go.wdy.de/nago/pkg/ndb/msgstore" // registers the "msgstore" engine via init
)

func TestEngineMessages(t *testing.T) {
	root := t.TempDir()

	db := option.Must(ndb.Open(root, ndb.Options{DefaultKind: "msgstore"}))
	defer func() { option.MustZero(db.Close()) }()

	eng, err := db.Engine("events", ndb.EngineOptions{})
	if err != nil {
		t.Fatalf("open engine: %v", err)
	}

	if eng.Kind() != "msgstore" {
		t.Fatalf("expected kind msgstore, got %q", eng.Kind())
	}
	if eng.Name() != "events" {
		t.Fatalf("expected name events, got %q", eng.Name())
	}

	me, ok := eng.(ndb.MessageEngine)
	if !ok {
		t.Fatal("expected a message engine capability")
	}
	if _, ok := eng.(ndb.BlobEngine); ok {
		t.Fatal("msgstore engine should not expose blob stores")
	}

	m := me.Messages()

	const typeID ndb.TypeID = 7
	var trace [16]byte

	seq, err := m.Append(typeID, trace, []byte("hello"))
	if err != nil {
		t.Fatalf("append: %v", err)
	}
	if seq != 1 {
		t.Fatalf("expected seq 1, got %d", seq)
	}

	var got []byte
	for tid, msg := range m.Replay([]ndb.TypeID{typeID}, 1, math.MaxUint64) {
		if tid != typeID {
			t.Fatalf("replay type: got %d want %d", tid, typeID)
		}
		if msg.Type != typeID {
			t.Fatalf("message Type field not populated: got %d", msg.Type)
		}
		got = slices.Clone(msg.Payload)
	}
	if string(got) != "hello" {
		t.Fatalf("replay payload: got %q want %q", got, "hello")
	}
}

func TestEngineCachedIdentity(t *testing.T) {
	root := t.TempDir()
	db := option.Must(ndb.Open(root, ndb.Options{}))
	defer func() { option.MustZero(db.Close()) }()

	a, err := db.Engine("events", ndb.EngineOptions{})
	if err != nil {
		t.Fatalf("open a: %v", err)
	}
	b, err := db.Engine("events", ndb.EngineOptions{})
	if err != nil {
		t.Fatalf("open b: %v", err)
	}
	if a != b {
		t.Fatal("expected the same cached engine instance for the same name")
	}
}

func TestEngineDefaultKindAndReopen(t *testing.T) {
	root := t.TempDir()

	// No explicit kind: the sole registered engine (msgstore) is used.
	db := option.Must(ndb.Open(root, ndb.Options{}))

	eng, err := db.Engine("retained", ndb.EngineOptions{})
	if err != nil {
		t.Fatalf("open engine: %v", err)
	}
	m := eng.(ndb.MessageEngine).Messages()

	const typeID ndb.TypeID = 3
	var trace [16]byte
	if _, err := m.Put(typeID, trace, []byte("v1")); err != nil {
		t.Fatalf("put: %v", err)
	}
	option.MustZero(db.Close())

	// Reopen: the instance must come back via the recorded engine marker, and
	// the retained value must still be there.
	db2 := option.Must(ndb.Open(root, ndb.Options{}))
	defer func() { option.MustZero(db2.Close()) }()

	var names []string
	for info, err := range db2.Engines() {
		if err != nil {
			t.Fatalf("engines: %v", err)
		}
		if info.Name == "retained" && info.Kind != "msgstore" {
			t.Fatalf("expected kind msgstore for 'retained', got %q", info.Kind)
		}
		names = append(names, info.Name)
	}
	if !slices.Contains(names, "retained") {
		t.Fatalf("expected engine 'retained' to persist, got %v", names)
	}

	eng2, err := db2.Engine("retained", ndb.EngineOptions{})
	if err != nil {
		t.Fatalf("reopen engine: %v", err)
	}
	if eng2.Kind() != "msgstore" {
		t.Fatalf("reopened kind: got %q", eng2.Kind())
	}

	cur := option.Must(eng2.(ndb.MessageEngine).Messages().Get(typeID))
	if cur.IsNone() {
		t.Fatal("expected retained value after reopen")
	}
	if string(cur.Unwrap().Payload) != "v1" {
		t.Fatalf("retained payload: got %q", cur.Unwrap().Payload)
	}
}

func TestEngineMultipleInstancesSameKind(t *testing.T) {
	root := t.TempDir()
	db := option.Must(ndb.Open(root, ndb.Options{DefaultKind: "msgstore"}))
	defer func() { option.MustZero(db.Close()) }()

	audit, err := db.Engine("audit", ndb.EngineOptions{})
	if err != nil {
		t.Fatalf("open audit: %v", err)
	}
	telemetry, err := db.Engine("telemetry", ndb.EngineOptions{})
	if err != nil {
		t.Fatalf("open telemetry: %v", err)
	}
	if audit == telemetry {
		t.Fatal("two differently named engines must be distinct instances")
	}

	const typeID ndb.TypeID = 1
	var trace [16]byte

	am := audit.(ndb.MessageEngine).Messages()
	tm := telemetry.(ndb.MessageEngine).Messages()

	if _, err := am.Append(typeID, trace, []byte("a")); err != nil {
		t.Fatalf("append audit: %v", err)
	}
	if _, err := tm.Append(typeID, trace, []byte("b")); err != nil {
		t.Fatalf("append telemetry: %v", err)
	}

	// Each instance keeps its own, independent stream.
	auditVal := option.Must(am.Get(typeID)).Unwrap()
	telVal := option.Must(tm.Get(typeID)).Unwrap()
	if string(auditVal.Payload) != "a" {
		t.Fatalf("audit payload: got %q", auditVal.Payload)
	}
	if string(telVal.Payload) != "b" {
		t.Fatalf("telemetry payload: got %q", telVal.Payload)
	}
}

func TestEngineWithDSNConfig(t *testing.T) {
	root := t.TempDir()
	db := option.Must(ndb.Open(root, ndb.Options{DefaultKind: "msgstore"}))
	defer func() { option.MustZero(db.Close()) }()

	// A DSN string is a valid EngineConfig for msgstore.
	eng, err := db.Engine("dsn", ndb.EngineOptions{
		Config: "?compress=s2&split=count:5&maxmsg=8mib",
	})
	if err != nil {
		t.Fatalf("open with dsn: %v", err)
	}

	m := eng.(ndb.MessageEngine).Messages()
	const typeID ndb.TypeID = 1
	var trace [16]byte
	if _, err := m.Append(typeID, trace, []byte("payload")); err != nil {
		t.Fatalf("append: %v", err)
	}
}

func TestEngineWithStructConfig(t *testing.T) {
	root := t.TempDir()
	db := option.Must(ndb.Open(root, ndb.Options{DefaultKind: "msgstore"}))
	defer func() { option.MustZero(db.Close()) }()

	// The native options struct is also a valid EngineConfig and allows custom
	// strategy functions that a DSN cannot express.
	eng, err := db.Engine("struct", ndb.EngineOptions{
		Config: msgstore.Options{
			Compress:    msgstore.NoCompression,
			ShouldSplit: msgstore.SplitByCount(2),
		},
	})
	if err != nil {
		t.Fatalf("open with struct config: %v", err)
	}
	if _, ok := eng.(ndb.MessageEngine); !ok {
		t.Fatal("expected message engine")
	}
}

func TestLookupEngineMissing(t *testing.T) {
	root := t.TempDir()
	db := option.Must(ndb.Open(root, ndb.Options{}))
	defer func() { option.MustZero(db.Close()) }()

	got, err := db.LookupEngine("does-not-exist")
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if got.IsSome() {
		t.Fatal("expected empty option for missing engine")
	}
}

func TestLookupEngineExisting(t *testing.T) {
	root := t.TempDir()
	db := option.Must(ndb.Open(root, ndb.Options{DefaultKind: "msgstore"}))
	defer func() { option.MustZero(db.Close()) }()

	if _, err := db.Engine("present", ndb.EngineOptions{}); err != nil {
		t.Fatalf("open: %v", err)
	}

	got, err := db.LookupEngine("present")
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if got.IsNone() {
		t.Fatal("expected to find existing engine")
	}
	if got.Unwrap().Name() != "present" {
		t.Fatalf("unexpected name %q", got.Unwrap().Name())
	}
}

func TestDBClosedIsSpent(t *testing.T) {
	root := t.TempDir()
	db := option.Must(ndb.Open(root, ndb.Options{DefaultKind: "msgstore"}))

	if _, err := db.Engine("a", ndb.EngineOptions{}); err != nil {
		t.Fatalf("open: %v", err)
	}
	option.MustZero(db.Close())

	if _, err := db.Engine("b", ndb.EngineOptions{}); err == nil {
		t.Fatal("expected error opening engine on a closed database")
	}
}

func TestRegisteredEngines(t *testing.T) {
	if !slices.Contains(ndb.RegisteredEngines(), "msgstore") {
		t.Fatalf("msgstore engine not registered: %v", ndb.RegisteredEngines())
	}
}
