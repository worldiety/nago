package alert

import (
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/presentation/core"
)

// IfPermissionDenied returns a view if either the user has not been authenticated or if not authorized.
// Note, that no given oneOf permission just validates for an authenticated user.
// If everything is fine, a nil interface is returned.
func IfPermissionDenied(wnd core.Window, oneOf ...permission.ID) core.View {
	if !wnd.Subject().Valid() {
		return Banner("Anmeldung erforderlich", "Um fortzufahren, müssen Sie sich am System anmelden.")
	}

	if len(oneOf) > 0 {
		if !auth.OneOf(wnd.Subject(), oneOf...) {
			return Banner("Zugriff verweigert", "Leider haben Sie nicht die nötigen Rechte, um auf diese Daten zuzugreifen.")
		}
	}

	return nil
}
