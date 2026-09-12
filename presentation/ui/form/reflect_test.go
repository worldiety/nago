// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package form

import (
	"errors"
	"reflect"
	"slices"
	"strings"
	"testing"

	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/pkg/xerrors"
)

type addr struct {
	Country string
	Zip     string
}

type seat struct {
	Address addr
	Name    string
}

type issuer struct {
	Name string
	Seat seat
}

type root struct {
	Title  string
	Issuer issuer
}

// Addr is exported on purpose: an embedded field of an unexported type is skipped by
// GroupsOf, which would not exercise the promotion path.
type Addr struct {
	Country string
	Zip     string
}

type embedded struct {
	Addr
	Title string
}

type recursive struct {
	Name  string
	Child *recursive
}

// d3 is a deeply nested chain used to verify that the descent respects maxNestingDepth.
type d3 struct{ Leaf string }
type d2 struct{ Next d3 }
type d1 struct{ Next d2 }
type d0 struct{ Next d1 }

func TestGroupsOfNonStruct(t *testing.T) {
	// used to panic and take the whole render down
	tests := []reflect.Type{
		reflect.TypeFor[int](),
		reflect.TypeFor[[]string](),
		reflect.TypeFor[map[string]string](),
	}

	for _, tt := range tests {
		t.Run(tt.String(), func(t *testing.T) {
			if got := GroupsOf(tt); got != nil {
				t.Errorf("want nil, got %v", got)
			}
		})
	}
}

func TestGroupsOfKeepsNamedNestedStructAsSingleField(t *testing.T) {
	groups := GroupsOf(reflect.TypeFor[root]())

	var names []string
	for _, g := range groups {
		for _, f := range g.Fields {
			names = append(names, f.Name)
		}
	}

	slices.Sort(names)
	want := []string{"Issuer", "Title"}

	if !slices.Equal(names, want) {
		t.Errorf("fields = %v, want %v", names, want)
	}
}

func TestGroupsOfPromotesEmbedded(t *testing.T) {
	groups := GroupsOf(reflect.TypeFor[embedded]())

	var names []string
	for _, g := range groups {
		for _, f := range g.Fields {
			names = append(names, f.Name)
		}
	}

	slices.Sort(names)

	// VisibleFields yields the embedded struct itself plus its promoted leaves
	for _, want := range []string{"Addr", "Country", "Title", "Zip"} {
		if !slices.Contains(names, want) {
			t.Errorf("missing %q in %v", want, names)
		}
	}
}

func TestFieldPaths(t *testing.T) {
	tests := []struct {
		name string
		typ  reflect.Type
		want []string
	}{
		{
			name: "flat",
			typ:  reflect.TypeFor[addr](),
			want: []string{"Country", "Zip"},
		},
		{
			name: "two levels",
			typ:  reflect.TypeFor[seat](),
			want: []string{"Address.Country", "Address.Zip", "Name"},
		},
		{
			name: "three levels",
			typ:  reflect.TypeFor[root](),
			want: []string{"Title", "Issuer.Name", "Issuer.Seat.Address.Country", "Issuer.Seat.Address.Zip", "Issuer.Seat.Name"},
		},
		{
			name: "embedded stays flat and the container is not addressable",
			typ:  reflect.TypeFor[embedded](),
			want: []string{"Country", "Zip", "Title"},
		},
		{
			name: "deep chain",
			typ:  reflect.TypeFor[d0](),
			want: []string{"Next.Next.Next.Leaf"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FieldPaths(tt.typ)

			slices.Sort(got)
			want := slices.Clone(tt.want)
			slices.Sort(want)

			if !slices.Equal(got, want) {
				t.Errorf("paths = %v, want %v", got, want)
			}
		})
	}
}

func TestFieldPathsHonoursIgnoreFields(t *testing.T) {
	got := FieldPaths(reflect.TypeFor[seat](), "Name")

	if slices.Contains(got, "Name") {
		t.Errorf("ignored field must not be addressable: %v", got)
	}

	if !slices.Contains(got, "Address.Country") {
		t.Errorf("nested fields must survive: %v", got)
	}
}

func TestFieldPathsNonStruct(t *testing.T) {
	if got := FieldPaths(reflect.TypeFor[int]()); len(got) != 0 {
		t.Errorf("want no paths, got %v", got)
	}
}

// TestDescendable pins the descent rule, which is what decides whether a nested model is
// rendered at all.
func TestDescendable(t *testing.T) {
	fieldOf := func(typ reflect.Type, name string) reflect.StructField {
		f, ok := typ.FieldByName(name)
		if !ok {
			t.Fatalf("no field %q on %v", name, typ)
		}
		return f
	}

	tests := []struct {
		name  string
		field reflect.StructField
		want  bool
	}{
		{"named nested struct", fieldOf(reflect.TypeFor[root](), "Issuer"), true},
		{"string", fieldOf(reflect.TypeFor[root](), "Title"), false},
		{"pointer to struct", fieldOf(reflect.TypeFor[recursive](), "Child"), false},
		{"embedded struct", fieldOf(reflect.TypeFor[embedded](), "Addr"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := descendable(tt.field); got != tt.want {
				t.Errorf("descendable = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestWalkFieldsDepthCap guards the only bound on the traversal. The visitor never claims a
// descendable field, so walkFields' own cap is the sole thing that can stop the recursion.
func TestWalkFieldsDepthCap(t *testing.T) {
	walk := func(startDepth int) []string {
		var out []string
		walkFields(GroupsOf(reflect.TypeFor[d0]())[0].Fields, nil, nil, "", startDepth, fieldVisitor{
			visit: func(field reflect.StructField, path string, depth int) bool {
				out = append(out, path)
				return !descendable(field)
			},
		})
		return out
	}

	want := []string{"Next", "Next.Next", "Next.Next.Next", "Next.Next.Next.Leaf"}
	if got := walk(0); !slices.Equal(got, want) {
		t.Errorf("full walk = %v, want %v", got, want)
	}

	if got := walk(maxNestingDepth); !slices.Equal(got, []string{"Next"}) {
		t.Errorf("at the cap the walk must not descend, got %v", got)
	}
}

func TestConsumedKeys(t *testing.T) {
	var c consumedKeys

	if c.has("Name") {
		t.Error("empty set must not report a key")
	}

	c.mark("Name")

	if !c.has("Name") {
		t.Error("marked key must be reported")
	}

	// marking the empty key is a no-op, a field without an error must not claim anything
	c.mark("")
	if c.has("") {
		t.Error("empty key must never be marked")
	}
}

func TestConsumedKeysNilSafe(t *testing.T) {
	var c *consumedKeys

	c.mark("Name") // must not panic
	if c.has("Name") {
		t.Error("nil set must not report a key")
	}
}

// TestResidualError covers the invariant that nothing the caller passed in disappears.
func TestResidualError(t *testing.T) {
	dbDown := errors.New("db down")

	tests := []struct {
		name      string
		errs      error
		claimed   []string
		wantNil   bool
		wantIsErr error
	}{
		{
			name:    "no error",
			errs:    nil,
			wantNil: true,
		},
		{
			name:    "every field message reached a field",
			errs:    xerrors.WithFields("", "Name", "empty"),
			claimed: []string{"Name"},
			wantNil: true,
		},
		{
			// the regression: the validation half made the technical half look handled
			name:      "technical error joined with a fully claimed validation error",
			errs:      errors.Join(xerrors.WithFields("", "Name", "empty"), dbDown),
			claimed:   []string{"Name"},
			wantIsErr: dbDown,
		},
		{
			name:      "purely technical error",
			errs:      dbDown,
			wantIsErr: dbDown,
		},
		{
			// a typo in the key, or a field hidden by IgnoreFields
			name:    "field message reached nobody",
			errs:    xerrors.WithFields("", "Naem", "empty"),
			claimed: nil,
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			consumed := &consumedKeys{}
			for _, k := range tt.claimed {
				consumed.mark(k)
			}

			fields, _ := xerrors.Collect(tt.errs)

			var auto TAuto[any]
			auto.opts.Errors = tt.errs

			got := auto.residualError(nil, reflect.TypeFor[root](), consumed, fields.Fields)

			if tt.wantNil {
				if got != nil {
					t.Fatalf("want nil, got %q", got.Error())
				}
				return
			}

			if got == nil {
				t.Fatal("want a residual error, got nil")
			}

			if tt.wantIsErr != nil && !errors.Is(got, tt.wantIsErr) {
				t.Errorf("residual must still match the original: %q", got.Error())
			}
		})
	}
}

// TestResidualErrorDoesNotLeakFieldKeysToTheUser pins that Go identifiers stay in the log.
func TestResidualErrorDoesNotLeakFieldKeysToTheUser(t *testing.T) {
	consumed := &consumedKeys{}
	errs := xerrors.WithFields("", "Issuer.Seat.Country", "must not be empty")
	fields, _ := xerrors.Collect(errs)

	var auto TAuto[any]
	auto.opts.Errors = errs

	got := auto.residualError(nil, reflect.TypeFor[root](), consumed, fields.Fields)
	if got == nil {
		t.Fatal("an unclaimed message must still produce a banner")
	}

	var localized std.LocalizedError
	if !errors.As(got, &localized) {
		t.Fatalf("want a localized error, got %T", got)
	}

	if strings.Contains(localized.Description(), "Issuer.Seat.Country") {
		t.Errorf("developer diagnostics must not reach the user: %q", localized.Description())
	}
}
