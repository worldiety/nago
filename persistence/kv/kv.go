package kv

import (
	"cmp"
	"encoding/json"
	"fmt"
	"go.wdy.de/nago/container/enum"
	"go.wdy.de/nago/container/slice"
	"go.wdy.de/nago/persistence"
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
	Writeable() bool // true, if this transaction can be written
}

type Bucket interface {
	Each(f func(key, value []byte) error) error // loops over a snapshot of all available entries, key and value must not escape.
	Delete(key []byte) error
	Put(key, value []byte) error
	Get(key []byte) ([]byte, error)
}

// A Collection is an abstraction layer over the key value store and assumes, that a bucket contains only elements
// of the same type. Keys are strings and values are json serialized.
type Collection[ID cmp.Ordered, E persistence.Entity[ID]] struct {
	db   Store
	name []byte
}

func NewCollection[E persistence.Entity[ID], ID cmp.Ordered](db Store, name string) Collection[ID, E] {
	return Collection[ID, E]{
		db:   db,
		name: []byte(name),
	}
}

// IntoSlice loads the entire set of all key/values atomically into memory and sorts it by identifier.
// See also Find.
func (c Collection[ID, E]) IntoSlice() (slice.Slice[E], error) {
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
			res = append(res, t)
			return nil
		})
	})

	if err != nil {
		return slice.Of[E](), err
	}

	slices.SortFunc(res, func(a, b E) int {
		return cmp.Compare(a.Identity(), b.Identity())
	})

	return slice.Of(res...), nil
}

// Delete removes the Entity within a distinct transaction. It is not an error if neither the collection exists nor
// the entity itself.
func (c Collection[ID, E]) Delete(id ID) error {
	return c.db.Update(func(tx Tx) error {
		bucket, err := tx.Bucket(c.name)
		if err != nil {
			return err
		}

		if bucket == nil {
			return nil
		}

		return bucket.Delete([]byte(fmt.Sprintf("%v", id)))
	})
}

// DeleteAll removes all those entities in a single transaction for which the predicate returns true.
// See also Find.
func (c Collection[ID, E]) DeleteAll(f func(E) bool) error {
	return c.db.Update(func(tx Tx) error {
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

	})
}

// Save creates or updates the entity by marshalling into JSON.
// It is a programming error causing a panic, if types are used which cannot be unmarshalled.
func (c Collection[ID, E]) Save(entity E) persistence.InfrastructureError {
	err := c.db.Update(func(tx Tx) error {
		buf, err := json.Marshal(entity)
		if err != nil {
			panic(fmt.Errorf("cannot marshal into type %T: %w", entity, err)) // this is very likely an unfixable programming error
		}

		bucket, err := tx.Bucket(c.name)
		if err != nil {
			return err
		}

		return bucket.Put([]byte(persistence.IdentString(entity.Identity())), buf)
	})

	return err
}

// Find returns the given entity or an error with a lookup failure. Unmarshalls from JSON.
// It is a programming error causing a panic, if types are used which cannot be unmarshalled.
func (c Collection[ID, E]) Find(id ID) (E, enum.Error[persistence.LookupFailure]) {
	var value E
	var errLookup enum.Error[persistence.LookupFailure]
	err := c.db.View(func(tx Tx) error {
		bucket, err := tx.Bucket(c.name)
		if err != nil {
			return err
		}

		buf, err := bucket.Get([]byte(persistence.IdentString(id)))
		if err != nil {
			return err
		}

		if bucket == nil || buf == nil {
			errLookup = enum.IntoErr(persistence.LookupFailure{}.With1(persistence.EntityNotFound(persistence.IdentString(id))))
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
		return value, enum.IntoErr(persistence.LookupFailure{}.With2(err))
	}

	return value, nil
}
