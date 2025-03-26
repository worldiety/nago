// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package usercircle

import (
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/std"
	"iter"
	"os"
	"slices"
	"sync"
)

// MyCircleMembers returns all those users defined in the nago IAM which are included by the circle member rules.
type MyCircleMembers func(subject auth.Subject, id ID) iter.Seq2[user.User, error]

type IsCircleAdmin func(uid user.ID) (bool, error)

// MyCircles returns all circles, in which the subject is declared as an administrator.
type MyCircles func(subject auth.Subject) iter.Seq2[Circle, error]

type MyRoles func(subject auth.Subject, circle ID) iter.Seq2[role.Role, error]
type MyGroups func(subject auth.Subject, circle ID) iter.Seq2[group.Group, error]

type IsMyCircleMember func(subject auth.Subject, circle ID, other user.ID) (bool, error)
type MyCircleRolesAdd func(subject auth.Subject, circleId ID, usrId user.ID, roles ...role.ID) error
type MyCircleRolesRemove func(subject auth.Subject, circleId ID, usrId user.ID, roles ...role.ID) error

type MyCircleGroupsAdd func(subject auth.Subject, circleId ID, usrId user.ID, groups ...group.ID) error
type MyCircleGroupsRemove func(subject auth.Subject, circleId ID, usrId user.ID, groups ...group.ID) error

type MyCircleUserRemove func(subject auth.Subject, circleId ID, usrId user.ID) error
type MyCircleUserUpdateStatus func(subject auth.Subject, circleId ID, usrId user.ID, status user.AccountStatus) error
type MyCircleUserVerified func(subject auth.Subject, circleId ID, usrId user.ID, emailVerified bool) error

// FindAll returns all known circles.
type FindAll func(subject auth.Subject) iter.Seq2[Circle, error]

type DeleteByID func(subject auth.Subject, id ID) error

type FindByID func(subject auth.Subject, id ID) (std.Option[Circle], error)

type Update func(subject auth.Subject, circle Circle) error

type Create func(subject auth.Subject, circle Circle) (ID, error)

type UseCases struct {
	MyCircleMembers          MyCircleMembers
	MyCircles                MyCircles
	MyCircleRolesAdd         MyCircleRolesAdd
	MyCircleRolesRemove      MyCircleRolesRemove
	MyCircleGroupsAdd        MyCircleGroupsAdd
	MyCircleGroupsRemove     MyCircleGroupsRemove
	MyCircleUserRemove       MyCircleUserRemove
	MyCircleUserUpdateStatus MyCircleUserUpdateStatus
	MyCircleUserVerified     MyCircleUserVerified
	MyRoles                  MyRoles
	MyGroups                 MyGroups
	IsMyCircleMember         IsMyCircleMember
	FindAll                  FindAll
	DeleteByID               DeleteByID
	FindByID                 FindByID
	Update                   Update
	Create                   Create
	IsCircleAdmin            IsCircleAdmin
}

func NewUseCases(
	repoCircle Repository,
	users user.UseCases,
	findGroupByID group.FindByID,
	findRoleByID role.FindByID,
) UseCases {
	var mutex sync.Mutex

	return UseCases{
		MyCircles:                NewMyCircles(repoCircle),
		MyCircleMembers:          NewMyCircleMembers(repoCircle, users.FindAll),
		Create:                   NewCreate(&mutex, repoCircle),
		Update:                   NewUpdate(&mutex, repoCircle),
		DeleteByID:               NewDeleteByID(&mutex, repoCircle),
		FindByID:                 NewFindByID(repoCircle),
		FindAll:                  NewFindAll(repoCircle),
		IsCircleAdmin:            NewIsCircleAdmin(repoCircle),
		IsMyCircleMember:         NewIsMyCircleMember(repoCircle, users.FindByID),
		MyCircleGroupsAdd:        NewMyCircleGroupsAdd(&mutex, repoCircle, users),
		MyCircleGroupsRemove:     NewMyCircleGroupsRemove(&mutex, repoCircle, users),
		MyCircleRolesRemove:      NewMyCircleRolesRemove(&mutex, repoCircle, users),
		MyCircleRolesAdd:         NewMyCircleRolesAdd(&mutex, repoCircle, users),
		MyGroups:                 NewMyGroups(repoCircle, users, findGroupByID),
		MyRoles:                  NewMyRoles(repoCircle, users, findRoleByID),
		MyCircleUserRemove:       NewMyCircleUserRemove(&mutex, repoCircle, users),
		MyCircleUserUpdateStatus: NewMyCircleUserUpdateStatus(&mutex, repoCircle, users),
		MyCircleUserVerified:     NewMyCircleUserVerified(&mutex, repoCircle, users),
	}
}

func myCircle(repo Repository, subject auth.Subject, id ID) (Circle, error) {
	if !subject.Valid() {
		return Circle{}, user.InvalidSubjectErr
	}
	optCircle, err := repo.FindByID(id)
	if err != nil {
		return Circle{}, err
	}

	if optCircle.IsNone() {
		return Circle{}, os.ErrNotExist
	}

	circle := optCircle.Unwrap()
	if !slices.Contains(circle.Administrators, subject.ID()) {
		return Circle{}, user.PermissionDeniedErr
	}

	return circle, nil
}

func myCircleAndUser(repo Repository, findUserByID user.FindByID, subject auth.Subject, cid ID, other user.ID) (Circle, user.User, error) {
	if !subject.Valid() {
		return Circle{}, user.User{}, user.InvalidSubjectErr
	}
	circle, err := myCircle(repo, subject, cid)
	if err != nil {
		return Circle{}, user.User{}, err
	}

	optUsr, err := findUserByID(user.SU(), other)
	if err != nil {
		return Circle{}, user.User{}, err
	}

	if optUsr.IsNone() {
		return Circle{}, user.User{}, os.ErrNotExist
	}

	if !circle.isMember(optUsr.Unwrap()) {
		return Circle{}, user.User{}, user.PermissionDeniedErr
	}

	return circle, optUsr.Unwrap(), nil
}
