// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gollama

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"
)

// GGUF metadata value types as defined by the GGUF specification
// (https://github.com/ggml-org/ggml/blob/master/docs/gguf.md).
const (
	ggufTypeUint8 uint32 = iota
	ggufTypeInt8
	ggufTypeUint16
	ggufTypeInt16
	ggufTypeUint32
	ggufTypeInt32
	ggufTypeFloat32
	ggufTypeBool
	ggufTypeString
	ggufTypeArray
	ggufTypeUint64
	ggufTypeInt64
	ggufTypeFloat64
)

// maxGGUFString caps the length of a single GGUF string to guard against corrupt/hostile files.
const maxGGUFString = 64 << 20 // 64 MiB

// ggufMetadata holds the subset of GGUF header metadata the provider needs to pick the right family adapter,
// detokenizer and stop tokens. It is extracted without reading any tensor data.
type ggufMetadata struct {
	architecture   string // general.architecture, e.g. "qwen2", "llama", "gemma2"
	name           string // general.name
	tokenizerModel string // tokenizer.ggml.model, e.g. "gpt2", "llama", "bert"
	chatTemplate   string // tokenizer.ggml.chat_template (Jinja); used only as a fingerprint

	// contextLength is the model's trained context window ({arch}.context_length). Because the binding cannot
	// safely set llama_context_params.n_ctx (see engine.go), the context is always created with n_ctx=0
	// ("from model"), so this value is the actual physical KV size and the upper bound for prompt+generation.
	// 0 means the key was absent.
	contextLength int

	bosTokenID int
	eosTokenID int
	eotTokenID int
	hasBOS     bool
	hasEOS     bool
	hasEOT     bool
}

// readGGUFMetadata parses just the metadata key/value section of a GGUF file. Tensor information and tensor
// data are never read, so this is cheap even for multi-gigabyte models.
func readGGUFMetadata(path string) (ggufMetadata, error) {
	f, err := os.Open(path)
	if err != nil {
		return ggufMetadata{}, err
	}
	defer func() { _ = f.Close() }()

	r := &ggufReader{r: bufio.NewReaderSize(f, 1<<20)}

	var magic [4]byte
	if _, err := io.ReadFull(r.r, magic[:]); err != nil {
		return ggufMetadata{}, fmt.Errorf("read gguf magic: %w", err)
	}
	if string(magic[:]) != "GGUF" {
		return ggufMetadata{}, fmt.Errorf("not a gguf file: bad magic %q", magic[:])
	}

	version, err := r.u32()
	if err != nil {
		return ggufMetadata{}, err
	}
	r.version = version

	if _, err := r.u64(); err != nil { // tensor_count (unused)
		return ggufMetadata{}, err
	}
	kvCount, err := r.u64()
	if err != nil {
		return ggufMetadata{}, err
	}
	if kvCount > 1<<20 {
		return ggufMetadata{}, fmt.Errorf("implausible gguf metadata count %d", kvCount)
	}

	kv := make(map[string]any, kvCount)
	for i := uint64(0); i < kvCount; i++ {
		key, err := r.str()
		if err != nil {
			return ggufMetadata{}, fmt.Errorf("read gguf key #%d: %w", i, err)
		}
		valueType, err := r.u32()
		if err != nil {
			return ggufMetadata{}, fmt.Errorf("read gguf value type for %q: %w", key, err)
		}
		val, err := r.value(valueType)
		if err != nil {
			return ggufMetadata{}, fmt.Errorf("read gguf value for %q: %w", key, err)
		}
		if val != nil { // arrays are consumed but not stored
			kv[key] = val
		}
	}

	arch := asString(kv["general.architecture"])
	return ggufMetadata{
		architecture:   arch,
		name:           asString(kv["general.name"]),
		tokenizerModel: asString(kv["tokenizer.ggml.model"]),
		chatTemplate:   asString(kv["tokenizer.ggml.chat_template"]),
		contextLength:  asInt(kv[arch+".context_length"]),
		bosTokenID:     asInt(kv["tokenizer.ggml.bos_token_id"]),
		eosTokenID:     asInt(kv["tokenizer.ggml.eos_token_id"]),
		eotTokenID:     asInt(kv["tokenizer.ggml.eot_token_id"]),
		hasBOS:         hasKey(kv, "tokenizer.ggml.bos_token_id"),
		hasEOS:         hasKey(kv, "tokenizer.ggml.eos_token_id"),
		hasEOT:         hasKey(kv, "tokenizer.ggml.eot_token_id"),
	}, nil
}

func hasKey(kv map[string]any, key string) bool {
	_, ok := kv[key]
	return ok
}

func asString(v any) string {
	s, _ := v.(string)
	return s
}

// asInt coerces a stored GGUF scalar into an int.
func asInt(v any) int {
	switch n := v.(type) {
	case int64:
		return int(n)
	case uint64:
		return int(n)
	case int32:
		return int(n)
	case uint32:
		return int(n)
	case int16:
		return int(n)
	case uint16:
		return int(n)
	case int8:
		return int(n)
	case uint8:
		return int(n)
	default:
		return 0
	}
}

// ggufReader reads little-endian GGUF primitives from a buffered stream.
type ggufReader struct {
	r       *bufio.Reader
	version uint32
}

func (g *ggufReader) u32() (uint32, error) {
	var b [4]byte
	if _, err := io.ReadFull(g.r, b[:]); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(b[:]), nil
}

func (g *ggufReader) u64() (uint64, error) {
	var b [8]byte
	if _, err := io.ReadFull(g.r, b[:]); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(b[:]), nil
}

// strLen reads the length prefix of a GGUF string or array. GGUF v1 used 32-bit lengths; v2+ uses 64-bit.
func (g *ggufReader) strLen() (uint64, error) {
	if g.version == 1 {
		n, err := g.u32()
		return uint64(n), err
	}
	return g.u64()
}

func (g *ggufReader) str() (string, error) {
	n, err := g.strLen()
	if err != nil {
		return "", err
	}
	if n > maxGGUFString {
		return "", fmt.Errorf("gguf string too long: %d", n)
	}
	buf := make([]byte, n)
	if _, err := io.ReadFull(g.r, buf); err != nil {
		return "", err
	}
	return string(buf), nil
}

// value reads a single metadata value. Scalars and strings are returned as Go values; arrays are fully
// consumed (so the stream stays aligned) but returned as nil because the provider does not need them.
func (g *ggufReader) value(valueType uint32) (any, error) {
	switch valueType {
	case ggufTypeUint8:
		b, err := g.r.ReadByte()
		return b, err
	case ggufTypeInt8:
		b, err := g.r.ReadByte()
		return int8(b), err
	case ggufTypeBool:
		b, err := g.r.ReadByte()
		return b != 0, err
	case ggufTypeUint16:
		var b [2]byte
		_, err := io.ReadFull(g.r, b[:])
		return binary.LittleEndian.Uint16(b[:]), err
	case ggufTypeInt16:
		var b [2]byte
		_, err := io.ReadFull(g.r, b[:])
		return int16(binary.LittleEndian.Uint16(b[:])), err
	case ggufTypeUint32:
		v, err := g.u32()
		return v, err
	case ggufTypeInt32:
		v, err := g.u32()
		return int32(v), err
	case ggufTypeFloat32:
		v, err := g.u32()
		return v, err // raw bits; not needed as float
	case ggufTypeUint64:
		v, err := g.u64()
		return v, err
	case ggufTypeInt64:
		v, err := g.u64()
		return int64(v), err
	case ggufTypeFloat64:
		v, err := g.u64()
		return v, err
	case ggufTypeString:
		return g.str()
	case ggufTypeArray:
		return nil, g.skipArray()
	default:
		return nil, fmt.Errorf("unknown gguf value type %d", valueType)
	}
}

// skipArray consumes an array value without retaining it.
func (g *ggufReader) skipArray() error {
	elemType, err := g.u32()
	if err != nil {
		return err
	}
	count, err := g.u64()
	if err != nil {
		return err
	}

	if size, fixed := ggufFixedSize(elemType); fixed {
		return g.discard(uint64(size) * count)
	}

	switch elemType {
	case ggufTypeString:
		for i := uint64(0); i < count; i++ {
			n, err := g.strLen()
			if err != nil {
				return err
			}
			if err := g.discard(n); err != nil {
				return err
			}
		}
		return nil
	case ggufTypeArray:
		for i := uint64(0); i < count; i++ {
			if err := g.skipArray(); err != nil {
				return err
			}
		}
		return nil
	default:
		return fmt.Errorf("unknown gguf array element type %d", elemType)
	}
}

// discard skips n bytes from the stream in bounded chunks.
func (g *ggufReader) discard(n uint64) error {
	for n > 0 {
		chunk := n
		const max = 1 << 30
		if chunk > max {
			chunk = max
		}
		skipped, err := g.r.Discard(int(chunk))
		if err != nil {
			return err
		}
		n -= uint64(skipped)
	}
	return nil
}

// ggufFixedSize returns the byte size of fixed-width GGUF element types.
func ggufFixedSize(elemType uint32) (int, bool) {
	switch elemType {
	case ggufTypeUint8, ggufTypeInt8, ggufTypeBool:
		return 1, true
	case ggufTypeUint16, ggufTypeInt16:
		return 2, true
	case ggufTypeUint32, ggufTypeInt32, ggufTypeFloat32:
		return 4, true
	case ggufTypeUint64, ggufTypeInt64, ggufTypeFloat64:
		return 8, true
	default:
		return 0, false
	}
}

// isEOG reports whether the token id is an end-of-generation token according to the GGUF metadata. Only the
// primary EOS and EOT ids are known here; additional end markers (e.g. Llama 3's <|eot_id|>/<|eom_id|> or
// Qwen's <|endoftext|>) are caught textually via the adapter's stop strings.
func (m ggufMetadata) isEOG(tok int) bool {
	if m.hasEOS && tok == m.eosTokenID {
		return true
	}
	if m.hasEOT && tok == m.eotTokenID {
		return true
	}
	return false
}

// isByteLevelBPE reports whether the tokenizer uses GPT-2/byte-level BPE detokenization (as opposed to
// SentencePiece). This drives detokenizer selection (see detok.go).
func (m ggufMetadata) isByteLevelBPE() bool {
	switch strings.ToLower(m.tokenizerModel) {
	case "gpt2", "bpe", "tekken":
		return true
	case "llama", "spm", "t5", "rwkv":
		return false
	default:
		// Llama 3 reports "gpt2"; most others without a clear marker are byte-level nowadays. Default to
		// byte-level BPE which is the most common case for current instruct GGUFs.
		return true
	}
}
