package eventstore

import (
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/blob/tdb"
	"io"
	"math/rand"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestNewID(t *testing.T) {
	for range 5000 {
		now := time.Now()
		id := timeIntoID(now)
		ti, err := id.Time(time.Local)
		if err != nil {
			t.Fatal(id, ti, err)
		}

		if ti.UnixMilli() != now.UnixMilli() {
			t.Fatal(id, ti, ti.UnixMilli())
		}
	}
}

func TestStore(t *testing.T) {
	db, err := tdb.Open(filepath.Join(t.TempDir(), "tdb-test"))

	if err != nil {
		t.Fatal(err)
	}

	// required 106.519584ms to insert 10000 entries, 93879.45tps
	store := blob.Store(tdb.NewBlobStore(db, "test")) // this requires 53MiB after write

	// required 848.175917ms to insert 10000 entries, 11790.01tps
	// store, _ = pebble.Open(filepath.Join(t.TempDir(), "pebble")) // this requires 69MiB after write

	// required 1m54.263868875s to insert 10000 entries, 87.52tps
	//boltdb, _ := bbolt.Open(filepath.Join(t.TempDir(), "bbolt-test"), os.ModePerm, nil)
	//store = bolt.NewBlobStore(boltdb, "test") // this requires 96MiB after write

	//required 119.342917ms to insert 10000 entries, 83792.15tps
	//store, _ = badger.Open(filepath.Join(t.TempDir(), "badger")) // this requires 2GiB BEFORE !!! any write

	defer func() {
		if c, ok := store.(io.Closer); ok {
			c.Close()
		}
	}()

	testSet := makeTestSet()
	events := NewStore(store)
	var ids []ID
	start := time.Now()
	for _, bytes := range testSet {
		id, err := events.Save("abc", bytes)
		if err != nil {
			t.Fatal(err)
		}

		ids = append(ids, id)
	}

	delta := time.Now().Sub(start)
	t.Logf("required %v to insert %d entries, %.2ftps\n", delta, len(testSet), float64(len(testSet))/float64(delta.Seconds()))

	var lastId ID
	for idx, id := range ids {
		if id <= lastId {
			t.Fatalf("invalid id sequence, must be strict monotonic: %v vs %v", lastId, id)
		}

		lastId = id

		msg, err := events.Load(id)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(msg.Unwrap().Data, testSet[idx]) {
			t.Fatalf("invalid payload: %v vs %v", msg.Unwrap().Data, testSet[idx])
		}

	}
}

func makeTestSet() [][]byte {
	var res [][]byte
	rnd := rand.New(rand.NewSource(1234))
	for range 10_000 {
		length := rnd.Intn(8 * 1024)
		buf := make([]byte, length)
		rnd.Read(buf)
		res = append(res, buf)
	}

	return res
}
