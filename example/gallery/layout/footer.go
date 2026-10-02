// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"go.wdy.de/nago/example/gallery/app"
	"go.wdy.de/nago/presentation/core"
	. "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/footer"
)

func init() {
	app.Register("footer", func(wnd core.Window) core.View {
		return VStack(
			footer.Footer().
				Logo(Text("ACME").Font(HeadlineSmall)).
				Slogan("Software that just works.").
				Impress("/impress").
				PrivacyPolicy("/privacy").
				GeneralTermsAndConditions("/gtc").
				TermsOfUse("/terms").
				ProviderName("ACME GmbH"),
		).Frame(Frame{Width: L880})
	})
}
