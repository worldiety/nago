package application

import (
	"fmt"
	"go.wdy.de/nago/application/group"
	uigroup "go.wdy.de/nago/application/group/ui"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data/json"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui/form"
	"iter"
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

		c.AddSystemService("nago.groups", form.AnyUseCaseList[group.Group, group.ID](func(subject auth.Subject) iter.Seq2[group.Group, error] {
			return c.groupManagement.UseCases.FindAll(subject)
		}))

		c.AddSystemService("nago.groups.find_by_id", c.groupManagement.UseCases.FindByID)
	}

	return *c.groupManagement, nil
}
