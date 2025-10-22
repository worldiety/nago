// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"time"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application"
	cfgai "go.wdy.de/nago/application/ai/cfg"
	"go.wdy.de/nago/application/ai/conversation"
	_ "go.wdy.de/nago/application/ai/provider/mistralai"
	_ "go.wdy.de/nago/application/ai/provider/openai"
	uiai "go.wdy.de/nago/application/ai/ui"
	"go.wdy.de/nago/application/drive"
	cfgdrive "go.wdy.de/nago/application/drive/cfg"
	cfginspector "go.wdy.de/nago/application/inspector/cfg"
	cfglocalization "go.wdy.de/nago/application/localization/cfg"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/events"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_77")
		cfg.Serve(vuejs.Dist())

		option.MustZero(cfg.StandardSystems())
		option.Must(option.Must(cfg.UserManagement()).UseCases.EnableBootstrapAdmin(time.Now().Add(time.Hour), "%6UbRsCuM8N$auy"))
		cfg.SetDecorator(cfg.NewScaffold().
			NoFooterContentPadding(".").
			Decorator())
		option.Must(cfginspector.Enable(cfg))
		option.Must(cfglocalization.Enable(cfg))
		drives := option.Must(cfgdrive.Enable(cfg))

		option.Must(cfgai.Enable(cfg))

		option.Must(drives.UseCases.OpenRoot(user.SU(), drive.OpenRootOptions{
			Create: true,
			Mode:   drive.OtherWrite | drive.OtherRead,
		}))

		cfg.RootViewWithDecoration(".", func(wnd core.Window) core.View {
			prompt := core.AutoState[string](wnd)
			conv := core.AutoState[conversation.ID](wnd)

			wnd.AddDestroyObserver(events.SubscribeFor[conversation.Updated](cfg.EventBus(), func(evt conversation.Updated) {
				if conv.Get() == evt.Conversation {
					conv.Invalidate()
				}
			}))

			const innerFullHeight = "calc(100vh - 16rem - 1px)"

			return ui.HStack(
				ui.ScrollView(
					uiai.Chats(conv).Frame(ui.Frame{Width: ui.L560, MinWidth: ui.L560}),
				).Axis(ui.ScrollViewAxisVertical).Frame(ui.Frame{Height: ui.Full, Width: ui.Full}),

				ui.ScrollView(
					uiai.Chat(conv, prompt).
						StartOptions(conversation.StartOptions{
							WorkspaceName: "Test",
							AgentName:     "Walter",
							CloudStore:    true,
						}),
				).ScrollToView("start-chat-button", ui.ScrollAnimationSmooth).
					Axis(ui.ScrollViewAxisVertical).Frame(ui.Frame{Height: ui.Full, Width: ui.Full}),
			).Alignment(ui.Top).Frame(ui.Frame{Width: ui.Full, Height: innerFullHeight})

		})
	}).Run()
}
