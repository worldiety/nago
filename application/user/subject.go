package user

import (
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/license"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/pkg/std/tick"
	"golang.org/x/text/language"
	"iter"
	"log/slog"
	"slices"
	"sync"
	"time"
)

type AuditableUser interface {
	permission.Auditable
	ID() ID
	Valid() bool
}

// Subject is a common contract for an authenticated identity, actor or subject.
// Different implementations may provide additional interfaces or
// expose concrete types behind it.
type Subject interface {
	permission.Auditable

	// ID is the unique actor id within a single NAGO instance. These IDs are generated in a secure way,
	// however, you must not expose that into the public or use it as a source of anonymization.
	// This ID will never change throughout the lifetime of the user and this instance.
	ID() ID

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

	// Avatar returns optionally a resource representing the avatar image. This may be an url, an uri or
	// any id. By default, Nago returns an [image.ID].
	Avatar() string

	// Roles yields over all associated roles. This is important if the domain needs to model
	// resource based access using role identifiers.
	Roles() iter.Seq[role.ID]

	// HasRole returns true, if the user has the associated role.
	HasRole(id role.ID) bool

	// Groups yields over all associated groups. This is important if the domain needs to model
	// resource based access using group identifiers.
	Groups() iter.Seq[group.ID]

	// HasGroup returns true, if the user is in the associated group.
	HasGroup(id group.ID) bool

	// HasLicense checks, if the Subject has the given license. This is usually used by the Domain and UI, to enable
	// or disable features based on contracts and payments.
	HasLicense(id license.ID) bool

	// Licenses returns all associated and applicable licenses which are not disabled. If a license is assigned
	// but [license.License.Enabled] is false, it is not contained. Consequently, [HasLicense] will return false.
	// Order is undefined.
	Licenses() iter.Seq[license.ID]

	// Valid tells us, if the subject has been authenticated and potentially contains permissions.
	// If the mail has never been verified, a user will not be valid.
	Valid() bool

	// Language returns the BCP47 language tag, which encodes a language and locale.
	Language() language.Tag
}

type viewImpl struct {
	user              User
	mutex             sync.Mutex
	repo              Repository
	roleRepo          data.ReadRepository[role.Role, role.ID]
	lastRefreshedAt   time.Time
	refreshInterval   time.Duration
	locale            language.Tag
	roles             []role.ID
	rolesLookup       map[role.ID]struct{}
	groups            []group.ID
	groupsLookup      map[group.ID]struct{}
	permissions       []permission.ID
	permissionsLookup map[permission.ID]struct{}
	licences          []license.ID
	licencesLookup    map[license.ID]struct{}
}

func newViewImpl(repo Repository, roles data.ReadRepository[role.Role, role.ID], user User) *viewImpl {
	v := &viewImpl{
		user:              user,
		lastRefreshedAt:   time.Now(),
		refreshInterval:   5 * time.Minute,
		roleRepo:          roles, // intentionally not the use case, we don't want that each user requires to read all permissions
		repo:              repo,
		groupsLookup:      make(map[group.ID]struct{}),
		permissionsLookup: make(map[permission.ID]struct{}),
		licencesLookup:    make(map[license.ID]struct{}),
		rolesLookup:       make(map[role.ID]struct{}),
	}

	v.load()

	return v
}

func (v *viewImpl) refresh() User {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	if v.refreshInterval == 0 {
		v.refreshInterval = 5 * time.Minute
	}

	now := tick.Now(tick.Minute)
	if now.Sub(v.lastRefreshedAt) >= v.refreshInterval {
		v.load()
	}

	return v.user
}

func (v *viewImpl) load() {
	v.lastRefreshedAt = tick.Now(tick.Minute)

	if v.user.ID == "" {
		slog.Error("user has no id")
		return
	}

	optUsr, err := v.repo.FindByID(v.user.ID)
	if err != nil {
		slog.Error("cannot refresh user", "id", v.user.ID, "err", err)
		v.user = User{ID: v.user.ID, Status: Disabled{}}
		return
	}

	if optUsr.IsNone() {
		slog.Error("user is gone", "id", v.user.ID, "err", err)
		v.user = User{ID: v.user.ID, Status: Disabled{}}
		return
	}

	if v.user.Contact.DisplayLanguage == "und" {
		v.locale = language.English
	} else {
		tag, err := language.Parse(v.user.Contact.DisplayLanguage)
		if err != nil {
			// this is just way to verbose and common not to have such value
			// let us simply ignore it.
			//slog.Error("cannot parse user preferred language", "id", v.user.ID, "err", err)
		}

		v.locale = tag
	}

	v.groups = v.user.Groups
	clear(v.groupsLookup)
	for _, id := range v.groups {
		v.groupsLookup[id] = struct{}{}
	}

	v.roles = v.user.Roles
	clear(v.rolesLookup)
	for _, id := range v.roles {
		v.rolesLookup[id] = struct{}{}
	}

	v.licences = v.user.Licenses
	clear(v.licencesLookup)
	for _, id := range v.licences {
		v.licencesLookup[id] = struct{}{}
	}

	clear(v.permissions)
	v.permissions = v.permissions[:0]
	for _, id := range v.user.Permissions {
		v.permissionsLookup[id] = struct{}{}
	}

	for _, roleId := range v.user.Roles {
		optRole, err := v.roleRepo.FindByID(roleId)
		if err != nil {
			slog.Error("referenced role in user not loadable", "id", v.user.ID, "roleID", roleId, "err", err)
			continue
		}

		if optRole.IsNone() {
			slog.Error("referenced role in user is orphaned", "id", v.user.ID, "roleID", roleId)
			continue
		}

		for _, pid := range optRole.Unwrap().Permissions {
			v.permissionsLookup[pid] = struct{}{}
		}
	}

	for id := range v.permissionsLookup {
		v.permissions = append(v.permissions, id)
	}

	slices.Sort(v.permissions)

}

func (v *viewImpl) Permissions() iter.Seq[permission.ID] {
	if !v.Valid() {
		return func(yield func(permission.ID) bool) {}
	}

	v.mutex.Lock()
	defer v.mutex.Unlock()

	return slices.Values(v.permissions)
}

func (v *viewImpl) Audit(permission permission.ID) error {
	usr := v.refresh()

	if v.user.ID == "" {
		return InvalidSubjectErr
	}

	if !v.Valid() {
		if !usr.EMailVerified {
			return std.NewLocalizedError("Keine Berechtigung", "Die Mail-Adresse zum Konto muss zuerst bestätigt werden.")
		}

		return std.NewLocalizedError("Keine Berechtigung", "Das Nutzerkonto ist nicht gültig.")
	}

	if !v.HasPermission(permission) {
		return PermissionDeniedErr
	}

	return nil
}

func (v *viewImpl) HasPermission(permission permission.ID) bool {
	v.refresh()

	if !v.Valid() {
		return false
	}

	v.mutex.Lock()
	defer v.mutex.Unlock()

	_, ok := v.permissionsLookup[permission]
	return ok
}

func (v *viewImpl) ID() ID {
	v.refresh()

	v.mutex.Lock()
	defer v.mutex.Unlock()

	return v.user.ID
}

func (v *viewImpl) Avatar() string {
	v.refresh()

	v.mutex.Lock()
	defer v.mutex.Unlock()

	return string(v.user.Contact.Avatar)
}

func (v *viewImpl) Name() string {
	v.refresh()

	v.mutex.Lock()
	defer v.mutex.Unlock()

	if v.user.Contact.Firstname == "" || v.user.Contact.Lastname == "" {
		if v.user.Contact.Firstname != "" {
			return v.user.Contact.Firstname
		}

		return v.user.Contact.Lastname
	}

	return v.user.Contact.Firstname + " " + v.user.Contact.Lastname
}

func (v *viewImpl) Firstname() string {
	v.refresh()

	v.mutex.Lock()
	defer v.mutex.Unlock()

	return v.user.Contact.Firstname
}

func (v *viewImpl) Lastname() string {
	v.refresh()

	v.mutex.Lock()
	defer v.mutex.Unlock()

	return v.user.Contact.Lastname
}

func (v *viewImpl) Email() string {
	v.refresh()

	v.mutex.Lock()
	defer v.mutex.Unlock()

	return string(v.user.Email)
}

func (v *viewImpl) Roles() iter.Seq[role.ID] {
	v.refresh()

	if !v.Valid() {
		return func(yield func(role.ID) bool) {}
	}

	v.mutex.Lock()
	defer v.mutex.Unlock()

	return slices.Values(v.roles)
}

func (v *viewImpl) HasRole(id role.ID) bool {
	v.refresh()

	if !v.Valid() {
		return false
	}

	v.mutex.Lock()
	defer v.mutex.Unlock()

	_, ok := v.rolesLookup[id]
	return ok
}

func (v *viewImpl) Groups() iter.Seq[group.ID] {
	v.refresh()

	if !v.Valid() {
		return func(yield func(group.ID) bool) {}
	}

	v.mutex.Lock()
	defer v.mutex.Unlock()

	return slices.Values(v.groups)
}

func (v *viewImpl) HasGroup(id group.ID) bool {
	v.refresh()

	if !v.Valid() {
		return false
	}

	v.mutex.Lock()
	defer v.mutex.Unlock()

	_, ok := v.groupsLookup[id]
	return ok
}

func (v *viewImpl) HasLicense(id license.ID) bool {
	v.refresh()

	if !v.Valid() {
		return false
	}

	v.mutex.Lock()
	defer v.mutex.Unlock()

	_, ok := v.licencesLookup[id]
	return ok
}

func (v *viewImpl) Licenses() iter.Seq[license.ID] {
	v.refresh()

	if !v.Valid() {
		return func(yield func(license.ID) bool) {}
	}

	v.mutex.Lock()
	defer v.mutex.Unlock()

	return slices.Values(v.licences)
}

func (v *viewImpl) Valid() bool {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	return v.user.EMailVerified && v.user.Enabled()
}

func (v *viewImpl) Language() language.Tag {
	v.refresh()

	return v.locale
}
