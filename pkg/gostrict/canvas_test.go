// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict_test

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/dop251/goja"
	"go.wdy.de/nago/pkg/gostrict"
	"go.wdy.de/nago/pkg/gostrict/engine/native"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/draw"
	"go.wdy.de/nago/pkg/gostrict/lib/canvas"
)

// fakeCanvas records the calls of the bindings of lib/canvas in goja like the native implementation, so that a
// test compares them. Objects are named after their kind and the order of their creation.
const fakeCanvas = `(function () {
	const log = [];
	let ids = 0;
	const named = (kind) => ({ name: kind + '#' + (++ids) });
	const arg = (a) => (a !== null && typeof a === 'object' && !Array.isArray(a) && a.name !== undefined ? a.name : a);
	const record = (target, name, args) => { log.push([target, name, Array.from(args).map(arg)]); };
	let font = '10px sans-serif';
	const gradient = (kind) => {
		const g = named(kind);
		g.addColorStop = (o, c) => record(g.name, 'addColorStop', [o, c]);
		return g;
	};
	const ctx = new Proxy({}, {
		get(t, p) {
			switch (p) {
				case 'canvas': return { width: 300, height: 150 };
				case 'createLinearGradient': return () => gradient('linear');
				case 'createRadialGradient': return () => gradient('radial');
				case 'createConicGradient': return () => gradient('conic');
				case 'createPattern': return () => named('pattern');
				case 'isPointInPath': case 'isPointInStroke': return () => false;
				case 'measureText': return (s) => {
					const m = /([0-9.]+)px/.exec(font);
					const px = m ? Number(m[1]) : 10;
					return { width: Array.from(s).length * px * 0.6, actualBoundingBoxAscent: px * 0.8, actualBoundingBoxDescent: px * 0.2 };
				};
			}
			return function () { record('ctx', p, arguments); };
		},
		set(t, p, v) {
			if (p === 'font') font = v;
			record('ctx', p, [v]);
			return true;
		},
	});
	globalThis.Path2D = function (d) {
		const o = named(d === undefined ? 'path' : 'path ' + d);
		return new Proxy(o, { get: (t, p) => (p === 'name' ? t.name : function () { record(t.name, p, arguments); }) });
	};
	const host = {
		context: ctx,
		dpr: 1,
		width: 300,
		height: 150,
		loadImage(url) {
			const o = named('image ' + url);
			o.bitmap = { name: o.name };
			o.width = 0;
			o.height = 0;
			return o;
		},
	};
	return { host, log };
})()`

// TestCanvas draws natively, in goja with the native implementation and in goja with the JavaScript bindings and
// a fake context, and compares the commands.
func TestCanvas(t *testing.T) {
	b := compile(t, "internal/examples/draw")
	ctx := context.Background()

	run := func(e gostrict.Engine, host gostrict.Funcs) []string {
		canvas.Reset()
		in, err := e.Load(ctx, b, host)
		if err != nil {
			t.Fatal(err)
		}
		client := draw.Client{Instance: in}
		s, err := client.New(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if err := client.Frame(ctx, s, 1.5); err != nil {
			t.Fatal(err)
		}
		return canvas.Commands()
	}

	want := run(native.Engine{Exports: draw.Exports()}, nil)
	got := run(goEngine, canvas.Host())
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("the commands differ\nnative:\n%s\ngoja:\n%s", strings.Join(want, "\n"), strings.Join(got, "\n"))
	}

	bound := runBindings(t, b)
	if strings.Join(bound, "\n") != strings.Join(want, "\n") {
		t.Fatalf("the bindings differ from the native implementation\nnative:\n%s\nbindings:\n%s", strings.Join(want, "\n"), strings.Join(bound, "\n"))
	}
	t.Logf("%d commands, e.g.\n%s", len(want), strings.Join(want[:8], "\n"))
	if len(want) < 60 {
		t.Fatalf("too few commands: %d", len(want))
	}
}

// runBindings evaluates the bundle in goja with the fake context as host and returns the recorded calls in the
// format of the native implementation.
func runBindings(t *testing.T, b *gostrict.Bundle) []string {
	vm := goja.New()
	fake, err := vm.RunString(fakeCanvas)
	if err != nil {
		t.Fatal(err)
	}
	fo := fake.ToObject(vm)
	bundle, err := vm.RunString(b.JS)
	if err != nil {
		t.Fatal(err)
	}
	mk, _ := goja.AssertFunction(bundle)
	exports, err := mk(goja.Undefined(), fo.Get("host"))
	if err != nil {
		t.Fatal(err)
	}
	eo := exports.ToObject(vm)
	newFn, _ := goja.AssertFunction(eo.Get("New"))
	frameFn, _ := goja.AssertFunction(eo.Get("Frame"))
	s, err := newFn(goja.Undefined())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := frameFn(goja.Undefined(), s, vm.ToValue(1.5)); err != nil {
		t.Fatal(err)
	}

	var res []string
	for _, e := range fo.Get("log").Export().([]any) {
		entry := e.([]any)
		var args []string
		for _, a := range entry[2].([]any) {
			args = append(args, formatArg(a))
		}
		res = append(res, fmt.Sprintf("%s.%s(%s)", entry[0], entry[1], strings.Join(args, ", ")))
	}
	return res
}

// formatArg formats a value of goja like the native implementation of lib/canvas.
func formatArg(a any) string {
	switch a := a.(type) {
	case int64:
		return strconv.FormatFloat(float64(a), 'g', 6, 64)
	case float64:
		return strconv.FormatFloat(a, 'g', 6, 64)
	case string:
		return strconv.Quote(a)
	case []any:
		var fs []float64
		for _, v := range a {
			switch v := v.(type) {
			case int64:
				fs = append(fs, float64(v))
			case float64:
				fs = append(fs, v)
			}
		}
		if fs == nil {
			fs = []float64{}
		}
		return fmt.Sprint(fs)
	}
	return fmt.Sprint(a)
}
