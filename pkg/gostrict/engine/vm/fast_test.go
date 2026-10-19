// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package vm_test

import (
	"context"
	"testing"

	"go.wdy.de/nago/pkg/gostrict"
	"go.wdy.de/nago/pkg/gostrict/engine/vm"
)

// writeRec writes the neutral value of a fastrec.Rec with the writer of a fast call.
func writeRec(w gostrict.Writer, v []any) {
	w.Record()
	w.String(v[0].(string))
	tags := v[1].([]any)
	w.List(len(tags))
	for _, t := range tags {
		w.String(t.(string))
	}
	w.End()
	scores := v[2].([]any)
	w.Map(len(scores))
	for _, e := range scores {
		kv := e.([]any)
		w.String(kv[0].(string))
		w.Int32(kv[1].(int32))
	}
	w.End()
	w.Big(v[3].(string))
	w.Record()
	w.Float64(v[4].([]any)[0].(float64))
	w.End()
	w.Record()
	for _, s := range v[5].([]any) {
		w.String(s.(string))
	}
	w.End()
	items := v[6].([]any)
	w.List(len(items))
	for _, it := range items {
		w.Record()
		w.Float64(it.([]any)[0].(float64))
		w.End()
	}
	w.End()
	w.Int32(v[7].(int32))
	w.Uint32(v[8].(uint32))
	w.Float64(v[9].(float64))
	w.Bool(v[10].(bool))
	w.End()
}

// readRec reads a fastrec.Rec from the reader of a fast call as its neutral value.
func readRec(r gostrict.Reader) []any {
	if kept, ok := r.Record(); ok {
		return kept.([]any)
	}
	var v []any
	v = append(v, r.String())
	var tags []any
	for range r.List() {
		tags = append(tags, r.String())
	}
	r.End()
	v = append(v, nonNil(tags))
	var scores []any
	for range r.Map() {
		k := r.String()
		scores = append(scores, []any{k, r.Int32()})
	}
	r.End()
	v = append(v, nonNil(scores), r.Big())
	if _, ok := r.Record(); ok {
		panic("Inner changes, it is not kept")
	}
	v = append(v, []any{r.Float64()})
	r.End()
	look, ok := r.Record()
	if !ok {
		var fs []any
		for range 9 {
			fs = append(fs, r.String())
		}
		if r.End() {
			r.Keep(fs)
		}
		look = fs
	}
	v = append(v, look)
	var items []any
	for range r.List() {
		r.Record()
		items = append(items, []any{r.Float64()})
		r.End()
	}
	r.End()
	v = append(v, nonNil(items), r.Int32(), r.Uint32(), r.Float64(), r.Bool())
	r.End()
	return v
}

func nonNil(v []any) []any {
	if v == nil {
		return []any{}
	}
	return v
}

// TestFastCall calls a function of the program with the values of each kind once with neutral values and once
// with the reader and writer of a fast call, which must agree.
func TestFastCall(t *testing.T) {
	b := compile(t, "../../testdata/headers/fastuse")
	look := []any{"1", "2", "3", "4", "5", "6", "7", "8", "9"}
	rec := []any{"rec", []any{"a", "b"}, []any{[]any{"k", int32(1)}}, "123456789012345678901234567890", []any{2.5}, look, []any{[]any{1.0}}, int32(4), uint32(1), 2.0, true}
	for _, fail := range []bool{false, true} {
		var neutral, fast []any
		var fastErr error
		host := gostrict.Funcs{"go.wdy.de/nago/pkg/gostrict/testdata/headers/fast.Apply": func(args []any) []any {
			cb := args[1].(gostrict.Callback)
			res, err := cb.Call(args[0], int32(1))
			if err != nil {
				t.Fatal(err)
			}
			neutral = res
			fc, ok := cb.(gostrict.FastCallback)
			if !ok {
				t.Fatal("the callback of the vm is not fast")
			}
			w := fc.Begin()
			writeRec(w, args[0].([]any))
			w.Int32(1)
			r, err := fc.Invoke()
			if err != nil {
				t.Fatal(err)
			}
			fast = []any{readRec(r)}
			fastErr = r.Error()
			fast = append(fast, fastErr)
			fc.End()
			return neutral
		}}
		in, err := new(vm.Engine).Load(context.Background(), b, host)
		if err != nil {
			t.Fatal(err)
		}
		res, err := in.Call(context.Background(), "Run", rec, fail)
		if err != nil {
			t.Fatal(err)
		}
		if !gostrict.SameValues(neutral, fast) {
			t.Fatalf("fail=%v: the fast call differs\n  neutral: %#v\n  fast:    %#v", fail, neutral, fast)
		}
		t.Logf("fail=%v: %v", fail, res)
		if fail && res[1] != "failed rec" || !fail && res[1] != "ok" {
			t.Fatalf("unexpected result %v", res)
		}
	}
}
