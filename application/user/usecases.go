package user

import (
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/std"
	"iter"
	"sync"
	"time"
)

type Repository data.Repository[User, ID]

type Create func(subject permission.Auditable, model ShortRegistrationUser) (User, error)
type FindByID func(subject permission.Auditable, id ID) (std.Option[User], error)
type FindByMail func(subject permission.Auditable, email Email) (std.Option[User], error)
type FindAll func(subject permission.Auditable) iter.Seq2[User, error]
type ChangeOtherPassword func(subject permission.Auditable, uid ID, pwd Password, pwdRepeated Password) error
type ChangeMyPassword func(subject AuditableUser, oldPassword, newPassword, newRepeated Password) error
type Delete func(subject permission.Auditable, id ID) error
type UpdateMyContact func(subject AuditableUser, contact Contact)
type UpdateOtherContact func(subject AuditableUser, id ID, contact Contact) error
type UpdateOtherRoles func(subject AuditableUser, id ID, roles []role.ID) error
type UpdateOtherPermissions func(subject AuditableUser, id ID, permissions []permission.ID) error
type UpdateOtherGroups func(subject AuditableUser, id ID, groups []group.ID) error
type ReadMyContact func(subject AuditableUser) (Contact, error)
type ViewOf func(subject permission.Auditable, id ID) (std.Option[Subject], error)
type EnableBootstrapAdmin func(aliveUntil time.Time, password Password) (ID, error)

// System returns the always mighty build-in system user. This user never authenticates but can always
// be used from the code side to invoke any auditable use case. Use it with caution and only if necessary.
// Never use it, if you could instead pass an authenticated user. Typical scenarios are
// automations, cron jobs, scheduled processes or data(base) migrations.
type System func() Subject

// AuthenticateByPassword checks mail and password and returns the view of the user to the caller.
type AuthenticateByPassword func(email Email, password Password) (std.Option[User], error)

type UseCases struct {
	Create                 Create
	FindByID               FindByID
	FindByMail             FindByMail
	FindAll                FindAll
	ChangeOtherPassword    ChangeOtherPassword
	ChangeMyPassword       ChangeMyPassword
	Delete                 Delete
	UpdateMyContact        UpdateMyContact
	UpdateOtherContact     UpdateOtherContact
	UpdateOtherRoles       UpdateOtherRoles
	UpdateOtherPermissions UpdateOtherPermissions
	UpdateOtherGroups      UpdateOtherGroups
	ReadMyContact          ReadMyContact
	ViewOf                 ViewOf
	EnableBootstrapAdmin   EnableBootstrapAdmin
	System                 System
	AuthenticateByPassword AuthenticateByPassword
}

func NewUseCases(users Repository, roles data.ReadRepository[role.Role, role.ID]) UseCases {
	findByMailFn := NewFindByMail(users)
	var createLock sync.Mutex
	createFn := NewCreate(&createLock, findByMailFn, users)

	findByIdFn := NewFindByID(users)
	findAllFn := NewFindAll(users)

	systemFn := NewSystem()
	enableBootstrapAdminFn := NewEnableBootstrapAdmin(users, systemFn, findByMailFn)

	changeMyPasswordFn := NewChangeMyPassword(&createLock, users)
	deleteFn := NewDelete(users)

	authenticateByPasswordFn := NewAuthenticatesByPassword(findByMailFn, systemFn)
	viewOfFn := NewViewOf(users, roles)

	readMyContactFn := NewReadMyContact(users)

	return UseCases{
		Create:                 createFn,
		FindByID:               findByIdFn,
		FindByMail:             findByMailFn,
		FindAll:                findAllFn,
		ChangeOtherPassword:    nil,
		ChangeMyPassword:       changeMyPasswordFn,
		Delete:                 deleteFn,
		UpdateMyContact:        nil,
		UpdateOtherContact:     nil,
		UpdateOtherRoles:       nil,
		UpdateOtherPermissions: nil,
		UpdateOtherGroups:      nil,
		ReadMyContact:          readMyContactFn,
		ViewOf:                 viewOfFn,
		EnableBootstrapAdmin:   enableBootstrapAdminFn,
		System:                 systemFn,
		AuthenticateByPassword: authenticateByPasswordFn,
	}
}
