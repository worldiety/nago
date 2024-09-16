package json

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/std"
	"io"
	"iter"
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

func (r *Repository[DomainModel, DomainID, PersistenceModel, PersistenceID]) FindByID(id DomainID) (std.Option[DomainModel], error) {
	var res std.Option[DomainModel]
	// we use the reader directly here, because it allows potential optimizations. Using blob.Get requires at least another additional full slice allocation
	optR, err := r.store.NewReader(context.Background(), data.Idtos(id))

	if err != nil {
		return res, err
	}

	if optR.IsNone() {
		return res, nil
	}

	var domainModel DomainModel
	reader := optR.Unwrap()
	defer reader.Close() // otherwise we get a deadlock

	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&domainModel); err != nil {
		return res, err
	}

	return std.Some(domainModel), nil

}

func (r *Repository[DomainModel, DomainID, PersistenceModel, PersistenceID]) decode(reader io.Reader) (DomainModel, error) {
	var persistenceModel PersistenceModel
	var zeroDomain DomainModel

	dec := json.NewDecoder(reader)
	err := dec.Decode(&persistenceModel)
	if err != nil {
		return zeroDomain, err
	}

	return r.intoDomain(persistenceModel)
}

func (r *Repository[DomainModel, DomainID, PersistenceModel, PersistenceID]) encode(domainModel DomainModel) (string, []byte, error) {
	persistenceModel, err := r.intoPersistence(domainModel)
	if err != nil {
		return "", nil, err
	}

	buf, err := json.Marshal(persistenceModel)
	if err != nil {
		return "", nil, err
	}

	return data.Idtos(domainModel.Identity()), buf, nil
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

func (r *Repository[DomainModel, DomainID, PersistenceModel, PersistenceID]) Each(yield func(DomainModel, error) bool) {
	var zeroDomain DomainModel
	for id, err := range r.store.List(context.Background(), blob.ListOptions{}) {
		if err != nil {
			if !yield(zeroDomain, err) {
				return
			}
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

func (r *Repository[DomainModel, DomainID, PersistenceModel, PersistenceID]) Count() (int, error) {
	count := 0
	for _, err := range r.store.List(context.Background(), blob.ListOptions{}) {
		if err != nil {
			return count, err
		}
		count++
	}

	return count, nil
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
	r.Each(func(d DomainModel, err error) bool {
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
