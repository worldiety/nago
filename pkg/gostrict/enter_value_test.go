// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict_test

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	"go.wdy.de/nago/pkg/gostrict"
	"go.wdy.de/nago/pkg/gostrict/engine/gojaengine"
	"go.wdy.de/nago/pkg/gostrict/engine/vm"
)

// markHost is the fake host of whole_test.go, whose aspect denies the entry points in marked with the denial marked
// by gostrict.AsResult, and the entry points in denied with the plain denial, and whose host function guard.Reject
// ends its entry point with gostrict.ReturnError.
type markHost struct {
	*fakeHost
	marked map[string]bool
}

func (h *markHost) Enter(entry string) error {
	if name := short(entry); h.marked[name] {
		h.log = append(h.log, "enter "+name)
		return gostrict.AsResult(errDenied)
	}
	return h.fakeHost.Enter(entry)
}

func (h *markHost) funcs() gostrict.Funcs {
	f := gostrict.Funcs{}
	for k, v := range h.fakeHost.funcs() {
		if k != gostrict.EnterFunc && k != gostrict.ExitFunc {
			f[k] = v
		}
	}
	f[wholePrefix+"n10/guard.Reject"] = func(args []any) []any {
		gostrict.ReturnError(errors.New(args[0].(string)))
		return nil
	}
	return f.WithAspect(h)
}

// TestEnterValue is the acceptance test of the mark of Aspect.Enter and of gostrict.ReturnError (wokoda N10,
// R-SCRIPT-GS-CALL, L-N10-2), in goja and in the vm. Each row stands next to its counter row:
//   - an entry point with (T, error), which Enter denies with gostrict.AsResult, returns zero values, whose list is
//     empty, and the unmarked error to its caller, without its body and without Exit; the caller handles the error as
//     a value and returns; counter row: without the mark the body runs. Beside the error a script may not read the
//     zero values (wokoda R-SCRIPT-GS-OPTIONAL, N11), so the entry point n10/pass passes them on to the host, which
//     reads them;
//   - an error of Enter without the mark aborts the whole call, the caller can not handle it; counter row: the same
//     error with the mark is a value;
//   - the mark at an entry point, whose last result is no error, aborts with the unmarked error; counter row: the same
//     entry point without denial returns its result;
//   - gostrict.ReturnError in an entry point with (T, error) ends it after Exit with the error with zero values and the
//     error, which the caller handles; in an entry point without an error result it aborts the whole call, also when
//     an entry point with an error result calls that entry point (counter row: the same call of an entry point with
//     an error result is a value).
func TestEnterValue(t *testing.T) {
	for name, e := range map[string]gostrict.Engine{"goja": new(gojaengine.Engine), "vm": new(vm.Engine)} {
		t.Run(name, func(t *testing.T) { testEnterValue(t, e) })
	}
}

func testEnterValue(t *testing.T, e gostrict.Engine) {
	prog := loadWhole(t, "n10/caller", "n10/target", "n10/plain", "n10/callplain", "n10/callerplain", "n10/pass")
	b, diags := gostrict.Emit(prog)
	if len(diags) > 0 {
		t.Fatal(diags.Text())
	}
	h := &markHost{fakeHost: newFakeHost(), marked: map[string]bool{}}
	ctx := context.Background()
	in, err := e.Load(ctx, b, h.funcs())
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()

	for _, c := range []struct {
		name   string
		marked string // an entry point, which Enter denies with the mark
		denied string // an entry point, which Enter denies without the mark
		entry  string
		args   []any
		want   string // the first result; "" expects an abort with abortMsg
		// pass reads the results (target.Info, error) of n10/pass as "value <error> zero <results>" resp.
		// "result <results>", with an empty, not a missing list
		pass    bool
		abort   string
		log     []string
		present []string // keys, which the call committed
		absent  []string // keys, which it did not
	}{
		{
			name: "marked denial of an entry point with (T, error): zero values and the error as a value", marked: "n10/target.Main",
			entry: "n10/caller.Main", args: []any{"a", false}, want: "value permission denied",
			log:    []string{"enter n10/caller.Main", "enter n10/target.Main", "exit n10/caller.Main ok"},
			absent: []string{"target/a"},
		},
		{
			name: "marked denial of an entry point with (T, error): the zero values beside the error, passed on to the host", marked: "n10/target.Main",
			entry: "n10/pass.Main", args: []any{"a1", false}, want: "value permission denied zero /0/0", pass: true,
			absent: []string{"target/a1"},
		},
		{
			name:  "counter row: the same entry point without denial runs its body",
			entry: "n10/caller.Main", args: []any{"a2", false}, want: "result ran/7/1",
			log:     []string{"enter n10/caller.Main", "enter n10/target.Main", "exit n10/target.Main ok", "exit n10/caller.Main ok"},
			present: []string{"target/a2"},
		},
		{
			name:  "counter row: the same entry point without denial, passed on to the host",
			entry: "n10/pass.Main", args: []any{"a3", false}, want: "result ran/7/1", pass: true,
			present: []string{"target/a3"},
		},
		{
			name: "unmarked error of Enter aborts the whole call, the caller can not handle it", denied: "n10/target.Main",
			entry: "n10/caller.Main", args: []any{"b", false}, abort: "permission denied",
			log:    []string{"enter n10/caller.Main", "enter n10/target.Main", "exit n10/caller.Main aborted"},
			absent: []string{"target/b"},
		},
		{
			name: "marked denial of an entry point without an error result aborts with the unmarked error", marked: "n10/plain.Main",
			entry: "n10/callplain.Main", args: []any{false}, abort: "permission denied",
			log:    []string{"enter n10/callplain.Main", "enter n10/plain.Main", "exit n10/callplain.Main aborted"},
			absent: []string{"plain"},
		},
		{
			name:  "counter row: the entry point without an error result and without denial returns its result",
			entry: "n10/callplain.Main", args: []any{false}, want: "plain 1",
			log:     []string{"enter n10/callplain.Main", "enter n10/plain.Main", "exit n10/plain.Main ok", "exit n10/callplain.Main ok"},
			present: []string{"plain"},
		},
		{
			name:  "ReturnError in an entry point with (T, error): Exit with the error, zero values and the error as a value",
			entry: "n10/caller.Main", args: []any{"d", true}, want: "value rejected d",
			log:     []string{"enter n10/caller.Main", "enter n10/target.Main", "exit n10/target.Main rejected d", "exit n10/caller.Main ok"},
			present: []string{"target/d"}, absent: []string{"after/d"},
		},
		{
			name:  "ReturnError in an entry point with (T, error): the zero values beside the error, passed on to the host",
			entry: "n10/pass.Main", args: []any{"d1", true}, want: "value rejected d1 zero /0/0", pass: true,
			absent: []string{"after/d1"},
		},
		{
			name:  "ReturnError in an entry point without an error result aborts the whole call",
			entry: "n10/callplain.Main", args: []any{true}, abort: "rejected plain",
		},
		{
			name:  "ReturnError in an entry point without an error result, which an entry point with an error result calls, aborts the whole call",
			entry: "n10/callerplain.Main", args: []any{true}, abort: "rejected plain",
			absent: []string{"plain"},
		},
		{
			name:  "counter row: ReturnError in an entry point with an error result, which the same kind of entry point calls, is a value",
			entry: "n10/caller.Main", args: []any{"e", true}, want: "value rejected e",
			log:     []string{"enter n10/caller.Main", "enter n10/target.Main", "exit n10/target.Main rejected e", "exit n10/caller.Main ok"},
			present: []string{"target/e"}, absent: []string{"after/e"},
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			h.marked = map[string]bool{}
			h.denied = map[string]bool{}
			if c.marked != "" {
				h.marked[c.marked] = true
			}
			if c.denied != "" {
				h.denied[c.denied] = true
			}
			h.log, h.data = nil, map[string]string{}
			res, err := in.Call(ctx, wholePrefix+c.entry, c.args...)
			if c.pass && err == nil {
				res = []any{n10Passed(t, res)}
			}
			if c.want != "" {
				if err != nil || len(res) == 0 || res[0] != c.want {
					t.Fatalf("got %v, %v; want %q", res, err, c.want)
				}
			} else {
				if err == nil || !strings.Contains(err.Error(), c.abort) {
					t.Fatalf("got %v, %v; want an abort with %q", res, err, c.abort)
				}
				if c.abort == "permission denied" && !errors.Is(err, errDenied) {
					t.Errorf("the abort %v is not the denial of the aspect", err)
				}
				var r *gostrict.ResultError
				if errors.As(err, &r) {
					t.Errorf("the abort %v carries the mark", err)
				}
			}
			if c.log != nil && !slices.Equal(h.log, c.log) {
				t.Errorf("unexpected frames\n got: %q\nwant: %q", h.log, c.log)
			}
			if len(h.frames) != 0 {
				t.Errorf("%d frames are left", len(h.frames))
			}
			for _, k := range c.present {
				if _, ok := h.data[k]; !ok {
					t.Errorf("%s was not committed: %v", k, h.data)
				}
			}
			for _, k := range c.absent {
				if _, ok := h.data[k]; ok {
					t.Errorf("%s was committed: %v", k, h.data)
				}
			}
		})
	}

	// the instance works after the aborts
	h.marked, h.denied = map[string]bool{}, map[string]bool{}
	if res, err := in.Call(ctx, wholePrefix+"n10/caller.Main", "z", false); err != nil || res[0] != "result ran/7/1" {
		t.Fatalf("after the aborts: %v, %v", res, err)
	}
}

// n10Passed is the text of the results (target.Info, error) of n10/pass, as the host reads them: "value <error> zero
// <name>/<n>/<number of tags>" beside an error, else "result <name>/<n>/<number of tags>". A missing list of tags
// fails the test.
func n10Passed(t *testing.T, res []any) string {
	t.Helper()
	if len(res) != 2 {
		t.Fatalf("n10/pass returns %d results, want 2: %v", len(res), res)
	}
	rec, ok := res[0].([]any)
	if !ok || len(rec) != 3 {
		t.Fatalf("the first result of n10/pass is no record of three fields: %#v", res[0])
	}
	tags, ok := rec[2].([]any)
	if !ok || tags == nil {
		t.Fatalf("the list of the tags is missing: %#v", rec[2])
	}
	got := fmt.Sprintf("%v/%v/%d", rec[0], rec[1], len(tags))
	if err := gostrict.ErrorOf(res[1]); err != nil {
		return "value " + err.Error() + " zero " + got
	}
	return "result " + got
}
