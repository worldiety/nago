// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package json

import (
	"context"
	"errors"
	"fmt"
	"io"
	"iter"
	"log/slog"

	"github.com/worldiety/enum/json"
	"github.com/worldiety/option"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/std"
)

type Repository[DomainModel data.Aggregate[DomainID], DomainID data.IDType, PersistenceModel data.Aggregate[PersistenceID], PersistenceID data.IDType] struct {
	store           blob.Store
	intoDomain      data.MapF[PersistenceModel, DomainModel]
	intoPersistence data.MapF[DomainModel, PersistenceModel]
}

func NewJSONRepository[DomainModel data.Aggregate[DomainID], DomainID data.IDType, PersistenceModel data.Aggregate[PersistenceID], PersistenceID data.IDType](store blob.Store, intoDomain data.MapF[PersistenceModel, DomainModel], intoPersistence data.MapF[DomainModel, PersistenceModel]) *Repository[DomainModel, DomainID, PersistenceModel, PersistenceID] {
	return &Repository[DomainModel, DomainID, PersistenceModel, PersistenceID]{
		store:           store,
		intoDomain:      intoDomain,
		intoPersistence: intoPersistence,
	}
}

// NewSloppyJSONRepository is a shorthand factory to create a sloppy repository which just serializes the domain model.
// This comes handy for fast iterating cycles of throw-away prototypes, especially if neither domain nor persistence model has been explored carefully and stabilized.
func NewSloppyJSONRepository[DomainModel data.Aggregate[DomainID], DomainID data.IDType](store blob.Store) *Repository[DomainModel, DomainID, DomainModel, DomainID] {
	return NewJSONRepository[DomainModel, DomainID, DomainModel, DomainID](
		store,
		func(model DomainModel) (DomainModel, error) {
			return model, nil
		},
		func(model DomainModel) (DomainModel, error) {
			return model, nil
		},
	)
}

func (r *Repository[DomainModel, DomainID, PersistenceModel, PersistenceID]) FindByID(id DomainID) (option.Opt[DomainModel], error) {
	var res std.Option[DomainModel]
	// we use the reader directly here, because it allows potential optimizations. Using blob.Get requires at least another additional full slice allocation
	optR, err := r.store.NewReader(context.Background(), data.Idtos(id))

	if err != nil {
		return res, err
	}

	if optR.IsNone() {
		return res, nil
	}

	reader := optR.Unwrap()
	defer reader.Close() // otherwise we get a deadlock

	domainModel, err := r.decode(data.Idtos(id), reader)
	if err != nil {
		return res, err
	}

	// TODO this is expensive, thus should be guarded to a debug build, because that normally does not happen
	if id != domainModel.Identity() {
		slog.Error("json repo found model identifier mismatch, this may happen due to a broken WithIdentity function", "id", id, "model", domainModel.Identity(), "type", fmt.Sprintf("%T", domainModel))
		if withId, ok := any(domainModel).(interface{ WithIdentity(DomainID) DomainModel }); ok {
			domainModel = withId.WithIdentity(id)
			slog.Error("fixed model", "id", id)
		}
	}

	return std.Some(domainModel), nil

}

// DecodeError is returned by the reading methods of a [Repository], when a stored value has been read but cannot be
// decoded or mapped into the domain model, unlike an error of reading it. Its message is the one of Err.
type DecodeError struct {
	Key string
	Err error
}

func (e *DecodeError) Error() string {
	return e.Err.Error()
}

func (e *DecodeError) Unwrap() error {
	return e.Err
}

func (r *Repository[DomainModel, DomainID, PersistenceModel, PersistenceID]) decode(key string, reader io.Reader) (DomainModel, error) {
	var persistenceModel PersistenceModel
	var zeroDomain DomainModel

	// the decoder returns an error of the reader unchanged, so it is told apart here
	src := &readErrors{r: reader}
	dec := json.NewDecoder(src)
	err := dec.Decode(&persistenceModel)
	if err != nil {
		if src.err != nil {
			return zeroDomain, src.err
		}

		return zeroDomain, &DecodeError{Key: key, Err: err}
	}

	model, err := r.intoDomain(persistenceModel)
	if err != nil {
		return zeroDomain, &DecodeError{Key: key, Err: err}
	}

	return model, nil
}

// readErrors remembers the first error of reading, but the end of the input.
type readErrors struct {
	r   io.Reader
	err error
}

func (e *readErrors) Read(p []byte) (int, error) {
	n, err := e.r.Read(p)
	if err != nil && !errors.Is(err, io.EOF) && e.err == nil {
		e.err = err
	}

	return n, err
}

func (r *Repository[DomainModel, DomainID, PersistenceModel, PersistenceID]) encode(domainModel DomainModel) (string, []byte, error) {
	persistenceModel, err := r.intoPersistence(domainModel)
	if err != nil {
		return "", nil, err
	}

	// important: use a pointer here, otherwise we will miss the raw-interface type, if the persistence model is polymorphic
	buf, err := json.Marshal(&persistenceModel)
	if err != nil {
		return "", nil, err
	}

	return data.Idtos(domainModel.Identity()), buf, nil
}

func (r *Repository[DomainModel, DomainID, PersistenceModel, PersistenceID]) FindAllByPrefix(prefix DomainID) iter.Seq2[DomainModel, error] {
	return func(yield func(DomainModel, error) bool) {
		var zeroDomain DomainModel
		for id, err := range r.store.List(context.Background(), blob.ListOptions{
			Prefix: data.Idtos(prefix),
		}) {
			if err != nil {
				if !yield(zeroDomain, err) {
					return
				}

				continue
			}

			did, err := data.Stoid[DomainID](id)
			if err != nil {
				if !yield(zeroDomain, err) {
					return
				}
			}

			optModel, err := r.FindByID(did)
			if err != nil {
				if !yield(zeroDomain, err) {
					return
				}
				continue
			}

			if !yield(optModel.Unwrap(), nil) {
				return
			}
		}
	}
}

func (r *Repository[DomainModel, DomainID, PersistenceModel, PersistenceID]) Identifiers() iter.Seq2[DomainID, error] {
	return func(yield func(DomainID, error) bool) {
		var zeroID DomainID
		for id, err := range r.store.List(context.Background(), blob.ListOptions{}) {
			if err != nil {
				if !yield(zeroID, err) {
					return
				}
			} else {
				ci, err := data.Stoid[DomainID](id)
				if !yield(ci, err) {
					return
				}
			}
		}
	}
}

func (r *Repository[DomainModel, DomainID, PersistenceModel, PersistenceID]) IdentifiersByPrefix(prefix DomainID) iter.Seq2[DomainID, error] {
	return func(yield func(DomainID, error) bool) {
		var zeroID DomainID
		for id, err := range r.store.List(context.Background(), blob.ListOptions{
			Prefix: data.Idtos(prefix),
		}) {
			if err != nil {
				if !yield(zeroID, err) {
					return
				}
			} else {
				ci, err := data.Stoid[DomainID](id)
				if !yield(ci, err) {
					return
				}
			}
		}
	}
}

func (r *Repository[DomainModel, DomainID, PersistenceModel, PersistenceID]) FindAllByID(ids iter.Seq[DomainID]) iter.Seq2[DomainModel, error] {
	return func(yield func(DomainModel, error) bool) {
		var zeroDomain DomainModel
		for id := range ids {
			optModel, err := r.FindByID(id)
			if err != nil {
				if !yield(zeroDomain, err) {
					return
				}
				continue
			}

			if optModel.IsNone() {
				continue
			}

			if !yield(optModel.Unwrap(), nil) {
				return
			}

		}
	}

}

func (r *Repository[DomainModel, DomainID, PersistenceModel, PersistenceID]) All() iter.Seq2[DomainModel, error] {
	return func(yield func(DomainModel, error) bool) {
		var zeroDomain DomainModel
		for id, err := range r.store.List(context.Background(), blob.ListOptions{}) {
			if err != nil {
				if !yield(zeroDomain, err) {
					return
				}
				continue
			}

			did, err := data.Stoid[DomainID](id)
			if err != nil {
				if !yield(zeroDomain, err) {
					return
				}
			}

			optModel, err := r.FindByID(did)
			if err != nil {
				if !yield(zeroDomain, err) {
					return
				}
			}

			if optModel.IsNone() {
				// was deleted in the meantime
				continue
			}

			if !yield(optModel.Unwrap(), nil) {
				return
			}

		}
	}
}

func (r *Repository[DomainModel, DomainID, PersistenceModel, PersistenceID]) Count() (int, error) {
	c, err := blob.Count(context.Background(), r.store)
	return int(c), err
}

func (r *Repository[DomainModel, DomainID, PersistenceModel, PersistenceID]) DeleteByID(id DomainID) error {
	return r.store.Delete(context.Background(), data.Idtos(id))
}

func (r *Repository[DomainModel, DomainID, PersistenceModel, PersistenceID]) DeleteByEntity(e DomainModel) error {
	return r.DeleteByID(e.Identity())
}

func (r *Repository[DomainModel, DomainID, PersistenceModel, PersistenceID]) DeleteAll() error {
	return blob.DeleteAll(r.store)
}

func (r *Repository[DomainModel, DomainID, PersistenceModel, PersistenceID]) DeleteAllByID(ids iter.Seq[DomainID]) error {
	for id := range ids {
		if err := r.store.Delete(context.Background(), data.Idtos(id)); err != nil {
			return err
		}
	}

	return nil
}

func (r *Repository[DomainModel, DomainID, PersistenceModel, PersistenceID]) Delete(predicate func(DomainModel) (bool, error)) error {
	var firstErr error
	r.All()(func(d DomainModel, err error) bool {
		doDelete, err := predicate(d)
		if err != nil {
			if errors.Is(err, data.SkipAll) {
				return false
			}

			firstErr = err
			return false
		}

		if doDelete {
			if err := r.DeleteByID(d.Identity()); err != nil {
				firstErr = err
				return false
			}
		}

		return true
	})

	if firstErr != nil {
		return firstErr
	}

	return nil
}

func (r *Repository[DomainModel, DomainID, PersistenceModel, PersistenceID]) Save(e DomainModel) error {
	key, buf, err := r.encode(e)
	if err != nil {
		return err
	}

	return blob.Put(r.store, key, buf)
}

func (r *Repository[DomainModel, DomainID, PersistenceModel, PersistenceID]) SaveAll(it iter.Seq[DomainModel]) error {
	for model := range it {
		if err := r.Save(model); err != nil {
			return fmt.Errorf("cannot save %v: %w", model.Identity(), err)
		}
	}

	return nil
}

func (r *Repository[DomainModel, DomainID, PersistenceModel, PersistenceID]) Name() string {
	return r.store.Name()
}
