package application

import (
	"fmt"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/auth/iam"
	"go.wdy.de/nago/auth/iam/iamui"
	"go.wdy.de/nago/pkg/data/json"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/icon"
	"go.wdy.de/nago/presentation/ora"
	"go.wdy.de/nago/presentation/ui"
)

type IAMSettings struct {
	Decorator   func(wnd core.Window, page *ui.Page, content core.Component)
	Permissions Permissions
	Users       Users
	Sessions    Sessions
	Login       Login
	Logout      Logout
}

type Login struct {
	// default to iam/login
	ID ora.ComponentFactoryId
}

type Logout struct {
	// default to iam/logout
	ID ora.ComponentFactoryId
}

type Users struct {
	// defaults to iam/users
	ID         ora.ComponentFactoryId
	Repository iam.UserRepository
}

type Permissions struct {
	// defaults to iam/permissions
	ID          ora.ComponentFactoryId
	Permissions *iam.Permissions
}

type Sessions struct {
	Repository iam.SessionRepository
}

func (c *Configurator) IAM(settings IAMSettings) IAMSettings {
	if settings.Permissions.ID == "" {
		settings.Permissions.ID = "iam/permissions"
	}

	if settings.Permissions.Permissions == nil {
		settings.Permissions.Permissions = iam.PermissionsFrom[iam.Permission](nil)
	}

	if settings.Users.ID == "" {
		settings.Users.ID = "iam/users"
	}

	if settings.Login.ID == "" {
		settings.Login.ID = "iam/login"
	}

	if settings.Logout.ID == "" {
		settings.Logout.ID = "iam/logout"
	}

	if settings.Decorator == nil {
		settings.Decorator = func(wnd core.Window, page *ui.Page, content core.Component) {
			page.Body().Set(ui.NewScaffold(func(scaffold *ui.Scaffold) {
				subject := wnd.Subject()
				scaffold.NavigationComponent().Set(ui.NewNavigationComponent(func(navigationComponent *ui.NavigationComponent) {
					navigationComponent.Alignment().Set(ora.AlignmentLeft)
					navigationComponent.Menu().Append(
						ui.NewMenuEntry(func(menuEntry *ui.MenuEntry) {
							menuEntry.Link(settings.Login.ID, wnd, nil)
							if subject.Valid() {
								menuEntry.Title().Set("Mit einem anderen Konto anmelden")
								menuEntry.Icon().Set(icon.UserPlus)
							} else {
								menuEntry.Title().Set("Anmelden")
								menuEntry.Icon().Set(icon.UserPlus)
							}
						}),
					)

					if subject.Valid() {
						navigationComponent.Menu().Append(
							ui.NewMenuEntry(func(menuEntry *ui.MenuEntry) {
								menuEntry.Link(settings.Logout.ID, wnd, nil)
								menuEntry.Title().Set("Abmelden")
								menuEntry.Icon().Set(icon.ArrowLeftStartOnRectangle)
							}),
						)
					}

					if auth.OneOf(subject, iam.ReadPermission, iam.ReadUser) {
						navigationComponent.Menu().Append(

							ui.NewMenuEntry(func(usm *ui.MenuEntry) {
								usm.Title().Set("Nutzerverwaltung")
								usm.Icon().Set(icon.Users)
								usm.Menu().Append(
									ui.NewMenuEntry(func(menuEntry *ui.MenuEntry) {
										menuEntry.Link(settings.Permissions.ID, wnd, nil)
										menuEntry.Title().Set("Berechtigungen")
										menuEntry.Icon().Set(icon.Users)
									}),
									ui.NewMenuEntry(func(menuEntry *ui.MenuEntry) {
										menuEntry.Link(settings.Users.ID, wnd, nil)
										menuEntry.Title().Set("Nutzerkonten")
										menuEntry.Icon().Set(icon.Users)
									}),
								)
							}),
						)
					}

				}))
				scaffold.Body().Set(content)
			}))
		}
	}

	if settings.Users.Repository == nil {
		settings.Users.Repository = json.NewSloppyJSONRepository[iam.User](c.EntityStore("iam.users"))
	}

	if settings.Sessions.Repository == nil {
		settings.Sessions.Repository = json.NewSloppyJSONRepository[iam.Session](c.EntityStore("iam.sessions"))
	}

	c.iamSettings = settings

	service := iam.NewService(settings.Permissions.Permissions, settings.Users.Repository, settings.Sessions.Repository)
	if err := service.Bootstrap(); err != nil {
		panic(fmt.Errorf("cannot bootstrap IAM service: %v", err))
	}
	c.Component(settings.Permissions.ID, func(wnd core.Window) core.Component {
		page := ui.NewPage(nil)
		settings.Decorator(wnd, page, iamui.Permissions(wnd.Subject(), service))
		return page
	})

	c.Component(settings.Login.ID, func(wnd core.Window) core.Component {
		page := ui.NewPage(nil)
		settings.Decorator(wnd, page, iamui.Login(wnd, page, service))
		return page
	})

	c.Component(settings.Logout.ID, func(wnd core.Window) core.Component {
		page := ui.NewPage(nil)
		settings.Decorator(wnd, page, iamui.Logout(wnd, service))
		return page
	})

	c.Component(settings.Users.ID, func(wnd core.Window) core.Component {
		page := ui.NewPage(nil)
		settings.Decorator(wnd, page, iamui.Users(wnd.Subject(), page, service))
		return page

	})

	c.AddOnWindowCreatedObserver(func(wnd core.Window) {
		wnd.UpdateSubject(service.Subject(wnd.SessionID()))
	})

	return settings
}
