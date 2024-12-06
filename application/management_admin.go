package application

import (
	uiadmin "go.wdy.de/nago/application/admin/ui"
	"go.wdy.de/nago/pkg/std"
)

func (c *Configurator) AdminPages() uiadmin.Pages {
	var pages uiadmin.Pages
	if c.HasMailManagement() {
		pages.Mail = std.Some(std.Must(c.MailManagement()).Pages)
	}

	return pages
}
