package json

import (
	"go.etcd.io/bbolt"
	"go.wdy.de/nago/pkg/blob/badger"
	"go.wdy.de/nago/pkg/blob/bolt"
	"go.wdy.de/nago/pkg/blob/fs"
	"go.wdy.de/nago/pkg/blob/mem"
	"go.wdy.de/nago/pkg/blob/pebble"
	"go.wdy.de/nago/pkg/data"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	slices2 "slices"
	"strings"
	"testing"
)

type Person struct {
	ID   string
	Name string
}

func (p Person) Identity() string {
	return p.ID
}

func TestNewSloppyJSONRepository(t *testing.T) {
	t.Run("pebble-prefix", func(t *testing.T) {
		db, err := pebble.Open(filepath.Join(t.TempDir(), "pebble-test-prefix"))
		if err != nil {
			t.Fatal(err)
		}

		db.SetPrefix("blub")
		testSuite(t, NewSloppyJSONRepository[Person, string](db))
		db.SetPrefix("blub2")
		testSuite(t, NewSloppyJSONRepository[Person, string](db))
	})

	t.Run("pebble", func(t *testing.T) {
		db, err := pebble.Open(filepath.Join(t.TempDir(), "pebble-test"))
		if err != nil {
			t.Fatal(err)
		}
		testSuite(t, NewSloppyJSONRepository[Person, string](db))
	})

	t.Run("badger-prefix", func(t *testing.T) {
		db, err := badger.Open(filepath.Join(t.TempDir(), "badger-test-prefix"))
		if err != nil {
			t.Fatal(err)
		}

		db.SetPrefix("blub")
		testSuite(t, NewSloppyJSONRepository[Person, string](db))
		db.SetPrefix("blub2")
		testSuite(t, NewSloppyJSONRepository[Person, string](db))
	})

	t.Run("badger", func(t *testing.T) {
		db, err := badger.Open(filepath.Join(t.TempDir(), "badger-test"))
		if err != nil {
			t.Fatal(err)
		}
		testSuite(t, NewSloppyJSONRepository[Person, string](db))
	})

	t.Run("bbolt", func(t *testing.T) {
		db, err := bbolt.Open(filepath.Join(t.TempDir(), "test.db"), os.ModePerm, nil)
		if err != nil {
			t.Fatal(err)
		}
		testSuite(t, NewSloppyJSONRepository[Person, string](bolt.NewBlobStore(db, "test")))
	})

	t.Run("mem", func(t *testing.T) {
		testSuite(t, NewSloppyJSONRepository[Person, string](mem.NewBlobStore()))
	})

	t.Run("fs", func(t *testing.T) {
		testSuite(t, NewSloppyJSONRepository[Person, string](unwrap(fs.NewBlobStore(t.TempDir()))))
	})

}

func BenchmarkNewSloppyJSONRepository(b *testing.B) {

	b.Run("bbolt", func(t *testing.B) {
		for n := 0; n < b.N; n++ {
			db, err := bbolt.Open(filepath.Join(t.TempDir(), "test.db"), os.ModePerm, nil)
			if err != nil {
				t.Fatal(err)
			}
			testSuite(t, NewSloppyJSONRepository[Person, string](bolt.NewBlobStore(db, "test")))
			db.Close()
		}

	})

	b.Run("mem", func(t *testing.B) {
		for n := 0; n < b.N; n++ {
			testSuite(t, NewSloppyJSONRepository[Person, string](mem.NewBlobStore()))
		}
	})

	b.Run("fs", func(t *testing.B) {
		for n := 0; n < b.N; n++ {
			store := unwrap(fs.NewBlobStore(t.TempDir()))
			testSuite(t, NewSloppyJSONRepository[Person, string](store))
			store.Close()
		}
	})
}

func testSuite(t interface {
	Fatalf(format string, args ...any)
	Fatal(...any)
}, repo data.Repository[Person, string]) {
	if v := unwrap(repo.Count()); v != 0 {
		t.Fatalf("expected 0 but got %v", v)
	}

	must(repo.Save(Person{
		ID:   "1234",
		Name: "Torben",
	}))

	if v := unwrap(repo.Count()); v != 1 {
		t.Fatalf("expected 1 but got %v", v)
	}

	if p := unwrap(repo.FindByID("1234")); p.Unwrap().ID != "1234" || p.Unwrap().Name != "Torben" {
		t.Fatalf("unexpected :%+v", p)
	}

	must(repo.DeleteByID("1"))
	if v := unwrap(repo.Count()); v != 1 {
		t.Fatalf("expected 1 but got %v", v)
	}

	must(repo.DeleteByID("1234"))
	if v := unwrap(repo.Count()); v != 0 {
		t.Fatalf("expected 0 but got %v", v)
	}

	testSet := []Person{
		{
			ID:   "1",
			Name: "Commander Sisko",
		},
		{
			ID:   "2",
			Name: "Captain Kirk",
		},
		{
			ID:   "3",
			Name: "Captain Picard",
		},
	}
	must(repo.SaveAll(slices2.Values(testSet)))

	if v := unwrap(repo.Count()); v != 3 {
		t.Fatalf("expected 3 but got %v", v)
	}

	var tmp []Person
	repo.Each(func(person Person, err error) bool {
		tmp = append(tmp, person)
		if err != nil {
			t.Fatal(err)
		}

		return true
	})

	slices.SortFunc(tmp, func(a, b Person) int {
		return strings.Compare(a.ID, b.ID)
	})

	if !reflect.DeepEqual(testSet, tmp) {
		t.Fatalf("unexpected %+v %+v", tmp, testSet)
	}

	// again but different
	tmp = nil
	for person, err := range repo.FindAllByID(slices.Values([]string{"3", "2", "1"})) {
		tmp = append(tmp, person)
		if err != nil {
			t.Fatal(err)
		}
	}

	slices.SortFunc(tmp, func(a, b Person) int {
		return strings.Compare(a.ID, b.ID)
	})

	if !reflect.DeepEqual(testSet, tmp) {
		t.Fatalf("unexpected %+v", tmp)
	}

	//
	must(repo.DeleteAllByID(slices2.Values([]string{"3", "2"})))
	if v := unwrap(repo.Count()); v != 1 {
		t.Fatalf("expected 1 but got %v", v)
	}

	//
	must(repo.DeleteAll())
	if v := unwrap(repo.Count()); v != 0 {
		t.Fatalf("expected 0 but got %v", v)
	}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func unwrap[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}

	return t
}
