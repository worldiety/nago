// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package nais

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
var ErrNotEnrolled = errors.New("nago ai service: not enrolled")

const stateID = "default"

// State is the persistent enrollment of the instance.
type State struct {
	ID           string    `json:"id"`
	Endpoint     string    `json:"endpoint"`
	Tenant       string    `json:"tenant,omitempty"`
	Refresh      string    `json:"refresh,omitempty"`
	Models       []Model   `json:"models,omitempty"`
	DefaultModel string    `json:"defaultModel,omitempty"`
	Limits       Limits    `json:"limits,omitzero"`
	EnrolledAt   time.Time `json:"enrolledAt,omitzero"`
	// TokenHash is the hash of the configured token, which has been taken over. A configured token is only taken over
	// once, so that a revoked token is not used again.
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
	// exchange is attempted for any http or https origin without path: whether a public origin is called back or the
	// instance is admitted by its address is the service's decision.
	Origin func() string
	// Nonces answers the call back of the service. A token exchange is only attempted, if set.
	Nonces *Nonces
	// States persists the enrollment. Optional, otherwise the enrollment is kept in memory.
	States StateRepository
	// HTTP is an optional client for the token handling.
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
	Models         []Model
	DefaultModel   string
	Limits         Limits
	EnrolledAt     time.Time
	LastError      string
	LastErrorAt    time.Time
	LastExchangeAt time.Time
}

// Service keeps the instance enrolled and hands out access tokens. It is safe for concurrent use.
type Service struct {
	opts   Options
	client Client
	now    func() time.Time

	ops sync.Mutex // serializes the token handling

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
			slog.Error("nago ai service: cannot load enrollment", "err", err)
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

// Endpoint is the configured endpoint.
func (s *Service) Endpoint() string {
	return s.opts.Endpoint
}

func (s *Service) Status() Status {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return Status{
		Endpoint:       s.opts.Endpoint,
		Enabled:        s.Enabled(),
		Enrolled:       s.state.Refresh != "",
		Tenant:         s.state.Tenant,
		Models:         append([]Model(nil), s.state.Models...),
		DefaultModel:   s.state.DefaultModel,
		Limits:         s.state.Limits,
		EnrolledAt:     s.state.EnrolledAt,
		LastError:      s.lastErr,
		LastErrorAt:    s.lastErrAt,
		LastExchangeAt: s.lastExchangeAt,
	}
}

// Access returns a valid access token. It refreshes it and enrolls, if necessary.
func (s *Service) Access(ctx context.Context) (string, error) {
	if !s.Enabled() {
		return "", ErrNotEnrolled
	}

	s.ops.Lock()
	defer s.ops.Unlock()

	access, err := s.ensureAccess(ctx)
	s.fail(err)
	return access, err
}

// Reject tells the service, that it refused the access token. The token is dropped, and with it the refresh token,
// if the service revoked it, so that the next [Service.Access] enrolls again.
func (s *Service) Reject(access string, e *Error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// another call may have replaced the token meanwhile
	if access != s.access {
		return
	}

	s.access = ""
	s.accessUntil = time.Time{}
	if e.Revoked() && s.state.Refresh != "" {
		slog.Warn("nago ai service revoked the token, enrolling again", "code", e.Code)
		s.state.Refresh = ""
		s.saveLocked()
	}
}

// Health asks the service about the tenant and keeps its models.
func (s *Service) Health(ctx context.Context) (Health, error) {
	access, err := s.Access(ctx)
	if err != nil {
		return Health{}, err
	}

	h, err := s.client.Health(ctx, access)
	if err != nil {
		return Health{}, err
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.state.Tenant = h.Tenant
	s.state.Models = h.Models
	s.state.DefaultModel = h.DefaultModel
	s.state.Limits.TokensPerDay = h.Limits.TokensPerDay
	s.saveLocked()
	return h, nil
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
			slog.Warn("nago ai service rejected the refresh token, enrolling again", "code", e.Code)
			s.dropRefresh()
			if err := s.exchange(ctx); err != nil {
				return "", err
			}
			exchanged = true
			continue
		}

		if err != nil {
			if errors.As(err, &e) && e.Status == http.StatusUnauthorized {
				s.dropRefresh()
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

	slog.Info("nago ai service: exchanging token", "endpoint", s.opts.Endpoint, "origin", origin)
	enrollment, err := s.client.Exchange(ctx, origin, nonce)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrNotEnrolled, err)
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.failedExchange = time.Time{}
	s.state.Tenant = enrollment.Tenant
	s.state.Refresh = enrollment.Refresh
	s.state.Models = enrollment.Models
	s.state.DefaultModel = enrollment.DefaultModel
	s.state.Limits = enrollment.Limits
	s.state.EnrolledAt = s.now()
	s.saveLocked()

	slog.Info("nago ai service: enrolled", "tenant", enrollment.Tenant, "defaultModel", enrollment.DefaultModel)
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

// dropRefresh forgets the access and the refresh token.
func (s *Service) dropRefresh() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.access = ""
	s.accessUntil = time.Time{}
	if s.state.Refresh != "" {
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
		slog.Error("nago ai service: cannot save enrollment", "err", err)
	}
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
