package auth

import (
	"go.wdy.de/nago/license"
	"golang.org/x/text/language"
	"iter"
)

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

	// Firstname contains the first name, if available.
	// May be something anonymous.
	// Depending on the provider, this may be even empty e.g. due to GDPR requirements.
	Firstname() string

	// Lastname contains the last name, if available.
	// May be something anonymous.
	// Depending on the provider, this may be even empty e.g. due to GDPR requirements.
	Lastname() string

	// Email contains the mail address, if available.
	// May be something anonymous.
	// Depending on the provider, this may be even empty e.g. due to GDPR requirements.
	// You probably should NEVER rely on this to verify that two identities or subjects are the same,
	// especially if the address has never been verified by a second factor (e.g. double opt-in or similar).
	// This is a string, because it remembers you, that at no time this returned value means that the mail
	// is valid in any way. Even if it has been verified once, the domain may have been deleted, or the mailbox is
	// full or locked or even worse, has been captured by a malicious party and compromised.
	Email() string

	// Roles yields over all associated roles. This is important if the domain needs to model
	// resource based access using role identifiers.
	Roles() iter.Seq[RID]

	// HasRole returns true, if the user has the associated role.
	HasRole(RID) bool

	// Groups yields over all associated groups. This is important if the domain needs to model
	// resource based access using group identifiers.
	Groups() iter.Seq[GID]

	// HasGroup returns true, if the user is in the associated group.
	HasGroup(GID) bool

	// Audit checks if this identity has the actual use case permission and may save the positive or
	// negative result in the audit log. An error indicates, that the Subject has not the given permission.
	Audit(permission string) error

	// HasPermission checks, if the Subject has the given permission. A regular use case
	// should use the [Subject.Audit]. However, this may be used e.g. by the UI to show or hide specific aspects.
	HasPermission(permission string) bool

	// HasLicense checks, if the Subject has the given license. This is usually used by the Domain and UI, to enable
	// or disable features based on contracts and payments.
	HasLicense(id license.ID) bool

	// Licenses returns all associated and applicable licenses which are not disabled. If a license is assigned
	// but [license.License.Enabled] is false, it is not contained. Consequently, [HasLicense] will return false.
	// Order is undefined.
	Licenses() iter.Seq[license.ID]

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
