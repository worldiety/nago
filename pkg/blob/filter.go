package blob

import (
	"fmt"
	"go.wdy.de/nago/pkg/std"
)

type filterStore struct {
	predicate func(Entry) bool
	store     Store
}

type filterViewTx struct {
	predicate func(Entry) bool
	tx        Tx
}

func (f filterViewTx) Each(yield func(Entry, error) bool) {
	f.tx.Each(func(entry Entry, err error) bool {
		if f.predicate(entry) {
			return yield(entry, err)
		}

		return true
	})
}

func (f filterViewTx) Delete(key string) error {
	return fmt.Errorf("a filtered store is readonly")
}

func (f filterViewTx) Put(entry Entry) error {
	return fmt.Errorf("a filtered store is readonly")
}

func (f filterViewTx) Get(key string) (std.Option[Entry], error) {
	optEnt, err := f.tx.Get(key)
	if err != nil {
		return optEnt, err
	}

	if !optEnt.Valid {
		return optEnt, nil
	}

	ent := optEnt.Unwrap()
	if f.predicate(ent) {
		return optEnt, nil
	}

	return std.None[Entry](), nil
}

func (s filterStore) Update(f func(Tx) error) error {
	return fmt.Errorf("a filtered store is readonly")
}

func (s filterStore) View(f func(Tx) error) error {
	return s.store.View(func(tx Tx) error {
		return f(filterViewTx{
			predicate: s.predicate,
			tx:        tx,
		})
	})
}

// Filter provides only a filtered and readonly subset from the given store.
// Only those entries are available, whose entries are accepted by the given predicate.
func Filter(store Store, p func(Entry) bool) Store {
	return filterStore{
		predicate: p,
		store:     store,
	}
}
