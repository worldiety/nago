// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gollama

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

// ggufBuilder assembles a minimal but spec-conformant GGUF file in memory for tests. Only the metadata
// key/value section is populated (tensor_count is 0); that is all readGGUFMetadata consumes.
type ggufBuilder struct {
	t  *testing.T
	kv bytes.Buffer
	n  uint64
}

func newGGUFBuilder(t *testing.T) *ggufBuilder {
	t.Helper()
	return &ggufBuilder{t: t}
}

func (b *ggufBuilder) put(v any) {
	if err := binary.Write(&b.kv, binary.LittleEndian, v); err != nil {
		b.t.Fatalf("binary.Write %T: %v", v, err)
	}
}

func (b *ggufBuilder) putStr(s string) {
	b.put(uint64(len(s)))
	b.kv.WriteString(s)
}

func (b *ggufBuilder) String(key, val string) *ggufBuilder {
	b.putStr(key)
	b.put(ggufTypeString)
	b.putStr(val)
	b.n++
	return b
}

func (b *ggufBuilder) Uint32(key string, val uint32) *ggufBuilder {
	b.putStr(key)
	b.put(ggufTypeUint32)
	b.put(val)
	b.n++
	return b
}

// ArrayString writes a string array value. The reader must skip it without misaligning the stream.
func (b *ggufBuilder) ArrayString(key string, vals ...string) *ggufBuilder {
	b.putStr(key)
	b.put(ggufTypeArray)
	b.put(ggufTypeString)
	b.put(uint64(len(vals)))
	for _, v := range vals {
		b.putStr(v)
	}
	b.n++
	return b
}

// ArrayInt32 writes a fixed-width array value (exercises the fast discard path).
func (b *ggufBuilder) ArrayInt32(key string, vals ...int32) *ggufBuilder {
	b.putStr(key)
	b.put(ggufTypeArray)
	b.put(ggufTypeInt32)
	b.put(uint64(len(vals)))
	for _, v := range vals {
		b.put(v)
	}
	b.n++
	return b
}

func (b *ggufBuilder) write(version uint32) string {
	var buf bytes.Buffer
	buf.WriteString("GGUF")
	if err := binary.Write(&buf, binary.LittleEndian, version); err != nil {
		b.t.Fatal(err)
	}
	if err := binary.Write(&buf, binary.LittleEndian, uint64(0)); err != nil { // tensor_count
		b.t.Fatal(err)
	}
	if err := binary.Write(&buf, binary.LittleEndian, b.n); err != nil { // kv_count
		b.t.Fatal(err)
	}
	buf.Write(b.kv.Bytes())

	path := filepath.Join(b.t.TempDir(), "model.gguf")
	if err := os.WriteFile(path, buf.Bytes(), 0o600); err != nil {
		b.t.Fatal(err)
	}
	return path
}

func TestReadGGUFMetadataChatML(t *testing.T) {
	// Arrays are interleaved between scalars so that a parsing/alignment bug would corrupt the fields read
	// after them (name, model, template, token ids).
	path := newGGUFBuilder(t).
		String("general.architecture", "qwen2").
		ArrayString("tokenizer.ggml.tokens", "<|endoftext|>", "hello", "world").
		ArrayInt32("tokenizer.ggml.token_type", 3, 1, 1).
		String("general.name", "Qwen2.5 Coder 7B Instruct").
		String("tokenizer.ggml.model", "gpt2").
		String("tokenizer.ggml.chat_template", "{% for m in messages %}<|im_start|>{{ m.role }}\n{{ m.content }}<|im_end|>{% endfor %}").
		Uint32("tokenizer.ggml.bos_token_id", 151643).
		Uint32("tokenizer.ggml.eos_token_id", 151645).
		write(3)

	meta, err := readGGUFMetadata(path)
	if err != nil {
		t.Fatalf("readGGUFMetadata: %v", err)
	}

	if meta.architecture != "qwen2" {
		t.Errorf("architecture = %q, want qwen2", meta.architecture)
	}
	if meta.name != "Qwen2.5 Coder 7B Instruct" {
		t.Errorf("name = %q (array skipping likely misaligned the stream)", meta.name)
	}
	if meta.tokenizerModel != "gpt2" {
		t.Errorf("tokenizerModel = %q, want gpt2", meta.tokenizerModel)
	}
	if !meta.hasBOS || meta.bosTokenID != 151643 {
		t.Errorf("bos = (%v,%d), want (true,151643)", meta.hasBOS, meta.bosTokenID)
	}
	if !meta.hasEOS || meta.eosTokenID != 151645 {
		t.Errorf("eos = (%v,%d), want (true,151645)", meta.hasEOS, meta.eosTokenID)
	}
	if meta.hasEOT {
		t.Errorf("hasEOT = true, want false (no eot key written)")
	}

	if !meta.isEOG(151645) {
		t.Error("isEOG(eos) = false, want true")
	}
	if meta.isEOG(151643) {
		t.Error("isEOG(bos) = true, want false")
	}
	if meta.isEOG(0) {
		t.Error("isEOG(0) = true, want false (eot absent)")
	}
	if got := detectFamily(meta); got != familyChatML {
		t.Errorf("detectFamily = %q, want chatml", got)
	}
	if !meta.isByteLevelBPE() {
		t.Error("isByteLevelBPE() = false, want true for gpt2 tokenizer")
	}
}

func TestReadGGUFMetadataLlamaSPM(t *testing.T) {
	path := newGGUFBuilder(t).
		String("general.architecture", "llama").
		String("general.name", "Mistral 7B Instruct v0.3").
		String("tokenizer.ggml.model", "llama").
		Uint32("tokenizer.ggml.eos_token_id", 2).
		write(3)

	meta, err := readGGUFMetadata(path)
	if err != nil {
		t.Fatalf("readGGUFMetadata: %v", err)
	}
	if meta.isByteLevelBPE() {
		t.Error("isByteLevelBPE() = true, want false for llama/spm tokenizer")
	}
	if got := detectFamily(meta); got != familyMistral {
		t.Errorf("detectFamily = %q, want mistral", got)
	}
}

func TestReadGGUFMetadataBadMagic(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.gguf")
	if err := os.WriteFile(path, []byte("NOPExxxxxxxxxxxx"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := readGGUFMetadata(path); err == nil {
		t.Fatal("readGGUFMetadata accepted a non-GGUF file, want error")
	}
}
