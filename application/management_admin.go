package application

import (
	"go.wdy.de/nago/application/admin"
	uiadmin "go.wdy.de/nago/application/admin/ui"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/presentation/core"
	"slices"
	"strings"
)

type AdminManagement struct {
	FindAll     admin.FindAllGroups
	QueryGroups admin.QueryGroups
	Pages       uiadmin.Pages
}

func (c *Configurator) AddAdminCenterGroup(group func(subject auth.Subject) admin.Group) *Configurator {
	c.adminManagementGroups = append(c.adminManagementGroups, group)
	return c
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

		if _, err := c.BillingManagement(); err != nil {
			return AdminManagement{}, err
		}

		c.adminManagement = &AdminManagement{
			Pages: uiadmin.Pages{
				AdminCenter: "admin",
			},

			FindAll: func(subject auth.Subject) []admin.Group {
				if !subject.Valid() {
					return nil
				}
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

				if c.licenseManagement != nil {
					pages.License = c.licenseManagement.Pages
				}

				if c.billingManagement != nil {
					pages.Billing = c.billingManagement.Pages
				}

				if c.backupManagement != nil {
					pages.Backup = c.backupManagement.Pages
				}

				if c.secretManagement != nil {
					pages.Secret = c.secretManagement.Pages
				}

				if c.templateManagement != nil {
					pages.Template = c.templateManagement.Pages
				}

				var groups []admin.Group
				for _, groupFn := range c.adminManagementGroups {
					groups = append(groups, groupFn(subject))
				}

				groups = append(groups, admin.DefaultGroups(pages)...)

				slices.SortFunc(groups, func(a, b admin.Group) int {
					return strings.Compare(a.Title, b.Title)
				})

				for _, group := range groups {
					slices.SortFunc(group.Entries, func(a, b admin.Card) int {
						return strings.Compare(a.Title, b.Title)
					})
				}

				return groups
			},
			QueryGroups: admin.NewGroups(func(subject auth.Subject) []admin.Group {
				// be invariant on replacements for FindAll, so that developer can inject custom admin groups
				return c.adminManagement.FindAll(subject)
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
