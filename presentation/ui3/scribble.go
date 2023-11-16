package main

import (
	"cmp"
	"go.wdy.de/nago/container/slice"
	dm "go.wdy.de/nago/domain"
)

type ListView[E dm.Entity[ID], ID cmp.Ordered, Params any] struct {
	Delete func(Params, id slice.Slice[ID]) error
}

type P struct{}

type Page[Params any] struct {
	Body any
}

type PersonID string

type Person struct {
	ID PersonID
}

func (p Person) Identity() PersonID {
	return p.ID
}

type UserParams struct {
	User any
	ID   PersonID
}

func main() {
	_ = Page[UserParams]{
		Body: ListView[Person, PersonID, UserParams]{
			Delete: func(Params, id slice.Slice[PersonID]) error {
				return nil
			},
		},
	}
}
