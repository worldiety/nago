// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package onboarding sets up the first account of an empty instance.
//
// A provisioning system like the wdyhub passes the mail address of the first user to the process, see
// [EnvInitialUserEmail] and [EnvHubOnboardingMail]. As long as there is no user at all, every page shows the setup
// instead. The visitor proves to own the mail address by a code, which is sent to it, and then creates the account.
// The account is verified, gets the configured roles, groups and permissions and the session is logged in.
//
// The code is the proof, not the session: a code is valid for every session, but it confirms only the session in
// which it is entered. Codes are global for the instance and rate limited, because the mail address is fixed and
// must not be flooded.
//
// Instead of a password, the first user may sign in by the single sign-on of a Nago login service (NLS), see
// [CompleteWithSSO]. Then, the setup configures the login service and decides who else may sign in by it.
//
// Before a code is requested, the setup tells whether the code can be sent at all, see [Delivery], and afterward,
// what has become of the mail with the code, see [CodeStatus].
package onboarding

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"go.wdy.de/nago/logging"

	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/session"
	"go.wdy.de/nago/application/settings"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/events"
	"golang.org/x/text/language"
)

// EnvInitialUserEmail is the environment variable with the mail address of the first user.
const EnvInitialUserEmail = "INITIAL_USER_EMAIL"

// EnvHubOnboardingMail is the environment variable with the mail address of the first user, which the wdyhub sets.
// It is read, if [EnvInitialUserEmail] is empty.
const EnvHubOnboardingMail = "HUB_ONBOARDING_MAIL"

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
	// SSOServer is the login service, which the setup offers for the single sign-on, see [CompleteWithSSO]. It is
	// empty, if the setup offers no single sign-on.
	SSOServer string
	// Delivery tells, whether the code can be sent. It is checked anew for every state, as long as the setup is
	// pending, and zero otherwise.
	Delivery Delivery
}

// DeliveryLevel tells, whether the code can be sent.
type DeliveryLevel int

const (
	// DeliveryNone means, that there is no way to send the code, thus [RequestCode] fails with [ErrNoDelivery].
	DeliveryNone DeliveryLevel = iota
	// DeliveryTryable means, that the code can be sent, but whether it arrives shows only afterward, e.g. because
	// the instance enrolls at the Nago Mail Service right now.
	DeliveryTryable
	// DeliveryAssured means, that there is a way to send the code, e.g. a custom [Deliver], an SMTP server or an
	// enrollment at the Nago Mail Service.
	DeliveryAssured
)

// Delivery of the code. It tells nothing about the delivery of a code, which has been sent, see [CodeStatus].
type Delivery struct {
	Level DeliveryLevel
	// ExchangeFailedAt is the time of the last failed token exchange with the Nago Mail Service and NextExchangeAt
	// the earliest time of the next one. Both are zero, if no exchange has failed or the instance holds a token.
	ExchangeFailedAt time.Time
	NextExchangeAt   time.Time
}

// CheckDelivery tells, whether the code can be sent right now. It is invoked anew by every [GetState] and
// [RequestCode], as long as the setup is pending, so that a new way to send the code applies at once.
type CheckDelivery func() Delivery

// CodeStatus tells, what has become of the mail with the code.
type CodeStatus int

const (
	// CodeStatusUnknown means, that nothing is known, e.g. because the code is delivered in a custom way.
	CodeStatusUnknown CodeStatus = iota
	// CodeStatusQueued means, that the mail waits to be sent.
	CodeStatusQueued
	// CodeStatusAccepted means, that the mail server or service has accepted the mail.
	CodeStatusAccepted
	// CodeStatusRetrying means, that sending has failed and is repeated.
	CodeStatusRetrying
	// CodeStatusFailed means, that the mail has not been sent and is not repeated.
	CodeStatusFailed
	// CodeStatusInLog means, that a custom [Deliver] has written the code to the log, e.g. during development.
	CodeStatusInLog
)

// ReadCodeStatus returns the status of the code, which has been sent to the address at the given time.
type ReadCodeStatus func(to user.Email, sentAt time.Time) CodeStatus

// CurrentCodeStatus returns the status of the code, which has been sent last, while it is valid, and
// [CodeStatusUnknown] otherwise.
type CurrentCodeStatus func() CodeStatus

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
	ErrNoSSO           = errors.New("onboarding: the setup offers no single sign-on")
	ErrNoDelivery      = errors.New("onboarding: there is no way to send the code")
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

// RequestCode sends a new code to the mail address of the first user. A former code becomes invalid. Without a way
// to send the code, see [DeliveryNone], it fails with [ErrNoDelivery] and keeps the former code.
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

// CompleteWithSSO creates the first user for a confirmed session like [Complete], but without a password: the user
// signs in by the single sign-on of the login service of [State.SSOServer], which it configures. The session is not
// logged in, it returns the address of the login service, which the browser opens next. Until the first user has
// signed in by it, only its address may sign in by single sign-on, then the access applies, see [ApplySSOAccess].
// Thus, the first user cannot end up in another account by signing in with another identity.
type CompleteWithSSO func(sid session.ID, access SSOAccess, preferredLanguage language.Tag) (loginURL string, err error)

// ApplySSOAccess grants the access, which has been chosen by [CompleteWithSSO], after the first user has signed in.
// It does nothing for another user. An allow list, which an administrator has changed meanwhile, is kept.
type ApplySSOAccess func(signedIn user.ID) error

type UseCases struct {
	State           GetState
	RequestCode     RequestCode
	CurrentCode     CurrentCode
	CodeStatus      CurrentCodeStatus
	VerifyCode      VerifyCode
	Verified        Verified
	Complete        Complete
	CompleteWithSSO CompleteWithSSO
	ApplySSOAccess  ApplySSOAccess
	Reset           Reset
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

	// NLSServer is the login service, which the setup offers for the single sign-on. Default is [DefaultNLSServer].
	// A login service, which the settings of the instance contain already, takes precedence.
	NLSServer string
	// DisableSSO offers no single sign-on, thus the first user always gets a password.
	DisableSSO bool
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
	// CheckDelivery is optional, without it the code can always be sent, see [DeliveryAssured].
	CheckDelivery CheckDelivery
	// CodeStatus is optional, without it the status of a code is always [CodeStatusUnknown].
	CodeStatus ReadCodeStatus

	// required for the single sign-on, unless [Options.DisableSSO] is set
	LoadGlobal   settings.LoadGlobal
	StoreGlobal  settings.StoreGlobal
	StartNLSFlow session.StartNLSFlow
	PendingSSO   PendingSSORepository
}

// PendingSSO keeps the access, which [CompleteWithSSO] has chosen, until the first user signs in.
type PendingSSO struct {
	ID PendingSSOID `json:"id"`
	// Admin is the first user.
	Admin user.ID `json:"admin"`
	// Gate is the allow list, which only allows the first user until it signs in.
	Gate []string `json:"gate"`
	// AllowList is the allow list afterward.
	AllowList []string `json:"allowList"`
}

type PendingSSOID string

func (p PendingSSO) Identity() PendingSSOID {
	return p.ID
}

type PendingSSORepository data.Repository[PendingSSO, PendingSSOID]

// PendingSSOKey is the id of the single [PendingSSO].
const PendingSSOKey PendingSSOID = "sso"

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

	if opts.NLSServer == "" {
		opts.NLSServer = DefaultNLSServer
	}

	if deps.LoadGlobal == nil || deps.StoreGlobal == nil || deps.StartNLSFlow == nil || deps.PendingSSO == nil {
		opts.DisableSSO = true
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
		State:           s.state,
		RequestCode:     s.requestCode,
		CurrentCode:     s.currentCode,
		CodeStatus:      s.codeStatus,
		VerifyCode:      s.verifyCode,
		Verified:        s.isVerified,
		Complete:        s.complete,
		CompleteWithSSO: s.completeWithSSO,
		ApplySSOAccess:  s.applySSOAccess,
		Reset:           s.reset,
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

	if st.Pending {
		st.SSOServer = s.ssoServer()
		st.Delivery = s.delivery()
	}

	return st
}

// delivery checks anew, whether the code can be sent.
func (s *setup) delivery() Delivery {
	if s.deps.CheckDelivery == nil {
		return Delivery{Level: DeliveryAssured}
	}

	return s.deps.CheckDelivery()
}

// ssoServer returns the login service, which the setup offers, or nothing.
func (s *setup) ssoServer() string {
	if s.opts.DisableSSO {
		return ""
	}

	if configured := settings.ReadGlobal[user.Settings](s.deps.LoadGlobal).SSONLSServer; configured != "" {
		return configured
	}

	return s.opts.NLSServer
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

	// checked outside the mutex, because it may read repositories
	if s.delivery().Level == DeliveryNone {
		return Code{}, ErrNoDelivery
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

func (s *setup) codeStatus() CodeStatus {
	if s.deps.CodeStatus == nil {
		return CodeStatusUnknown
	}

	c, ok := s.currentCode()
	if !ok {
		return CodeStatusUnknown
	}

	return s.deps.CodeStatus(s.email, c.SentAt)
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

	usr, err := s.createFirstUser(profile)
	if err != nil {
		return user.User{}, err
	}

	s.finish(usr)
	if err := s.deps.LoginUser(sid, usr.ID); err != nil {
		return usr, fmt.Errorf("onboarding: the user has been created, but the session cannot be logged in: %w", err)
	}

	return usr, nil
}

// createFirstUser creates and configures the first user. The caller holds the mutex.
func (s *setup) createFirstUser(profile Profile) (user.User, error) {
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
		s.removeFirstUser(usr)
		return user.User{}, err
	}

	return usr, nil
}

func (s *setup) removeFirstUser(usr user.User) {
	if err := s.deps.Delete(user.SU(), usr.ID); err != nil {
		slog.Error("onboarding: cannot remove the half configured first user", "user", usr.ID, "err", err)
	}
}

// finish ends the setup after the first user has been created. The caller holds the mutex.
func (s *setup) finish(usr user.User) {
	s.done.Store(true)
	s.verified = map[session.ID]time.Time{}
	s.current = nil

	slog.Info("onboarding: first user created", "user", usr.ID)
	if s.deps.Bus != nil {
		s.deps.Bus.Publish(Completed{User: usr.ID, Email: usr.Email, At: s.now()})
	}
}

func (s *setup) completeWithSSO(sid session.ID, access SSOAccess, preferredLanguage language.Tag) (string, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if err := s.check(); err != nil {
		return "", err
	}

	if !s.verifiedLocked(sid) {
		return "", ErrNotVerified
	}

	server := s.ssoServer()
	if server == "" {
		return "", ErrNoSSO
	}

	// without a password, the user gets a random one, which it can reset by mail, if the single sign-on fails
	usr, err := s.createFirstUser(Profile{PreferredLanguage: preferredLanguage})
	if err != nil {
		return "", err
	}

	before := settings.ReadGlobal[user.Settings](s.deps.LoadGlobal)
	usrSettings := before
	usrSettings.SSONLSServer = server
	gate := []string{emailPattern(s.email)}
	usrSettings.SSOAllowList = gate
	if err := s.deps.StoreGlobal(user.SU(), usrSettings); err != nil {
		s.removeFirstUser(usr)
		return "", fmt.Errorf("onboarding: cannot configure the single sign-on: %w", err)
	}

	allowList := access.AllowList(s.email)
	if !slices.Equal(allowList, gate) {
		pending := PendingSSO{ID: PendingSSOKey, Admin: usr.ID, Gate: gate, AllowList: allowList}
		if err := s.deps.PendingSSO.Save(pending); err != nil {
			if err := s.deps.StoreGlobal(user.SU(), before); err != nil {
				slog.Error("onboarding: cannot restore the settings of the single sign-on", "err", err)
			}

			s.removeFirstUser(usr)
			return "", fmt.Errorf("onboarding: cannot keep the access of the single sign-on: %w", err)
		}
	}

	s.finish(usr)
	slog.Info("onboarding: single sign-on configured", "server", server, "allowList", allowList)

	loginURL, err := s.deps.StartNLSFlow(sid)
	if err != nil {
		return "", fmt.Errorf("onboarding: the account has been created, but the single sign-on cannot be started: %w", err)
	}

	return loginURL, nil
}

func (s *setup) applySSOAccess(signedIn user.ID) error {
	if s.opts.DisableSSO {
		return nil
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	optPending, err := s.deps.PendingSSO.FindByID(PendingSSOKey)
	if err != nil {
		return fmt.Errorf("onboarding: cannot find the access of the single sign-on: %w", err)
	}

	if optPending.IsNone() || optPending.Unwrap().Admin != signedIn {
		return nil
	}

	pending := optPending.Unwrap()
	usrSettings := settings.ReadGlobal[user.Settings](s.deps.LoadGlobal)
	if slices.Equal(usrSettings.SSOAllowList, pending.Gate) {
		usrSettings.SSOAllowList = pending.AllowList
		if err := s.deps.StoreGlobal(user.SU(), usrSettings); err != nil {
			return fmt.Errorf("onboarding: cannot grant the access of the single sign-on: %w", err)
		}

		slog.Info("onboarding: the first user has signed in, the single sign-on is open", "allowList", pending.AllowList)
	}

	return s.deps.PendingSSO.DeleteByID(PendingSSOKey)
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
