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
	"strings"
	"testing"

	"github.com/dop251/goja"
	"go.wdy.de/nago/pkg/gostrict"
	"go.wdy.de/nago/pkg/gostrict/engine/native"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/spin"
	"go.wdy.de/nago/pkg/gostrict/lib/three"
)

// TestThree runs a three.js program natively and in goja with the native scene graph as host, which receives the
// handles back through goja, and compares the scenes.
func TestThree(t *testing.T) {
	b := compile(t, "internal/examples/spin")
	run := func(e gostrict.Engine, host gostrict.Funcs) string {
		in, err := e.Load(context.Background(), b, host)
		if err != nil {
			t.Fatal(err)
		}
		client := spin.Client{Instance: in}
		n, err := client.Setup(context.Background(), 16.0/9)
		if err != nil {
			t.Fatal(err)
		}
		if n != 13 {
			t.Fatalf("got %d meshes", n)
		}
		if err := client.Frame(context.Background(), 1.5); err != nil {
			t.Fatal(err)
		}
		return three.Dump()
	}

	want := run(native.Engine{Exports: spin.Exports()}, nil)
	got := run(goEngine, three.Host())
	if got != want {
		t.Fatalf("the scenes differ\nnative:\n%s\ngoja:\n%s", want, got)
	}
	if !strings.Contains(want, "Mesh pos=[") || strings.Count(want, "\n") != 16 {
		t.Fatalf("unexpected scene:\n%s", want)
	}
	t.Log("\n" + want)
}

// fakeThree is a minimal three.js in goja: its objects keep, what the native scene graph of lib/three keeps.
const fakeThree = `(function () {
	class Vec { constructor(x, y, z) { this.x = x; this.y = y; this.z = z; } set(x, y, z) { this.x = x; this.y = y; this.z = z; } setScalar(s) { this.set(s, s, s); } }
	class Color { constructor(c) { this.hex = c; } setHex(c) { this.hex = c; } }
	class Object3D {
		constructor(kind) { this.kind = kind; this.children = []; this.position = new Vec(0, 0, 0); this.rotation = new Vec(0, 0, 0); this.scale = new Vec(1, 1, 1); this.name = ''; this.visible = true; }
		add(c) { this.children.push(c); }
		remove(c) { this.children = this.children.filter((x) => x !== c); }
		lookAt() {}
	}
	const THREE = {
		Scene: class extends Object3D { constructor() { super('Scene'); } },
		Color,
		PerspectiveCamera: class extends Object3D { constructor() { super('Camera'); } updateProjectionMatrix() {} },
		Group: class extends Object3D { constructor() { super('Group'); } },
		Mesh: class extends Object3D { constructor(g, m) { super('Mesh'); this.material = m; } },
		DirectionalLight: class extends Object3D { constructor() { super('DirectionalLight'); } },
		AmbientLight: class extends Object3D { constructor() { super('AmbientLight'); } },
		BoxGeometry: class {}, SphereGeometry: class {}, TorusKnotGeometry: class {},
		MeshStandardMaterial: class { constructor(o) { this.color = new Color(o.color); } },
	};
	let shown = null;
	const dump = (o) => [o.kind, [o.position.x, o.position.y, o.position.z], [o.rotation.x, o.rotation.y, o.rotation.z], o.scale.x, o.material ? o.material.color.hex : -1, o.name, !o.visible, o.children.map(dump)];
	return { THREE, host: { show(s) { shown = s; } }, dump: () => dump(shown) };
})()`

// TestThreeBindings runs the program with the bindings of lib/three in goja against a fake three.js and compares
// the scene with the native scene graph.
func TestThreeBindings(t *testing.T) {
	b := compile(t, "internal/examples/spin")
	ctx := context.Background()

	// goja with the native implementation builds a fresh native scene, while the package variables of the native
	// engine are shared with other tests
	in, err := goEngine.Load(ctx, b, three.Host())
	if err != nil {
		t.Fatal(err)
	}
	client := spin.Client{Instance: in}
	if _, err := client.Setup(ctx, 16.0/9); err != nil {
		t.Fatal(err)
	}
	if err := client.Frame(ctx, 1.5); err != nil {
		t.Fatal(err)
	}
	want := three.Dump()

	vm := goja.New()
	fake, err := vm.RunString(fakeThree)
	if err != nil {
		t.Fatal(err)
	}
	fo := fake.ToObject(vm)
	if err := vm.Set("THREE", fo.Get("THREE")); err != nil {
		t.Fatal(err)
	}
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
	setup, _ := goja.AssertFunction(eo.Get("Setup"))
	frame, _ := goja.AssertFunction(eo.Get("Frame"))
	if _, err := setup(goja.Undefined(), vm.ToValue(16.0/9)); err != nil {
		t.Fatal(err)
	}
	if _, err := frame(goja.Undefined(), vm.ToValue(1.5)); err != nil {
		t.Fatal(err)
	}
	dumpFn, _ := goja.AssertFunction(fo.Get("dump"))
	d, err := dumpFn(goja.Undefined())
	if err != nil {
		t.Fatal(err)
	}

	var sb strings.Builder
	dumpFake(&sb, d.Export().([]any), 0)
	if got := sb.String(); got != want {
		t.Fatalf("the scenes differ\nnative:\n%s\nbindings:\n%s", want, got)
	}
}

// dumpFake formats a node of the fake three.js like three.Dump.
func dumpFake(sb *strings.Builder, n []any, depth int) {
	num := func(v any) float64 {
		switch v := v.(type) {
		case int64:
			return float64(v)
		case float64:
			return v
		}
		return 0
	}
	vec := func(v any) [3]float64 {
		l := v.([]any)
		return [3]float64{num(l[0]), num(l[1]), num(l[2])}
	}
	fmt.Fprintf(sb, "%s%s pos=%.2f rot=%.2f scale=%.2f", strings.Repeat("  ", depth), n[0], vec(n[1]), vec(n[2]), num(n[3]))
	if c := num(n[4]); c >= 0 {
		fmt.Fprintf(sb, " color=%06x", uint32(c))
	}
	if name := n[5].(string); name != "" {
		fmt.Fprintf(sb, " name=%q", name)
	}
	if n[6].(bool) {
		sb.WriteString(" hidden")
	}
	sb.WriteByte('\n')
	for _, c := range n[7].([]any) {
		dumpFake(sb, c.([]any), depth+1)
	}
}
