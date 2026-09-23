// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package nprotoc

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// varintCases are the values both sides must agree on. They cover the zigzag boundaries, the 7 bit group
// boundaries and values beyond 32 bit, which JavaScript bitwise operators silently truncate.
var varintCases = []int64{
	0, 1, -1, 2, -2, 63, -64, 64, -65, 127, 128, 250, 403, 500, -500,
	1<<31 - 1, -(1 << 31), 1 << 31, 1 << 32, -(1 << 32), 1 << 40, -(1 << 40),
	1<<52 - 1, -(1<<52 - 1),
}

var uvarintCases = []uint64{
	0, 1, 127, 128, 16383, 16384, 1<<31 - 1, 1 << 31, 1<<32 - 1, 1 << 32, 1 << 40, 1<<53 - 1,
}

// runTSBinary transpiles ts_binary.ts with the typescript compiler of the frontend and runs script against it.
// The test is skipped if node or the frontend dependencies are not installed.
func runTSBinary(t *testing.T, script string) string {
	t.Helper()

	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node not installed")
	}

	tsDir, err := filepath.Abs("../../web/vuejs/node_modules/typescript")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(tsDir); err != nil {
		t.Skip("frontend dependencies not installed (npm install in web/vuejs)")
	}

	dir := t.TempDir()
	runner := `
const ts = require(` + jsString(tsDir) + `);
const fs = require('fs');
const src = fs.readFileSync(` + jsString(mustAbs(t, "ts_binary.ts")) + `, 'utf8');
const js = ts.transpileModule(src, { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2020 } }).outputText;
const mod = { exports: {} };
new Function('module', 'exports', 'require', js)(mod, mod.exports, require);
const { BinaryWriter, BinaryReader } = mod.exports;
` + script

	file := filepath.Join(dir, "run.cjs")
	if err := os.WriteFile(file, []byte(runner), 0o600); err != nil {
		t.Fatal(err)
	}

	out, err := exec.Command(node, file).CombinedOutput()
	if err != nil {
		t.Fatalf("node failed: %v\n%s", err, out)
	}

	return strings.TrimSpace(string(out))
}

func mustAbs(t *testing.T, p string) string {
	a, err := filepath.Abs(p)
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func jsString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func jsonNumbers[T int64 | uint64](v []T) string {
	b, _ := json.Marshal(v)
	return string(b)
}

// The frontend must encode signed values exactly like encoding/binary.PutVarint, otherwise the backend reads
// garbage: without zigzag a 500 arrives as 250 and -1 does not arrive at all.
func TestTSWriteVarintMatchesGo(t *testing.T) {
	out := runTSBinary(t, `
const w = new BinaryWriter();
for (const v of `+jsonNumbers(varintCases)+`) w.writeVarint(v);
console.log(Buffer.from(w.getBuffer()).toString('hex'));
`)

	buf, err := hex.DecodeString(out)
	if err != nil {
		t.Fatalf("unexpected output %q: %v", out, err)
	}

	r := bytes.NewReader(buf)
	for _, want := range varintCases {
		got, err := binary.ReadVarint(r)
		if err != nil {
			t.Fatalf("cannot read %d: %v", want, err)
		}
		if got != want {
			t.Errorf("writeVarint(%d) decoded by Go as %d", want, got)
		}
	}
}

func TestTSReadVarintMatchesGo(t *testing.T) {
	var buf []byte
	for _, v := range varintCases {
		buf = binary.AppendVarint(buf, v)
	}

	out := runTSBinary(t, `
const r = new BinaryReader(Buffer.from(`+jsString(hex.EncodeToString(buf))+`, 'hex'));
const res = [];
for (let i = 0; i < `+strconv.Itoa(len(varintCases))+`; i++) res.push(r.readVarint());
console.log(JSON.stringify(res));
`)

	var got []int64
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("unexpected output %q: %v", out, err)
	}

	for i, want := range varintCases {
		if got[i] != want {
			t.Errorf("readVarint of Go %d returned %d", want, got[i])
		}
	}
}

func TestTSUvarintRoundTripsWithGo(t *testing.T) {
	var buf []byte
	for _, v := range uvarintCases {
		buf = binary.AppendUvarint(buf, v)
	}

	out := runTSBinary(t, `
const w = new BinaryWriter();
for (const v of `+jsonNumbers(uvarintCases)+`) w.writeUvarint(v);
const r = new BinaryReader(Buffer.from(`+jsString(hex.EncodeToString(buf))+`, 'hex'));
const res = [];
for (let i = 0; i < `+strconv.Itoa(len(uvarintCases))+`; i++) res.push(r.readUvarint());
console.log(Buffer.from(w.getBuffer()).toString('hex') + ' ' + JSON.stringify(res));
`)

	written, read, ok := strings.Cut(out, " ")
	if !ok {
		t.Fatalf("unexpected output %q", out)
	}

	if written != hex.EncodeToString(buf) {
		t.Errorf("writeUvarint differs from Go:\n got %s\nwant %s", written, hex.EncodeToString(buf))
	}

	var got []uint64
	if err := json.Unmarshal([]byte(read), &got); err != nil {
		t.Fatalf("unexpected output %q: %v", read, err)
	}

	for i, want := range uvarintCases {
		if got[i] != want {
			t.Errorf("readUvarint of Go %d returned %d", want, got[i])
		}
	}
}
