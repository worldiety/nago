package main

import (
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/pkg/workspace"
	"go.wdy.de/nago/pkg/workspace/workspaceui"
	"go.wdy.de/nago/presentation/core"
	heroSolid "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		iamCfg := application.IAMSettings{}
		iamCfg.Decorator = cfg.NewScaffold().Decorator()
		iamCfg = cfg.IAM(iamCfg)

		workspaceRepo := application.SloppyRepository[workspace.Workspace](cfg)
		wsList := workspace.NewList(workspaceRepo)
		wsSave := workspace.NewSave(workspaceRepo)

		workspaceTypes := []workspaceui.DashboardType{
			{Icon: heroSolid.AcademicCap, Type: "kernel", Name: "worldiety Rechenkernel", Description: "Die Metriksuchmaschine wird mit eigenen Rechenkerneln bestückt, die in Go geschrieben werden und aus mehreren Go-Source Code Dateien bestehen kann."},
			{Icon: heroSolid.GlobeEuropeAfrica, Type: "invoice", Name: "Rechungsvorlagen", Description: "Eine Latex-Vorlage, die Logos und latexmk Dateien enthält, um eine PDF Rechnung zu erstellen, was wegen XRechnung aber keiner mehr braucht."},
			{Icon: heroSolid.Trash, Type: "offer", Name: "Angebotsvorlagen", Description: "Wir machen jetzt die Vorlagen in Typst, weil man damit viel schneller Templates schreiben kann und das Kompilieren auch viel schneller ist als in Latex."},
		}

		cfg.RootView(".", iamCfg.DecorateRootView(func(wnd core.Window) core.View {
			return workspaceui.Dashboard(wnd, workspaceui.DashboardOptions{
				Title:            "Übersicht Vorlagen",
				Types:            workspaceTypes,
				OverviewListPath: "workspace/list",
			})
		}))

		cfg.RootView("workspace/list", iamCfg.DecorateRootView(func(wnd core.Window) core.View {
			return workspaceui.WorkspaceList(wnd, workspaceui.WorkspaceListOptions{
				ListWorkspaces: wsList,
				SaveWorkspace:  wsSave,
				Types:          workspaceTypes,
			})
		}))

	}).Run()
}
