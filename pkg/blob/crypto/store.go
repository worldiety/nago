// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package crypto

import (
	"bytes"
	"context"
	"fmt"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/std"
	"io"
	"iter"
)

// header definition
var magic = [7]byte{'n', 'a', 'g', 'o', 'c', 'r', 'y'} // 7 bytes for magic nago crypto box header
const aesGCM256RandNonce uint8 = 1                     // 8th byte is the type
// other 8 bytes are random to increase boundary
var boundary = [8]byte{0x3c, 0x52, 0xc, 0x87, 0xfe, 0x3d, 0xba, 0} // ends like a c string and indicates a binary format

// EncryptionKey is a pointer to a 32 byte secret. This avoids to expose multiple copies of the secret in memory
// and simplifies the deletion of it.
type EncryptionKey *[32]byte

func NewEncryptionKey() EncryptionKey {
	return newEncryptionKey()
}

// NewBlobStore creates a backwards compatible encrypted store instance.
// It uses at least a signed cypher mode, to detect attacks or corruptions.
//
// The implementation applies transparently the encryption, so that you can introduce encryption on any
// store at any time later. To accomplish this, a 16 byte header is prepended with some magic bytes, to detect
// the format. For sure, this won't work for arbitrary binary data in general, but it will be likely enough
// to distinguish it from any reasonable defined file format. As soon as new data is written, the encryption is
// applied.
//
// The most important limitation is, that the keys are not encrypted, just as known from encrypted zip files.
// Therefore, be sure not to expose important aspects through the keys (which are normally random anyways).
//
// Note: The current implementation does not support streaming and needs to buffer the entire payload in memory,
// thus do not use it for large files, to avoid running out of memory. This may be fixed in the future.
func NewBlobStore(delegate blob.Store, key EncryptionKey) blob.Store {
	if *key == [32]byte{} {
		panic("NewBlobStore called with empty key")
	}

	return storeAdapter{delegate, key}
}

// IsEncrypted returns true, if the value of the associated key has the magic prefix.
func IsEncrypted(buf []byte) bool {
	if len(buf) < 16 {
		return false
	}

	return bytes.HasPrefix(buf, magic[:]) && bytes.HasSuffix(buf[:16], boundary[:])
}

// we could have used JWE but that standard looks like a mess and the implementations for it are absolutely
// ridiculous blown up (in a double sense)
// and cannot be reviewed in a reasonable time and context. Let us stick to the reviewed AEAD stdlib
// implementations. An alternative could be the NaCL box. Note, that we already have a lot of bad stuff here, which
// we cannot fix so easy:
//   - tdb keeps a lot (sometimes all) revisions
//   - we do not encrypt ids
//   - an attackers knows how many files and the id of them. Just like an encrypted zip file.
//   - we cannot hold incrementing nonce, thus we use random nonces which will cause birthday paradoxons
//   - jwe cannot be streamed and is very inefficient, must be kept entirely in memory, consider video files
type storeAdapter struct {
	delegate blob.Store
	key      EncryptionKey
}

// Name returns the distinct name. Stores with the same name are considered equal.
func (c storeAdapter) Name() string {
	return c.delegate.Name()
}

func (c storeAdapter) List(ctx context.Context, opts blob.ListOptions) iter.Seq2[string, error] {
	return c.delegate.List(ctx, opts)
}

func (c storeAdapter) Exists(ctx context.Context, key string) (bool, error) {
	return c.delegate.Exists(ctx, key)
}

func (c storeAdapter) Delete(ctx context.Context, key string) error {
	return c.delegate.Delete(ctx, key)
}

func (c storeAdapter) NewReader(ctx context.Context, key string) (std.Option[io.ReadCloser], error) {
	optReader, err := c.delegate.NewReader(ctx, key)
	if err != nil {
		return optReader, err
	}

	if optReader.IsNone() {
		return optReader, nil
	}

	reader := optReader.Unwrap()
	cypher, err := io.ReadAll(reader)
	if err != nil {
		return optReader, err
	}

	plainBuf, err := c.decode(cypher)
	if err != nil {
		return optReader, err
	}

	return std.Some[io.ReadCloser](readerCloser{bytes.NewReader(plainBuf)}), nil
}

func (c storeAdapter) NewWriter(ctx context.Context, key string) (io.WriteCloser, error) {
	return &writer{parent: c, key: key, ctx: ctx}, nil
}

func (c storeAdapter) encode(buf []byte) ([]byte, error) {
	tmp := make([]byte, 0, len(buf)+16)
	tmp = append(tmp, magic[:]...)
	tmp = append(tmp, aesGCM256RandNonce)
	tmp = append(tmp, boundary[:]...)

	cypher, err := encrypt(buf, c.key)
	if err != nil {
		return nil, err
	}

	tmp = append(tmp, cypher...)

	return tmp, nil
}

func (c storeAdapter) decode(buf []byte) ([]byte, error) {
	if !IsEncrypted(buf) {
		return buf, nil
	}

	encoding := buf[7]
	cypher := buf[16:]

	var plain []byte
	switch encoding {
	case aesGCM256RandNonce:
		p, err := decrypt(cypher, c.key)
		if err != nil {
			return nil, err
		}

		plain = p
	default:
		return nil, fmt.Errorf("unknown encryption key type: %v", encoding)
	}

	return plain, nil
}
