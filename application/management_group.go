package application

import (
	"fmt"
	"go.wdy.de/nago/application/group"
	uigroup "go.wdy.de/nago/application/group/ui"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/data/json"
	"go.wdy.de/nago/presentation/core"
)

type GroupManagement struct {
	UseCases group.UseCases
	Pages    uigroup.Pages
}

func (c *Configurator) GroupManagement() (GroupManagement, error) {
	if c.groupManagement == nil {
		groupStore, err := c.EntityStore("nago.iam.group")
		if err != nil {
			return GroupManagement{}, err
		}

		groupRepo := json.NewSloppyJSONRepository[group.Group, group.ID](groupStore)

		c.groupManagement = &GroupManagement{
			UseCases: group.NewUseCases(groupRepo),
			Pages: uigroup.Pages{
				Groups: "admin/groups",
			},
		}

		if _, err := c.groupManagement.UseCases.Upsert(user.NewSystem()(), group.Group{
			ID:          group.System,
			Name:        "System",
			Description: "Die Systemgruppe ist eine interne Gruppe, die nicht für reale Nutzer bestimmt ist und von automatisierten systemrelevanten Diensten verwendet wird.",
		}); err != nil {
			return GroupManagement{}, fmt.Errorf("cannot upsert system group: %w", err)
		}

		c.RootView(c.groupManagement.Pages.Groups, c.DecorateRootView(func(wnd core.Window) core.View {
			return uigroup.Groups(wnd, c.groupManagement.UseCases)
		}))
	}

	return *c.groupManagement, nil
}
