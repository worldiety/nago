package alert

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"slices"
)

type Message struct {
	Title   string
	Message string
}

// MessageList may return nil, if no information needs to be displayed. Otherwise, it appends to
// the modal overlay.
func MessageList(wnd core.Window) core.View {
	messages := core.TransientStateOf[[]Message](wnd, ".nago-messages")
	if len(messages.Get()) == 0 {
		return nil
	}

	return ui.Overlay(
		ui.ScrollView(
			ui.VStack(
				ui.Each(slices.Values(messages.Get()), func(t Message) core.View {
					presented := core.StateOf[bool](wnd, ".msg-"+t.Title+t.Message).Init(func() bool {
						return true
					})

					return Banner(t.Title, t.Message).
						Closeable(presented).
						OnClosed(func() {
							messages.Set(slices.DeleteFunc(messages.Get(), func(message Message) bool {
								return message == t
							}))
						})
				})...,
			).Gap(ui.L8).Padding(ui.Padding{Right: ui.L16}),
		).
			Frame(ui.Frame{MaxHeight: "calc(100dvh - 8rem)"}),
	).Right(ui.L8).Top(ui.L120)
}

// ShowMessage puts the given msg into the global messages state list.
// Just include [MessageList] always in your view tree, which will overlay the message as required.
func ShowMessage(wnd core.Window, msg Message) {
	messages := core.TransientStateOf[[]Message](wnd, ".nago-messages")
	if slices.Contains(messages.Get(), msg) {
		return
	}

	messages.Set(append(messages.Get(), msg))
}
