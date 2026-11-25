// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package webpush

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/asn1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/url"
	"strings"
	"time"
)

// GenerateKey creates a new VAPID key pair on the curve P-256, see RFC 8292.
func GenerateKey() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}

// MarshalPrivateKey returns the private scalar of the key as unpadded base64url, the usual notation of VAPID keys.
func MarshalPrivateKey(key *ecdsa.PrivateKey) (string, error) {
	buf, err := key.Bytes()
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// ParsePrivateKey parses a private VAPID key as returned by [MarshalPrivateKey].
func ParsePrivateKey(s string) (*ecdsa.PrivateKey, error) {
	buf, err := decodeBase64(s)
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}

	return ecdsa.ParseRawPrivateKey(elliptic.P256(), buf)
}

// PublicKey returns the public key as an uncompressed point in unpadded base64url. This is the application server
// key, which a browser requires to subscribe.
func PublicKey(key *ecdsa.PrivateKey) (string, error) {
	buf, err := key.PublicKey.Bytes()
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// vapidTokenLifetime is the lifetime of a token, which must not exceed 24 hours, see RFC 8292 section 2.
const vapidTokenLifetime = 12 * time.Hour

// authorization returns the value of the authorization header for a request to the given endpoint.
func authorization(key *ecdsa.PrivateKey, endpoint string, subject string, now time.Time) (string, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return "", fmt.Errorf("invalid endpoint: %w", err)
	}

	type claims struct {
		Aud string `json:"aud"`
		Exp int64  `json:"exp"`
		Sub string `json:"sub,omitempty"`
	}

	header := base64.RawURLEncoding.EncodeToString([]byte(`{"typ":"JWT","alg":"ES256"}`))
	payload, err := json.Marshal(claims{
		Aud: u.Scheme + "://" + u.Host,
		Exp: now.Add(vapidTokenLifetime).Unix(),
		Sub: subject,
	})
	if err != nil {
		return "", err
	}

	unsigned := header + "." + base64.RawURLEncoding.EncodeToString(payload)
	digest := sha256.Sum256([]byte(unsigned))
	der, err := ecdsa.SignASN1(rand.Reader, key, digest[:])
	if err != nil {
		return "", fmt.Errorf("cannot sign the token: %w", err)
	}

	// a JWS carries the signature as r || s with 32 bytes each instead of ASN.1, see RFC 7518 section 3.4
	var sig struct{ R, S *big.Int }
	if _, err := asn1.Unmarshal(der, &sig); err != nil {
		return "", fmt.Errorf("cannot decode the signature: %w", err)
	}

	raw := make([]byte, 64)
	sig.R.FillBytes(raw[:32])
	sig.S.FillBytes(raw[32:])

	public, err := PublicKey(key)
	if err != nil {
		return "", err
	}

	return "vapid t=" + unsigned + "." + base64.RawURLEncoding.EncodeToString(raw) + ", k=" + public, nil
}

// decodeBase64 accepts base64url and standard base64, both with and without padding, because browsers and
// libraries differ in what they produce.
func decodeBase64(s string) ([]byte, error) {
	s = strings.TrimRight(strings.TrimSpace(s), "=")
	s = strings.NewReplacer("+", "-", "/", "_").Replace(s)
	return base64.RawURLEncoding.DecodeString(s)
}
