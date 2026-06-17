// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gollama

import (
	"strconv"
	"strings"
	"unicode/utf8"
)

// pieceDecoder turns the raw vocabulary pieces returned by gollama.Token_to_piece into proper UTF-8 text.
//
// This abstraction exists because the gollama binding exposes only llama_vocab_get_text (not
// llama_token_to_piece / llama_detokenize), so the byte-level decoding has to be done in Go, and the scheme
// differs per tokenizer family (GPT-2/byte-level BPE vs SentencePiece). Decoders are stateful: a single
// multibyte rune (or a byte-fallback emoji) may be split across several tokens, so an incomplete trailing
// byte sequence is buffered until it completes.
type pieceDecoder interface {
	// decode appends one raw piece and returns any text that became complete.
	decode(piece string) string
	// flush returns any buffered remainder (called once after the final token).
	flush() string
}

// newPieceDecoder selects the decoder matching the model's tokenizer.
func newPieceDecoder(meta ggufMetadata) pieceDecoder {
	if meta.isByteLevelBPE() {
		return &bpeDecoder{}
	}
	return &spmDecoder{}
}

// ----- byte boundary buffer shared by both decoders -----

// byteEmitter accumulates raw bytes and only releases complete UTF-8 sequences, holding back an incomplete
// trailing multibyte sequence until the next bytes arrive.
type byteEmitter struct {
	buf []byte
}

func (e *byteEmitter) push(b []byte) string {
	e.buf = append(e.buf, b...)
	n := completeUTF8Prefix(e.buf)
	if n == 0 {
		return ""
	}
	out := string(e.buf[:n])
	e.buf = append(e.buf[:0], e.buf[n:]...)
	return out
}

func (e *byteEmitter) flush() string {
	if len(e.buf) == 0 {
		return ""
	}
	out := string(e.buf)
	e.buf = e.buf[:0]
	return out
}

// completeUTF8Prefix returns the length of the leading byte slice that contains only complete UTF-8 runes,
// holding back a trailing sequence that is a valid but incomplete prefix of a multibyte rune.
func completeUTF8Prefix(b []byte) int {
	if len(b) == 0 {
		return 0
	}
	max := 4
	if len(b) < max {
		max = len(b)
	}
	for i := 1; i <= max; i++ {
		c := b[len(b)-i]
		if utf8.RuneStart(c) { // not a continuation byte
			need := utf8ExpectedLen(c)
			if i < need {
				return len(b) - i // incomplete trailing rune: hold it back
			}
			return len(b)
		}
	}
	// only continuation bytes in the tail (malformed) -> emit everything to avoid stalling
	return len(b)
}

// utf8ExpectedLen returns the total length of the UTF-8 sequence started by lead byte c.
func utf8ExpectedLen(c byte) int {
	switch {
	case c < 0x80:
		return 1
	case c&0xE0 == 0xC0:
		return 2
	case c&0xF0 == 0xE0:
		return 3
	case c&0xF8 == 0xF0:
		return 4
	default:
		return 1 // invalid lead byte
	}
}

// ----- byte-level BPE (GPT-2 / Qwen / Llama 3) -----

// bpeDecoder reverses the GPT-2 byte-level BPE "bytes_to_unicode" mapping: printable substitute runes (e.g.
// 'Ġ' for a space) are mapped back to their original byte.
type bpeDecoder struct {
	byteEmitter
}

func (d *bpeDecoder) decode(piece string) string {
	b := make([]byte, 0, len(piece))
	for _, r := range piece {
		if by, ok := byteDecoder[r]; ok {
			b = append(b, by)
		} else {
			// rune is not part of the byte mapping (e.g. control token text) -> keep as-is
			b = append(b, []byte(string(r))...)
		}
	}
	return d.push(b)
}

// byteDecoder is the inverse of the GPT-2 / Qwen byte-level BPE bytes_to_unicode mapping.
var byteDecoder = buildByteDecoder()

func buildByteDecoder() map[rune]byte {
	// Identical to the reference implementation of GPT-2 bytes_to_unicode().
	var bs []int
	for b := int('!'); b <= int('~'); b++ {
		bs = append(bs, b)
	}
	for b := int('¡'); b <= int('¬'); b++ {
		bs = append(bs, b)
	}
	for b := int('®'); b <= int('ÿ'); b++ {
		bs = append(bs, b)
	}

	inBs := make(map[int]bool, len(bs))
	for _, b := range bs {
		inBs[b] = true
	}

	cs := append([]int(nil), bs...)
	n := 0
	for b := 0; b < 256; b++ {
		if !inBs[b] {
			bs = append(bs, b)
			cs = append(cs, 256+n)
			n++
		}
	}

	dec := make(map[rune]byte, 256)
	for i := range bs {
		dec[rune(cs[i])] = byte(bs[i])
	}
	return dec
}

// ----- SentencePiece (Llama 2 / Mistral v0.x) -----

// spmDecoder handles SentencePiece detokenization: the meta space '▁' (U+2581) maps to a regular space and
// byte-fallback tokens of the form "<0xHH>" map to the raw byte.
type spmDecoder struct {
	byteEmitter
	started bool
}

const spmSpace = '\u2581' // '▁'

func (d *spmDecoder) decode(piece string) string {
	// byte fallback token, e.g. "<0x0A>"
	if b, ok := parseByteToken(piece); ok {
		return d.push([]byte{b})
	}

	s := strings.ReplaceAll(piece, string(spmSpace), " ")
	// SentencePiece prefixes the very first token with a space; llama.cpp strips it.
	if !d.started {
		s = strings.TrimPrefix(s, " ")
		d.started = true
	}
	return d.push([]byte(s))
}

// parseByteToken decodes a SentencePiece byte-fallback token "<0xHH>" into its byte value.
func parseByteToken(piece string) (byte, bool) {
	if len(piece) == 6 && strings.HasPrefix(piece, "<0x") && piece[5] == '>' {
		if v, err := strconv.ParseUint(piece[3:5], 16, 8); err == nil {
			return byte(v), true
		}
	}
	return 0, false
}
