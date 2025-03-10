package permission

import "iter"

// Auditable is used to bootstrap without package cycles.
type Auditable interface {
	// Audit checks if this identity, subject or context has the actual use case permission and may save the positive or
	// negative result in the audit log. An error indicates, that the Subject has not the given permission. The error
	// may just be promoted into error receiving ui components like [alert.BannerError] or [alert.ShowBannerError].
	Audit(permission ID) error

	// HasPermission checks, if the Subject or context has the given permission. A regular use case
	// should use the [Auditable.Audit]. However, this may be used e.g. by the UI to show or hide specific aspects.
	HasPermission(permission ID) bool

	Permissions() iter.Seq[ID]
}

// SU returns the auditable system user or super user. It is not the same user.SU and required to bootstrap
// packages based only on this permission package.
func SU() Auditable {
	return suAuditable{}
}

type suAuditable struct {
}

func (s suAuditable) Audit(permission ID) error {
	return nil
}

func (s suAuditable) HasPermission(permission ID) bool {
	return true
}

func (s suAuditable) Permissions() iter.Seq[ID] {
	return func(yield func(ID) bool) {
		for permission := range All() {
			if !yield(permission.ID) {
				return
			}
		}
	}
}
