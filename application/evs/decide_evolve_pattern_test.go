// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package evs_test

import (
	"context"
	"fmt"
	"iter"
	"testing"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/evs"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
)

type PID string
type Person struct {
	id        PID
	firstname string
	lastname  string
	deleted   bool
}

func (p *Person) Clone() *Person {
	return &Person{id: p.id, firstname: p.firstname, lastname: p.lastname, deleted: p.deleted}
}

func (p *Person) IsDeleted() bool {
	return p.deleted
}

type UpdateFirstnameCmd struct {
	ID        PID
	Firstname string
}

func (cmd UpdateFirstnameCmd) Decide(subject auth.Subject, aggregate *Person) ([]Evt, error) {
	if cmd.Firstname == aggregate.firstname {
		return nil, fmt.Errorf("firstname did not change")
	}

	if cmd.Firstname == "" {
		return nil, fmt.Errorf("firstname cannot be empty")
	}

	return []Evt{FirstnameUpdated{Person: cmd.ID, Firstname: cmd.Firstname}}, nil
}

type Evt = evs.Evt[*Person]

type FirstnameUpdated struct {
	Person    PID    `json:"person"`
	Firstname string `json:"firstname"`
}

func (f FirstnameUpdated) Discriminator() evs.Discriminator {
	return "FirstnameUpdated"
}

func (f FirstnameUpdated) Evolve(ctx context.Context, mut *Person) error {
	mut.firstname = f.Firstname
	return nil
}

type LastnameUpdated struct {
	Person PID    `json:"person"`
	Name   string `json:"name"`
}

func (f LastnameUpdated) Discriminator() evs.Discriminator {
	return "LastnameUpdated"
}

func (f LastnameUpdated) Evolve(ctx context.Context, mut *Person) error {
	mut.lastname = f.Name
	return nil
}

// memBackend is a minimal in-memory [evs.Backend] test double. Append assigns a
// monotonic sequence and stores the envelope; ReplayAll yields them in order.
type memBackend struct {
	events []evs.Envelope[Evt]
}

func (b *memBackend) Append(subject auth.Subject, e Evt) (evs.Envelope[Evt], error) {
	env := evs.Envelope[Evt]{
		Sequence: evs.SeqID(len(b.events) + 1),
		Data:     e,
	}
	b.events = append(b.events, env)
	return env, nil
}

func (b *memBackend) ReplayAll(subject auth.Subject) iter.Seq2[evs.Envelope[Evt], error] {
	return func(yield func(evs.Envelope[Evt], error) bool) {
		for _, env := range b.events {
			if !yield(env, nil) {
				return
			}
		}
	}
}

func TestHandler(t *testing.T) {
	backend := &memBackend{}
	handler := evs.NewHandler[*Person](backend, func(e Evt) (PID, bool) {
		switch evt := e.(type) {
		case FirstnameUpdated:
			return evt.Person, evt.Person != ""
		case LastnameUpdated:
			return evt.Person, evt.Person != ""
		default:
			return "", false
		}
	}, nil)

	handler.RegisterEvents(
		FirstnameUpdated{},
		LastnameUpdated{},
	)

	if err := handler.Handle(user.SU(), "1", UpdateFirstnameCmd{ID: "1", Firstname: "John"}); err != nil {
		t.Error(err)
	}

	if option.Must(handler.Aggregate(context.Background(), "1")).firstname != "John" {
		t.Error("firstname did not change")
	}

	t.Log(handler.Aggregate(context.Background(), "1"))
}
