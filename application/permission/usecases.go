package permission

import "iter"

type FindAll func(subject Auditable) iter.Seq2[Permission, error]

type UseCases struct {
	FindAll FindAll
}

func NewUseCases() UseCases {
	return UseCases{
		FindAll: NewFindAll(),
	}
}
