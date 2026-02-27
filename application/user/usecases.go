// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"context"
	"iter"
	"strings"
	"sync"
	"time"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/consent"
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/image"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/settings"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/events"
	"go.wdy.de/nago/pkg/std"
	"golang.org/x/text/language"
)

const Namespace rebac.Namespace = "nago.iam.user"

type Repository data.Repository[User, ID]

type Create func(subject permission.Auditable, model ShortRegistrationUser) (User, error)
type FindByID func(subject permission.Auditable, id ID) (option.Opt[User], error)
type FindByMail func(subject permission.Auditable, email Email) (option.Opt[User], error)

type EMailUsed func(email Email) (bool, error)
type FindAll func(subject permission.Auditable) iter.Seq2[User, error]

type FindAllIdentifiers func(subject permission.Auditable) iter.Seq2[ID, error]
type ChangeOtherPassword func(subject AuditableUser, uid ID, pwd Password, pwdRepeated Password) error
type ChangeMyPassword func(subject AuditableUser, oldPassword, newPassword, newRepeated Password) error
type Delete func(subject permission.Auditable, id ID) error
type UpdateMyContact func(subject AuditableUser, contact Contact) error
type UpdateOtherContact func(subject AuditableUser, id ID, contact Contact) error
type UpdateOtherRoles func(subject AuditableUser, id ID, roles []role.ID) error
type UpdateOtherPermissions func(subject AuditableUser, id ID, permissions []permission.ID) error
type UpdateOtherGroups func(subject AuditableUser, id ID, groups []group.ID) error

type AddUserToGroup func(subject AuditableUser, id ID, group group.ID) error

type ReadMyContact func(subject AuditableUser) (Contact, error)

// SubjectFromUser returns a subject view for the given user ID. This view can be leaked as long as required and
// updates itself automatically, if the underlying data changes. But keep in mind, that this does not
// mean, that a User (or Subject) is actually logged into a [core.Window]. A window, a session and a subject
// view have NO direct relationship with each other. The process of session handling and logging in and out
// will update the Window reference and the session persistence. It does never make a user invalid.
type SubjectFromUser func(subject permission.Auditable, id ID) (option.Opt[Subject], error)
type EnableBootstrapAdmin func(aliveUntil time.Time, password Password) (ID, error)

// SysUser returns the always mighty build-in system user. This user never authenticates but can always
// be used from the code side to invoke any auditable use case. Use it with caution and only if necessary.
// Never use it, if you could instead pass an authenticated user. Typical scenarios are
// automations, cron jobs, scheduled processes or data(base) migrations.
type SysUser func() Subject

// GetAnonUser returns an anonymous user which has all the declared permissions, roles and groups defined by the
// settings. Note, that the anon user has english as its default language. We currently don't plan to
// change that, for logical and performance implications.
type GetAnonUser func() Subject

// AuthenticateByPassword checks mail and password and returns the view of the user to the caller.
type AuthenticateByPassword func(email Email, password Password) (std.Option[User], error)

type ConfirmMail func(userId ID, code string) error

type ResetVerificationCode func(uid ID, lifetime time.Duration) (code string, err error)

type ResetPasswordRequestCode func(mail Email, lifetime time.Duration) (code string, err error)

type RequiresPasswordChange func(uid ID) (bool, error)

type ChangePasswordWithCode func(uid ID, code string, newPassword Password, newRepeated Password) error

// DisplayName leaks information details about the given user, if you already know that the ID is there.
// The returned information may be stale, to improve performance.
type DisplayName func(uid ID) Compact

type UpdateAccountStatus func(subject permission.Auditable, id ID, status AccountStatus) error

type UpdateVerification func(subject permission.Auditable, id ID, verified bool) error

type UpdateVerificationByMail func(subject permission.Auditable, mail Email, verified bool) error

type CountUsers func() (int, error)

// Consent either approves or revokes a given consent. Usually, this is something caused by GDPR concerns.
// If the [consent.Action.At] time is the zero value, the current time stamp is inserted.
type Consent func(subject AuditableUser, user ID, consentID consent.ID, action consent.Action) error

type LoadSettings func(subject AuditableUser) Settings

// ListGroups returns all groups which the given user is a member of according to the ReBAC system.
// A valid user can always read its own groups.
type ListGroups func(subject AuditableUser, uid ID) iter.Seq2[group.ID, error]

// ListRoles returns all roles which the given user is a member of according to the ReBAC system.
// A valid user can always read its own roles.
type ListRoles func(subject AuditableUser, uid ID) iter.Seq2[role.ID, error]

// ListGlobalPermissions returns all global permissions which the given user is a member of according to the ReBAC
// system. A global permission has the permission id as a relation and as target [rebac.Global] and [rebac.AllInstances].
// A valid user can always read its own global permissions.
type ListGlobalPermissions func(subject AuditableUser, uid ID) iter.Seq2[permission.ID, error]

type SingleSignOnUser struct {
	Firstname         string
	Lastname          string
	Name              string
	Email             Email
	PreferredLanguage string

	Salutation        string
	Title             string
	Position          string
	CompanyName       string
	City              string
	PostalCode        string
	State             string
	Country           string
	ProfessionalGroup string
	MobilePhone       string
	AboutMe           string
}

func (u SingleSignOnUser) FirstName() string {
	if u.Firstname != "" {
		return u.Firstname
	}

	return strings.Split(u.Name, " ")[0]
}

func (u SingleSignOnUser) LastName() string {
	if u.Lastname != "" {
		return u.Lastname
	}

	if tokens := strings.Split(u.Name, " ")[0]; len(tokens) > 1 {
		return strings.Split(u.Name, " ")[1]
	}

	return u.Name
}

// MergeSingleSignOnUser accepts the given user credentials as verified and trusted. If any existing user
// is found with the same mail address, it will be marked as SSO-managed and the password-login and profile editing
// is disabled.
type MergeSingleSignOnUser func(user SingleSignOnUser, avatar []byte) (ID, error)

type ExportFormat int

const (
	ExportCSV ExportFormat = iota
)

type ExportUsersOptions struct {
	Format   ExportFormat
	Language language.Tag
}

type ExportUsers func(subject Subject, users []ID, opts ExportUsersOptions) ([]byte, error)

type Compact struct {
	ID          ID
	Avatar      image.ID
	Displayname string
	Mail        Email
	Valid       bool
}
type UseCases struct {
	Create                   Create
	FindByID                 FindByID
	FindByMail               FindByMail
	FindAll                  FindAll
	ChangeOtherPassword      ChangeOtherPassword
	ChangeMyPassword         ChangeMyPassword
	ChangePasswordWithCode   ChangePasswordWithCode
	Delete                   Delete
	UpdateMyContact          UpdateMyContact
	UpdateOtherContact       UpdateOtherContact
	UpdateOtherRoles         UpdateOtherRoles
	UpdateOtherPermissions   UpdateOtherPermissions
	UpdateOtherGroups        UpdateOtherGroups
	ReadMyContact            ReadMyContact
	SubjectFromUser          SubjectFromUser
	EnableBootstrapAdmin     EnableBootstrapAdmin
	SysUser                  SysUser
	AuthenticateByPassword   AuthenticateByPassword
	ConfirmMail              ConfirmMail
	ResetVerificationCode    ResetVerificationCode
	RequiresPasswordChange   RequiresPasswordChange
	ResetPasswordRequestCode ResetPasswordRequestCode
	DisplayName              DisplayName
	UpdateAccountStatus      UpdateAccountStatus
	AddUserToGroup           AddUserToGroup
	UpdateVerification       UpdateVerification
	UpdateVerificationByMail UpdateVerificationByMail
	FindAllIdentifiers       FindAllIdentifiers
	EMailUsed                EMailUsed
	CountUsers               CountUsers
	GetAnonUser              GetAnonUser
	Consent                  Consent
	ExportUsers              ExportUsers
	MergeSingleSignOnUser    MergeSingleSignOnUser

	// deprecated: use rules api
	AddResourcePermissions AddResourcePermissions
	// deprecated: use rules api
	RemoveResourcePermissions RemoveResourcePermissions
	// deprecated: use rules api
	ListResourcePermissions ListResourcePermissions
	// deprecated: use rules api
	GrantPermissions GrantPermissions
	// deprecated: use rules api
	ListGrantedPermissions ListGrantedPermissions
	// deprecated: use rules api
	ListGrantedUsers ListGrantedUsers

	ListGroups            ListGroups
	ListRoles             ListRoles
	ListGlobalPermissions ListGlobalPermissions

	Resources rebac.Resources
}

func NewUseCases(ctx context.Context, eventBus events.EventBus, rdb *rebac.DB, loadGlobal settings.LoadGlobal, users data.NotifyRepository[User, ID], roles data.ReadRepository[role.Role, role.ID], groups group.FindAll, findRoleByID role.FindByID, listRolePerms role.ListPermissions, createSrcSet image.CreateSrcSet) UseCases {
	findByMailFn := NewFindByMail(users)
	var globalLock sync.Mutex
	createFn := NewCreate(&globalLock, rdb, loadGlobal, eventBus, findByMailFn, users)

	findByIdFn := NewFindByID(users)
	findAllFn := NewFindAll(users)

	systemFn := NewSystem(ctx)
	enableBootstrapAdminFn := NewEnableBootstrapAdmin(users, systemFn, findByMailFn, rdb)

	changeMyPasswordFn := NewChangeMyPassword(&globalLock, users)
	changeOtherPasswordFn := NewChangeOtherPassword(&globalLock, users)
	changePasswordWithCodeFn := NewChangePasswordWithCode(&globalLock, systemFn, users, changeOtherPasswordFn)
	deleteFn := NewDelete(users)

	authenticateByPasswordFn := NewAuthenticatesByPassword(findByMailFn, systemFn)
	subjectFromUserFn := NewViewOf(ctx, eventBus, users, rdb)

	readMyContactFn := NewReadMyContact(users)

	updateMyContactFn := NewUpdateMyContact(&globalLock, eventBus, users)
	updateOtherContactFn := NewUpdateOtherContact(&globalLock, eventBus, users)
	updateOtherRolesFn := NewUpdateOtherRoles(rdb)
	updateOtherPermissionsFn := NewUpdateOtherPermissions(rdb)
	updateOtherGroupsFn := NewUpdateOtherGroups(rdb, groups)

	confirmMailFn := NewConfirmMail(&globalLock, users)
	resetVerificationCodeFn := NewResetVerificationCode(&globalLock, users)
	requiresPasswordChangeFn := NewRequiresPasswordChange(systemFn, findByIdFn)
	resetPasswordRequestCodeFn := NewResetPasswordRequestCode(&globalLock, systemFn, users, findByMailFn)

	findAllIdentsFn := NewFindAllIdentifiers(users)

	return UseCases{
		Create:                    createFn,
		FindByID:                  findByIdFn,
		FindByMail:                findByMailFn,
		FindAll:                   findAllFn,
		ChangeOtherPassword:       changeOtherPasswordFn,
		ChangeMyPassword:          changeMyPasswordFn,
		Delete:                    deleteFn,
		UpdateMyContact:           updateMyContactFn,
		UpdateOtherContact:        updateOtherContactFn,
		UpdateOtherRoles:          updateOtherRolesFn,
		UpdateOtherPermissions:    updateOtherPermissionsFn,
		UpdateOtherGroups:         updateOtherGroupsFn,
		ReadMyContact:             readMyContactFn,
		SubjectFromUser:           subjectFromUserFn,
		EnableBootstrapAdmin:      enableBootstrapAdminFn,
		SysUser:                   systemFn,
		AuthenticateByPassword:    authenticateByPasswordFn,
		ConfirmMail:               confirmMailFn,
		ResetVerificationCode:     resetVerificationCodeFn,
		RequiresPasswordChange:    requiresPasswordChangeFn,
		ResetPasswordRequestCode:  resetPasswordRequestCodeFn,
		ChangePasswordWithCode:    changePasswordWithCodeFn,
		DisplayName:               NewDisplayName(users, time.Minute*5),
		UpdateAccountStatus:       NewUpdateAccountStatus(&globalLock, users),
		AddUserToGroup:            NewAddUserToGroup(rdb),
		UpdateVerification:        NewUpdateVerification(&globalLock, users),
		UpdateVerificationByMail:  NewUpdateVerificationByMail(&globalLock, users, findByMailFn),
		FindAllIdentifiers:        findAllIdentsFn,
		EMailUsed:                 NewEMailUsed(users),
		CountUsers:                NewCountUsers(users),
		GetAnonUser:               NewGetAnonUser(ctx, eventBus, loadGlobal, findRoleByID, listRolePerms),
		Consent:                   NewConsent(&globalLock, eventBus, users),
		AddResourcePermissions:    NewAddResourcePermissions(rdb),
		RemoveResourcePermissions: NewRemoveResourcePermissions(rdb),
		ListResourcePermissions:   NewListResourcePermissions(rdb),
		GrantPermissions:          NewGrantPermissions(&globalLock, users, findByIdFn, rdb),
		ListGrantedPermissions:    NewListGrantedPermissions(rdb),
		ListGrantedUsers:          NewListGrantedUsers(rdb),
		ExportUsers:               NewExportUsers(users),
		MergeSingleSignOnUser:     NewMergeSingleSignOnUser(&globalLock, users, findByMailFn, loadGlobal, createSrcSet, rdb),
		ListGroups:                NewListGroups(rdb),
		ListRoles:                 NewListRoles(rdb),
		ListGlobalPermissions:     NewListGlobalPermissions(rdb),
		Resources:                 NewResources(findAllIdentsFn, findByIdFn),
	}
}
