// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package webpush

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/hkdf"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
)

const (
	// recordSize is the record size of the single record of a message. Push services accept at most 4096 bytes
	// of payload, so a message always fits into one record.
	recordSize = 4096

	saltLen      = 16
	authLen      = 16
	tagLen       = 16
	publicKeyLen = 65
	headerLen    = saltLen + 4 + 1 + publicKeyLen

	// MaxPayloadSize is the largest plaintext, which fits into the 4096 bytes of a push message after the
	// encryption: the header, the padding delimiter and the authentication tag take the rest.
	MaxPayloadSize = recordSize - headerLen - tagLen - 1
)

// Encrypt encrypts the plaintext for the user agent, which subscribed with the given keys, as described by RFC 8291
// in the content coding aes128gcm of RFC 8188. A new ephemeral key and salt are used for each message.
func Encrypt(keys Keys, plaintext []byte) ([]byte, error) {
	asPrivate, err := ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("cannot generate the ephemeral key: %w", err)
	}

	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("cannot generate the salt: %w", err)
	}

	return encrypt(keys, plaintext, asPrivate, salt)
}

// encrypt is like [Encrypt] but with a given ephemeral key and salt, as the test vectors of RFC 8291 require.
func encrypt(keys Keys, plaintext []byte, asPrivate *ecdh.PrivateKey, salt []byte) ([]byte, error) {
	if len(plaintext) > MaxPayloadSize {
		return nil, fmt.Errorf("payload has %d bytes but at most %d bytes fit into a push message", len(plaintext), MaxPayloadSize)
	}

	uaPublicBytes, err := decodeBase64(keys.P256dh)
	if err != nil {
		return nil, fmt.Errorf("invalid p256dh key: %w", err)
	}

	authSecret, err := decodeBase64(keys.Auth)
	if err != nil {
		return nil, fmt.Errorf("invalid auth secret: %w", err)
	}

	if len(authSecret) != authLen {
		return nil, fmt.Errorf("auth secret has %d bytes but must have %d bytes", len(authSecret), authLen)
	}

	uaPublic, err := ecdh.P256().NewPublicKey(uaPublicBytes)
	if err != nil {
		return nil, fmt.Errorf("invalid p256dh key: %w", err)
	}

	ecdhSecret, err := asPrivate.ECDH(uaPublic)
	if err != nil {
		return nil, fmt.Errorf("cannot agree on a secret: %w", err)
	}

	asPublic := asPrivate.PublicKey().Bytes()

	// key_info = "WebPush: info" || 0x00 || ua_public || as_public
	keyInfo := make([]byte, 0, 14+2*publicKeyLen)
	keyInfo = append(keyInfo, "WebPush: info\x00"...)
	keyInfo = append(keyInfo, uaPublicBytes...)
	keyInfo = append(keyInfo, asPublic...)

	prkKey, err := hkdf.Extract(sha256.New, ecdhSecret, authSecret)
	if err != nil {
		return nil, err
	}

	ikm, err := hkdf.Expand(sha256.New, prkKey, string(keyInfo), 32)
	if err != nil {
		return nil, err
	}

	prk, err := hkdf.Extract(sha256.New, ikm, salt)
	if err != nil {
		return nil, err
	}

	cek, err := hkdf.Expand(sha256.New, prk, "Content-Encoding: aes128gcm\x00", 16)
	if err != nil {
		return nil, err
	}

	nonce, err := hkdf.Expand(sha256.New, prk, "Content-Encoding: nonce\x00", 12)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(cek)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// the single record is the last one, which the delimiter 0x02 marks, and needs no further padding
	padded := make([]byte, 0, len(plaintext)+1)
	padded = append(padded, plaintext...)
	padded = append(padded, 0x02)

	// header = salt || rs || idlen || keyid
	out := make([]byte, headerLen, headerLen+len(padded)+tagLen)
	copy(out, salt)
	binary.BigEndian.PutUint32(out[saltLen:], recordSize)
	out[saltLen+4] = publicKeyLen
	copy(out[saltLen+5:], asPublic)

	return gcm.Seal(out, nonce, padded, nil), nil
}

// Decrypt decrypts a message as the user agent does, which subscribed with the private key uaPrivate and the auth
// secret auth. A server never needs it, but it lets a test check what a browser would receive.
func Decrypt(uaPrivate *ecdh.PrivateKey, auth []byte, msg []byte) ([]byte, error) {
	if len(msg) < headerLen+tagLen {
		return nil, fmt.Errorf("message too short")
	}

	salt := msg[:saltLen]
	idLen := int(msg[saltLen+4])
	if idLen != publicKeyLen || len(msg) < saltLen+5+idLen+tagLen {
		return nil, fmt.Errorf("unexpected key id of %d bytes", idLen)
	}

	asPublicBytes := msg[saltLen+5 : saltLen+5+idLen]
	asPublic, err := ecdh.P256().NewPublicKey(asPublicBytes)
	if err != nil {
		return nil, fmt.Errorf("invalid key of the application server: %w", err)
	}

	secret, err := uaPrivate.ECDH(asPublic)
	if err != nil {
		return nil, err
	}

	keyInfo := append([]byte("WebPush: info\x00"), uaPrivate.PublicKey().Bytes()...)
	keyInfo = append(keyInfo, asPublicBytes...)

	prkKey, err := hkdf.Extract(sha256.New, secret, auth)
	if err != nil {
		return nil, err
	}

	ikm, err := hkdf.Expand(sha256.New, prkKey, string(keyInfo), 32)
	if err != nil {
		return nil, err
	}

	prk, err := hkdf.Extract(sha256.New, ikm, salt)
	if err != nil {
		return nil, err
	}

	cek, err := hkdf.Expand(sha256.New, prk, "Content-Encoding: aes128gcm\x00", 16)
	if err != nil {
		return nil, err
	}

	nonce, err := hkdf.Expand(sha256.New, prk, "Content-Encoding: nonce\x00", 12)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(cek)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	plain, err := gcm.Open(nil, nonce, msg[saltLen+5+idLen:], nil)
	if err != nil {
		return nil, err
	}

	// the last record ends with the delimiter 0x02, followed by optional zero padding
	end := len(plain) - 1
	for end >= 0 && plain[end] == 0 {
		end--
	}

	if end < 0 || plain[end] != 0x02 {
		return nil, fmt.Errorf("missing delimiter of the last record")
	}

	return plain[:end], nil
}
