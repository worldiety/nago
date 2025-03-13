package uiusercircles

import (
	"fmt"
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/application/usercircle"
	"go.wdy.de/nago/presentation/core"
	heroSolid "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/list"
)

func PageMyCircleGroups(wnd core.Window, pages Pages, useCases usercircle.UseCases, findGroupById group.FindByID) core.View {
	circle, err := loadMyCircle(wnd, useCases)
	if err != nil {
		return alert.BannerError(err)
	}

	var groups []group.Group
	for _, id := range circle.Groups {
		optRole, err := findGroupById(user.SU(), id) // security note: we are allowed by user circle definition
		if err != nil {
			return alert.BannerError(err)
		}

		if optRole.IsNone() {
			continue
		}

		groups = append(groups, optRole.Unwrap())
	}

	return ui.VStack(
		ui.H1(circle.Name+" / Gruppen"),
		list.List(ui.ForEach(groups, func(t group.Group) core.View {
			return list.Entry().
				Headline(t.Name).
				SupportingText(t.Description).
				Trailing(ui.ImageIcon(heroSolid.ChevronRight))
		})...).OnEntryClicked(func(idx int) {
			grp := groups[idx]
			wnd.Navigation().ForwardTo(pages.MyCircleGroupsUsers, core.Values{"circle": string(circle.ID), "group": string(grp.ID)})
		}).
			Caption(ui.Text("In diesem Kreis sichtbare Gruppen")).
			Footer(ui.Text(fmt.Sprintf("%d Gruppen sind zur Verwaltung verfügbar", len(groups)))).
			Frame(ui.Frame{}.FullWidth()),
	).Alignment(ui.Leading).FullWidth()
}
