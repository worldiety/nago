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
	"reflect"
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
}

func (p *Person) Clone() *Person {
	return &Person{id: p.id, firstname: p.firstname, lastname: p.lastname}
}

type UpdateFirstnameCmd struct {
	Firstname string
}

func (cmd UpdateFirstnameCmd) Decide(subject auth.Subject, aggregate *Person) ([]Evt, error) {
	if cmd.Firstname == aggregate.firstname {
		return nil, fmt.Errorf("firstname did not change")
	}

	if cmd.Firstname == "" {
		return nil, fmt.Errorf("firstname cannot be empty")
	}

	return []Evt{FirstnameUpdated{Firstname: cmd.Firstname}}, nil
}

type Evt = evs.Evt[*Person]

type FirstnameUpdated struct {
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
	Name string `json:"name"`
}

func (f LastnameUpdated) Discriminator() evs.Discriminator {
	return "LastnameUpdated"
}

func (f LastnameUpdated) Evolve(ctx context.Context, mut *Person) error {
	mut.lastname = f.Name
	return nil
}

func TestHandler(t *testing.T) {
	handler := evs.NewHandler[*Person, Evt, PID](
		func(t reflect.Type, discriminatorName evs.Discriminator) error {
			return nil
		},
		func(subject auth.Subject, primary PID, opts evs.ReplayOptions) iter.Seq2[evs.Envelope[Evt], error] {
			return func(yield func(evs.Envelope[Evt], error) bool) {

			}
		},

		func(subject auth.Subject, evt Evt, opts evs.StoreOptions) (evs.Envelope[Evt], error) {
			return evs.Envelope[Evt]{Data: evt}, nil
		},
	)

	handler.RegisterEvents(
		FirstnameUpdated{},
		LastnameUpdated{},
	)

	if err := handler.Handle(user.SU(), "1", UpdateFirstnameCmd{Firstname: "John"}); err != nil {
		t.Error(err)
	}

	if option.Must(handler.Aggregate(context.Background(), "1")).firstname != "John" {
		t.Error("firstname did not change")
	}

	t.Log(handler.Aggregate(context.Background(), "1"))
}
