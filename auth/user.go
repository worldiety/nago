package auth

import (
	"context"
	"regexp"
)

type UserID string
type RoleID string

const (
	Admin RoleID = "admin"
)

type EMail string

var regexMail = regexp.MustCompile(`^[\w-.]+@([\w-]+\.)+[\w-]{2,4}$`)

// Valid checks if the Mail looks like structural valid mail. It does not mean that the address actually exists
// or has been verified.
func (e EMail) Valid() bool {
	// see https://cheatsheetseries.owasp.org/cheatsheets/Input_Validation_Cheat_Sheet.html#email-address-validation
	if len(e) > 254 {
		return false
	}

	return regexMail.FindString(string(e)) == string(e)
}

// User is a common contract for an authenticated user. Different implementations may provide additional interfaces or
// expose concrete types behind it.
// To get access to a user, a [presentation.PageHandler] must be configured to be authenticated.
type User interface {
	// UserID is the unique user id which is never empty but its nature is totally undefined and depends on the provider.
	// Its value is constant between different sessions, e.g. keycloak provides a UUID in the sub(ject) claim.
	// If you need resource-based authorization use this id for association.
	// Use [User.Verified] and [User.Valid] in your domain logic.
	UserID() UserID

	// Verified is true if the user has been verified in some way, e.g. by double opt-in or by an administrator.
	// This is often important, to only allow critical operations on verified users.
	Verified() bool

	// Roles contains an unspecified bunch of associated role names which can be used to distinguish between
	// different static user authorization, like users having administrator privileges.
	// However, it is not suited to implement resource based authorizations which must be usually modelled explicitly
	// in the domain layer.
	// This confirms to iter.Seq[RoleID].
	Roles(yield func(RoleID) bool)

	// Email contains the primary unparsed and unvalidated mail address, if available. If no mail is provided,
	// it is empty, e.g. due to GDPR requirements. Note also, that a user can always change its mail address, so
	// using this as a primary key in your domain logic is probably always wrong.
	// It is not a dedicated email type, because we cannot trust external implementations but they often provide
	// this kind of property.
	Email() EMail

	// Name contains the natural name of the user, e.g. a firstname lastname tuple. Depending on the provider,
	// this may be empty e.g. due to GDPR requirements.
	Name() string

	// Valid returns only true, if the user has been authenticated and the lease and internal session
	// of e.g. a token is still valid.
	Valid() bool
}

type userKey string

// FromContext extracts the current user from the context or returns nil if not authenticated.
// Unauthenticated means also that the session or token lifetime has been expired. See also [WithContext].
func FromContext(ctx context.Context) User {
	user := ctx.Value(userKey("user"))
	if user == nil {
		return nil
	}

	return user.(User)
}

// WithContext allocates a new context with the supplied user value. See also [FromContext].
func WithContext(ctx context.Context, user User) context.Context {
	return context.WithValue(ctx, userKey("user"), user)
}
