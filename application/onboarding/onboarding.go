// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package onboarding sets up the first account of an empty instance.
//
// A provisioning system like the wdyhub passes the mail address of the first user to the process, see
// [EnvInitialUserEmail]. As long as there is no user at all, every page shows the setup instead. The visitor
// proves to own the mail address by a code, which is sent to it, and then creates the account. The account is
// verified, gets the configured roles, groups and permissions and the session is logged in.
//
// The code is the proof, not the session: a code is valid for every session, but it confirms only the session in
// which it is entered. Codes are global for the instance and rate limited, because the mail address is fixed and
// must not be flooded.
package onboarding

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"go.wdy.de/nago/logging"

	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/session"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/events"
	"golang.org/x/text/language"
)

// EnvInitialUserEmail is the environment variable with the mail address of the first user.
const EnvInitialUserEmail = "HUB_ONBOARDING_MAIL"

type Problem string

const (
	NoProblem    Problem = ""
	NoEmail      Problem = "no_email"      // the mail address has not been configured
	InvalidEmail Problem = "invalid_email" // the configured value is not a valid mail address
)

// State of the setup.
type State struct {
	// Pending is true, as long as the instance has no user.
	Pending bool
	// Email is the configured mail address of the first user. It is empty, if there is a problem.
	Email user.Email
	// Problem tells, why the setup cannot be done.
	Problem Problem
}

// MaskedEmail returns the mail address with most characters of the local part replaced, so that a visitor of the
// public setup page does not learn the address, but its owner recognizes it.
func (s State) MaskedEmail() string {
	local, domain, ok := strings.Cut(string(s.Email), "@")
	if !ok || local == "" {
		return ""
	}

	return local[:1] + "***@" + domain
}

// Profile of the first user.
type Profile struct {
	Firstname         string
	Lastname          string
	Password          user.Password
	PasswordRepeated  user.Password
	PreferredLanguage language.Tag
}

// Code describes the code, which has been sent last.
type Code struct {
	SentAt       time.Time
	ValidUntil   time.Time
	ResendAt     time.Time // earliest time to request a new code
	AttemptsLeft int
}

// Completed is published after the first user has been created.
type Completed struct {
	User  user.ID
	Email user.Email
	At    time.Time
}

var (
	ErrNotPending      = errors.New("onboarding: the instance has already been set up")
	ErrNoEmail         = errors.New("onboarding: no valid mail address of the first user has been configured")
	ErrNoCode          = errors.New("onboarding: no code has been requested")
	ErrInvalidCode     = errors.New("onboarding: the code is not valid")
	ErrCodeExpired     = errors.New("onboarding: the code has expired")
	ErrTooManyAttempts = errors.New("onboarding: too many wrong codes, request a new one")
	ErrNotVerified     = errors.New("onboarding: the session has not been confirmed by a code")
)

// CooldownError is returned, if a code is requested too early.
type CooldownError struct {
	RetryAt time.Time
}

func (e CooldownError) Error() string {
	return fmt.Sprintf("onboarding: a new code can be requested at %s", e.RetryAt.Format(time.RFC3339))
}

// GetState returns the state of the setup.
type GetState func() State

// RequestCode sends a new code to the mail address of the first user. A former code becomes invalid.
type RequestCode func(sid session.ID) (Code, error)

// CurrentCode returns the code, which has been sent last, if it is still valid.
type CurrentCode func() (Code, bool)

// VerifyCode confirms the session, if the code is valid.
type VerifyCode func(sid session.ID, code string) error

// Verified tells, if the session has been confirmed.
type Verified func(sid session.ID) bool

// Complete creates the first user for a confirmed session and logs the session in. If the login fails, the user
// is returned together with the error, because the account exists and can be used by password.
type Complete func(sid session.ID, profile Profile) (user.User, error)

// Reset forgets the confirmation of the session.
type Reset func(sid session.ID)

type UseCases struct {
	State       GetState
	RequestCode RequestCode
	CurrentCode CurrentCode
	VerifyCode  VerifyCode
	Verified    Verified
	Complete    Complete
	Reset       Reset
}

// Options of the setup.
type Options struct {
	// Email is the raw mail address of the first user, usually from [EnvInitialUserEmail].
	Email string

	Roles       []role.ID
	Groups      []group.ID
	Permissions []permission.ID
	// ConfigurePermissions is invoked after the user has been created and got the roles, groups and permissions. An
	// error aborts the setup and the user is removed again.
	ConfigurePermissions func(usr user.User) error

	CodeLifetime    time.Duration // default 15 minutes
	ResendCooldown  time.Duration // default 1 minute
	MaxCodesPerHour int           // default 10
	MaxAttempts     int           // wrong codes per code, default 5
	// VerifiedLifetime is the time, a confirmed session has to complete the setup. Default is 1 hour.
	VerifiedLifetime time.Duration
}

// Deliver sends the code to the mail address.
type Deliver func(to user.Email, code string, validUntil time.Time) error

// Deps are the use cases the setup builds on.
type Deps struct {
	CountUsers             user.CountUsers
	Create                 user.Create
	UpdateVerification     user.UpdateVerification
	UpdateOtherRoles       user.UpdateOtherRoles
	UpdateOtherGroups      user.UpdateOtherGroups
	UpdateOtherPermissions user.UpdateOtherPermissions
	Delete                 user.Delete
	LoginUser              session.LoginUser
	Deliver                Deliver
	Bus                    events.EventBus // optional
}

type challenge struct {
	hash       [32]byte
	salt       [16]byte
	sentAt     time.Time
	validUntil time.Time
	attempts   int
}

type setup struct {
	opts  Options
	deps  Deps
	email user.Email
	issue Problem
	now   func() time.Time

	// done is set, once a user exists. It is never reset, so the hot path does not query the repository.
	done atomic.Bool

	mutex    sync.Mutex
	current  *challenge
	sent     []time.Time // send times of the last hour
	verified map[session.ID]time.Time
}

// NewUseCases creates the use cases. It subscribes for created users, to end the setup as soon as any user exists.
func NewUseCases(opts Options, deps Deps) UseCases {
	return newSetup(opts, deps).useCases()
}

func newSetup(opts Options, deps Deps) *setup {
	if opts.CodeLifetime <= 0 {
		opts.CodeLifetime = 15 * time.Minute
	}

	if opts.ResendCooldown <= 0 {
		opts.ResendCooldown = time.Minute
	}

	if opts.MaxCodesPerHour <= 0 {
		opts.MaxCodesPerHour = 10
	}

	if opts.MaxAttempts <= 0 {
		opts.MaxAttempts = 5
	}

	if opts.VerifiedLifetime <= 0 {
		opts.VerifiedLifetime = time.Hour
	}

	s := &setup{opts: opts, deps: deps, now: time.Now, verified: map[session.ID]time.Time{}}
	s.email, s.issue = parseEmail(opts.Email)

	if deps.Bus != nil {
		events.SubscribeFor[user.Created](deps.Bus, func(evt user.Created) {
			s.done.Store(true)
		})
	}

	return s
}

func (s *setup) useCases() UseCases {
	return UseCases{
		State:       s.state,
		RequestCode: s.requestCode,
		CurrentCode: s.currentCode,
		VerifyCode:  s.verifyCode,
		Verified:    s.isVerified,
		Complete:    s.complete,
		Reset:       s.reset,
	}
}

func parseEmail(raw string) (user.Email, Problem) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", NoEmail
	}

	email := user.Email(strings.ToLower(raw))
	if !email.Valid() {
		return "", InvalidEmail
	}

	return email, NoProblem
}

// pending checks the repository until the first user exists.
func (s *setup) pending() bool {
	if s.done.Load() {
		return false
	}

	n, err := s.deps.CountUsers()
	if err != nil {
		slog.Error("onboarding: cannot count users", "err", err)
		return false // fail closed: never offer the setup on an instance, which may have users
	}

	if n > 0 {
		s.done.Store(true)
		return false
	}

	return true
}

func (s *setup) state() State {
	st := State{Pending: s.pending(), Problem: s.issue}
	if st.Problem == NoProblem {
		st.Email = s.email
	}

	return st
}

func (s *setup) check() error {
	if !s.pending() {
		return ErrNotPending
	}

	if s.issue != NoProblem {
		return ErrNoEmail
	}

	return nil
}

func (s *setup) requestCode(sid session.ID) (Code, error) {
	if err := s.check(); err != nil {
		return Code{}, err
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	now := s.now()
	s.sent = pruneBefore(s.sent, now.Add(-time.Hour))

	if s.current != nil && now.Before(s.current.sentAt.Add(s.opts.ResendCooldown)) {
		return Code{}, CooldownError{RetryAt: s.current.sentAt.Add(s.opts.ResendCooldown)}
	}

	if len(s.sent) >= s.opts.MaxCodesPerHour {
		return Code{}, CooldownError{RetryAt: s.sent[0].Add(time.Hour)}
	}

	code, err := newCode()
	if err != nil {
		return Code{}, err
	}

	c := &challenge{sentAt: now, validUntil: now.Add(s.opts.CodeLifetime)}
	if _, err := rand.Read(c.salt[:]); err != nil {
		return Code{}, err
	}

	c.hash = hashCode(c.salt, code)

	if err := s.deps.Deliver(s.email, code, c.validUntil); err != nil {
		return Code{}, fmt.Errorf("onboarding: cannot send the code: %w", err)
	}

	s.current = c
	s.sent = append(s.sent, now)
	slog.Info("onboarding: code sent", "session", logging.Secret(string(sid)), "validUntil", c.validUntil)

	return s.codeLocked(), nil
}

func (s *setup) currentCode() (Code, bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.current == nil || !s.now().Before(s.current.validUntil) {
		return Code{}, false
	}

	return s.codeLocked(), true
}

func (s *setup) codeLocked() Code {
	c := s.current
	return Code{
		SentAt:       c.sentAt,
		ValidUntil:   c.validUntil,
		ResendAt:     c.sentAt.Add(s.opts.ResendCooldown),
		AttemptsLeft: max(0, s.opts.MaxAttempts-c.attempts),
	}
}

func (s *setup) verifyCode(sid session.ID, code string) error {
	if err := s.check(); err != nil {
		return err
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	c := s.current
	switch {
	case c == nil:
		return ErrNoCode
	case !s.now().Before(c.validUntil):
		return ErrCodeExpired
	case c.attempts >= s.opts.MaxAttempts:
		return ErrTooManyAttempts
	}

	given := hashCode(c.salt, normalizeCode(code))
	if subtle.ConstantTimeCompare(given[:], c.hash[:]) != 1 {
		c.attempts++
		if c.attempts >= s.opts.MaxAttempts {
			return ErrTooManyAttempts
		}

		return ErrInvalidCode
	}

	s.current = nil // a code confirms a single session
	s.verified[sid] = s.now()
	slog.Info("onboarding: session confirmed", "session", logging.Secret(string(sid)))
	return nil
}

func (s *setup) isVerified(sid session.ID) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.verifiedLocked(sid)
}

func (s *setup) verifiedLocked(sid session.ID) bool {
	at, ok := s.verified[sid]
	if ok && s.now().Sub(at) > s.opts.VerifiedLifetime {
		delete(s.verified, sid)
		return false
	}

	return ok
}

func (s *setup) reset(sid session.ID) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.verified, sid)
}

func (s *setup) complete(sid session.ID, profile Profile) (user.User, error) {
	// the lock also serializes concurrent completions of different sessions
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if err := s.check(); err != nil {
		return user.User{}, err
	}

	if !s.verifiedLocked(sid) {
		return user.User{}, ErrNotVerified
	}

	su := user.SU()
	usr, err := s.deps.Create(su, user.ShortRegistrationUser{
		Firstname:         strings.TrimSpace(profile.Firstname),
		Lastname:          strings.TrimSpace(profile.Lastname),
		Email:             s.email,
		Password:          profile.Password,
		PasswordRepeated:  profile.PasswordRepeated,
		PreferredLanguage: profile.PreferredLanguage,
		Verified:          true,
		NotifyUser:        false,
	})
	if err != nil {
		return user.User{}, err
	}

	// from here on, the user must not survive a failure, otherwise the setup is gone and the instance is unusable
	if err := s.configure(su, usr); err != nil {
		if delErr := s.deps.Delete(su, usr.ID); delErr != nil {
			slog.Error("onboarding: cannot remove the half configured first user", "user", usr.ID, "err", delErr)
		}

		return user.User{}, err
	}

	s.done.Store(true)
	s.verified = map[session.ID]time.Time{}
	s.current = nil

	slog.Info("onboarding: first user created", "user", usr.ID)
	if s.deps.Bus != nil {
		s.deps.Bus.Publish(Completed{User: usr.ID, Email: usr.Email, At: s.now()})
	}

	if err := s.deps.LoginUser(sid, usr.ID); err != nil {
		return usr, fmt.Errorf("onboarding: the user has been created, but the session cannot be logged in: %w", err)
	}

	return usr, nil
}

func (s *setup) configure(su user.Subject, usr user.User) error {
	if err := s.deps.UpdateVerification(su, usr.ID, true); err != nil {
		return fmt.Errorf("cannot verify the user: %w", err)
	}

	if len(s.opts.Roles) > 0 {
		if err := s.deps.UpdateOtherRoles(su, usr.ID, s.opts.Roles); err != nil {
			return fmt.Errorf("cannot assign the roles: %w", err)
		}
	}

	if len(s.opts.Groups) > 0 {
		if err := s.deps.UpdateOtherGroups(su, usr.ID, s.opts.Groups); err != nil {
			return fmt.Errorf("cannot assign the groups: %w", err)
		}
	}

	if len(s.opts.Permissions) > 0 {
		if err := s.deps.UpdateOtherPermissions(su, usr.ID, s.opts.Permissions); err != nil {
			return fmt.Errorf("cannot assign the permissions: %w", err)
		}
	}

	if s.opts.ConfigurePermissions != nil {
		if err := s.opts.ConfigurePermissions(usr); err != nil {
			return fmt.Errorf("cannot configure the permissions: %w", err)
		}
	}

	return nil
}

// codeDigits is the length of a code.
const codeDigits = 6

// newCode returns 6 uniformly distributed digits.
func newCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%0*d", codeDigits, n.Int64()), nil
}

// normalizeCode removes separators and restores the leading zeros of a purely numeric input, which integer
// input fields drop, e.g. 031796 becomes 31796.
func normalizeCode(code string) string {
	code = strings.Map(func(r rune) rune {
		if r == ' ' || r == '-' || r == '\t' {
			return -1
		}
		return r
	}, code)

	if len(code) > 0 && len(code) < codeDigits && strings.Trim(code, "0123456789") == "" {
		code = strings.Repeat("0", codeDigits-len(code)) + code
	}

	return code
}

func hashCode(salt [16]byte, code string) [32]byte {
	return sha256.Sum256(append(salt[:], code...))
}

func pruneBefore(times []time.Time, limit time.Time) []time.Time {
	i := 0
	for i < len(times) && times[i].Before(limit) {
		i++
	}

	return times[i:]
}
