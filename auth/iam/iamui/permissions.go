package iamui

import (
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/auth/iam"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

func Permissions(subject auth.Subject, service *iam.Service) core.Component {
	return ui.NewTable(func(table *ui.Table) {
		table.Header().Append(ui.NewTextCell("Anwendungsfall"), ui.NewTextCell("Beschreibung"))
		service.AllPermissions(subject)(func(permission iam.Permission, err error) bool {
			table.Rows().Append(ui.NewTableRow(func(row *ui.TableRow) {
				row.Cells().Append(
					ui.NewTextCell(permission.Name()),
					ui.NewTextCell(permission.Desc()),
				)
			}))

			return true
		})

	})
}
