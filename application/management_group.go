// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application

import (
	"fmt"
	"iter"

	"go.wdy.de/nago/application/group"
	uigroup "go.wdy.de/nago/application/group/ui"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data/json"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui/form"
)

// GroupManagement is a nago system(Group Management).
// It provides UseCases for creating and managing user groups.
// Groups are used to bundle users together and control access to certain pages or resources.
// They can be created, edited, and deleted via UI or code.
// Group membership is managed through UserManagement.
// A special "System" group is created automatically for internal services
// and is not intended for real users.
type GroupManagement struct {
	UseCases group.UseCases
	Pages    uigroup.Pages
}

func (c *Configurator) GroupManagement() (GroupManagement, error) {
	if c.groupManagement == nil {
		groupStore, err := c.EntityStore(string(group.Namespace))
		if err != nil {
			return GroupManagement{}, err
		}

		groupRepo := json.NewSloppyJSONRepository[group.Group, group.ID](groupStore)

		c.groupManagement = &GroupManagement{
			UseCases: group.NewUseCases(c.EventBus(), groupRepo),
			Pages: uigroup.Pages{
				Groups: "admin/groups",
			},
		}

		if _, err := c.groupManagement.UseCases.Upsert(c.SysUser(), group.Group{
			ID:          group.System,
			Name:        "System",
			Description: "Die Systemgruppe ist eine interne Gruppe, die nicht für reale Nutzer bestimmt ist und von automatisierten systemrelevanten Diensten verwendet wird.",
		}); err != nil {
			return GroupManagement{}, fmt.Errorf("cannot upsert system group: %w", err)
		}

		c.RootView(c.groupManagement.Pages.Groups, c.DecorateRootView(func(wnd core.Window) core.View {
			return uigroup.Groups(wnd, c.groupManagement.UseCases)
		}))

		c.AddContextValue(core.ContextValue("nago.groups", form.AnyUseCaseList[group.Group, group.ID](func(subject auth.Subject) iter.Seq2[group.Group, error] {
			return c.groupManagement.UseCases.FindAll(subject)
		})))

		c.AddContextValue(core.ContextValue("nago.groups.find_by_id", c.groupManagement.UseCases.FindByID))
	}

	return *c.groupManagement, nil
}
