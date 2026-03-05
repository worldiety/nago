// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"context"
	"errors"
	"slices"
	"time"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/evs"
	cfgevs "go.wdy.de/nago/application/evs/cfg"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/xerrors"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/web/vuejs"
)

type EID string

type EmployeeAggregate struct {
	ID       EID
	Name     string
	Meetings []MeetingID
}

func (e *EmployeeAggregate) Clone() *EmployeeAggregate {
	return &EmployeeAggregate{
		ID:       e.ID,
		Name:     e.Name,
		Meetings: slices.Clone(e.Meetings),
	}
}

type EmployeeEvent interface {
	evs.Evt[*EmployeeAggregate]
	Employee() EID
}

type EmployeeIntroduced struct {
	EmployeeID EID
	Name       string
}

func (e EmployeeIntroduced) Employee() EID {
	return e.EmployeeID
}

func (e EmployeeIntroduced) Evolve(ctx context.Context, aggregate *EmployeeAggregate) error {
	aggregate.ID = e.EmployeeID
	aggregate.Name = e.Name
	return nil
}

func (e EmployeeIntroduced) Discriminator() evs.Discriminator {
	return "EmployeeIntroduced"
}

// MeetingID references another aggregate root. Replaying and folding/evolving
// a meeting event into different aggregate roots is considered an architectural
// smell
type MeetingID string

// MeetingPlanned just references an external aggregate root for Meetings.
type MeetingPlanned struct {
	Participant EID
	Meeting     MeetingID
}

func (e MeetingPlanned) Evolve(ctx context.Context, aggregate *EmployeeAggregate) error {
	aggregate.Meetings = append(aggregate.Meetings, e.Meeting)
	return nil
}

func (e MeetingPlanned) Discriminator() evs.Discriminator {
	return "MeetingPlanned"
}

func (e MeetingPlanned) Employee() EID {
	return e.Participant
}

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_85_2")
		cfg.Serve(vuejs.Dist())

		option.MustZero(cfg.StandardSystems())
		option.Must(option.Must(cfg.UserManagement()).UseCases.EnableBootstrapAdmin(time.Now().Add(time.Hour), "%6UbRsCuM8N$auy"))
		cfg.SetDecorator(cfg.NewScaffold().Decorator())

		handler := option.Must(cfgevs.NewHandler[*EmployeeAggregate, EmployeeEvent, EID](
			cfg,
			"test.employee",
			"Mitarbeiter",
			func(e evs.Envelope[EmployeeEvent]) (EID, error) {
				return e.Data.Employee(), nil
			},
			[]EmployeeEvent{
				EmployeeIntroduced{},
				MeetingPlanned{},
			},
		))

		ucIntroduce := NewIntroduceEmployee(handler)

		cfg.RootViewWithDecoration(".", func(wnd core.Window) core.View {
			name := core.AutoState[string](wnd)

			return ui.VStack(
				ui.TextField("Name", name.Get()).InputValue(name),
				ui.PrimaryButton(func() {
					if err := ucIntroduce(wnd.Subject(), IntroduceEmployeeCmd{Name: name.Get()}); err != nil {
						alert.ShowBannerError(wnd, err)
					}
				}).Title("Introduce Employee"),
			)
		})
	}).Run()
}

// Architecture style note:
// To align your code better with Event Storming keep your use case function declaration but summarize
// your arguments into a single struct called <UseCaseName>Cmd. The Result of that function is
// the Event which has been stored. See the example below

type IntroduceEmployeeCmd struct {
	ID   EID
	Name string
}

func (cmd IntroduceEmployeeCmd) Decide(subject auth.Subject, aggregate *EmployeeAggregate) ([]EmployeeEvent, error) {
	var errs xerrors.Errors

	if cmd.Name == "" {
		errs.Add("Name", "must not be empty")
	}

	if errs.Has() {
		return nil, errs.Error()
	}

	return []EmployeeEvent{EmployeeIntroduced{EmployeeID: cmd.ID, Name: cmd.Name}}, nil
}

type IntroduceEmployee func(subject auth.Subject, cmd IntroduceEmployeeCmd) error

func NewIntroduceEmployee(handler *evs.Handler[*EmployeeAggregate, EmployeeEvent, EID]) IntroduceEmployee {
	return func(subject auth.Subject, cmd IntroduceEmployeeCmd) error {
		if !subject.Valid() {
			// apply domain specific authorization here
			return errors.New("invalid subject")
		}

		if cmd.ID == "" {
			cmd.ID = data.RandIdent[EID]()
		}

		return handler.Handle(subject, cmd.ID, cmd)
	}
}
