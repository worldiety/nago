package workspaceui

import (
	"fmt"
	"go.wdy.de/nago/pkg/workspace"
	"go.wdy.de/nago/presentation/core"
	heroSolid "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/crud"
	"go.wdy.de/nago/presentation/ui/list"
)

type WorkspaceListOptions struct {
	ListWorkspaces workspace.List
	SaveWorkspace  workspace.Save
	Types          []DashboardType
}

func WorkspaceList(wnd core.Window, opts WorkspaceListOptions) core.View {
	typeId := workspace.Type(wnd.Values()["type"])
	var selectedType DashboardType
	for _, dashboardType := range opts.Types {
		if dashboardType.Type == typeId {
			selectedType = dashboardType
			break
		}
	}

	if selectedType.Type == "" {
		return alert.Banner("Workspace-Typ nicht gefunden", fmt.Sprintf("Der Typ '%s' ist nicht bekannt", typeId))
	}

	bnd := newBinding(wnd, selectedType)
	return ui.VStack(
		ui.H1(selectedType.Name),
		ui.Text(selectedType.Description).Padding(ui.Padding{Bottom: ui.L16}),
		crud.View(crud.Options(bnd).
			FindAll(opts.ListWorkspaces(wnd.Subject(), selectedType.Type)).
			ViewStyle(crud.ViewStyleListOnly).
			Actions(crud.ButtonCreate(bnd, workspace.Workspace{}, func(w workspace.Workspace) (errorText string, infrastructureError error) {
				return "", opts.SaveWorkspace(wnd.Subject(), w)
			})),
		).Frame(ui.Frame{}.FullWidth()),
	).Alignment(ui.Leading).FullWidth()
}

func newBinding(wnd core.Window, dt DashboardType) *crud.Binding[workspace.Workspace] {
	bnd := crud.NewBinding[workspace.Workspace](wnd)
	bnd.IntoListEntry(func(entity workspace.Workspace) list.TEntry {
		var sumMib int64
		for _, file := range entity.Files {
			sumMib += file.Size
		}

		return list.Entry().
			Leading(ui.ImageIcon(dt.Icon)).
			Headline(entity.Name).
			SupportingText(fmt.Sprintf("%d Dateien / %.2f MiB", len(entity.Files), float64(sumMib)/1024/1024)).
			Trailing(ui.ImageIcon(heroSolid.ArrowRight))
	})

	bnd.Add(crud.Text(crud.TextOptions{Label: "Name"}, crud.Ptr(func(model *workspace.Workspace) *string {
		return &model.Name
	})))

	return bnd
}
