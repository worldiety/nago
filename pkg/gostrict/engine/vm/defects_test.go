// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package vm_test

import (
	"context"
	"errors"
	"math"
	"testing"
	"testing/fstest"
	"time"

	"go.wdy.de/nago/pkg/gostrict"
	"go.wdy.de/nago/pkg/gostrict/engine/vm"
	"go.wdy.de/nago/pkg/gostrict/lib/ui"
)

// compileSrc compiles the source of package s.
func compileSrc(t *testing.T, src string) *gostrict.Bundle {
	t.Helper()
	b, diags, err := gostrict.NewCache(1).Compile("llm.local/s", gostrict.Module{Path: "llm.local/s", FS: fstest.MapFS{"s.go": {Data: []byte(src)}}})
	if err != nil || len(diags) > 0 {
		t.Fatal(err, diags.Text())
	}
	if _, err := b.VM(); err != nil {
		t.Fatal(err)
	}
	return b
}

// TestLibraryInt calls the functions of the library, whose parameters have the type int.
func TestLibraryInt(t *testing.T) {
	b := compileSrc(t, `//gostrict:terminating
package s

import "math"

func Inf(neg bool) float64 {
	if neg {
		return math.Inf(1) + math.Inf(-1)
	}
	return math.Inf(1)
}

func IsInf(x float64) bool { return math.IsInf(x, 0) && !math.IsInf(x, -1) }
`)
	if got := call(t, new(vm.Engine), b, nil, "Inf", []any{true}); !math.IsNaN(got[0].(float64)) {
		t.Fatalf("Inf(-1) = %v", got)
	}
	if got := call(t, new(vm.Engine), b, nil, "Inf", []any{false}); !math.IsInf(got[0].(float64), 1) {
		t.Fatalf("Inf(1) = %v", got)
	}
	if got := call(t, new(vm.Engine), b, nil, "IsInf", []any{math.Inf(1)}); got[0] != true {
		t.Fatalf("IsInf(+Inf) = %v", got)
	}
	if got := call(t, new(vm.Engine), b, nil, "IsInf", []any{1.0}); got[0] != false {
		t.Fatalf("IsInf(1) = %v", got)
	}
}

// TestLibraryValue uses functions of the library as values, also one, which creates an error.
func TestLibraryValue(t *testing.T) {
	b := compileSrc(t, `package s

import (
	"errors"
	"strings"
)

var newErr = errors.New

func apply(f func(string) string, s string) string { return f(s) }

func Trim(s string) string {
	f := strings.TrimSpace
	return apply(f, s) + apply(strings.TrimSpace, s+"y ")
}

func Fail(msg string) error { return newErr(msg) }
`)
	if got := call(t, new(vm.Engine), b, nil, "Trim", []any{" x "}); got[0] != "xx y" {
		t.Fatalf("Trim = %q", got)
	}
	if got := call(t, new(vm.Engine), b, nil, "Fail", []any{"boom"}); gostrict.ErrorOf(got[0]) == nil || gostrict.ErrorOf(got[0]).Error() != "boom" {
		t.Fatalf("Fail = %v", got)
	}
}

// TestRunCallback ends a callback, which the host calls between two calls, with the context of Run.
func TestRunCallback(t *testing.T) {
	b := compileSrc(t, `package s

import "go.wdy.de/nago/pkg/gostrict/lib/ui"

func spin() {
	x := int32(0)
	for i := int32(0); i < 2000000000; i++ {
		x += i
	}
}

func Keep(slow bool) ui.AlertOption {
	if slow {
		return ui.AlertDelete(spin)
	}
	return ui.AlertDelete(func() {})
}
`)
	const keep = "go.wdy.de/nago/pkg/gostrict/lib/ui.AlertDelete"
	var kept gostrict.Callback
	h := gostrict.Funcs{}
	for name, f := range ui.Host() {
		h[name] = f
	}
	h[keep] = func(args []any) []any {
		kept = args[0].(gostrict.Callback)
		return ui.Host()[keep](args)
	}
	in, err := new(vm.Engine).Load(context.Background(), b, h)
	if err != nil {
		t.Fatal(err)
	}
	vin := in.(*vm.Instance)

	if _, err := in.Call(context.Background(), "Keep", false); err != nil {
		t.Fatal(err)
	}
	if err := vin.Run(context.Background(), func() { _, err = kept.Call() }); err != nil {
		t.Fatal(err)
	}

	if _, err := in.Call(context.Background(), "Keep", true); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	var cbErr error
	runErr := vin.Run(ctx, func() { _, cbErr = kept.Call() })
	if !errors.Is(cbErr, context.DeadlineExceeded) || !errors.Is(runErr, context.DeadlineExceeded) {
		t.Fatalf("the context did not end the callback: %v, %v", cbErr, runErr)
	}
	// Run has reported the error, not the next call, and the instance is usable
	if _, err := in.Call(context.Background(), "Keep", false); err != nil {
		t.Fatal(err)
	}
	if err := vin.Run(context.Background(), func() { _, err = kept.Call() }); err != nil {
		t.Fatal(err)
	}
}

// TestCancelledInit loads a bundle, whose first initialization of the package variables the context ends.
func TestCancelledInit(t *testing.T) {
	b := compileSrc(t, `package s

var total = sum()

func sum() int32 {
	s := int32(0)
	for i := int32(0); i < 3000000; i++ {
		s += i % 7
	}
	return s
}

func Total() int32 { return total }
`)
	e := new(vm.Engine)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := e.Load(ctx, b, nil); !errors.Is(err, context.Canceled) {
		t.Fatalf("Load with an ended context = %v", err)
	}
	if got := call(t, e, b, nil, "Total", nil); got[0] != int32(8999994) {
		t.Fatalf("Total = %v", got)
	}
}

// TestErrorType tells a failure of the program from an error of the API.
func TestErrorType(t *testing.T) {
	b := compileSrc(t, `package s

import "go.wdy.de/nago/pkg/gostrict/lib/ui"

func Alert() ui.AlertOption { return ui.AlertDelete(func() {}) }
`)
	broken := errors.New("broken")
	h := gostrict.Funcs{}
	for name, f := range ui.Host() {
		h[name] = f
	}
	h["go.wdy.de/nago/pkg/gostrict/lib/ui.AlertDelete"] = func(args []any) []any { panic(broken) }
	in, err := new(vm.Engine).Load(context.Background(), b, h)
	if err != nil {
		t.Fatal(err)
	}
	var ve *vm.Error
	if _, err := in.Call(context.Background(), "Alert"); !errors.As(err, &ve) || !errors.Is(err, broken) {
		t.Fatalf("the failure of the program is %T: %v", err, err)
	}
	if _, err := in.Call(context.Background(), "Nope"); err == nil || errors.As(err, &ve) {
		t.Fatalf("the error of the API is %T: %v", err, err)
	}
	if _, err := in.Call(context.Background(), "Alert", "x"); err == nil || errors.As(err, &ve) {
		t.Fatalf("the error of the API is %T: %v", err, err)
	}
}
