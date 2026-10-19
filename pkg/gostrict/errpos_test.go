// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict_test

import (
	"context"
	"testing"

	"go.wdy.de/nago/pkg/gostrict"
)

func expectPos(t *testing.T, err error, file string, line int) {
	t.Helper()
	if err == nil {
		t.Fatal("expected an error")
	}
	f, l, ok := gostrict.ErrorPos(err)
	if !ok || f != file || l != line {
		t.Fatalf("%q at %s:%d (%v), want %s:%d", err, f, l, ok, file, line)
	}
}

// TestErrorPos returns an error of errors.New from an export: goja reports where the program created it, the
// native engine does not, and both compare equal.
func TestErrorPos(t *testing.T) {
	b := compile(t, "internal/examples/errs")
	args := []any{int32(1), int32(0)}
	got := call(t, goEngine, b, nil, "Average", args)
	expectPos(t, got[1].(error), "go.wdy.de/nago/pkg/gostrict/internal/examples/errs/errs.go", 34)
	if got[1].(error).Error() != "no values" {
		t.Fatalf("unexpected message %q", got[1])
	}

	want := call(t, nativeEngine("errs"), b, nil, "Average", args)
	if _, _, ok := gostrict.ErrorPos(want[1].(error)); ok {
		t.Fatal("the native engine has no positions")
	}
	if !gostrict.SameValues(got, want) {
		t.Fatalf("goja %v, native %v", got, want)
	}
}

// TestErrorPosScript covers package variables and the functions of the library.
func TestErrorPosScript(t *testing.T) {
	cache := gostrict.NewCache(1)
	b, diags, err := cache.Compile("llm.local/s", script(`package s

import (
	"errors"

	"go.wdy.de/nago/pkg/gostrict/std/text"
)

var errEmpty = errors.New("empty")

func Parse(s string) (int32, error) {
	if s == "" {
		return 0, errEmpty
	}
	return text.ParseInt(s)
}

func Wrap(s string) error {
	_, err := Parse(s)
	if err != nil {
		return errors.New("wrapped: " + err.Error())
	}
	return nil
}
`))
	if err != nil || len(diags) > 0 {
		t.Fatal(err, diags)
	}
	in := mustLoad(t, goEngine, b)
	ctx := context.Background()
	run := func(name string, args ...any) error {
		t.Helper()
		res, err := in.Call(ctx, name, args...)
		if err != nil {
			t.Fatal(err)
		}
		err, _ = res[len(res)-1].(error)
		return err
	}

	expectPos(t, run("Parse", ""), "llm.local/s/s.go", 9)
	expectPos(t, run("Parse", "x"), "llm.local/s/s.go", 15)
	expectPos(t, run("Wrap", "x"), "llm.local/s/s.go", 21)
	if err := run("Parse", "7"); err != nil {
		t.Fatal(err)
	}
}
