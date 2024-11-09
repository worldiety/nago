package iamui

import (
	"go.wdy.de/nago/auth/iam"
	"go.wdy.de/nago/presentation/core"
)

func Account(wnd core.Window, service *iam.Service) core.View {
	if !wnd.Subject().Valid() {

	}

	return nil
}
