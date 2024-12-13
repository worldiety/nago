package application

import (
	"go.wdy.de/nago/application/group"
	uigroup "go.wdy.de/nago/application/group/ui"
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

		c.RootView(c.groupManagement.Pages.Groups, c.DecorateRootView(func(wnd core.Window) core.View {
			return uigroup.Groups(wnd, c.groupManagement.UseCases)
		}))
	}

	return *c.groupManagement, nil
}
