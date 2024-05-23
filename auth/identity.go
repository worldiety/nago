package auth

import "golang.org/x/text/language"

// UID is the User or Subject ID.
type UID string

// GID is the Group ID.
type GID string

// RID is the role ID.
type RID string

// Subject is a common contract for an authenticated identity, actor or subject.
// Different implementations may provide additional interfaces or
// expose concrete types behind it.
type Subject interface {
	// ID is the unique actor id which is never empty but its nature is totally undefined and depends on the provider.
	// Its value is constant between different sessions, e.g. keycloak provides a UUID in the sub(ject) claim.
	// If you need resource-based authorization use this id for association in your domain.
	ID() UID

	// Name contains an arbitrary non-unique calling name of the identity.
	// May be something anonymous.
	// Depending on the provider, this may be even empty e.g. due to GDPR requirements.
	Name() string

	// Roles yields over all associated roles. This is important if the domain needs to model
	// resource based access using role identifiers.
	Roles(yield func(RID) bool)

	// Groups yields over all associated groups. This is important if the domain needs to model
	// resource based access using group identifiers.
	Groups(yield func(GID) bool)

	// Audit checks if this identity has the actual use case permission and may save the positive or
	// negative result in the audit log. An error indicates, that the Subject has not the given permission.
	Audit(permission string) error

	// HasPermission checks, if the Subject has the given permission. A regular use case
	// should use the [Subject.Audit]. However, this may be used e.g. by the UI to show or hide specific aspects.
	HasPermission(permission string) bool

	// Valid tells us, if the subject has been authenticated and potentially contains permissions.
	Valid() bool

	// Language returns the BCP47 language tag, which encodes a language and locale.
	Language() language.Tag
}

func OneOf(subject Subject, permissions ...string) bool {
	for _, permission := range permissions {
		if subject.HasPermission(permission) {
			return true
		}
	}

	return false
}
