// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package form

import (
	"iter"
	"sync"
	"time"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/user"
)

// Choice is one selectable entry of a [Source]: a stored value and the label a human reads.
type Choice struct {
	Value string
	Label string
}

// Identity makes the choice pickable.
func (c Choice) Identity() string { return c.Value }

// String is what the picker displays.
func (c Choice) String() string { return c.Label }

// NewStaticSource builds a [Source] over a fixed set of choices, for the closed sets a domain declares as
// constants (internal or external, quarterly or yearly).
//
// Spelling the labels out here rather than deriving them from the constant keeps the stored value stable
// while the wording stays free to change. The order of the given choices is preserved.
func NewStaticSource(choices ...Choice) Source {
	byID := make(map[string]Choice, len(choices))
	for _, c := range choices {
		byID[c.Value] = c
	}

	ordered := make([]Choice, len(choices))
	copy(ordered, choices)

	return funcAnySource{
		findAll: func(user.Subject) iter.Seq2[string, error] {
			return func(yield func(string, error) bool) {
				for _, c := range ordered {
					if !yield(c.Value, nil) {
						return
					}
				}
			}
		},
		findByID: func(_ user.Subject, id string) (option.Opt[Entity], error) {
			c, ok := byID[id]
			if !ok {
				return option.None[Entity](), nil
			}

			return option.Some(Entity{ID: c.Value, Value: c}), nil
		},
	}
}

// NewQuerySource builds a [Source] over anything that can list itself for a subject - typically a use case
// returning a plain slice, which is the shape [NewSource] cannot take because it requires a repository.
//
// The subject is passed through, so a picker can never offer something the actor is not allowed to see.
//
// # Why there is an index
//
// The form renderers walk a source as one FindAll followed by one FindByID for every identity it yielded. A
// source that answers FindByID by listing again therefore costs one listing plus one per entry. For a listing
// that is a local map lookup this is irrelevant; for one that is an HTTP call it is the difference between a
// responsive form and one that takes seconds. FindAll leaves its result behind so the lookups that follow are
// map reads.
//
// The index is kept per subject. A source is typically built once at start-up and shared by every window, so
// an index left by one actor must never answer another's lookup: two people opening the same form would
// otherwise see each other's permitted set.
func NewQuerySource[T any](list func(subject user.Subject) ([]T, error), id func(T) string, label func(T) string) Source {
	idx := &choiceIndex{entries: map[user.ID]choiceIndexEntry{}}

	return funcAnySource{
		findAll: func(subject user.Subject) iter.Seq2[string, error] {
			items, err := list(subject)
			if err != nil {
				return func(yield func(string, error) bool) { yield("", err) }
			}

			choices := make([]Choice, 0, len(items))
			for _, item := range items {
				choices = append(choices, Choice{Value: id(item), Label: label(item)})
			}

			idx.put(subjectID(subject), choices)

			return func(yield func(string, error) bool) {
				for _, c := range choices {
					if !yield(c.Value, nil) {
						return
					}
				}
			}
		},
		findByID: func(subject user.Subject, wanted string) (option.Opt[Entity], error) {
			if c, ok := idx.get(subjectID(subject), wanted); ok {
				return option.Some(Entity{ID: c.Value, Value: c}), nil
			}

			// A miss is not necessarily an error: it happens when a form shows a value that was not in the
			// listing - a retired group on an existing profile, for instance - or when the index expired
			// between the two calls. Ask the query itself before giving up.
			items, err := list(subject)
			if err != nil {
				return option.None[Entity](), err
			}

			for _, item := range items {
				if id(item) == wanted {
					c := Choice{Value: wanted, Label: label(item)}
					return option.Some(Entity{ID: c.Value, Value: c}), nil
				}
			}

			return option.None[Entity](), nil
		},
	}
}

// subjectID tolerates a nil subject, which occurs in tests and in views rendered before login.
func subjectID(subject user.Subject) user.ID {
	if subject == nil {
		return ""
	}

	return subject.ID()
}

// choiceIndexEntry is one subject's last listing.
type choiceIndexEntry struct {
	choices map[string]Choice
	at      time.Time
}

// choiceIndex remembers the last listing per subject so the lookups that follow a traversal are map reads.
type choiceIndex struct {
	mutex   sync.RWMutex
	entries map[user.ID]choiceIndexEntry
}

// choiceIndexTTL is how long an entry is worth keeping. It only has to survive the lookups of one render; the
// rest is housekeeping.
const choiceIndexTTL = time.Minute

func (i *choiceIndex) put(subject user.ID, choices []Choice) {
	byID := make(map[string]Choice, len(choices))
	for _, c := range choices {
		byID[c.Value] = c
	}

	i.mutex.Lock()
	defer i.mutex.Unlock()

	now := time.Now()
	i.entries[subject] = choiceIndexEntry{choices: byID, at: now}

	// Prune while we hold the lock. Without it the map would keep an entry for every subject that has ever
	// opened a form.
	for id, e := range i.entries {
		if now.Sub(e.at) > choiceIndexTTL {
			delete(i.entries, id)
		}
	}
}

func (i *choiceIndex) get(subject user.ID, wanted string) (Choice, bool) {
	i.mutex.RLock()
	defer i.mutex.RUnlock()

	entry, ok := i.entries[subject]
	if !ok || time.Since(entry.at) > choiceIndexTTL {
		return Choice{}, false
	}

	c, ok := entry.choices[wanted]
	return c, ok
}
