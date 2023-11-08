package kv

import (
	"cmp"
	"encoding/json"
	"fmt"
	"go.wdy.de/nago/container/enum"
	"go.wdy.de/nago/container/serrors"
	"go.wdy.de/nago/container/slice"
	"go.wdy.de/nago/domain"
	"slices"
)

// Store is a contract for a generic transaction-based key-value store. See [Collection] for an easy-to-use abstraction.
type Store interface {
	Update(f func(Tx) error) error // executes synchronously a writeable transaction closure
	View(f func(Tx) error) error   // executes synchronously a read-only transaction closure
}

// Tx is a transaction contract.
type Tx interface {
	Each(f func(name []byte, c Bucket) error) error // loops over all available buckets.
	Bucket(name []byte) (Bucket, error)             // returns or creates the according bucket. If read-only and not exists, it is nil.
	DeleteBucket(name []byte) error
	Writable() bool // true, if this transaction can be written
}

type Bucket interface {
	Each(f func(key, value []byte) error) error // loops over a snapshot of all available entries, key and value must not escape.
	Delete(key []byte) error
	Put(key, value []byte) error
	Get(key []byte) ([]byte, error)
}

// A Collection is an abstraction layer over the key value store and assumes, that a bucket contains only elements
// of the same type. Keys are strings and values are json serialized.
type Collection[E dm.Entity[ID], ID cmp.Ordered] struct {
	db   Store
	name []byte
}

func NewCollection[E dm.Entity[ID], ID cmp.Ordered](db Store, name string) Collection[E, ID] {
	return Collection[E, ID]{
		db:   db,
		name: []byte(name),
	}
}

// IntoSlice loads the entire set of all key/values atomically into memory and sorts it by identifier.
// See also Find.
func (c Collection[E, ID]) IntoSlice() (slice.Slice[E], serrors.InfrastructureError) {
	return c.Filter(func(e E) bool {
		return true
	})
}

// Filter collects all entities for which the given predicate returns true. The result ist sorted by identifier.
// Even though it unmarshalls each entity, only the collected entities are held in memory.
func (c Collection[E, ID]) Filter(p func(E) bool) (slice.Slice[E], serrors.InfrastructureError) {
	var res []E
	err := c.db.View(func(tx Tx) error {
		bucket, err := tx.Bucket(c.name)
		if err != nil {
			return err
		}

		return bucket.Each(func(key, value []byte) error {
			var t E
			if err := json.Unmarshal(value, &t); err != nil {
				return err
			}

			if p(t) {
				res = append(res, t)
			}
			return nil
		})
	})

	if err != nil {
		return slice.Of[E](), serrors.IntoInfrastructure(err)
	}

	slices.SortFunc(res, func(a, b E) int {
		return cmp.Compare(a.Identity(), b.Identity())
	})

	return slice.Of(res...), nil
}

// Delete removes the Entity within a distinct transaction. It is not an error if neither the collection exists nor
// the entity itself.
func (c Collection[E, ID]) Delete(id ID) serrors.InfrastructureError {
	return serrors.IntoInfrastructure(c.db.Update(func(tx Tx) error {
		bucket, err := tx.Bucket(c.name)
		if err != nil {
			return err
		}

		if bucket == nil {
			return nil
		}

		return bucket.Delete([]byte(fmt.Sprintf("%v", id)))
	}))
}

// DeleteAll removes all those entities in a single transaction for which the predicate returns true.
// See also Find.
func (c Collection[E, ID]) DeleteAll(f func(E) bool) serrors.InfrastructureError {
	return serrors.IntoInfrastructure(c.db.Update(func(tx Tx) error {
		bucket, err := tx.Bucket(c.name)
		if err != nil {
			return err
		}

		if bucket == nil {
			return nil
		}

		return bucket.Each(func(key, value []byte) error {
			var t E
			if err := json.Unmarshal(value, &t); err != nil {
				return err
			}

			if f(t) {
				if err := bucket.Delete([]byte(fmt.Sprintf("%v", t.Identity()))); err != nil {
					return err
				}
			}

			return nil
		})

	}))
}

// Save creates or updates the entity by marshalling into JSON.
// It is a programming error causing a panic, if types are used which cannot be unmarshalled.
func (c Collection[E, ID]) Save(entities ...E) serrors.InfrastructureError {
	err := c.db.Update(func(tx Tx) error {
		for _, entity := range entities {
			buf, err := json.Marshal(entity)
			if err != nil {
				panic(fmt.Errorf("cannot marshal into type %T: %w", entity, err)) // this is very likely an unfixable programming error
			}

			bucket, err := tx.Bucket(c.name)
			if err != nil {
				return err
			}

			if err := bucket.Put([]byte(dm.IdentString(entity.Identity())), buf); err != nil {
				return err
			}
		}

		return nil
	})

	return serrors.IntoInfrastructure(err)
}

// Find returns the given entity or an error with a lookup failure. Unmarshalls from JSON.
// It is a programming error causing a panic, if types are used which cannot be unmarshalled.
func (c Collection[E, ID]) Find(id ID) (E, enum.Error[dm.LookupFailure]) {
	var value E
	var errLookup enum.Error[dm.LookupFailure]
	err := c.db.View(func(tx Tx) error {
		bucket, err := tx.Bucket(c.name)
		if err != nil {
			return err
		}

		buf, err := bucket.Get([]byte(dm.IdentString(id)))
		if err != nil {
			return err
		}

		if bucket == nil || buf == nil {
			errLookup = enum.IntoErr(dm.LookupFailure{}.With1(dm.EntityNotFound(dm.IdentString(id))))
			return errLookup
		}

		if err := json.Unmarshal(buf, &value); err != nil {
			panic(fmt.Errorf("cannot unmarshal into type %T: %w", value, err)) // this is very likely an unfixable programming error
		}

		return nil
	})

	if errLookup != nil {
		return value, errLookup
	}

	if err != nil {
		return value, enum.IntoErr(dm.LookupFailure{}.With2(serrors.IntoInfrastructure(err)))
	}

	return value, nil
}
