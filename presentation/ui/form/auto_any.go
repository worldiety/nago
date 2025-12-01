// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package form

import (
	"fmt"
	"iter"
	"reflect"

	"go.wdy.de/nago/application/ent"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/xiter"
)

// AnyEntity is a wrapper (Entity).
// It provides a type-erased representation of an aggregate entity
// with a string identity and a stored aggregate of any type.
type AnyEntity struct {
	id        string               // string identity of the entity
	aggregate any                  // wrapped aggregate instance
	setId     *func(id string) any // we need a pointer to a func (which is usually already a pointer?) so that deep equal works transparently
}

// Equals checks whether this entity is equal to another.
// Equality is based on matching id and deep equality of the aggregate.
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

// UnwrapEntity returns the underlying aggregate.
func (a AnyEntity) UnwrapEntity() any {
	return a.aggregate
}

// Identity returns the string identity of the entity.
func (a AnyEntity) Identity() string {
	return a.id
}

// WithIdentity returns a copy of the entity with the given identity.
func (a AnyEntity) WithIdentity(id string) AnyEntity {
	fn := *(a.setId)
	a.aggregate = fn(id)
	a.id = id
	return a
}

// String returns a string representation of the wrapped aggregate.
func (a AnyEntity) String() string {
	return fmt.Sprintf("%v", a.aggregate)
}

// AnyEntityOf wraps a strongly-typed aggregate into an AnyEntity.
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

// AnySeq2Of maps a sequence of typed aggregates into a sequence of AnyEntity.
func AnySeq2Of[E Aggregate[E, ID], ID ~string](it iter.Seq2[E, error]) iter.Seq2[AnyEntity, error] {
	return xiter.Map2(func(in E, err error) (AnyEntity, error) {
		return AnyEntityOf(in), err
	}, it)
}

// UseCaseListAny is a function type returning a sequence of AnyEntity for a subject.
type UseCaseListAny = ent.FindAll[AnyEntity, string]

// AnyUseCaseList converts a typed use case list function into a UseCaseListAny
// that produces type-erased AnyEntity results.
func AnyUseCaseList[E Aggregate[E, ID], ID ~string](list func(subject auth.Subject) iter.Seq2[E, error]) UseCaseListAny {
	return func(subject auth.Subject) iter.Seq2[AnyEntity, error] {
		return AnySeq2Of(list(subject))
	}
}
