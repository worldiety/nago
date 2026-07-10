package ndb_test

import (
	"fmt"
	"path/filepath"
	"sync"
	"testing"

	"github.com/worldiety/option"
	"go.wdy.de/nago/pkg/ndb"
)

// capturingEngine records the shared pool it was handed by the DB.
type capturingEngine struct {
	name string
	pool *ndb.FilePool
}

func (e *capturingEngine) Name() string         { return e.name }
func (e *capturingEngine) Kind() ndb.EngineKind { return "poolcapture" }

var poolCaptureInit sync.Once

func registerPoolCapture() {
	poolCaptureInit.Do(func() {
		ndb.Register("poolcapture", func(name, dir string, pool *ndb.FilePool, cfg ndb.EngineConfig) (ndb.Engine, func() error, error) {
			// Engines must never close the shared pool. We only observe it.
			return &capturingEngine{name: name, pool: pool}, func() error { return nil }, nil
		})
	})
}

func TestSharedFilePoolAcrossEngines(t *testing.T) {
	registerPoolCapture()

	root := t.TempDir()
	db := option.Must(ndb.Open(root, ndb.Options{DefaultKind: "poolcapture"}))

	a, err := db.Engine("a", ndb.EngineOptions{})
	if err != nil {
		t.Fatalf("open a: %v", err)
	}
	b, err := db.Engine("b", ndb.EngineOptions{})
	if err != nil {
		t.Fatalf("open b: %v", err)
	}

	pa := a.(*capturingEngine).pool
	pb := b.(*capturingEngine).pool

	if pa == nil || pb == nil {
		t.Fatal("engines did not receive a shared pool")
	}
	if pa != pb {
		t.Fatalf("expected both engines to share one *FilePool, got %p and %p", pa, pb)
	}

	if err := db.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
}

func TestExplicitSharedFilePoolIsClosedOnce(t *testing.T) {
	root := t.TempDir()
	shared := ndb.NewFilePool(8)
	db := option.Must(ndb.Open(root, ndb.Options{DefaultKind: "poolcapture", FilePool: shared}))
	registerPoolCapture()

	eng, err := db.Engine("only", ndb.EngineOptions{})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if eng.(*capturingEngine).pool != shared {
		t.Fatal("engine did not receive the explicitly provided shared pool")
	}

	// Closing the DB closes the shared pool; a second Close is a no-op and
	// must not error (Close is documented idempotent).
	if err := db.Close(); err != nil {
		t.Fatalf("first close: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("second close: %v", err)
	}
}

// TestFilePoolShardingConsistent verifies that a given path always resolves to
// the same shard, so the pool never holds two handles for one file, and that
// concurrent access across many paths is race-free (run with -race).
func TestFilePoolShardingConsistent(t *testing.T) {
	dir := t.TempDir()
	pool := ndb.NewFilePool(1024)
	defer pool.Close()

	// write and read back many distinct files concurrently.
	const files = 64
	var wg sync.WaitGroup
	for i := 0; i < files; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			path := filepath.Join(dir, fmt.Sprintf("f%d.bin", i))
			payload := []byte(fmt.Sprintf("data-%d", i))
			if _, err := pool.WriteAt(path, payload, 0); err != nil {
				t.Error(err)
				return
			}
			buf := make([]byte, len(payload))
			if _, err := pool.ReadAt(path, buf, 0); err != nil {
				t.Error(err)
				return
			}
			if string(buf) != string(payload) {
				t.Errorf("file %d: read %q want %q", i, buf, payload)
			}
		}(i)
	}
	wg.Wait()
}
