package json

import (
	"go.etcd.io/bbolt"
	"go.wdy.de/nago/pkg/blob/bolt"
	"go.wdy.de/nago/pkg/blob/fs"
	"go.wdy.de/nago/pkg/blob/mem"
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
	t.Run("mem", func(t *testing.T) {
		testSuite(t, NewSloppyJSONRepository[Person, string](mem.NewBlobStore()))
	})

	t.Run("fs", func(t *testing.T) {
		testSuite(t, NewSloppyJSONRepository[Person, string](fs.NewBlobStore(t.TempDir())))
	})

	t.Run("bbolt", func(t *testing.T) {
		db, err := bbolt.Open(filepath.Join(t.TempDir(), "test.db"), os.ModePerm, nil)
		if err != nil {
			t.Fatal(err)
		}
		testSuite(t, NewSloppyJSONRepository[Person, string](bolt.NewBlobStore(db, "test")))
	})

}

func testSuite(t *testing.T, repo data.Repository[Person, string]) {
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
		t.Fatalf("unexpected %+v", tmp)
	}

	// again but different
	tmp = nil
	repo.FindAllByID(slices2.Values([]string{"3", "2", "1"}), func(person Person, err error) bool {
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
