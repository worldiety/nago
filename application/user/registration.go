// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"go.wdy.de/nago/application/settings"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/std"
)

// The self registration starts with the mail address: a visitor requests a link, which is sent to the address,
// and fills in the rest of the account after opening it. The account is created only at the end, already verified,
// so there are no half accounts of addresses which have never been confirmed. A request holds nothing but the
// mail address and the hash of the code, it ends with the account or expires.

// DefaultRegistrationLifeTime is the time a registration link stays valid.
const DefaultRegistrationLifeTime = time.Hour * 24

type RegistrationID string

// RegistrationRequest is a pending self registration. The code itself is never persisted, only its salted hash.
type RegistrationRequest struct {
	ID         RegistrationID `json:"id"`
	Email      Email          `json:"email"`
	Salt       []byte         `json:"salt"`
	CodeHash   []byte         `json:"codeHash"`
	CreatedAt  time.Time      `json:"createdAt"`
	ValidUntil time.Time      `json:"validUntil"`
	Attempts   int            `json:"attempts,omitempty"`
}

func (r RegistrationRequest) Identity() RegistrationID {
	return r.ID
}

type RegistrationRepository data.Repository[RegistrationRequest, RegistrationID]

const (
	// ErrRegistrationDisabled is the cause of the error, which is returned, if self registration is not allowed.
	ErrRegistrationDisabled std.Error = "registration: self registration is not allowed"
	// ErrRegistrationInvalid is the cause of the error, which is returned for a wrong, expired or used up link.
	ErrRegistrationInvalid std.Error = "registration: the link is invalid or has expired"
	// ErrRegistrationDelivery is the cause of the error, which is returned, if the mail cannot be sent.
	ErrRegistrationDelivery std.Error = "registration: the mail cannot be sent"
)

func errRegistrationDisabled() error {
	return std.NewLocalizedError("Registrierung nicht möglich", "Die Registrierung ist derzeit nicht möglich.").WithError(ErrRegistrationDisabled)
}

func errRegistrationInvalid() error {
	return std.NewLocalizedError("Link nicht gültig", "Der Link ist ungültig, abgelaufen oder wurde bereits verwendet.").WithError(ErrRegistrationInvalid)
}

func errRegistrationDelivery() error {
	return std.NewLocalizedError("Link nicht gesendet", "Der Link konnte nicht gesendet werden. Bitte versuchen Sie es später erneut.").WithError(ErrRegistrationDelivery)
}

// RegistrationCooldownError is returned, if a link is requested too often.
type RegistrationCooldownError struct {
	RetryAt time.Time
}

func (e RegistrationCooldownError) Error() string {
	return fmt.Sprintf("registration: a new link can be requested at %s", e.RetryAt.Format(time.RFC3339))
}

// RequestRegistration sends a registration link to the given address. To not expose, which addresses are already
// registered, it behaves the same for those, but sends a notice instead, which points to the login and the password
// reset. Only an invalid or not allowed address, the rate limit and a failed delivery are reported back.
//
// The client is an opaque key of the requester, e.g. the session ID, to limit the links a single client can
// request. An empty client is only limited by the address and the instance.
type RequestRegistration func(client string, email Email) error

// FindRegistration returns the mail address of a valid registration link, without using it up.
type FindRegistration func(id RegistrationID, code string) (Email, error)

// CompleteRegistration creates the account of a valid registration link with the mail address of the request.
// The account is verified, because opening the link proved the ownership of the address. The link is used up, but
// only if the account has been created, so that e.g. a rejected password can be corrected.
type CompleteRegistration func(id RegistrationID, code string, model ShortRegistrationUser) (User, error)

// DeliverRegistration sends the link of a registration to the address.
type DeliverRegistration func(to Email, id RegistrationID, code string, validUntil time.Time) error

// DeliverAlreadyRegistered tells the owner of an already registered address, that someone tried to register it.
type DeliverAlreadyRegistered func(to Email) error

// Registration groups the use cases of the self registration. Like the onboarding, they share a state (the rate
// limits) and send their mails by injected functions instead of events, because a failed delivery must be reported
// back to the visitor. Thus, they are not part of [UseCases], see [NewRegistration].
type Registration struct {
	Request  RequestRegistration
	Find     FindRegistration
	Complete CompleteRegistration
}

type RegistrationOptions struct {
	Lifetime          time.Duration // default DefaultRegistrationLifeTime
	ResendCooldown    time.Duration // per address, default 1 minute
	MaxPerAddressHour int           // links per address and hour, default 5
	MaxPerClientHour  int           // links per client and hour, default 10
	// MaxPerHour limits the links of the whole instance per hour, default 200. It protects the mail reputation of
	// the instance, but once reached, nobody can register until the hour is over.
	MaxPerHour  int
	MaxAttempts int // wrong codes per request, default 5
}

type RegistrationDeps struct {
	Repository        RegistrationRepository
	LoadGlobal        settings.LoadGlobal
	EMailUsed         EMailUsed
	Create            Create
	Deliver           DeliverRegistration
	DeliverRegistered DeliverAlreadyRegistered
}

type registration struct {
	opts RegistrationOptions
	deps RegistrationDeps
	now  func() time.Time

	// the mutex is also held while a mail is handed over to the mail queue, which is cheap
	mutex     sync.Mutex
	sent      []time.Time            // send times of the whole instance in the last hour
	perMail   map[Email][]time.Time  // send times per address in the last hour
	perClient map[string][]time.Time // send times per client in the last hour
}

// NewRegistration creates the use cases of the self registration. Expired requests are removed at once and with
// every new request.
func NewRegistration(opts RegistrationOptions, deps RegistrationDeps) Registration {
	r := newRegistration(opts, deps)
	r.removeExpired()
	return r.useCases()
}

func newRegistration(opts RegistrationOptions, deps RegistrationDeps) *registration {
	if opts.Lifetime <= 0 {
		opts.Lifetime = DefaultRegistrationLifeTime
	}

	if opts.ResendCooldown <= 0 {
		opts.ResendCooldown = time.Minute
	}

	if opts.MaxPerAddressHour <= 0 {
		opts.MaxPerAddressHour = 5
	}

	if opts.MaxPerClientHour <= 0 {
		opts.MaxPerClientHour = 10
	}

	if opts.MaxPerHour <= 0 {
		opts.MaxPerHour = 200
	}

	if opts.MaxAttempts <= 0 {
		opts.MaxAttempts = 5
	}

	return &registration{
		opts:      opts,
		deps:      deps,
		now:       time.Now,
		perMail:   map[Email][]time.Time{},
		perClient: map[string][]time.Time{},
	}
}

func (r *registration) useCases() Registration {
	return Registration{
		Request:  r.request,
		Find:     r.find,
		Complete: r.complete,
	}
}

// checkAllowed validates the address against the settings, the same way [Create] does.
func (r *registration) checkAllowed(email Email) error {
	s := settings.ReadGlobal[Settings](r.deps.LoadGlobal)
	if !s.SelfRegistration {
		return errRegistrationDisabled()
	}

	if !email.Valid() {
		return std.NewLocalizedError("Eingabebeschränkung", "Dieses E-Mail Format wird nicht akzeptiert.")
	}

	return checkAllowedDomain(s, email)
}

func (r *registration) request(client string, email Email) error {
	email = normalizeEmail(email)
	if err := r.checkAllowed(email); err != nil {
		return err
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	now := r.now()
	if err := r.rateLimitLocked(client, email, now); err != nil {
		return err
	}

	used, err := r.deps.EMailUsed(email)
	if err != nil {
		return fmt.Errorf("cannot check mail: %w", err)
	}

	// the limits count for registered addresses as well, otherwise their behavior would tell them apart
	r.sent = append(r.sent, now)
	r.perMail[email] = append(r.perMail[email], now)
	if client != "" {
		r.perClient[client] = append(r.perClient[client], now)
	}

	if used {
		if r.deps.DeliverRegistered != nil {
			if err := r.deps.DeliverRegistered(email); err != nil {
				// security note: report the same error as for a new address, which tells nothing
				slog.Error("registration: cannot send the notice to a registered address", "err", err)
				return errRegistrationDelivery()
			}
		}

		return nil
	}

	if err := r.removeLocked(func(req RegistrationRequest) bool {
		return req.Email == email || !now.Before(req.ValidUntil)
	}); err != nil {
		return err
	}

	code := data.RandIdent[string]()
	req := RegistrationRequest{
		ID:         data.RandIdent[RegistrationID](),
		Email:      email,
		Salt:       make([]byte, 16),
		CreatedAt:  now,
		ValidUntil: now.Add(r.opts.Lifetime),
	}

	if _, err := rand.Read(req.Salt); err != nil {
		return err
	}

	req.CodeHash = hashRegistrationCode(req.Salt, code)

	if err := r.deps.Repository.Save(req); err != nil {
		return fmt.Errorf("cannot save registration request: %w", err)
	}

	if err := r.deps.Deliver(email, req.ID, code, req.ValidUntil); err != nil {
		_ = r.deps.Repository.DeleteByID(req.ID)
		slog.Error("registration: cannot send the link", "err", err)
		return errRegistrationDelivery()
	}

	return nil
}

func (r *registration) rateLimitLocked(client string, email Email, now time.Time) error {
	hourAgo := now.Add(-time.Hour)
	r.sent = pruneTimesBefore(r.sent, hourAgo)
	pruneTimesByKey(r.perMail, hourAgo)
	pruneTimesByKey(r.perClient, hourAgo)

	mine := r.perMail[email]
	if len(mine) > 0 {
		if retryAt := mine[len(mine)-1].Add(r.opts.ResendCooldown); now.Before(retryAt) {
			return RegistrationCooldownError{RetryAt: retryAt}
		}
	}

	if len(mine) >= r.opts.MaxPerAddressHour {
		return RegistrationCooldownError{RetryAt: mine[0].Add(time.Hour)}
	}

	if client != "" {
		if theirs := r.perClient[client]; len(theirs) >= r.opts.MaxPerClientHour {
			return RegistrationCooldownError{RetryAt: theirs[0].Add(time.Hour)}
		}
	}

	if len(r.sent) >= r.opts.MaxPerHour {
		return RegistrationCooldownError{RetryAt: r.sent[0].Add(time.Hour)}
	}

	return nil
}

func pruneTimesBefore(times []time.Time, limit time.Time) []time.Time {
	i := 0
	for i < len(times) && times[i].Before(limit) {
		i++
	}

	return times[i:]
}

func pruneTimesByKey[K comparable](m map[K][]time.Time, limit time.Time) {
	for key, times := range m {
		if times = pruneTimesBefore(times, limit); len(times) == 0 {
			delete(m, key)
		} else {
			m[key] = times
		}
	}
}

// removeExpired removes the requests, which have expired, e.g. while the instance was not running.
func (r *registration) removeExpired() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	now := r.now()
	if err := r.removeLocked(func(req RegistrationRequest) bool {
		return !now.Before(req.ValidUntil)
	}); err != nil {
		slog.Error("registration: cannot remove expired requests", "err", err)
	}
}

func (r *registration) removeLocked(match func(RegistrationRequest) bool) error {
	var ids []RegistrationID
	for req, err := range r.deps.Repository.All() {
		if err != nil {
			return fmt.Errorf("cannot read registration requests: %w", err)
		}

		if match(req) {
			ids = append(ids, req.ID)
		}
	}

	for _, id := range ids {
		if err := r.deps.Repository.DeleteByID(id); err != nil {
			return fmt.Errorf("cannot remove registration request: %w", err)
		}
	}

	return nil
}

func hashRegistrationCode(salt []byte, code string) []byte {
	h := sha256.New()
	h.Write(salt)
	h.Write([]byte(code))
	return h.Sum(nil)
}

// validLocked returns the request of a valid link. A wrong code counts as attempt, too many remove the request.
func (r *registration) validLocked(id RegistrationID, code string) (RegistrationRequest, error) {
	if id == "" || code == "" {
		return RegistrationRequest{}, errRegistrationInvalid()
	}

	optReq, err := r.deps.Repository.FindByID(id)
	if err != nil {
		return RegistrationRequest{}, fmt.Errorf("cannot find registration request: %w", err)
	}

	if optReq.IsNone() {
		return RegistrationRequest{}, errRegistrationInvalid()
	}

	req := optReq.Unwrap()
	if !r.now().Before(req.ValidUntil) || req.Attempts >= r.opts.MaxAttempts {
		_ = r.deps.Repository.DeleteByID(req.ID)
		return RegistrationRequest{}, errRegistrationInvalid()
	}

	if subtle.ConstantTimeCompare(hashRegistrationCode(req.Salt, code), req.CodeHash) != 1 {
		req.Attempts++
		if req.Attempts >= r.opts.MaxAttempts {
			_ = r.deps.Repository.DeleteByID(req.ID)
		} else if err := r.deps.Repository.Save(req); err != nil {
			return RegistrationRequest{}, fmt.Errorf("cannot save registration request: %w", err)
		}

		return RegistrationRequest{}, errRegistrationInvalid()
	}

	return req, nil
}

func (r *registration) find(id RegistrationID, code string) (Email, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	req, err := r.validLocked(id, code)
	if err != nil {
		return "", err
	}

	return req.Email, nil
}

func (r *registration) complete(id RegistrationID, code string, model ShortRegistrationUser) (User, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	req, err := r.validLocked(id, code)
	if err != nil {
		return User{}, err
	}

	if err := r.checkAllowed(req.Email); err != nil {
		return User{}, err
	}

	// security note: the address and the verification come from the request, never from the caller
	model.Email = req.Email
	model.SelfRegistered = true
	model.Verified = true
	model.NotifyUser = false

	usr, err := r.deps.Create(SU(), model)
	if err != nil {
		// the link stays valid, so that e.g. a rejected password can be corrected
		return User{}, err
	}

	if err := r.deps.Repository.DeleteByID(req.ID); err != nil {
		// the account is complete and the address is used now, so the request cannot create another one
		slog.Error("registration: account created but cannot remove the request", "err", err)
	}

	return usr, nil
}
