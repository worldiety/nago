package application

import (
	"go.wdy.de/nago/application/admin"
	uiadmin "go.wdy.de/nago/application/admin/ui"
	"go.wdy.de/nago/presentation/core"
)

type AdminManagement struct {
	FindAll     admin.FindAllGroups
	QueryGroups admin.QueryGroups
	Pages       uiadmin.Pages
}

// WithAdminManagement installs a mutator for a future invocation or immediately mutates the current configuration.
// Note, that even though most build-in implementations will perform a dynamic lookup, you may still want to install
// the handler BEFORE any *Management system has been initialized.
func (c *Configurator) WithAdminManagement(fn func(m *AdminManagement)) *Configurator {
	c.adminManagementMutator = fn
	if c.adminManagement != nil {
		fn(c.adminManagement)
	}
	
	return c
}

func (c *Configurator) AdminManagement() (AdminManagement, error) {
	if c.adminManagement == nil {

		c.adminManagement = &AdminManagement{
			Pages: uiadmin.Pages{
				AdminCenter: "admin",
			},

			FindAll: func() []admin.Group {
				// be invariant on execution order
				var pages admin.Pages
				if c.mailManagement != nil {
					pages.Mail = c.mailManagement.Pages
				}

				if c.roleManagement != nil {
					pages.Role = c.roleManagement.Pages
				}

				if c.sessionManagement != nil {
					pages.Session = c.sessionManagement.Pages
				}

				if c.userManagement != nil {
					pages.User = c.userManagement.Pages
				}

				if c.permissionManagement != nil {
					pages.Permission = c.permissionManagement.Pages
				}

				if c.groupManagement != nil {
					pages.Group = c.groupManagement.Pages
				}

				return admin.DefaultGroups(pages)
			},
			QueryGroups: admin.NewGroups(func() []admin.Group {
				// be invariant on replacements for FindAll, so that developer can inject custom admin groups
				return c.adminManagement.FindAll()
			}),
		}

		if c.adminManagementMutator != nil {
			c.adminManagementMutator(c.adminManagement)
		}

		c.RootView(c.adminManagement.Pages.AdminCenter, c.DecorateRootView(func(wnd core.Window) core.View {
			return uiadmin.AdminCenter(wnd, c.adminManagement.QueryGroups)
		}))
	}

	return *c.adminManagement, nil
}
