// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package webpush sends messages to the push services of browsers, as specified by the Web Push protocol of
// RFC 8030, the message encryption of RFC 8291 and the voluntary application server identification (VAPID) of
// RFC 8292. It depends on the standard library only.
package webpush

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// ErrGone is matched by the error of [Send], if the push service does not know the subscription anymore, i.e.
// it answered with 404 or 410. The subscription must not be used again.
var ErrGone = errors.New("webpush: subscription is gone")

// Keys are the keys of a subscription, as the browser returns them by PushSubscription.toJSON().
type Keys struct {
	// P256dh is the public key of the user agent as an uncompressed point in base64url.
	P256dh string
	// Auth is the authentication secret of 16 bytes in base64url.
	Auth string
}

// Subscription is the subscription of a browser, as returned by PushManager.subscribe.
type Subscription struct {
	Endpoint string
	Keys     Keys
}

// Urgency tells the push service how important a message is, so that it may save the battery of a device for
// unimportant messages, see RFC 8030 section 5.3.
type Urgency string

const (
	UrgencyVeryLow Urgency = "very-low"
	UrgencyLow     Urgency = "low"
	UrgencyNormal  Urgency = "normal"
	UrgencyHigh    Urgency = "high"
)

// Options of [Send].
type Options struct {
	// Key is the private VAPID key, whose public key the browser got for the subscription.
	Key *ecdsa.PrivateKey
	// Subject is the contact of the application server, a mailto: or https: URL. The push service of Apple rejects
	// a request without it.
	Subject string
	// TTL is how long the push service keeps a message for a device, which is offline. Zero means that the message
	// is dropped, if it cannot be delivered at once.
	TTL time.Duration
	// Urgency is optional and normal by default.
	Urgency Urgency
	// Topic optionally replaces a pending message with the same topic at the push service. It must consist of at
	// most 32 characters of the base64url alphabet.
	Topic string
	// Client is optional and [http.DefaultClient] by default.
	Client *http.Client
	// Now is optional and [time.Now] by default.
	Now func() time.Time
}

// StatusError is returned by [Send], if the push service did not accept the message.
type StatusError struct {
	StatusCode int
	Body       string
	// RetryAfter is the time, the push service asks to wait before the next request, e.g. after 429.
	RetryAfter time.Duration
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("webpush: push service answered %d: %s", e.StatusCode, e.Body)
}

// Is reports whether a 404 or 410 matches [ErrGone].
func (e *StatusError) Is(target error) bool {
	return target == ErrGone && (e.StatusCode == http.StatusNotFound || e.StatusCode == http.StatusGone)
}

// Send encrypts the payload for the subscription and posts it to the push service of the subscription. The
// payload must not exceed [MaxPayloadSize]. If the push service does not accept the message, the error is a
// [*StatusError], which matches [ErrGone] for a subscription, which must be deleted.
func Send(ctx context.Context, sub Subscription, payload []byte, opts Options) error {
	if opts.Key == nil {
		return fmt.Errorf("webpush: no VAPID key")
	}

	now := time.Now
	if opts.Now != nil {
		now = opts.Now
	}

	client := http.DefaultClient
	if opts.Client != nil {
		client = opts.Client
	}

	body, err := Encrypt(sub.Keys, payload)
	if err != nil {
		return fmt.Errorf("webpush: %w", err)
	}

	auth, err := authorization(opts.Key, sub.Endpoint, opts.Subject, now())
	if err != nil {
		return fmt.Errorf("webpush: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, sub.Endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("webpush: invalid request: %w", err)
	}

	req.Header.Set("Authorization", auth)
	req.Header.Set("Content-Encoding", "aes128gcm")
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("TTL", strconv.FormatInt(int64(opts.TTL/time.Second), 10))
	if opts.Urgency != "" {
		req.Header.Set("Urgency", string(opts.Urgency))
	}

	if opts.Topic != "" {
		req.Header.Set("Topic", opts.Topic)
	}

	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("webpush: cannot reach the push service: %w", err)
	}
	defer res.Body.Close()

	// the body of an error is short, and a success has none, which would keep the connection from being reused
	msg, _ := io.ReadAll(io.LimitReader(res.Body, 1024))
	if res.StatusCode >= 200 && res.StatusCode < 300 {
		return nil
	}

	statusErr := &StatusError{StatusCode: res.StatusCode, Body: string(bytes.TrimSpace(msg))}
	if secs, err := strconv.Atoi(res.Header.Get("Retry-After")); err == nil && secs > 0 {
		statusErr.RetryAfter = time.Duration(secs) * time.Second
	}

	return statusErr
}
