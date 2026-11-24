// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package onboarding

import (
	"errors"
	"testing"
	"time"

	"go.wdy.de/nago/application/user"
)

// A code is only requested, if it can be sent. Otherwise, neither a code is created nor the delivery invoked.
func TestRequestCodeByDelivery(t *testing.T) {
	for _, tc := range []struct {
		name  string
		check CheckDelivery
		want  error
	}{
		{name: "no check is assured"},
		{name: "assured", check: func() Delivery { return Delivery{Level: DeliveryAssured} }},
		{name: "tryable", check: func() Delivery { return Delivery{Level: DeliveryTryable} }},
		{name: "none", check: func() Delivery { return Delivery{Level: DeliveryNone} }, want: ErrNoDelivery},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, f, _ := newTestSetup(t, Options{})
			s.deps.CheckDelivery = tc.check

			_, err := s.requestCode("s1")
			if !errors.Is(err, tc.want) {
				t.Fatalf("expected %v, got %v", tc.want, err)
			}

			wantSent := tc.want == nil
			if _, current := s.currentCode(); current != wantSent || (len(f.codes) == 1) != wantSent {
				t.Fatalf("expected sent=%v, got a current code %v and the deliveries %v", wantSent, current, f.codes)
			}
		})
	}
}

// A former code survives a refused request, so a visitor in the code step can still enter it.
func TestRequestCodeWithoutDeliveryKeepsFormerCode(t *testing.T) {
	s, f, c := newTestSetup(t, Options{})
	level := DeliveryAssured
	s.deps.CheckDelivery = func() Delivery { return Delivery{Level: level} }

	first, err := s.requestCode("s1")
	if err != nil {
		t.Fatal(err)
	}

	level = DeliveryNone
	c.add(time.Minute) // after the cooldown, the request would succeed otherwise
	if _, err := s.requestCode("s1"); !errors.Is(err, ErrNoDelivery) {
		t.Fatalf("expected no delivery, got %v", err)
	}

	if cur, ok := s.currentCode(); !ok || !cur.SentAt.Equal(first.SentAt) || len(f.codes) != 1 {
		t.Fatalf("the former code must stay: %+v %v %v", cur, ok, f.codes)
	}

	if err := s.verifyCode("s1", f.lastCode()); err != nil {
		t.Fatalf("the former code must still be accepted: %v", err)
	}

	level = DeliveryAssured
	if _, err := s.requestCode("s2"); err != nil {
		t.Fatalf("the same request must succeed with a delivery: %v", err)
	}
}

// The state checks the delivery anew for each call while the setup is pending, also with a problem of the address,
// so a page tells both reasons. Afterward, it is not checked anymore.
func TestStateChecksDeliveryAnew(t *testing.T) {
	failed := time.Date(2026, 10, 1, 11, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		name  string
		email string
	}{
		{name: "valid address", email: "torben@example.com"},
		{name: "invalid address", email: "no-mail"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, f, _ := newTestSetup(t, Options{Email: tc.email})
			calls := 0
			next := Delivery{Level: DeliveryNone, ExchangeFailedAt: failed, NextExchangeAt: failed.Add(time.Hour)}
			s.deps.CheckDelivery = func() Delivery {
				calls++
				return next
			}

			if st := s.state(); st.Delivery != next || calls != 1 {
				t.Fatalf("unexpected delivery %+v after %d calls", st.Delivery, calls)
			}

			next = Delivery{Level: DeliveryTryable}
			if st := s.state(); st.Delivery != next || calls != 2 {
				t.Fatalf("the delivery must be checked anew, got %+v after %d calls", st.Delivery, calls)
			}

			f.users["u0"] = user.User{ID: "u0"}
			if st := s.state(); st.Pending || st.Delivery != (Delivery{}) || calls != 2 {
				t.Fatalf("no check after the setup, got %+v after %d calls", st, calls)
			}
		})
	}

	s, _, _ := newTestSetup(t, Options{})
	if st := s.state(); st.Delivery.Level != DeliveryAssured {
		t.Fatalf("without a check, the delivery must be assured: %+v", st.Delivery)
	}
}

// The status of a code is read for the address and the time of the code, which is valid, and unknown otherwise.
func TestCodeStatus(t *testing.T) {
	type call struct {
		to     user.Email
		sentAt time.Time
	}

	for _, tc := range []struct {
		name    string
		read    bool          // a ReadCodeStatus is given
		request bool          // a code has been requested
		later   time.Duration // time after the request
		verify  bool          // the code has been entered
		want    CodeStatus
	}{
		{name: "no dependency", request: true, want: CodeStatusUnknown},
		{name: "no code", read: true, want: CodeStatusUnknown},
		{name: "valid code", read: true, request: true, want: CodeStatusRetrying},
		{name: "expired code", read: true, request: true, later: 15 * time.Minute, want: CodeStatusUnknown},
		{name: "used code", read: true, request: true, verify: true, want: CodeStatusUnknown},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, f, c := newTestSetup(t, Options{})
			var calls []call
			if tc.read {
				s.deps.CodeStatus = func(to user.Email, sentAt time.Time) CodeStatus {
					calls = append(calls, call{to, sentAt})
					return CodeStatusRetrying
				}
			}

			sentAt := c.t
			if tc.request {
				if _, err := s.requestCode("s1"); err != nil {
					t.Fatal(err)
				}
			}

			if tc.verify {
				if err := s.verifyCode("s1", f.lastCode()); err != nil {
					t.Fatal(err)
				}
			}

			c.add(tc.later)
			if got := s.codeStatus(); got != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}

			if tc.want != CodeStatusUnknown && (len(calls) != 1 || calls[0] != (call{"torben@example.com", sentAt})) {
				t.Fatalf("expected the status of the code at %v, got %v", sentAt, calls)
			}
		})
	}
}
