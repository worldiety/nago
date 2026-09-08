// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package form

import (
	"errors"
	"iter"
	"slices"
	"testing"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/user"
)

// funcSource is a Source built from plain closures, so a test can make it misbehave in the specific ways a
// real source does.
type funcSource struct {
	all    func(user.Subject) iter.Seq2[string, error]
	byID   func(user.Subject, string) (option.Opt[Entity], error)
	lookup int
}

func (f *funcSource) FindAll(s user.Subject) iter.Seq2[string, error] { return f.all(s) }

func (f *funcSource) FindByID(s user.Subject, id string) (option.Opt[Entity], error) {
	f.lookup++
	return f.byID(s, id)
}

func idsOf(entities []Entity) []string {
	got := make([]string, 0, len(entities))
	for _, e := range entities {
		got = append(got, e.ID)
	}
	return got
}

func seqOf(ids ...string) func(user.Subject) iter.Seq2[string, error] {
	return func(user.Subject) iter.Seq2[string, error] {
		return func(yield func(string, error) bool) {
			for _, id := range ids {
				if !yield(id, nil) {
					return
				}
			}
		}
	}
}

// TestCollectSourceEntities_SkipsUnresolvableIDs is the regression for a crash, not a cosmetic issue: the
// renderers used to call Unwrap on the optional without checking it, so one entry that could not be resolved
// panicked with "unwrapped invalid option" and took the entire page down.
//
// It happens for ordinary reasons - an entry removed between the listing and the lookup, or a source that
// lists more than it resolves - so the picker must simply show one option less.
func TestCollectSourceEntities_SkipsUnresolvableIDs(t *testing.T) {
	src := &funcSource{
		all: seqOf("a", "gone", "b"),
		byID: func(_ user.Subject, id string) (option.Opt[Entity], error) {
			if id == "gone" {
				return option.None[Entity](), nil
			}
			return option.Some(Entity{ID: id, Value: id}), nil
		},
	}

	got, err := collectSourceEntities(src, nil)
	if err != nil {
		t.Fatalf("collect failed: %v", err)
	}

	want := []string{"a", "b"}
	if !slices.Equal(idsOf(got), want) {
		t.Errorf("got %v, want %v", idsOf(got), want)
	}
}

// TestCollectSourceEntities_PropagatesErrors keeps a genuine backend failure distinguishable from an empty
// picker, which would otherwise look like "there is nothing to choose".
func TestCollectSourceEntities_PropagatesErrors(t *testing.T) {
	boom := errors.New("backend unavailable")

	listing := &funcSource{
		all: func(user.Subject) iter.Seq2[string, error] {
			return func(yield func(string, error) bool) { yield("", boom) }
		},
		byID: func(user.Subject, string) (option.Opt[Entity], error) { return option.None[Entity](), nil },
	}

	if _, err := collectSourceEntities(listing, nil); !errors.Is(err, boom) {
		t.Errorf("got %v, want the listing error", err)
	}

	lookup := &funcSource{
		all:  seqOf("a"),
		byID: func(user.Subject, string) (option.Opt[Entity], error) { return option.None[Entity](), boom },
	}

	if _, err := collectSourceEntities(lookup, nil); !errors.Is(err, boom) {
		t.Errorf("got %v, want the lookup error", err)
	}
}

// TestNewStaticSource covers the closed sets a domain declares as constants.
func TestNewStaticSource(t *testing.T) {
	src := NewStaticSource(
		Choice{Value: "internal", Label: "Intern"},
		Choice{Value: "external", Label: "Extern"},
	)

	got, err := collectSourceEntities(src, nil)
	if err != nil {
		t.Fatalf("collect failed: %v", err)
	}

	if want := []string{"internal", "external"}; !slices.Equal(idsOf(got), want) {
		t.Errorf("got %v, want %v (order must be preserved)", idsOf(got), want)
	}

	optE, err := src.FindByID(nil, "internal")
	if err != nil || optE.IsNone() {
		t.Fatalf("cannot resolve a known choice: %v", err)
	}

	if label := optE.Unwrap().Value.(Choice).Label; label != "Intern" {
		t.Errorf("got label %q, want Intern", label)
	}

	if optE, err := src.FindByID(nil, "nope"); err != nil || optE.IsSome() {
		t.Errorf("an unknown choice resolved to something: %v %v", optE, err)
	}
}

// TestNewQuerySource_IndexesTheListing pins down the reason the constructor exists at all. The renderers walk
// a source as one FindAll followed by one FindByID per entry; without an index, a source whose listing is a
// network call pays for that listing once per entry.
func TestNewQuerySource_IndexesTheListing(t *testing.T) {
	listings := 0

	src := NewQuerySource(
		func(user.Subject) ([]string, error) {
			listings++
			return []string{"a", "b", "c"}, nil
		},
		func(s string) string { return s },
		func(s string) string { return "Label " + s },
	)

	if _, err := collectSourceEntities(src, nil); err != nil {
		t.Fatalf("collect failed: %v", err)
	}

	if listings != 1 {
		t.Errorf("the listing ran %d times for 3 entries, want exactly 1", listings)
	}
}

// TestNewQuerySource_ResolvesUnlistedIDs covers the case the index cannot answer: a form showing a value that
// is no longer offered, such as a retired group on an existing profile.
func TestNewQuerySource_ResolvesUnlistedIDs(t *testing.T) {
	src := NewQuerySource(
		func(user.Subject) ([]string, error) { return []string{"a"}, nil },
		func(s string) string { return s },
		func(s string) string { return "Label " + s },
	)

	optE, err := src.FindByID(nil, "a")
	if err != nil || optE.IsNone() {
		t.Fatalf("cannot resolve without a preceding listing: %v", err)
	}

	if optE, err := src.FindByID(nil, "unknown"); err != nil || optE.IsSome() {
		t.Errorf("an id the query does not know resolved to something: %v %v", optE, err)
	}
}

// TestNewQuerySource_PropagatesErrors makes sure a failing query is not presented as an empty choice list.
func TestNewQuerySource_PropagatesErrors(t *testing.T) {
	boom := errors.New("backend unavailable")

	src := NewQuerySource(
		func(user.Subject) ([]string, error) { return nil, boom },
		func(s string) string { return s },
		func(s string) string { return s },
	)

	if _, err := collectSourceEntities(src, nil); !errors.Is(err, boom) {
		t.Errorf("got %v, want the query error", err)
	}

	if _, err := src.FindByID(nil, "a"); !errors.Is(err, boom) {
		t.Errorf("got %v, want the query error", err)
	}
}
