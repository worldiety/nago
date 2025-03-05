package cfgusercircle

import (
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/admin"
	"go.wdy.de/nago/application/rcrud"
	"go.wdy.de/nago/application/usercircle"
	uiusercircles "go.wdy.de/nago/application/usercircle/ui"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data/json"
	"go.wdy.de/nago/presentation/core"
	"log/slog"
)

type Management struct {
	circleRepo usercircle.Repository
	UseCases   usercircle.UseCases
	Pages      uiusercircles.Pages
}

func Enable(cfg *application.Configurator) (Management, error) {
	management, ok := application.SystemServiceFor[Management](cfg, "")
	if ok {
		return management, nil
	}

	users, err := cfg.UserManagement()
	if err != nil {
		return Management{}, err
	}

	roles, err := cfg.RoleManagement()
	if err != nil {
		return Management{}, err
	}

	groups, err := cfg.GroupManagement()
	if err != nil {
		return Management{}, err
	}

	entityStore, err := cfg.EntityStore("nago.usercircle.circle")
	if err != nil {
		return Management{}, err
	}

	circleRepo := json.NewSloppyJSONRepository[usercircle.Circle, usercircle.ID](entityStore)
	useCases := usercircle.NewUseCases(circleRepo, users.UseCases, groups.UseCases.FindByID, roles.UseCases.FindByID)
	funcs := rcrud.Funcs[usercircle.Circle, usercircle.ID]{
		PermFindByID:   usercircle.PermFindByID,
		PermFindAll:    usercircle.PermFindAll,
		PermDeleteByID: usercircle.PermDeleteByID,
		PermCreate:     usercircle.PermCreate,
		PermUpdate:     usercircle.PermUpdate,
		FindByID:       useCases.FindByID,
		FindAll:        useCases.FindAll,
		DeleteByID:     useCases.DeleteByID,
		Create:         useCases.Create,
		Update:         useCases.Update,
		Upsert:         nil,
	}

	management = Management{
		circleRepo: circleRepo,
		UseCases:   useCases,
		Pages: uiusercircles.Pages{
			CirclesAdmin: "admin/user/circles",
			MyCircle:     "admin/user/my-circle",
		},
	}

	cfg.RootViewWithDecoration(management.Pages.CirclesAdmin, func(wnd core.Window) core.View {
		return uiusercircles.PageOverview(wnd, rcrud.UseCasesFrom(&funcs))
	})

	cfg.RootViewWithDecoration(management.Pages.MyCircle, func(wnd core.Window) core.View {
		return uiusercircles.PageMyCircle(wnd, useCases, roles.UseCases.FindByID, groups.UseCases.FindByID)
	})

	cfg.AddAdminCenterGroup(func(subject auth.Subject) admin.Group {
		group := admin.Group{
			Title: "Nutzerkreise",
			Entries: []admin.Card{
				{
					Title:      "Nutzerkreise",
					Text:       "Nutzerkreise verwalten, d.h. anlegen, bearbeiten und löschen.",
					Target:     management.Pages.CirclesAdmin,
					Permission: usercircle.PermFindAll,
				},
			},
		}

		for circle, _ := range useCases.MyCircles(subject) {
			group.Entries = append(group.Entries, admin.Card{
				Title:        circle.Name,
				Text:         circle.Description,
				Target:       management.Pages.MyCircle,
				TargetParams: core.Values{"id": string(circle.ID)},
			})
		}

		return group
	})
	cfg.AddSystemService("nago.usercircles", management)

	slog.Info("installed user circle management")

	return management, nil
}
