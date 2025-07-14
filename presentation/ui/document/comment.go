// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package document

import (
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/xtime"
	"go.wdy.de/nago/presentation/core"
	flowbiteOutline "go.wdy.de/nago/presentation/icons/flowbite/outline"
	flowbiteSolid "go.wdy.de/nago/presentation/icons/flowbite/solid"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/avatar"
	"time"
)

type Message struct {
	User    user.ID
	Message string
	Time    time.Time
}

type TComment struct {
	view         func(c TComment) core.View
	selected     *core.State[bool]
	comment      *core.State[string]
	needsSepLine bool
	resolve      func()
}

func Thread(wnd core.Window, messages ...Message) TComment {
	displayName, _ := core.FromContext[user.DisplayName](wnd.Context(), "")

	first := true
	return TComment{
		view: func(c TComment) core.View {
			border := ui.Border{}.Width(ui.L1).Color(ui.ColorIconsMuted).Radius(ui.L8)
			if c.selected != nil && c.selected.Get() {
				border = ui.Border{}.Width(ui.L1).Color(ui.ColorInteractive).Radius(ui.L8).Shadow(ui.L8)
			}

			return ui.VStack(
				ui.ForEach(messages, func(msg Message) core.View {
					usr := displayName(msg.User)
					var space ui.Length
					if !first {
						space = ui.L16
					}

					defer func() {
						first = false
					}()

					return ui.VStack(
						ui.HStack(
							avatar.TextOrImage(usr.Displayname, usr.Avatar),
							ui.Text(usr.Displayname).Font(ui.BodyMedium),
							ui.Spacer(),
							ui.IfFunc(c.resolve != nil && first, func() core.View {
								return ui.TertiaryButton(c.resolve).PreIcon(flowbiteOutline.Check).AccessibilityLabel("Als erledigt markieren")
							}),
						).FullWidth().Alignment(ui.Leading).Gap(ui.L8),
						ui.Text(msg.Message).Font(ui.BodySmall),
						ui.Text(msg.Time.Format(xtime.GermanDateTime)).Font(ui.BodySmall).Color(ui.ColorIconsMuted),
					).FullWidth().
						Alignment(ui.Leading).
						Gap(ui.L8).
						Padding(ui.Padding{Left: space})
				})...,
			).Append(func() core.View {
				// this is the comment field
				if c.comment != nil {
					tmp := core.DerivedState[string](c.comment, "tmp").Init(func() string {
						return c.comment.Get()
					})
					return ui.VStack(
						ui.TextField("", tmp.Get()).InputValue(tmp).FullWidth(),
						ui.SecondaryButton(func() {
							c.comment.Set(tmp.Get())
							c.comment.Notify()
							tmp.Set("")
						}).Enabled(tmp.Get() != "").PreIcon(flowbiteOutline.PaperPlane).AccessibilityLabel("Kommentar hinzufügen"),
					).Gap(ui.L8).FullWidth().Alignment(ui.Trailing)
				}

				return nil
			}()).
				Gap(ui.L8).
				Alignment(ui.Leading).
				FullWidth().
				NoClip(true).
				Border(border).
				Padding(ui.Padding{}.All(ui.L8))
		},
	}
}

func LogEntry(message string, who string, when time.Time) TComment {
	return TComment{
		view: func(c TComment) core.View {
			return ui.VStack(
				ui.Text(message),
				ui.Text(who).Font(ui.BodySmall),
				ui.Text(when.Format(xtime.GermanDateTime)).Font(ui.BodySmall),
			).Alignment(ui.Leading)
		},
		needsSepLine: true,
	}
}

func Comment(v core.View) TComment {
	return TComment{
		view: func(c TComment) core.View {
			return c
		},
	}
}

func (c TComment) InputSelectedValue(selected *core.State[bool]) TComment {
	c.selected = selected
	return c
}

func (c TComment) InputValue(comment *core.State[string]) TComment {
	c.comment = comment
	return c
}

func (c TComment) Resolve(action func()) TComment {
	c.resolve = action
	return c
}

func (c TComment) Render(ctx core.RenderContext) core.RenderNode {
	v := ui.HStack(
		c.view(c),
	).FullWidth().Alignment(ui.Leading).NoClip(true)

	if c.selected != nil {
		v = v.Action(func() {
			c.selected.Set(!c.selected.Get())
			c.selected.Notify()
		})

		if c.selected.Get() {
			v = v.FullWidth()
		}
	}

	return v.Render(ctx)
}

func AttachComment(selection *core.State[bool], v ui.DecoredView) ui.DecoredView {
	if selection == nil {
		return v
	}

	if selection.Get() {
		return ui.VStack(
			v,
			ui.VStack(
				ui.ImageIcon(flowbiteSolid.Annotation)).Position(ui.Position{Type: ui.PositionAbsolute, Top: "-1rem", Right: "0rem"}).
				NoClip(true),
		).BackgroundColor(ui.ColorCardFooter).Position(ui.Position{Type: ui.PositionOffset}).NoClip(true)
	} else {
		return v
	}
}

// NewCommentDialog is a real dialog, because we are limited scrollwise. It is not clear in large documents,
// where the new comment field at the side would be created, and it may be entirely off the actual screen.
// This helps a bit to reduce that problem, by showing the dialog always modal in the screen center.
func NewCommentDialog(presented *core.State[bool], addComment func(text string)) core.View {
	wnd := presented.Window()
	displayUser, _ := core.FromContext[user.DisplayName](wnd.Context(), "")
	usr := displayUser(wnd.Subject().ID())

	commentText := core.DerivedState[string](presented, "-cmt-text")
	return alert.Dialog(
		"Kommentar hinzufügen",
		ui.VStack(
			ui.HStack(
				avatar.TextOrImage(usr.Displayname, usr.Avatar),
				ui.Text(usr.Displayname),
			).Gap(ui.L8),
			ui.TextField("", commentText.Get()).InputValue(commentText).FullWidth(),
		).Gap(ui.L8).FullWidth().Alignment(ui.Leading),
		presented,
		alert.Closeable(),
		alert.Custom(func(close func(closeDlg bool)) core.View {
			return ui.PrimaryButton(func() {
				close(true)
				addComment(commentText.Get())
			}).PreIcon(flowbiteOutline.PaperPlane).Enabled(commentText.Get() != "").AccessibilityLabel("Kommentar hinzufügen")
		}),
	)
}
