package application

import (
	uiadmin "go.wdy.de/nago/application/admin/ui"
	"go.wdy.de/nago/pkg/std"
)

type AdminManagement struct {
	Pages uiadmin.Pages
}

func (c *Configurator) AdminPages() uiadmin.Pages {
	var pages uiadmin.Pages
	if c.HasMailManagement() {
		pages.Mail = std.Some(std.Must(c.MailManagement()).Pages)
	}

	return pages
}

func (c *Configurator) AdminManagement() (AdminManagement, error) {
	if c.adminManagement == nil {
		c.adminManagement = &AdminManagement{
			Pages: uiadmin.Pages{
				Dashboard: "admin",
			},
		}
	}

	return *c.adminManagement, nil
}
