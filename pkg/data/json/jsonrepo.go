package json

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/std"
	"io"
	"iter"
	"slices"
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
	err := r.store.View(func(tx blob.Tx) error {
		entry, err := tx.Get(data.Idtos(id))
		if err != nil {
			return fmt.Errorf("cannot retrieve data from store: %w", err)
		}

		if !entry.Valid {
			return nil // this means that we just return a none option => non-existing entry
		}

		domainModel, err := r.decode(entry.Unwrap())
		if err != nil {
			return err
		}

		res = std.Some(domainModel)

		return nil
	})

	return res, err
}

func (r *Repository[DomainModel, DomainID, PersistenceModel, PersistenceID]) decode(entry blob.Entry) (DomainModel, error) {
	var persistenceModel PersistenceModel
	var zeroDomain DomainModel

	reader, err := entry.Open()
	if err != nil {
		return zeroDomain, err
	}

	defer reader.Close() // don't care about read-closer errors

	dec := json.NewDecoder(reader)
	if err := dec.Decode(&persistenceModel); err != nil {
		return zeroDomain, err
	}

	return r.intoDomain(persistenceModel)
}

func (r *Repository[DomainModel, DomainID, PersistenceModel, PersistenceID]) encode(domainModel DomainModel) (blob.Entry, error) {
	persistenceModel, err := r.intoPersistence(domainModel)
	if err != nil {
		return blob.Entry{}, err
	}

	buf, err := json.Marshal(persistenceModel)
	if err != nil {
		return blob.Entry{}, err
	}

	return blob.Entry{
		Key: data.Idtos(domainModel.Identity()),
		Open: func() (io.ReadCloser, error) {
			return readerCloser{Reader: bytes.NewReader(buf)}, nil
		},
	}, nil

}

func (r *Repository[DomainModel, DomainID, PersistenceModel, PersistenceID]) FindAllByID(ids iter.Seq[DomainID]) iter.Seq2[DomainModel, error] {

	return func(yield func(DomainModel, error) bool) {
		var zeroDomain DomainModel
		idents := slices.Collect(ids)
		stopIter := false
		err := r.store.View(func(tx blob.Tx) error {
			for _, ident := range idents {
				optEnt, err := tx.Get(data.Idtos(ident))
				if err != nil {
					// continue iteration, perhaps a little more robust
					if !yield(zeroDomain, err) {
						stopIter = true
						return err
					}
				}

				if !optEnt.Valid {
					continue // just a not-found case
				}

				domainModel, err := r.decode(optEnt.Unwrap())
				if err != nil {
					// continue iteration, perhaps a little more robust e.g. due to JSON unmarshal incompatibility
					if !yield(zeroDomain, err) {
						stopIter = true
						return err
					}

					continue
				}

				if !yield(domainModel, nil) {
					stopIter = true
					return nil
				}
			}

			return nil
		})

		// errors after iteration shall stop, are suppressed. That seems the proposal design
		if !stopIter && err != nil {
			yield(zeroDomain, err)
		}
	}
}

func (r *Repository[DomainModel, DomainID, PersistenceModel, PersistenceID]) Each(yield func(DomainModel, error) bool) {
	var zeroDomain DomainModel
	stopIter := false
	err := r.store.View(func(tx blob.Tx) error {
		tx.Each(func(entry blob.Entry, err error) bool {
			if err != nil {
				// continue iteration, perhaps a little more robust e.g. due to JSON unmarshal incompatibility
				if !yield(zeroDomain, err) {
					stopIter = true
					return false
				}
				return true
			}

			if !yield(r.decode(entry)) {
				stopIter = true
				return false
			}

			return true
		})
		return nil
	})

	// errors after iteration shall stop, are suppressed. That seems the proposal design
	if !stopIter && err != nil {
		yield(zeroDomain, err)
	}
}

func (r *Repository[DomainModel, DomainID, PersistenceModel, PersistenceID]) Count() (int, error) {
	var count int
	var firstErr error
	err := r.store.View(func(tx blob.Tx) error {
		tx.Each(func(entry blob.Entry, err error) bool {
			if err != nil && firstErr == nil {
				firstErr = err
			} else if err == nil {
				count++
			}

			return true
		})

		return nil
	})

	if firstErr == nil && err != nil {
		firstErr = err
	}

	return count, firstErr
}

func (r *Repository[DomainModel, DomainID, PersistenceModel, PersistenceID]) DeleteByID(id DomainID) error {
	return r.store.Update(func(tx blob.Tx) error {
		return tx.Delete(data.Idtos(id))
	})
}

func (r *Repository[DomainModel, DomainID, PersistenceModel, PersistenceID]) DeleteByEntity(e DomainModel) error {
	return r.DeleteByID(e.Identity())
}

func (r *Repository[DomainModel, DomainID, PersistenceModel, PersistenceID]) DeleteAll() error {
	return r.store.Update(func(tx blob.Tx) error {
		var ids []string
		var firstErr error
		tx.Each(func(entry blob.Entry, err error) bool {
			if err != nil {
				firstErr = err
				return false
			}

			ids = append(ids, entry.Key)
			return true
		})

		if firstErr == nil {
			for _, id := range ids {
				if err := tx.Delete(id); err != nil {
					return err
				}
			}
		}

		return firstErr
	})
}

func (r *Repository[DomainModel, DomainID, PersistenceModel, PersistenceID]) DeleteAllByID(ids iter.Seq[DomainID]) error {
	return r.store.Update(func(tx blob.Tx) error {
		var err error
		ids(func(id DomainID) bool {
			err = tx.Delete(data.Idtos(id))
			if err != nil {
				return false
			}

			return true
		})
		return err
	})

}

func (r *Repository[DomainModel, DomainID, PersistenceModel, PersistenceID]) Delete(predicate func(DomainModel) (bool, error)) error {
	var firstErr error
	var deleteList []DomainID
	// we avoid intentionally the nesting of transactions
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
			deleteList = append(deleteList, d.Identity())
		}

		return true
	})

	if firstErr != nil {
		return firstErr
	}

	return r.DeleteAllByID(slices.Values(deleteList))
}

func (r *Repository[DomainModel, DomainID, PersistenceModel, PersistenceID]) Save(e DomainModel) error {
	return r.store.Update(func(tx blob.Tx) error {
		entry, err := r.encode(e)
		if err != nil {
			return err
		}

		return tx.Put(entry)
	})
}

func (r *Repository[DomainModel, DomainID, PersistenceModel, PersistenceID]) SaveAll(it iter.Seq[DomainModel]) error {
	return r.store.Update(func(tx blob.Tx) error {
		var firstError error
		it(func(model DomainModel) bool {
			entry, err := r.encode(model)
			if err != nil {
				firstError = err
				return false
			}

			if err := tx.Put(entry); err != nil {
				firstError = err
				return false
			}

			return true
		})

		return firstError
	})
}
