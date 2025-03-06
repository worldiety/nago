package cfgscheduler

import (
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/admin"
	"go.wdy.de/nago/application/scheduler"
	uischeduler "go.wdy.de/nago/application/scheduler/ui"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data/json"
	"go.wdy.de/nago/presentation/core"
	"log/slog"
)

type SchedulerManagement struct {
	settingsRepo scheduler.SettingsRepository
	UseCases     scheduler.UseCases
	Pages        uischeduler.Pages
}

func Enable(cfg *application.Configurator) (SchedulerManagement, error) {
	management, ok := application.SystemServiceFor[SchedulerManagement](cfg, "")
	if ok {
		return management, nil
	}

	settingsStore, err := cfg.EntityStore("nago.scheduler.settings")
	if err != nil {
		return SchedulerManagement{}, err
	}

	settingsRepo := json.NewSloppyJSONRepository[scheduler.Settings, scheduler.ID](settingsStore)

	management = SchedulerManagement{
		settingsRepo: settingsRepo,
		UseCases:     scheduler.NewUseCases(cfg.Context(), settingsRepo),
		Pages: uischeduler.Pages{
			SchedulerDashboard: "admin/scheduler/overview",
		},
	}

	cfg.RootViewWithDecoration(management.Pages.SchedulerDashboard, func(wnd core.Window) core.View {
		return uischeduler.PageOverview(wnd, management.UseCases)
	})

	cfg.AddAdminCenterGroup(func(subject auth.Subject) admin.Group {
		group := admin.Group{
			Title: "Hintergrundprozesse",
		}

		for opt, err := range management.UseCases.ListSchedulers(user.SU()) {
			if err != nil {
				slog.Error("failed to add scheduler to admin center group", "err", err)
				return group
			}

			group.Entries = append(group.Entries, admin.Card{
				Title:        opt.Name,
				Text:         opt.Description,
				Target:       management.Pages.SchedulerDashboard,
				TargetParams: core.Values{"id": string(opt.ID)},
				Permission:   scheduler.PermStatus,
			})
		}

		return group
	})
	cfg.AddSystemService("nago.scheduler", management)

	slog.Info("installed scheduler management")

	return management, nil
}
