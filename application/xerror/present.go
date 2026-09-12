// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package xerror turns an error into something that is safe to show to a human or to hand to
// a language model.
//
// Before this package existed, the classification lived unexported inside the alert package.
// As a consequence every other consumer had to improvise: the AI tool loop sent the raw
// err.Error() to the model, while banners deliberately withheld unknown technical detail. The
// same error was therefore presented inconsistently depending on the channel. [Present] is
// the single place which decides what a given error means and what may be said about it.
package xerror

import (
	"crypto/sha3"
	"encoding/hex"
	"errors"
	"os"
	"sync"

	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/pkg/xerrors"
	"golang.org/x/text/language"
)

// Presentation is the classified, localized and privacy checked view of an error.
//
// Everything except Cause is safe to display. Cause is the original error and must only be
// used for logging, never for rendering and never for sending to a third party such as a
// language model.
type Presentation struct {
	Title   string
	Message string

	// Kind is the category the error was classified into. Exactly one applies, which is why
	// this is an enum and not a set of booleans: the classification branches in [Present] are
	// mutually exclusive by construction, and a bool per category would make 2^n states
	// representable when only n are legal.
	Kind Kind

	// Fields maps field names or field paths to messages, see [xerrors.ErrorWithFields].
	//
	// This is orthogonal to Kind rather than a category of its own: an error can be a denial
	// and still carry field messages, for example when a validation error and an
	// authorization error were joined.
	Fields map[string]string

	// cause is the original error. It is deliberately unexported: everything else on this
	// struct is safe to render, and an exported error field sitting next to Title and Message
	// is an invitation to print it into a template. Use [Presentation.Cause] for logging.
	cause error
}

// Kind is the category an error was classified into.
type Kind int

const (
	// KindUnknown means the error matched no known category. Title and Message are then the
	// generic fallback and disclose nothing about the original error.
	KindUnknown Kind = iota
	// KindNotLoggedIn is a missing authentication.
	KindNotLoggedIn
	// KindDenied is a failed authorization of an authenticated subject.
	KindDenied
	// KindNotFound is a missing element.
	KindNotFound
	// KindAlreadyExists is an element that must not exist but does.
	KindAlreadyExists
	// KindValidation is a field bound validation failure, see [Presentation.Fields].
	KindValidation
	// KindPasswordTooWeak is a rejected password.
	KindPasswordTooWeak
	// KindLocalized is an error which already carries its own presentation, see
	// [std.LocalizedError].
	KindLocalized
)

func (k Kind) String() string {
	switch k {
	case KindNotLoggedIn:
		return "not_logged_in"
	case KindDenied:
		return "denied"
	case KindNotFound:
		return "not_found"
	case KindAlreadyExists:
		return "already_exists"
	case KindValidation:
		return "validation"
	case KindPasswordTooWeak:
		return "password_too_weak"
	case KindLocalized:
		return "localized"
	default:
		return "unknown"
	}
}

// Cause returns the original error. For logging only, never render it.
func (p Presentation) Cause() error {
	return p.cause
}

// Recognized reports whether the error matched a known category.
func (p Presentation) Recognized() bool {
	return p.Kind != KindUnknown
}

// Denied reports an authorization failure, which includes a missing authentication. Use it
// when the distinction between the two does not matter, and compare Kind when it does.
func (p Presentation) Denied() bool {
	return p.Kind == KindDenied || p.Kind == KindNotLoggedIn
}

// Validation reports whether field bound messages are available in Fields.
func (p Presentation) Validation() bool {
	return len(p.Fields) > 0
}

// Token returns a short stable hash of the underlying error, suitable to be shown to a user
// so that support can correlate the report with a log entry without the message itself
// having to be disclosed.
func (p Presentation) Token() string {
	if p.cause == nil {
		return ""
	}

	sum := sha3.Sum224([]byte(p.cause.Error()))
	return hex.EncodeToString(sum[:16])
}

// Present classifies err and produces a localized, display safe [Presentation].
//
// The second return value reports whether the error was recognized. If it is false, the
// caller is responsible for the fallback, usually a generic message plus [Presentation.Token].
// Use [PresentOrGeneric] to get that fallback applied already.
//
// b may be nil or may yield a nil bundle, in which case a default bundle is used. This keeps
// the classifier usable from non UI contexts such as the HTTP layer or a background job.
func Present(b i18n.Bundler, err error) (Presentation, bool) {
	if err == nil {
		return Presentation{}, false
	}

	b = bundlerOrDefault(b)

	p := Presentation{cause: err}

	// Field bound messages are collected regardless of which branch below matches, so that a
	// validation error joined with a technical error does not lose either half. This is why
	// Fields is orthogonal to Kind.
	if ewf, ok := xerrors.Collect(err); ok {
		p.Fields = ewf.Fields
	}

	var notLoggedIn interface{ NotLoggedIn() bool }
	if errors.As(err, &notLoggedIn) && notLoggedIn.NotLoggedIn() {
		p.Kind = KindNotLoggedIn
		p.Title = StrAccessDenied.Get(b)
		p.Message = StrNotLoggedInMsg.Get(b)
		return p, true
	}

	if permission.IsDenied(err) {
		p.Kind = KindDenied
		p.Title = StrAccessDenied.Get(b)
		p.Message = StrDeniedMsg.Get(b)

		// Only name a permission if one can actually be named and localized. Previously a
		// literal "?" was substituted here, which produced the unusable instruction
		// "a rights holder must grant ? first" for every denial that is not exactly a
		// user.PermissionDeniedError carrying a registered id.
		if perm, ok := permission.RequiredOf(err); ok {
			if name := perm.LocalizedName(b); name != "" {
				p.Message += " " + StrDeniedGrantHintX.Get(b, i18n.String("permission", name))
			}
		}

		return p, true
	}

	var localized std.LocalizedError
	if errors.As(err, &localized) {
		p.Kind = KindLocalized
		p.Title = localized.Title()
		p.Message = localized.Description()
		return p, true
	}

	if errors.Is(err, os.ErrNotExist) {
		p.Kind = KindNotFound
		p.Title = StrNotFound.Get(b)
		p.Message = StrNotFoundMsg.Get(b)
		return p, true
	}

	if errors.Is(err, os.ErrExist) {
		p.Kind = KindAlreadyExists
		p.Title = StrAlreadyExists.Get(b)
		p.Message = StrAlreadyExistsMsg.Get(b)
		return p, true
	}

	var pwErr user.PasswordStrengthError
	if errors.As(err, &pwErr) {
		p.Kind = KindPasswordTooWeak
		p.Title = StrPasswordTooWeak.Get(b)
		p.Message = passwordMessage(b, pwErr)
		return p, true
	}

	// A pure validation error is recognized, but its Message is a technical dump and must not
	// be shown. The field messages carry the information instead.
	if p.Validation() {
		p.Kind = KindValidation
		p.Title = StrValidationFailed.Get(b)
		p.Message = StrValidationFailedMsg.Get(b)
		return p, true
	}

	return p, false
}

// PresentOrGeneric is like [Present] but never fails. Unrecognized errors are reduced to a
// generic message plus a support token, so that no unvetted technical detail is disclosed.
func PresentOrGeneric(b i18n.Bundler, err error) Presentation {
	if err == nil {
		return Presentation{}
	}

	p, ok := Present(b, err)
	if ok {
		return p
	}

	b = bundlerOrDefault(b)

	p.Title = StrUnexpected.Get(b)
	p.Message = StrUnexpectedMsgX.Get(b, i18n.String("token", p.Token()))

	return p
}

// FallbackLanguage is used whenever a caller has no bundle of its own, for example the HTTP
// layer or a background job.
//
// This must be deterministic: iterating i18n.Default would pick a random language per call,
// so the same API endpoint could answer in German on one request and English on the next.
var FallbackLanguage = language.English

// defaultBundler resolves against the global i18n resources. i18n.StrHnd.Get dereferences the
// bundle without a nil check, so callers which have no window must still be given something
// usable rather than a panic inside the error handler.
type defaultBundler struct{}

func (defaultBundler) Bundle() *i18n.Bundle {
	if b, ok := i18n.Default.MatchBundle(FallbackLanguage); ok {
		return b
	}

	// No bundle at all for the fallback language. Pick the lowest sorted tag so the choice is
	// at least stable across processes.
	var (
		best    *i18n.Bundle
		bestTag string
	)

	for tag, b := range i18n.Default.All() {
		if best == nil || tag.String() < bestTag {
			best, bestTag = b, tag.String()
		}
	}

	return best
}

// BundlerOrDefault returns a bundler which is always safe to pass to i18n. Handles resolved
// against a nil bundler panic, so any caller outside a window must route through this.
func BundlerOrDefault(b i18n.Bundler) i18n.Bundler {
	return bundlerOrDefault(b)
}

func bundlerOrDefault(b i18n.Bundler) i18n.Bundler {
	if b != nil && b.Bundle() != nil {
		return b
	}

	if d := (defaultBundler{}); d.Bundle() != nil {
		return d
	}

	// Nothing is registered at all. Returning a nil bundle would panic in i18n, so fall back
	// to the compiled in literals.
	return literalBundler{}
}

// literalBundler is the last resort when no localization resources exist at all. i18n resolves
// a handle against a nil bundle by returning a placeholder rather than panicking only if the
// bundle itself is non nil, so we hand out an empty but valid bundle.
type literalBundler struct{}

func (literalBundler) Bundle() *i18n.Bundle {
	fallbackOnce.Do(func() {
		fallbackBundle, _ = i18n.Default.AddLanguage(FallbackLanguage)
	})

	return fallbackBundle
}

var (
	fallbackOnce   sync.Once
	fallbackBundle *i18n.Bundle
)

func passwordMessage(b i18n.Bundler, err user.PasswordStrengthError) string {
	switch s := err.Strength; {
	case s.Complexity < user.Strong:
		return StrPwComplexity.Get(b)
	case !s.ContainsUpperAndLowercase:
		return StrPwUpperLower.Get(b)
	case !s.ContainsMinLength:
		return StrPwMinLengthX.Get(b, i18n.Int("min", s.MinLengthRequired))
	case !s.ContainsSpecial:
		return StrPwSpecial.Get(b)
	case !s.ContainsBelowMaxLength:
		return StrPwTooLong.Get(b)
	case !s.ContainsNumber:
		return StrPwNumber.Get(b)
	default:
		return StrPwUnusable.Get(b)
	}
}
