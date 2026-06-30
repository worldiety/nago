// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cfgevs

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/admin"
	"go.wdy.de/nago/application/evs"
	uievs "go.wdy.de/nago/application/evs/ui"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/events"
	"go.wdy.de/nago/presentation/core"
)

type Module[Evt any] struct {
	Mutex       *sync.Mutex // Mutex used by the UseCases to protect critical write sections
	UseCases    evs.UseCases[Evt]
	Pages       uievs.Pages
	Permissions evs.Permissions
}

type AdminCenter struct {
	// Description is shown on the card. If empty, a default CRUD text is shown.
	Description string
}

type Options[Evt any] struct {
	// Mutex to protect the default critical sections. If nil, a new mutex is allocated as required.
	// If you don't know or care, just leave it nil.
	Mutex *sync.Mutex

	// Bus to be passed into all use cases. May be nil.
	Bus events.Bus

	// AdminCenter configuration for this entity type.
	AdminCenter AdminCenter

	// Schema maps discriminators to concrete types implementing Evt.
	Schema map[evs.Discriminator]reflect.Type

	// DecorateUseCases is invoked before the use cases are passed into all generated and dependent code fragments
	// thus you can customize, intercept or replace any standard use case here. For example, you can
	// apply custom validation and return [xerrors.WithFields].
	DecorateUseCases func(uc evs.UseCases[Evt]) evs.UseCases[Evt]

	HideInAdminCenter bool

	onCreated []func(uc evs.UseCases[Evt]) error
}

func (o Options[Evt]) WithOptions(opts ...Opt[Evt]) Options[Evt] {
	for _, opt := range opts {
		opt(&o)
	}

	return o
}

type Opt[Evt any] func(*Options[Evt])

func Schema[T any, Evt any](d evs.Discriminator) Opt[Evt] {
	return func(o *Options[Evt]) {
		if o.Schema == nil {
			o.Schema = make(map[evs.Discriminator]reflect.Type)
		}

		if _, ok := o.Schema[d]; ok {
			panic(fmt.Errorf("duplicate discriminator: %s", d))
		}

		o.Schema[d] = reflect.TypeFor[T]()
	}
}

// Enable configures an event sourcing module instance. See also [evs.UseCases] and [evs.DeclarePermissions] for details.
func Enable[Evt any](cfg *application.Configurator, prefix permission.ID, entityName string, opts Options[Evt]) (Module[Evt], error) {
	mod, ok := core.FromContext[Module[Evt]](cfg.Context(), "")
	if ok {
		return mod, nil
	}

	if !prefix.Valid() {
		return Module[Evt]{}, fmt.Errorf("prefix is not valid")
	}

	eventsBucketName := string(prefix) + ".event"
	eventStore, err := cfg.EntityStore(eventsBucketName)
	if err != nil {
		return mod, fmt.Errorf("failed to open entity store: %w", err)
	}

	if opts.Mutex == nil {
		opts.Mutex = &sync.Mutex{}
	}

	perms := evs.DeclarePermissions[Evt](prefix, entityName)

	uc := evs.NewUseCases[Evt](perms, eventStore, evs.Options[Evt]{
		Mutex: opts.Mutex,
		Bus:   opts.Bus,
	})

	for discriminator, r := range opts.Schema {
		if err := uc.Register(r, discriminator); err != nil {
			return mod, err
		}
	}

	for _, fn := range opts.onCreated {
		if err := fn(uc); err != nil {
			return mod, err
		}
	}

	mod = configureMod(cfg, perms, uc, opts)
	return mod, nil
}

// NewHandler enables a new event sourcing module instance but just returns
// the according handler instance. This is a convenience method for all domain implementations which just
// need the handler as foundation. aggID routes each event to its aggregate (see [evs.AggregateID]).
func NewHandler[Aggregate evs.Aggregate[Aggregate], SuperEvt evs.Evt[Aggregate], Primary ~string](cfg *application.Configurator, prefix permission.ID, entityName string, aggID evs.AggregateID[SuperEvt, Primary], events []SuperEvt) (*evs.Handler[Aggregate, SuperEvt, Primary], error) {
	if !prefix.Valid() {
		return nil, fmt.Errorf("prefix is not valid")
	}

	evsSchemas := map[evs.Discriminator]reflect.Type{}
	for _, ev := range events {
		if other, ok := evsSchemas[ev.Discriminator()]; ok {
			return nil, fmt.Errorf("duplicate event discriminator: %s defined by both %v and %T", ev.Discriminator(), other, ev)
		}

		evsSchemas[ev.Discriminator()] = reflect.TypeOf(ev)
	}

	// Enable wires the admin UI / use cases on top of the same event store the
	// backend below persists into.
	if _, err := Enable[SuperEvt](cfg, prefix, entityName, Options[SuperEvt]{Schema: evsSchemas}); err != nil {
		return nil, err
	}

	eventStore, err := cfg.EntityStore(string(prefix) + ".event")
	if err != nil {
		return nil, fmt.Errorf("failed to open entity store: %w", err)
	}

	backend := evs.NewBlobBackend[SuperEvt, Aggregate](eventStore)
	handler := evs.NewHandler[Aggregate](backend, aggID, backend.Register)

	for _, event := range events {
		handler.RegisterEvents(event)
	}

	return handler, nil
}

func configureMod[Evt any](cfg *application.Configurator, perms evs.Permissions, uc evs.UseCases[Evt], opts Options[Evt]) Module[Evt] {
	if opts.DecorateUseCases != nil {
		uc = opts.DecorateUseCases(uc)
	}

	mod := Module[Evt]{

		UseCases: uc,
		Pages: uievs.Pages{
			Audit:  "admin/events/" + makeFactoryID(perms.Prefix) + "/audit",
			Create: "admin/events/" + makeFactoryID(perms.Prefix) + "/create",
		},
		Permissions: perms,
	}

	cfg.RootViewWithDecoration(mod.Pages.Audit, func(wnd core.Window) core.View {
		return uievs.PageAudit(wnd, mod.UseCases, uievs.PageAuditOptions[Evt]{
			EntityName: perms.EntityName,
			Perms:      perms,
			Pages:      mod.Pages,
			Prefix:     perms.Prefix,
		})
	})

	for rT := range uc.RegisteredTypes() {
		cfg.RootViewWithDecoration(mod.Pages.Create.Join(string(rT.Discriminator)), func(wnd core.Window) core.View {
			return uievs.PageCreate(wnd, mod.UseCases, uievs.PageCreateOptions[Evt]{
				EntityName: perms.EntityName,
				Perms:      perms,
				Pages:      mod.Pages,
				Prefix:     perms.Prefix,
			})
		})
	}

	cfg.AddAdminCenterGroup(func(subject auth.Subject) admin.Group {
		var res admin.Group

		if opts.HideInAdminCenter {
			return res
		}

		if !(subject.HasPermission(perms.ReadAll)) {
			return res
		}

		cardText := subject.Bundle().Resolve(opts.AdminCenter.Description)
		if cardText == "" {
			cardText = uievs.StrManageEntitiesX.Get(subject, i18n.String("name", perms.EntityName))
		}

		groupTitle := perms.EntityName

		res.Title = groupTitle
		res.Entries = append(res.Entries, admin.Card{
			Title:  perms.EntityName,
			Text:   cardText,
			Target: mod.Pages.Audit,
			ID:     string(perms.Prefix),
		})

		return res
	})

	cfg.AddContextValue(core.ContextValue(string("module-"+perms.Prefix), mod))
	//cfg.AddContextValue(core.ContextValue(string(perms.Prefix), form.AnyUseCaseList[T, ID](uc.FindAll)))

	return mod
}

func makeFactoryID(prefix permission.ID) core.NavigationPath {
	return core.NavigationPath(strings.ReplaceAll(string(prefix), ".", "-"))
}
