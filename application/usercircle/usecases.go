package usercircle

import (
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/std"
	"iter"
	"sync"
)

// MyCircleMembers returns all those users defined in the nago IAM which are included by the circle member rules.
type MyCircleMembers func(uid user.ID, id ID) iter.Seq2[user.User, error]

type IsCircleAdmin func(uid user.ID) (bool, error)

// MyCircles returns all circles, in which the subject is declared as an administrator.
type MyCircles func(uid user.ID) iter.Seq2[Circle, error]

// FindAll returns all known circles.
type FindAll func(subject auth.Subject) iter.Seq2[Circle, error]

type DeleteByID func(subject auth.Subject, id ID) error

type FindByID func(subject auth.Subject, id ID) (std.Option[Circle], error)

type Update func(subject auth.Subject, circle Circle) error

type Create func(subject auth.Subject, circle Circle) (ID, error)
type UseCases struct {
	MyCircleMembers MyCircleMembers
	MyCircles       MyCircles
	FindAll         FindAll
	DeleteByID      DeleteByID
	FindByID        FindByID
	Update          Update
	Create          Create
	IsCircleAdmin   IsCircleAdmin
}

func NewUseCases(repoCircle Repository, findAllUsers user.FindAll) UseCases {
	var mutex sync.Mutex

	return UseCases{
		MyCircles:       NewMyCircles(repoCircle),
		MyCircleMembers: NewMyCircleMembers(repoCircle, findAllUsers),
		Create:          NewCreate(&mutex, repoCircle),
		Update:          NewUpdate(&mutex, repoCircle),
		DeleteByID:      NewDeleteByID(&mutex, repoCircle),
		FindByID:        NewFindByID(repoCircle),
		FindAll:         NewFindAll(repoCircle),
		IsCircleAdmin:   NewIsCircleAdmin(repoCircle),
	}
}
