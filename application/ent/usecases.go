package ent

import (
	"iter"
	"sync"

	"github.com/worldiety/option"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
)

type Aggregate[A any, ID comparable] interface {
	data.Aggregate[ID]
	WithIdentity(ID) A
}

type Create[T Aggregate[T, ID], ID data.IDType] func(subject auth.Subject, entity T) (ID, error)

type FindByID[T Aggregate[T, ID], ID data.IDType] func(subject auth.Subject, id ID) (option.Opt[T], error)

type FindAllIdentifiers[T Aggregate[T, ID], ID data.IDType] func(subject auth.Subject) iter.Seq2[ID, error]

type FindAll[T Aggregate[T, ID], ID data.IDType] func(subject auth.Subject) iter.Seq2[T, error]

type Update[T Aggregate[T, ID], ID data.IDType] func(subject auth.Subject, entity T) error

type DeleteByID[T Aggregate[T, ID], ID data.IDType] func(subject auth.Subject, id ID) error

// UseCases encapsulates a few default create-read-update-delete use cases using a standard implementation.
// You may start over by just creating a new named type with this as your underlying type, e.g.
//
//	package myaggregate
//	import (
//		  	...
//	)
//
//	var Permissions = crud.DeclarePermissions[Network, ID](<my.prefix>, "My Aggregate")
//
//	type Repository data.Repository[MyAggregate,ID]
//	type UseCases crud.UseCases[MyAggregate,ID]
//
//	func NewUseCases(repo Repository) UseCases {
//	  return UseCases(crud.NewUseCases(Permissions, repo, crud.Options{}))
//	}
//
// See also [cfgcrud.Enable] to bootstrap a default user interface.
type UseCases[T Aggregate[T, ID], ID data.IDType] struct {
	Create             Create[T, ID]
	FindByID           FindByID[T, ID]
	FindAll            FindAll[T, ID]
	FindAllIdentifiers FindAllIdentifiers[T, ID]
	Update             Update[T, ID]
	DeleteByID         DeleteByID[T, ID]
}

type Options struct {
	// Mutex is optional and may be nil. If nil, a new mutex is created automatically to protect critical section
	// under currency pressure. Otherwise, updates and creates may accidentally recreate or overwrite entities.
	Mutex *sync.Mutex
}

func NewUseCases[T Aggregate[T, ID], ID ~string](perms Permissions, repo data.Repository[T, ID], opts Options) UseCases[T, ID] {
	if opts.Mutex == nil {
		opts.Mutex = new(sync.Mutex)
	}

	return UseCases[T, ID]{
		Create:             NewCreate(opts, perms, repo),
		DeleteByID:         NewDeleteByID(opts, perms, repo),
		FindByID:           NewFindByID(opts, perms, repo),
		FindAllIdentifiers: NewFindAllIdentifiers(opts, perms, repo),
		Update:             NewUpdate(opts, perms, repo),
		FindAll:            NewFindAll(opts, perms, repo),
	}
}
