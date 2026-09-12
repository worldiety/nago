// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package xerrors

import (
	"errors"
	"fmt"
	"maps"
	"testing"
	"time"
)

func TestCollect(t *testing.T) {
	sentinel := errors.New("boom")

	tests := []struct {
		name       string
		err        error
		wantFound  bool
		wantFields map[string]string
	}{
		{
			name:      "nil",
			err:       nil,
			wantFound: false,
		},
		{
			name:      "plain error",
			err:       sentinel,
			wantFound: false,
		},
		{
			name:       "direct",
			err:        WithFields("invalid", "Name", "must not be empty"),
			wantFound:  true,
			wantFields: map[string]string{"Name": "must not be empty"},
		},
		{
			name:       "wrapped with percent w",
			err:        fmt.Errorf("cannot decide: %w", WithFields("invalid", "Name", "empty")),
			wantFound:  true,
			wantFields: map[string]string{"Name": "empty"},
		},
		{
			// this is the case errors.AsType gets wrong: it returns the first match only
			name: "joined with a technical error",
			err: errors.Join(
				WithFields("invalid", "Name", "empty"),
				sentinel,
			),
			wantFound:  true,
			wantFields: map[string]string{"Name": "empty"},
		},
		{
			name: "two joined field errors are merged",
			err: errors.Join(
				WithFields("a", "Name", "empty"),
				WithFields("b", "Age", "too small"),
			),
			wantFound:  true,
			wantFields: map[string]string{"Name": "empty", "Age": "too small"},
		},
		{
			name: "same key from two branches is joined",
			err: errors.Join(
				WithFields("a", "Name", "empty"),
				WithFields("b", "Name", "too short"),
			),
			wantFound:  true,
			wantFields: map[string]string{"Name": "empty\ntoo short"},
		},
		{
			name: "nested through cause",
			err: ErrorWithFields{
				Message: "outer",
				Fields:  map[string]string{"Outer": "bad"},
				Cause:   WithFields("inner", "Inner", "worse"),
			},
			wantFound:  true,
			wantFields: map[string]string{"Outer": "bad", "Inner": "worse"},
		},
		{
			name:       "deeply wrapped join",
			err:        fmt.Errorf("layer: %w", errors.Join(sentinel, WithFields("x", "Seat.Country", "unknown"))),
			wantFound:  true,
			wantFields: map[string]string{"Seat.Country": "unknown"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, found := Collect(tt.err)
			if found != tt.wantFound {
				t.Fatalf("found = %v, want %v", found, tt.wantFound)
			}

			if !found {
				return
			}

			if !maps.Equal(got.Fields, tt.wantFields) {
				t.Errorf("fields = %v, want %v", got.Fields, tt.wantFields)
			}

			// the original error must remain reachable, nothing may be lost
			if got.Cause == nil {
				t.Error("cause not preserved")
			}
		})
	}
}

func TestCollectPreservesSentinel(t *testing.T) {
	sentinel := errors.New("db down")
	err := errors.Join(WithFields("invalid", "Name", "empty"), sentinel)

	got, ok := Collect(err)
	if !ok {
		t.Fatal("want found")
	}

	if !errors.Is(got, sentinel) {
		t.Error("errors.Is must still find the joined technical error through the result")
	}
}

func TestCollectCycleSafe(t *testing.T) {
	c := &cyclic{}
	c.self = c

	// must terminate
	if _, ok := Collect(c); ok {
		t.Error("cyclic plain error must not report fields")
	}
}

type cyclic struct{ self error }

func (c *cyclic) Error() string { return "cyclic" }
func (c *cyclic) Unwrap() error { return c.self }

func TestErrorWithFieldsError(t *testing.T) {
	tests := []struct {
		name string
		err  ErrorWithFields
		want string
	}{
		{"empty", ErrorWithFields{}, "field validation failed"},
		{"message only", ErrorWithFields{Message: "invalid"}, "invalid"},
		{"cause only", ErrorWithFields{Cause: errors.New("boom")}, "boom"},
		{"both", ErrorWithFields{Message: "invalid", Cause: errors.New("boom")}, "invalid: boom"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestScoped(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
		in     map[string]string
		want   map[string]string
	}{
		{
			name:   "empty prefix is identity",
			prefix: "",
			in:     map[string]string{"Country": "empty"},
			want:   map[string]string{"Country": "empty"},
		},
		{
			name:   "single segment",
			prefix: "Issuer",
			in:     map[string]string{"Country": "empty"},
			want:   map[string]string{"Issuer.Country": "empty"},
		},
		{
			name:   "multi segment prefix",
			prefix: "Issuer.Seat",
			in:     map[string]string{"Country": "empty", "Zip": "bad"},
			want:   map[string]string{"Issuer.Seat.Country": "empty", "Issuer.Seat.Zip": "bad"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ErrorWithFields{Fields: tt.in}.Scoped(tt.prefix)
			if !maps.Equal(got.Fields, tt.want) {
				t.Errorf("fields = %v, want %v", got.Fields, tt.want)
			}
		})
	}
}

func TestScopedDoesNotMutateReceiver(t *testing.T) {
	orig := ErrorWithFields{Fields: map[string]string{"Country": "empty"}}
	_ = orig.Scoped("Issuer")

	if _, ok := orig.Fields["Country"]; !ok {
		t.Error("Scoped must not mutate the receivers map")
	}
}

func TestFieldBuilderNested(t *testing.T) {
	var errs FieldBuilder
	errs.Add("Name", "empty")
	errs.Nested("Issuer").Add("Country", "empty")
	errs.Nested("Issuer").Nested("Seat").Add("Zip", "bad")

	if !errs.Has() {
		t.Fatal("want Has")
	}

	err := errs.Error()
	ewf, ok := Collect(err)
	if !ok {
		t.Fatal("want ErrorWithFields")
	}

	want := map[string]string{
		"Name":            "empty",
		"Issuer.Country":  "empty",
		"Issuer.Seat.Zip": "bad",
	}

	if !maps.Equal(ewf.Fields, want) {
		t.Errorf("fields = %v, want %v", ewf.Fields, want)
	}
}

func TestFieldBuilderNestedHasAndError(t *testing.T) {
	var root FieldBuilder
	nested := root.Nested("Issuer")

	if root.Has() || nested.Has() {
		t.Fatal("empty builders must not report Has")
	}

	if root.Error() != nil || nested.Error() != nil {
		t.Fatal("empty builders must return a nil error")
	}

	nested.Add("Country", "empty")

	if !root.Has() {
		t.Error("adding via a nested builder must be visible on the root")
	}

	if !nested.Has() {
		t.Error("nested builder must delegate Has to the root")
	}
}

func TestFieldBuilderErrorIsDeterministic(t *testing.T) {
	build := func() string {
		var errs FieldBuilder
		errs.Add("Zulu", "z")
		errs.Add("Alpha", "a")
		errs.Add("Mike", "m")
		return errs.Error().Error()
	}

	want := "field validation failed: Alpha: a; Mike: m; Zulu: z"

	// a Go map dump would vary between iterations
	for i := 0; i < 50; i++ {
		if got := build(); got != want {
			t.Fatalf("iteration %d: got %q, want %q", i, got, want)
		}
	}
}

func TestFieldBuilderSetMessage(t *testing.T) {
	var errs FieldBuilder
	errs.Add("Name", "empty")
	errs.SetMessage("Die Eingaben sind unvollständig.")

	if got, want := errs.Error().Error(), "Die Eingaben sind unvollständig."; got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestFieldBuilderErrorSnapshotsFields(t *testing.T) {
	var errs FieldBuilder
	errs.Add("Name", "empty")

	err, _ := Collect(errs.Error())
	errs.Add("Age", "missing")

	if _, ok := err.Fields["Age"]; ok {
		t.Error("Error must snapshot the field map, not alias it")
	}
}

func TestWithFieldsOddArgs(t *testing.T) {
	err, ok := Collect(WithFields("m", "Name"))
	if !ok {
		t.Fatal("want found")
	}

	if got := err.Fields["Name"]; got != "!MISSING_VALUE" {
		t.Errorf("got %q, want the missing value marker", got)
	}
}

func TestResidual(t *testing.T) {
	dbDown := errors.New("db down")

	tests := []struct {
		name string
		err  error
		want string // "" means nil
	}{
		{
			name: "nil",
			err:  nil,
			want: "",
		},
		{
			name: "pure field error is fully handled by the fields",
			err:  WithFields("", "Name", "empty"),
			want: "",
		},
		{
			name: "technical error survives",
			err:  dbDown,
			want: "db down",
		},
		{
			// the regression: the validation half used to make the whole join look handled
			name: "field error joined with a technical error",
			err:  errors.Join(WithFields("", "Name", "empty"), dbDown),
			want: "db down",
		},
		{
			name: "two joined field errors leave nothing",
			err: errors.Join(
				WithFields("", "Name", "empty"),
				WithFields("", "Zip", "bad"),
			),
			want: "",
		},
		{
			name: "wrapper around a pure field error is dropped",
			err:  fmt.Errorf("cannot decide command: %w", WithFields("", "Name", "empty")),
			want: "",
		},
		{
			name: "deliberate message survives",
			err:  ErrorWithFields{Message: "Die Eingaben sind unvollständig.", Fields: map[string]string{"Name": "empty"}},
			want: "Die Eingaben sind unvollständig.",
		},
		{
			name: "technical cause of a field error survives",
			err:  ErrorWithFields{Fields: map[string]string{"Name": "empty"}, Cause: dbDown},
			want: "db down",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Residual(tt.err)

			if tt.want == "" {
				if got != nil {
					t.Fatalf("want nil, got %q", got.Error())
				}
				return
			}

			if got == nil {
				t.Fatalf("want %q, got nil", tt.want)
			}

			if got.Error() != tt.want {
				t.Errorf("got %q, want %q", got.Error(), tt.want)
			}
		})
	}
}

func TestResidualPreservesErrorsIs(t *testing.T) {
	sentinel := errors.New("db down")
	got := Residual(errors.Join(WithFields("", "Name", "empty"), sentinel))

	if !errors.Is(got, sentinel) {
		t.Error("the surviving technical error must still be matchable")
	}
}

func TestCollectMergesSubstringMessages(t *testing.T) {
	// regression: a substring check discarded the shorter, genuinely different message
	got, ok := Collect(errors.Join(
		WithFields("a", "Name", "must not be empty"),
		WithFields("b", "Name", "empty"),
	))

	if !ok {
		t.Fatal("want found")
	}

	if want := "must not be empty\nempty"; got.Fields["Name"] != want {
		t.Errorf("got %q, want %q", got.Fields["Name"], want)
	}
}

func TestCollectDeduplicatesIdenticalMessages(t *testing.T) {
	got, _ := Collect(errors.Join(
		WithFields("a", "Name", "empty"),
		WithFields("b", "Name", "empty"),
	))

	if got.Fields["Name"] != "empty" {
		t.Errorf("identical messages must not be repeated, got %q", got.Fields["Name"])
	}
}

func TestFieldBuilderErrorHasNoMessageUnlessSet(t *testing.T) {
	var errs FieldBuilder
	errs.Add("Name", "empty")

	ewf, ok := errs.Error().(ErrorWithFields)
	if !ok {
		t.Fatal("want ErrorWithFields")
	}

	if ewf.Message != "" {
		t.Errorf("Message must stay empty so consumers can tell it from a deliberate one, got %q", ewf.Message)
	}

	// but the technical rendering must still be deterministic and useful in a log
	if got, want := ewf.Error(), "field validation failed: Name: empty"; got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

// fanout branches on every unwrap, which makes a depth-only bound useless.
type fanout struct{ child error }

func (f fanout) Error() string   { return "fanout" }
func (f fanout) Unwrap() []error { return []error{f.child, f.child} }

func TestCollectTerminatesOnExponentialFanout(t *testing.T) {
	var e error = WithFields("leaf", "Name", "empty")
	for i := 0; i < 80; i++ {
		e = fanout{child: e}
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		Collect(e)
		Residual(e)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("traversal did not terminate: the node budget is not effective")
	}
}
