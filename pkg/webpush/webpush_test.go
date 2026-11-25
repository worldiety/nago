// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package webpush

import (
	"context"
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func mustDecode(t *testing.T, s string) []byte {
	t.Helper()
	buf, err := decodeBase64(s)
	if err != nil {
		t.Fatal(err)
	}
	return buf
}

// TestEncryptRFC8291 checks the example of RFC 8291 appendix A byte by byte.
func TestEncryptRFC8291(t *testing.T) {
	asPrivate, err := ecdh.P256().NewPrivateKey(mustDecode(t, "yfWPiYE-n46HLnH0KqZOF1fJJU3MYrct3AELtAQ-oRw"))
	if err != nil {
		t.Fatal(err)
	}

	keys := Keys{
		P256dh: "BCVxsr7N_eNgVRqvHtD0zTZsEc6-VV-JvLexhqUzORcxaOzi6-AYWXvTBHm4bjyPjs7Vd8pZGH6SRpkNtoIAiw4",
		Auth:   "BTBZMqHH6r4Tts7J_aSIgg",
	}

	got, err := encrypt(keys, []byte("When I grow up, I want to be a watermelon"), asPrivate, mustDecode(t, "DGv6ra1nlYgDCS1FRnbzlw"))
	if err != nil {
		t.Fatal(err)
	}

	want := "DGv6ra1nlYgDCS1FRnbzlwAAEABBBP4z9KsN6nGRTbVYI_c7VJSPQTBtkgcy27mlmlMoZIIgDll6e3vCYLocInmYWAmS6TlzAC8wEqKK6PBru3jl7A_yl95bQpu6cVPTpK4Mqgkf1CXztLVBSt2Ks3oZwbuwXPXLWyouBWLVWGNWQexSgSxsj_Qulcy4a-fN"
	if base64.RawURLEncoding.EncodeToString(got) != want {
		t.Fatalf("got %s\nwant %s", base64.RawURLEncoding.EncodeToString(got), want)
	}

	plain := decrypt(t, mustDecode(t, "q1dXpw3UpT5VOmu_cf_v6ih07Aems3njxI-JWgLcM94"), keys, got)
	if string(plain) != "When I grow up, I want to be a watermelon" {
		t.Fatalf("unexpected plaintext %q", plain)
	}
}

// decrypt is the user agent side of RFC 8291.
func decrypt(t *testing.T, uaPrivateBytes []byte, keys Keys, msg []byte) []byte {
	t.Helper()

	uaPrivate, err := ecdh.P256().NewPrivateKey(uaPrivateBytes)
	if err != nil {
		t.Fatal(err)
	}

	if rs := binary.BigEndian.Uint32(msg[16:20]); rs != recordSize {
		t.Fatalf("unexpected record size %d", rs)
	}

	plain, err := Decrypt(uaPrivate, mustDecode(t, keys.Auth), msg)
	if err != nil {
		t.Fatal(err)
	}

	return plain
}

func newUserAgent(t *testing.T) (*ecdh.PrivateKey, Keys) {
	t.Helper()
	ua, err := ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	auth := make([]byte, 16)
	_, _ = rand.Read(auth)
	return ua, Keys{
		P256dh: base64.RawURLEncoding.EncodeToString(ua.PublicKey().Bytes()),
		// browsers of the past returned padded standard base64, which is accepted as well
		Auth: base64.StdEncoding.EncodeToString(auth),
	}
}

func TestEncryptLimits(t *testing.T) {
	ua, keys := newUserAgent(t)

	full := strings.Repeat("x", MaxPayloadSize)
	msg, err := Encrypt(keys, []byte(full))
	if err != nil {
		t.Fatal(err)
	}

	if len(msg) != recordSize {
		t.Fatalf("a full message must take exactly the record size but has %d bytes", len(msg))
	}

	if string(decrypt(t, ua.Bytes(), keys, msg)) != full {
		t.Fatal("round trip failed")
	}

	if _, err := Encrypt(keys, []byte(full+"x")); err == nil {
		t.Fatal("expected an error for a payload, which is too large")
	}

	if _, err := Encrypt(Keys{P256dh: keys.P256dh, Auth: "AAAA"}, []byte("x")); err == nil {
		t.Fatal("expected an error for a short auth secret")
	}

	if _, err := Encrypt(Keys{P256dh: "AAAA", Auth: keys.Auth}, []byte("x")); err == nil {
		t.Fatal("expected an error for an invalid public key")
	}
}

func TestPrivateKeyRoundTrip(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatal(err)
	}

	s, err := MarshalPrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}

	parsed, err := ParsePrivateKey(s)
	if err != nil {
		t.Fatal(err)
	}

	if !parsed.Equal(key) {
		t.Fatal("parsed key differs")
	}

	pub, err := PublicKey(key)
	if err != nil {
		t.Fatal(err)
	}

	if raw := mustDecode(t, pub); len(raw) != 65 || raw[0] != 0x04 {
		t.Fatalf("public key must be an uncompressed point, got %d bytes", len(raw))
	}
}

// verifyVAPID checks the authorization header as a push service does and returns the claims.
func verifyVAPID(t *testing.T, header string) map[string]any {
	t.Helper()

	rest, ok := strings.CutPrefix(header, "vapid t=")
	if !ok {
		t.Fatalf("unexpected scheme: %s", header)
	}

	token, k, ok := strings.Cut(rest, ", k=")
	if !ok {
		t.Fatalf("missing key: %s", header)
	}

	pub, err := ecdsa.ParseUncompressedPublicKey(elliptic.P256(), mustDecode(t, k))
	if err != nil {
		t.Fatal(err)
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("not a JWT: %s", token)
	}

	var head map[string]any
	if err := json.Unmarshal(mustDecode(t, parts[0]), &head); err != nil || head["alg"] != "ES256" {
		t.Fatalf("unexpected header %v %v", head, err)
	}

	sig := mustDecode(t, parts[2])
	if len(sig) != 64 {
		t.Fatalf("signature must have 64 bytes but has %d", len(sig))
	}

	digest := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	if !ecdsa.Verify(pub, digest[:], new(big.Int).SetBytes(sig[:32]), new(big.Int).SetBytes(sig[32:])) {
		t.Fatal("invalid signature")
	}

	var claims map[string]any
	if err := json.Unmarshal(mustDecode(t, parts[1]), &claims); err != nil {
		t.Fatal(err)
	}

	return claims
}

func TestSend(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatal(err)
	}

	ua, keys := newUserAgent(t)
	now := time.Date(2026, 11, 25, 12, 0, 0, 0, time.UTC)

	var got *http.Request
	var gotBody []byte
	status := http.StatusCreated
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r
		gotBody, _ = io.ReadAll(r.Body)
		if status == http.StatusTooManyRequests {
			w.Header().Set("Retry-After", "30")
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte("details\n"))
	}))
	defer srv.Close()

	sub := Subscription{Endpoint: srv.URL + "/push/abc", Keys: keys}
	opts := Options{Key: key, Subject: "mailto:ops@example.com", TTL: time.Hour, Urgency: UrgencyHigh, Topic: "chat", Now: func() time.Time { return now }}

	if err := Send(context.Background(), sub, []byte(`{"title":"Hi"}`), opts); err != nil {
		t.Fatal(err)
	}

	if got.Method != http.MethodPost || got.URL.Path != "/push/abc" {
		t.Fatalf("unexpected request %s %s", got.Method, got.URL.Path)
	}

	for name, want := range map[string]string{"Content-Encoding": "aes128gcm", "TTL": "3600", "Urgency": "high", "Topic": "chat"} {
		if got.Header.Get(name) != want {
			t.Fatalf("header %s is %q but want %q", name, got.Header.Get(name), want)
		}
	}

	claims := verifyVAPID(t, got.Header.Get("Authorization"))
	if claims["aud"] != srv.URL || claims["sub"] != "mailto:ops@example.com" || int64(claims["exp"].(float64)) != now.Add(12*time.Hour).Unix() {
		t.Fatalf("unexpected claims %v", claims)
	}

	if string(decrypt(t, ua.Bytes(), keys, gotBody)) != `{"title":"Hi"}` {
		t.Fatal("push service got an unexpected payload")
	}

	for _, code := range []int{http.StatusNotFound, http.StatusGone} {
		status = code
		err := Send(context.Background(), sub, []byte("x"), opts)
		if !errors.Is(err, ErrGone) {
			t.Fatalf("%d must match ErrGone but got %v", code, err)
		}
	}

	status = http.StatusTooManyRequests
	err = Send(context.Background(), sub, []byte("x"), opts)
	var statusErr *StatusError
	if !errors.As(err, &statusErr) || errors.Is(err, ErrGone) || statusErr.RetryAfter != 30*time.Second || statusErr.Body != "details" {
		t.Fatalf("unexpected error %#v", err)
	}

	if err := Send(context.Background(), sub, []byte("x"), Options{}); err == nil {
		t.Fatal("expected an error without key")
	}
}
