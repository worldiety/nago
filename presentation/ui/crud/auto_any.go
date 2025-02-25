package crud

import (
	"fmt"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/xiter"
	"iter"
	"reflect"
)

type AnyEntity struct {
	id        string
	aggregate any
	setId     *func(id string) any // we need a pointer to a func (which is usually already a pointer?) so that deep equal works transparently
}

func (a AnyEntity) Equals(other any) bool {
	otherAe, ok := other.(AnyEntity)
	if !ok {
		return false
	}

	if a.id != otherAe.id {
		return false
	}

	return reflect.DeepEqual(a.aggregate, otherAe.aggregate)
}

func (a AnyEntity) UnwrapEntity() any {
	return a.aggregate
}

func (a AnyEntity) Identity() string {
	return a.id
}

func (a AnyEntity) WithIdentity(id string) AnyEntity {
	fn := *(a.setId)
	a.aggregate = fn(id)
	a.id = id
	return a
}

func (a AnyEntity) String() string {
	return fmt.Sprintf("%v", a.aggregate)
}

func AnyEntityOf[E Aggregate[E, ID], ID ~string](entity E) AnyEntity {
	fn := func(id string) any {
		return entity.WithIdentity(ID(id))
	}
	return AnyEntity{
		id:        string(entity.Identity()),
		aggregate: entity,
		setId:     &fn,
	}
}

func AnySeq2Of[E Aggregate[E, ID], ID ~string](it iter.Seq2[E, error]) iter.Seq2[AnyEntity, error] {
	return xiter.Map2(func(in E, err error) (AnyEntity, error) {
		return AnyEntityOf(in), err
	}, it)
}

type UseCaseListAny func(subject auth.Subject) iter.Seq2[AnyEntity, error]

func AnyUseCaseList[E Aggregate[E, ID], ID ~string](list func(subject auth.Subject) iter.Seq2[E, error]) UseCaseListAny {
	return func(subject auth.Subject) iter.Seq2[AnyEntity, error] {
		return AnySeq2Of(list(subject))
	}
}
