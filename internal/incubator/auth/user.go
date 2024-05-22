package auth

import (
	"time"
)

// ???
type UserN interface {
	// Firstname of the user, if any.
	Firstname() string

	// Lastname is the family name, if any.
	Lastname() string

	// Email contains the primary unparsed and (maybe un)validated but unique mail address, if available.
	// If no mail is provided, it is empty, e.g. due to GDPR requirements.
	// Note also, that a user can always change its mail address, so
	// using this as a primary key in your domain logic is probably always wrong.
	// It is almost always better to use UserID as a foreign key in your domain.
	// An Email must be compared in an insensitive way, so use [Email.Equals].
	Email() Email

	// EmailVerified is true if the user has been verified in some way, e.g. by double opt-in or by an administrator.
	// This is often important, to only allow critical operations on verified users or to send confidential content.
	EmailVerified() bool

	// LastAuthenticatedAt returns the time at which the user has been (re-)authenticated the last time.
	// Returns [time.Time.IsZero] if not authenticated at all. Note, that risky operations
	LastAuthenticatedAt() time.Time

	// Verified is true if the identity has been verified in some way, e.g. by double opt-in or by an administrator.
	// This is often important, to only allow critical operations on verified users or to send confidential content.
	Verified() bool
}

type GroupID string
type ResourceID string

const (
	Global ResourceID = "ora.resource.global"
)

type RoleID string
