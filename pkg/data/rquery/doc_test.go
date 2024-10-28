package rquery

import (
	"go.wdy.de/nago/pkg/xiter"
	"reflect"
	"slices"
	"strings"
	"testing"
)

type MyRef string
type Addr struct {
	ID     MyRef
	Street string
}
type Person struct {
	ID      int
	Name    string
	Tags    []string
	Addr    []*Addr
	Enabled bool
}

func TestPredicate(t *testing.T) {
	items := []Person{
		{
			ID:   456,
			Name: "Efeu",
			Tags: []string{"giftig", "grün", "robust"},
			Addr: []*Addr{
				{ID: "werty", Street: "Nordseestr"},
				{ID: "asdf", Street: "Marie-Curie-Str"},
			},
		},
		{
			ID:   5,
			Name: "Kaktus",
			Tags: []string{"weich", "grün", "robust"},
			Addr: []*Addr{
				{ID: "werty", Street: "Nordseestr"},
				{ID: "asdf", Street: "Marie-Curie-Str"},
			},
		},
		{
			ID:   6,
			Name: "irgendwas grünes",
			Tags: []string{"weich", "grün", "robust", "kaktus-like"},
			Addr: []*Addr{
				{ID: "werty", Street: "Nordseestr"},
				{ID: "asdf", Street: "Marie-Curie-Str"},
			},
		},
		{
			ID:   8,
			Name: "Palme",
			Tags: []string{"lecker", "grün", "Kokosnuss"},
			Addr: []*Addr{
				{ID: "werty", Street: "Nordseestr"},
				{ID: "asdf", Street: "Marie-Curie-Str"},
			},
		},
	}

	p := SimplePredicate[Person]("kaktus")
	values := slices.Collect(xiter.Filter(p, slices.Values(items)))

	if !reflect.DeepEqual(values, []Person{items[1], items[2]}) {
		t.Fatal(values)
	}

	values = slices.Collect(xiter.Filter(SimplePredicate[Person]("palm grün"), slices.Values(items)))

	if !reflect.DeepEqual(values, []Person{items[3]}) {
		t.Fatal(values)
	}
}

func TestContains(t *testing.T) {
	if !contains("hello world", "wor") {
		t.Fatal()
	}

	if contains("hello world", "wabc") {
		t.Fatal()
	}

	if !contains("1234", "23") {
		t.Fatal()
	}

	data := Person{
		ID:   456,
		Name: "Efeu",
		Tags: []string{"giftig", "grün", "robust"},
		Addr: []*Addr{
			{ID: "werty", Street: "Nordseestr"},
			{ID: "asdf", Street: "Marie-Curie-Str"},
		},
		Enabled: false,
	}

	mustContainTable := []string{
		"werty",
		"asdf",
		"efeu",
		"4",
		"56",
		"456",
		"Efeu",
		" EFEU  ",
		"grün",
		"nord",
		"curie",
	}

	for _, str := range mustContainTable {
		if !contains(data, strings.TrimSpace(strings.ToLower(str))) {
			t.Fatalf("expected to match %v", str)
		}
	}

}
