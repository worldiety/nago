package data

import (
	"go.wdy.de/nago/pkg/events"
	"go.wdy.de/nago/pkg/std"
	"iter"
	"sync"
)

type Deleted[ID IDType] struct {
	ID ID
}

type DeletedObserver[E Aggregate[ID], ID IDType] func(repository Repository[E, ID], deleted Deleted[ID]) error

type Saved[E Aggregate[ID], ID IDType] struct {
	ID    ID
	Value E
}

type SavedObserver[E Aggregate[ID], ID IDType] func(repository Repository[E, ID], saved Saved[E, ID]) error

// NewNotifyRepository decorates the given Repository and emits Deleted and Saved events into the given bus.
// Note, that this is not for free. The Bus (which may be nil) may cause itself a
// slow-down or resource leak, depending on how
// many events are generated or consumed. In general, short bursts or a few hundred events per second will not
// cause any problems.
// Note also, that batch operations are split up, thus optimizations of underlying implementations which may
// get otherwise executed within a single transaction (e.g. DeleteAll or SaveAll) are executed per item,
// to generate more correct events.
//
// See also [NotifyRepository.AddDeletedObserver] and [NotifyRepository.AddSavedObserver] for a blocking and more
// efficient way of getting notified. These observers do not need to pass the bus (which can be nil).
// The current implementation imposes a lock while executing the observer, thus do not add or remove from
// an observer, otherwise you are trapped into a deadlock. Any error returned from an Observer, will cancel the
// current run of calling the observers and the error is returned to the caller. The invocation order of observers
// is not defined.
func NewNotifyRepository[E Aggregate[ID], ID IDType](bus events.Bus, other Repository[E, ID]) NotifyRepository[E, ID] {
	if n, ok := other.(NotifyRepository[E, ID]); ok {
		return n
	}
	
	return &eventRepository[E, ID]{
		bus:   bus,
		other: other,
	}
}

type NotifyRepository[E Aggregate[ID], ID IDType] interface {
	Repository[E, ID]
	// AddDeletedObserver is called after the entity as been removed. The repository method is blocked, until
	// the observer returns.
	AddDeletedObserver(fn DeletedObserver[E, ID]) (close func())

	// AddSavedObserver is called after the entity as been saved. The repository method is blocked, until
	// the observer returns.
	AddSavedObserver(fn SavedObserver[E, ID]) (close func())
}

type eventRepository[E Aggregate[ID], ID IDType] struct {
	observerMutex      sync.Mutex
	nextObserverHandle int
	observersDeleted   map[int]DeletedObserver[E, ID]
	observersSaved     map[int]SavedObserver[E, ID]
	bus                events.Bus
	other              Repository[E, ID]
}

func (e *eventRepository[E, ID]) AddDeletedObserver(fn DeletedObserver[E, ID]) (close func()) {
	e.observerMutex.Lock()
	defer e.observerMutex.Unlock()

	if e.observersDeleted == nil {
		e.observersDeleted = make(map[int]DeletedObserver[E, ID])
	}

	e.nextObserverHandle++
	e.observersDeleted[e.nextObserverHandle] = fn

	return func() {
		e.observerMutex.Lock()
		defer e.observerMutex.Unlock()
		delete(e.observersDeleted, e.nextObserverHandle)
	}
}

func (e *eventRepository[E, ID]) AddSavedObserver(fn SavedObserver[E, ID]) (close func()) {
	e.observerMutex.Lock()
	defer e.observerMutex.Unlock()

	if e.observersSaved == nil {
		e.observersSaved = make(map[int]SavedObserver[E, ID])
	}

	e.nextObserverHandle++
	e.observersSaved[e.nextObserverHandle] = fn
	return func() {
		e.observerMutex.Lock()
		defer e.observerMutex.Unlock()
		delete(e.observersSaved, e.nextObserverHandle)
	}
}

func (e *eventRepository[E, ID]) FindByID(id ID) (std.Option[E], error) {
	return e.other.FindByID(id)
}

func (e *eventRepository[E, ID]) FindAllByPrefix(prefix ID) iter.Seq2[E, error] {
	return e.other.FindAllByPrefix(prefix)
}

func (e *eventRepository[E, ID]) Identifiers() iter.Seq2[ID, error] {
	return e.other.Identifiers()
}

func (e *eventRepository[E, ID]) FindAllByID(ids iter.Seq[ID]) iter.Seq2[E, error] {
	return e.other.FindAllByID(ids)
}

func (e *eventRepository[E, ID]) All() iter.Seq2[E, error] {
	return e.other.All()
}

func (e *eventRepository[E, ID]) Count() (int, error) {
	return e.other.Count()
}

func (e *eventRepository[E, ID]) DeleteByID(id ID) error {
	err := e.other.DeleteByID(id)
	if err != nil {
		return err
	}

	if e.bus != nil {
		e.bus.Publish(Deleted[ID]{ID: id})
	}

	e.observerMutex.Lock()
	defer e.observerMutex.Unlock()

	for _, s := range e.observersDeleted {
		if err := s(e, Deleted[ID]{
			ID: id,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (e *eventRepository[E, ID]) DeleteAll() error {
	for id, err := range e.Identifiers() {
		if err != nil {
			return err
		}

		if err := e.DeleteByID(id); err != nil {
			return err
		}
	}

	return nil
}

func (e *eventRepository[E, ID]) DeleteAllByID(ids iter.Seq[ID]) error {
	for id := range ids {
		if err := e.DeleteByID(id); err != nil {
			return err
		}
	}

	return nil
}

func (e *eventRepository[E, ID]) Delete(predicate func(E) (bool, error)) error {
	for entity, err := range e.All() {
		if err != nil {
			return err
		}

		accept, err := predicate(entity)
		if err != nil {
			return err
		}

		if accept {
			if err := e.DeleteByID(entity.Identity()); err != nil {
				return err
			}
		}
	}

	return nil
}

func (e *eventRepository[E, ID]) DeleteByEntity(e2 E) error {
	return e.DeleteByID(e2.Identity())
}

func (e *eventRepository[E, ID]) Save(e2 E) error {
	if err := e.other.Save(e2); err != nil {
		return err
	}

	if e.bus != nil {
		e.bus.Publish(Saved[E, ID]{
			ID:    e2.Identity(),
			Value: e2,
		})
	}

	e.observerMutex.Lock()
	defer e.observerMutex.Unlock()

	for _, s := range e.observersSaved {
		if err := s(e, Saved[E, ID]{
			ID:    e2.Identity(),
			Value: e2,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (e *eventRepository[E, ID]) SaveAll(it iter.Seq[E]) error {
	for entity := range it {
		if err := e.Save(entity); err != nil {
			return err
		}
	}

	return nil
}
