// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"github.com/worldiety/option"
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/image"
	"go.wdy.de/nago/application/license"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/settings"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/events"
	"go.wdy.de/nago/pkg/std"
	"iter"
	"sync"
	"time"
)

type Repository data.Repository[User, ID]

type Create func(subject permission.Auditable, model ShortRegistrationUser) (User, error)
type FindByID func(subject permission.Auditable, id ID) (option.Opt[User], error)
type FindByMail func(subject permission.Auditable, email Email) (std.Option[User], error)

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

type UpdateOtherLicenses func(subject AuditableUser, id ID, licenses []license.ID) error
type ReadMyContact func(subject AuditableUser) (Contact, error)

// SubjectFromUser returns a subject view for the given user ID. This view can be leaked as long as required and
// updates itself automatically, if the underlying data changes. But keep in mind, that this does not
// mean, that a User (or Subject) is actually logged into a [core.Window]. A window, a session and a subject
// view have NO direct relationship with each other. The process of session handling and logging in and out
// will update the Window reference and the session persistence. It does never make a user invalid.
type SubjectFromUser func(subject permission.Auditable, id ID) (option.Opt[Subject], error)
type EnableBootstrapAdmin func(aliveUntil time.Time, password Password) (ID, error)

// CountAssignedUserLicense counts how many licenses of the given id have been assigned.
type CountAssignedUserLicense func(auditable permission.Auditable, id license.ID) (int, error)

// RevokeAssignedUserLicense can be used to ensure a correctly assigned amount of licenses.
// If a license is removed or the MaxUser limit is lowered, the given count of licenses can be revoked. It
// is undefined which users get revoked. A negative amount will remove the license from all users.
type RevokeAssignedUserLicense func(auditable permission.Auditable, id license.ID, count int) error

// AssignUserLicense tries to assign the given license to the user. If there are not enough free licenses,
// false is returned, otherwise true. If the license is already assigned, this is a no-op and true is returned.
type AssignUserLicense func(auditable permission.Auditable, id ID, license license.ID) (bool, error)

// UnassignUserLicense removes the given license from the user. If no such license is assigned, this is a no-op.
type UnassignUserLicense func(auditable permission.Auditable, id ID, license license.ID) error

// SysUser returns the always mighty build-in system user. This user never authenticates but can always
// be used from the code side to invoke any auditable use case. Use it with caution and only if necessary.
// Never use it, if you could instead pass an authenticated user. Typical scenarios are
// automations, cron jobs, scheduled processes or data(base) migrations.
type SysUser func() Subject

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
type Compact struct {
	Avatar      image.ID
	Displayname string
	Mail        Email
	Valid       bool
}
type UseCases struct {
	Create                    Create
	FindByID                  FindByID
	FindByMail                FindByMail
	FindAll                   FindAll
	ChangeOtherPassword       ChangeOtherPassword
	ChangeMyPassword          ChangeMyPassword
	ChangePasswordWithCode    ChangePasswordWithCode
	Delete                    Delete
	UpdateMyContact           UpdateMyContact
	UpdateOtherContact        UpdateOtherContact
	UpdateOtherRoles          UpdateOtherRoles
	UpdateOtherPermissions    UpdateOtherPermissions
	UpdateOtherLicenses       UpdateOtherLicenses
	UpdateOtherGroups         UpdateOtherGroups
	ReadMyContact             ReadMyContact
	SubjectFromUser           SubjectFromUser
	EnableBootstrapAdmin      EnableBootstrapAdmin
	SysUser                   SysUser
	AuthenticateByPassword    AuthenticateByPassword
	CountAssignedUserLicense  CountAssignedUserLicense
	RevokeAssignedUserLicense RevokeAssignedUserLicense
	ConfirmMail               ConfirmMail
	ResetVerificationCode     ResetVerificationCode
	RequiresPasswordChange    RequiresPasswordChange
	ResetPasswordRequestCode  ResetPasswordRequestCode
	DisplayName               DisplayName
	UpdateAccountStatus       UpdateAccountStatus
	AddUserToGroup            AddUserToGroup
	UpdateVerification        UpdateVerification
	UpdateVerificationByMail  UpdateVerificationByMail
	FindAllIdentifiers        FindAllIdentifiers
	EMailUsed                 EMailUsed
	AssignUserLicense         AssignUserLicense
	UnassignUserLicense       UnassignUserLicense
	CountUsers                CountUsers
}

func NewUseCases(eventBus events.EventBus, loadGlobal settings.LoadGlobal, users Repository, roles data.ReadRepository[role.Role, role.ID], findUserLicenseByID license.FindUserLicenseByID) UseCases {
	findByMailFn := NewFindByMail(users)
	var globalLock sync.Mutex
	createFn := NewCreate(&globalLock, loadGlobal, eventBus, findByMailFn, users)

	findByIdFn := NewFindByID(users)
	findAllFn := NewFindAll(users)

	systemFn := NewSystem()
	enableBootstrapAdminFn := NewEnableBootstrapAdmin(users, systemFn, findByMailFn)

	changeMyPasswordFn := NewChangeMyPassword(&globalLock, users)
	changeOtherPasswordFn := NewChangeOtherPassword(&globalLock, users)
	changePasswordWithCodeFn := NewChangePasswordWithCode(&globalLock, systemFn, users, changeOtherPasswordFn)
	deleteFn := NewDelete(users)

	authenticateByPasswordFn := NewAuthenticatesByPassword(findByMailFn, systemFn)
	subjectFromUserFn := NewViewOf(users, roles)

	readMyContactFn := NewReadMyContact(users)

	updateMyContactFn := NewUpdateMyContact(&globalLock, users)
	updateOtherContactFn := NewUpdateOtherContact(&globalLock, users)
	updateOtherRolesFn := NewUpdateOtherRoles(&globalLock, users)
	updateOtherPermissionsFn := NewUpdateOtherPermissions(&globalLock, users)
	updateOtherGroupsFn := NewUpdateOtherGroups(&globalLock, users)
	updateOtherLicenseFn := NewUpdateOtherLicenses(eventBus, &globalLock, users)

	countAssignedUserLicenseFn := NewCountAssignedUserLicense(users)
	revokeAssignedUserLicenseFn := NewRevokeAssignedUserLicense(&globalLock, users)

	confirmMailFn := NewConfirmMail(&globalLock, users)
	resetVerificationCodeFn := NewResetVerificationCode(&globalLock, users)
	requiresPasswordChangeFn := NewRequiresPasswordChange(systemFn, findByIdFn)
	resetPasswordRequestCodeFn := NewResetPasswordRequestCode(&globalLock, systemFn, users, findByMailFn)

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
		CountAssignedUserLicense:  countAssignedUserLicenseFn,
		RevokeAssignedUserLicense: revokeAssignedUserLicenseFn,
		UpdateOtherLicenses:       updateOtherLicenseFn,
		ConfirmMail:               confirmMailFn,
		ResetVerificationCode:     resetVerificationCodeFn,
		RequiresPasswordChange:    requiresPasswordChangeFn,
		ResetPasswordRequestCode:  resetPasswordRequestCodeFn,
		ChangePasswordWithCode:    changePasswordWithCodeFn,
		DisplayName:               NewDisplayName(users, time.Minute*5),
		UpdateAccountStatus:       NewUpdateAccountStatus(&globalLock, users),
		AddUserToGroup:            NewAddUserToGroup(&globalLock, users),
		UpdateVerification:        NewUpdateVerification(&globalLock, users),
		UpdateVerificationByMail:  NewUpdateVerificationByMail(&globalLock, users, findByMailFn),
		FindAllIdentifiers:        NewFindAllIdentifiers(users),
		EMailUsed:                 NewEMailUsed(users),
		AssignUserLicense:         NewAssignUserLicense(&globalLock, users, countAssignedUserLicenseFn, findUserLicenseByID),
		UnassignUserLicense:       NewUnassignUserLicense(&globalLock, users),
		CountUsers:                NewCountUsers(users),
	}
}
