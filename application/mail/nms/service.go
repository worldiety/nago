// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package nms

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"go.wdy.de/nago/pkg/data"
)

// ErrNotEnrolled is returned, if the instance has no token and cannot get one right now.
var ErrNotEnrolled = errors.New("nago mail service: not enrolled")

// ErrNoSender is returned, if a message has no sender and the tenant has no sender domain to derive one.
var ErrNoSender = errors.New("nago mail service: no sender address")

const stateID = "default"

// State is the persistent enrollment of the instance.
type State struct {
	ID            string    `json:"id"`
	Endpoint      string    `json:"endpoint"`
	Tenant        string    `json:"tenant,omitempty"`
	Refresh       string    `json:"refresh,omitempty"`
	SenderDomains []string  `json:"senderDomains,omitempty"`
	Limits        Limits    `json:"limits,omitzero"`
	EnrolledAt    time.Time `json:"enrolledAt,omitzero"`
	// TokenHash is the hash of the configured token, which has been taken over. A configured token is only
	// taken over once, so that a revoked token is not used again.
	TokenHash string `json:"tokenHash,omitempty"`
}

func (s State) Identity() string {
	return s.ID
}

type StateRepository data.Repository[State, string]

type Options struct {
	// Endpoint of the service. Empty disables the service.
	Endpoint string
	// Token is an optional refresh token configured by the operator. It is taken over once.
	Token string
	// Origin returns the origin of the instance like https://my-app.example.com or http://localhost:3000. A token
	// exchange is attempted for any http or https origin without path: whether a public origin is called back or
	// the instance is admitted by its address is the service's decision, see the package documentation.
	Origin func() string
	// Nonces answers the call back of the service. A token exchange is only attempted, if set.
	Nonces *Nonces
	// States persists the enrollment. Optional, otherwise the enrollment is kept in memory.
	States StateRepository
	// HTTP is an optional client.
	HTTP *http.Client
	// ExchangeInterval is the minimum time after a failed exchange, before the next one is attempted. Default is 1 hour.
	ExchangeInterval time.Duration
}

// Status is a snapshot for the presentation.
type Status struct {
	Endpoint       string
	Enabled        bool
	Enrolled       bool
	Tenant         string
	SenderDomains  []string
	Limits         Limits
	EnrolledAt     time.Time
	LastError      string
	LastErrorAt    time.Time
	LastExchangeAt time.Time
}

// Service keeps the instance enrolled and sends messages. It is safe for concurrent use.
type Service struct {
	opts   Options
	client Client
	now    func() time.Time

	ops sync.Mutex // serializes the token handling and sending

	mutex          sync.Mutex // protects the fields below
	state          State
	access         string
	accessUntil    time.Time
	lastExchangeAt time.Time
	failedExchange time.Time // time of the last failed exchange, zero after a successful one
	lastErr        string
	lastErrAt      time.Time
}

func NewService(opts Options) *Service {
	if opts.ExchangeInterval <= 0 {
		opts.ExchangeInterval = time.Hour
	}

	if opts.Origin == nil {
		opts.Origin = func() string { return "" }
	}

	opts.Endpoint = strings.TrimRight(strings.TrimSpace(opts.Endpoint), "/")

	s := &Service{
		opts:   opts,
		client: Client{Endpoint: opts.Endpoint, HTTP: opts.HTTP},
		now:    time.Now,
	}

	if opts.Endpoint == "" {
		return s
	}

	s.state = State{ID: stateID, Endpoint: opts.Endpoint}
	if opts.States != nil {
		optState, err := opts.States.FindByID(stateID)
		if err != nil {
			slog.Error("nago mail service: cannot load enrollment", "err", err)
		} else if optState.IsSome() && optState.Unwrap().Endpoint == opts.Endpoint {
			s.state = optState.Unwrap()
		}
	}

	if opts.Token != "" {
		if hash := tokenHash(opts.Token); s.state.TokenHash != hash {
			s.state = State{ID: stateID, Endpoint: opts.Endpoint, Refresh: opts.Token, TokenHash: hash, EnrolledAt: s.now()}
			s.saveLocked()
		}
	}

	return s
}

func tokenHash(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

// Enabled is true, if an endpoint is configured.
func (s *Service) Enabled() bool {
	return s.opts.Endpoint != ""
}

// Available is true, if the service can be used right now: either it has a token or a token exchange is
// possible and due.
func (s *Service) Available() bool {
	if !s.Enabled() {
		return false
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.state.Refresh != "" {
		return true
	}

	return s.exchangeDueLocked() && s.canExchange()
}

// Limits returns the limits of the tenant.
func (s *Service) Limits() Limits {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.state.Limits
}

// SenderDomains returns the domains the instance may send from.
func (s *Service) SenderDomains() []string {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return append([]string(nil), s.state.SenderDomains...)
}

func (s *Service) Status() Status {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return Status{
		Endpoint:       s.opts.Endpoint,
		Enabled:        s.Enabled(),
		Enrolled:       s.state.Refresh != "",
		Tenant:         s.state.Tenant,
		SenderDomains:  append([]string(nil), s.state.SenderDomains...),
		Limits:         s.state.Limits,
		EnrolledAt:     s.state.EnrolledAt,
		LastError:      s.lastErr,
		LastErrorAt:    s.lastErrAt,
		LastExchangeAt: s.lastExchangeAt,
	}
}

// Send sends a single message. A revoked token is replaced once by a new one. A sender outside the sender domains
// of the tenant is replaced by a noreply address of the first sender domain and becomes the reply-to address.
func (s *Service) Send(ctx context.Context, msg Message) (MessageResult, error) {
	if !s.Enabled() {
		return MessageResult{}, ErrNotEnrolled
	}

	s.ops.Lock()
	defer s.ops.Unlock()

	for attempt := 0; ; attempt++ {
		access, err := s.ensureAccess(ctx)
		if err != nil {
			s.fail(err)
			return MessageResult{}, err
		}

		adjusted, err := withSender(msg, s.SenderDomains())
		if err != nil {
			return MessageResult{}, err
		}

		res, err := s.client.Send(ctx, access, adjusted)
		var e *Error
		if attempt == 0 && errors.As(err, &e) && e.Status == http.StatusUnauthorized {
			slog.Warn("nago mail service rejected the access token", "code", e.Code)
			s.dropAccess(e.Revoked())
			continue
		}

		if err != nil {
			s.fail(err)
			return MessageResult{}, err
		}

		s.fail(nil)
		return res[0], nil
	}
}

// Health asks the service about the tenant.
func (s *Service) Health(ctx context.Context) (Health, error) {
	if !s.Enabled() {
		return Health{}, ErrNotEnrolled
	}

	s.ops.Lock()
	defer s.ops.Unlock()

	access, err := s.ensureAccess(ctx)
	if err != nil {
		return Health{}, err
	}

	return s.client.Health(ctx, access)
}

// ensureAccess returns a valid access token. It refreshes it and enrolls, if necessary.
func (s *Service) ensureAccess(ctx context.Context) (string, error) {
	s.mutex.Lock()
	if s.access != "" && s.now().Before(s.accessUntil) {
		access := s.access
		s.mutex.Unlock()
		return access, nil
	}
	refresh := s.state.Refresh
	s.mutex.Unlock()

	exchanged := false
	if refresh == "" {
		if err := s.exchange(ctx); err != nil {
			return "", err
		}
		exchanged = true
	}

	for {
		s.mutex.Lock()
		refresh = s.state.Refresh
		s.mutex.Unlock()

		tokens, err := s.client.Token(ctx, refresh)
		var e *Error
		if errors.As(err, &e) && e.Status == http.StatusUnauthorized && !exchanged {
			slog.Warn("nago mail service rejected the refresh token, enrolling again", "code", e.Code)
			s.dropAccess(true)
			if err := s.exchange(ctx); err != nil {
				return "", err
			}
			exchanged = true
			continue
		}

		if err != nil {
			if errors.As(err, &e) && e.Status == http.StatusUnauthorized {
				s.dropAccess(true)
			}
			return "", err
		}

		s.mutex.Lock()
		defer s.mutex.Unlock()

		lifetime := time.Duration(tokens.ExpiresIn) * time.Second
		if lifetime <= 0 {
			lifetime = 5 * time.Minute
		}

		// renew a minute early, so that a token does not expire in flight
		s.access = tokens.Access
		s.accessUntil = s.now().Add(max(lifetime-time.Minute, lifetime/2))
		if tokens.Refresh != "" && tokens.Refresh != s.state.Refresh {
			s.state.Refresh = tokens.Refresh
			s.saveLocked()
		}

		return tokens.Access, nil
	}
}

// exchange enrolls the instance by a token exchange.
func (s *Service) exchange(ctx context.Context) error {
	s.mutex.Lock()
	if !s.exchangeDueLocked() {
		next := s.failedExchange.Add(s.opts.ExchangeInterval)
		s.mutex.Unlock()
		return fmt.Errorf("%w: next token exchange at %s", ErrNotEnrolled, next.Format(time.RFC3339))
	}

	if !s.canExchange() {
		s.mutex.Unlock()
		return fmt.Errorf("%w: a token exchange requires an http or https origin without path, got '%s'", ErrNotEnrolled, s.opts.Origin())
	}

	s.lastExchangeAt = s.now()
	s.failedExchange = s.lastExchangeAt // until proven otherwise
	s.mutex.Unlock()

	origin, _ := exchangeOrigin(s.opts.Origin())
	nonce := s.opts.Nonces.Issue()
	defer s.opts.Nonces.Forget(nonce)

	slog.Info("nago mail service: exchanging token", "endpoint", s.opts.Endpoint, "origin", origin)
	enrollment, err := s.client.Exchange(ctx, origin, nonce)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrNotEnrolled, err)
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.failedExchange = time.Time{}
	s.state.Tenant = enrollment.Tenant
	s.state.Refresh = enrollment.Refresh
	s.state.SenderDomains = enrollment.SenderDomains
	s.state.Limits = enrollment.Limits
	s.state.EnrolledAt = s.now()
	s.saveLocked()

	slog.Info("nago mail service: enrolled", "tenant", enrollment.Tenant, "senderDomains", enrollment.SenderDomains)
	return nil
}

func (s *Service) exchangeDueLocked() bool {
	return s.failedExchange.IsZero() || s.now().Sub(s.failedExchange) >= s.opts.ExchangeInterval
}

func (s *Service) canExchange() bool {
	if s.opts.Nonces == nil {
		return false
	}

	_, ok := exchangeOrigin(s.opts.Origin())
	return ok
}

// dropAccess forgets the access token and, if the refresh token is revoked, also the refresh token.
func (s *Service) dropAccess(revoked bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.access = ""
	s.accessUntil = time.Time{}
	if revoked && s.state.Refresh != "" {
		s.state.Refresh = ""
		s.saveLocked()
	}
}

func (s *Service) fail(err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if err == nil {
		s.lastErr = ""
		s.lastErrAt = time.Time{}
		return
	}

	s.lastErr = err.Error()
	s.lastErrAt = s.now()
}

func (s *Service) saveLocked() {
	if s.opts.States == nil {
		return
	}

	if err := s.opts.States.Save(s.state); err != nil {
		slog.Error("nago mail service: cannot save enrollment", "err", err)
	}
}

// withSender replaces a sender outside the sender domains, see [Service.Send].
func withSender(msg Message, domains []string) (Message, error) {
	_, domain, _ := strings.Cut(msg.From.Email, "@")
	if len(domains) > 0 && !containsFold(domains, domain) {
		if msg.From.Email != "" && msg.ReplyTo == nil {
			replyTo := msg.From
			msg.ReplyTo = &replyTo
		}

		msg.From.Email = "noreply@" + domains[0]
	}

	if msg.From.Email == "" {
		return Message{}, ErrNoSender
	}

	return msg, nil
}

func containsFold(list []string, s string) bool {
	for _, e := range list {
		if strings.EqualFold(e, s) {
			return true
		}
	}

	return false
}

// exchangeOrigin returns the normalized origin, if it is an http or https origin without path.
//
// A local origin like http://localhost:3000 is deliberately included: it cannot be called back, but the service may
// admit the instance by the address it calls from. A failed exchange is repeated only after
// [Options.ExchangeInterval], so an instance the service does not know asks once an hour at most.
func exchangeOrigin(origin string) (string, bool) {
	u, err := url.Parse(strings.TrimSpace(origin))
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return "", false
	}

	if u.Path != "" && u.Path != "/" {
		return "", false
	}

	return u.Scheme + "://" + u.Host, true
}
