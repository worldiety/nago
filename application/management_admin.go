// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application

import (
	"slices"
	"strings"

	"go.wdy.de/nago/application/admin"
	uiadmin "go.wdy.de/nago/application/admin/ui"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/presentation/core"
)

// AdminManagement is a nago system(Admin Management).
// It provides the Admin Center, a central hub that aggregates
// and displays the administration pages of all connected subsystems.
//
// Key features include:
//   - Centralized access to management systems
//   - Automatic integration of built-in systems (User, Role, Session, Permission, etc.)
//   - Role- and permission-based access control
//   - Extensibility by allowing developers to register custom groups and cards
//
// Admin Management is automatically initialized when the application starts
// and makes the admin center available at the navigation path "admin".
type AdminManagement struct {
	FindAll     admin.FindAllGroups
	QueryGroups admin.QueryGroups
	HasEntries  admin.HasEntries
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

		c.adminManagement = &AdminManagement{
			Pages: uiadmin.Pages{
				AdminCenter: "admin",
			},

			FindAll: func(subject auth.Subject) []admin.Group {
				if !subject.Valid() {
					return nil
				}

				var groups []admin.Group
				for _, groupFn := range c.adminManagementGroups {
					grp := groupFn(subject)
					merged := false
					// hm, this is quadratic effort. we should keep an eye on large admin centers
					for i, group := range groups {
						if group.Title == grp.Title {
							group.Entries = append(group.Entries, grp.Entries...)
							groups[i] = group
							merged = true
							break
						}
					}

					if !merged {
						groups = append(groups, grp)
					}
				}

				groups = append(groups, admin.DefaultGroups(c.adminPages())...)

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
			HasEntries: func(subject auth.Subject) bool {
				if !subject.Valid() {
					return false
				}

				// the groups check a lot of permissions, which are mostly denied and therefore the most expensive ones
				subject = user.WithGlobalGrants(subject)

				if c.adminManagementMutator != nil {
					// FindAll or QueryGroups may have been replaced, thus only the installed use case knows the answer
					return len(c.adminManagement.QueryGroups(subject, "")) > 0
				}

				// the same as the default FindAll and QueryGroups, but stops at the first visible card
				if admin.AnyVisible(subject, admin.DefaultGroups(c.adminPages())...) {
					return true
				}

				for _, groupFn := range c.adminManagementGroups {
					if admin.AnyVisible(subject, groupFn(subject)) {
						return true
					}
				}

				return false
			},
		}

		if c.adminManagementMutator != nil {
			c.adminManagementMutator(c.adminManagement)
		}

		c.RootView(c.adminManagement.Pages.AdminCenter, c.DecorateRootView(func(wnd core.Window) core.View {
			return uiadmin.AdminCenter(wnd, c.adminManagement.QueryGroups)
		}))

		c.AddContextValue(core.ContextValue("", c.adminManagement.Pages))
		c.AddContextValue(core.ContextValue("", c.adminManagement.QueryGroups))
		c.AddContextValue(core.ContextValue("", admin.HasEntries(func(subject auth.Subject) bool {
			// be invariant on replacements by WithAdminManagement after the initialization
			return c.adminManagement.HasEntries(subject)
		})))
	}

	return *c.adminManagement, nil
}

// adminPages collects the pages of the default admin center groups. Be invariant on the execution order, the
// management systems may be configured before or after the admin management.
func (c *Configurator) adminPages() admin.Pages {
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
		pages.UsersOverview = c.userManagement.Pages.Users
	}

	if c.permissionManagement != nil {
		pages.Permission = c.permissionManagement.Pages
	}

	if c.groupManagement != nil {
		pages.Group = c.groupManagement.Pages
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

	return pages
}
