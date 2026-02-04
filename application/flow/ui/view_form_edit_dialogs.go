// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiflow

import (
	"strings"

	"go.wdy.de/nago/application/flow"
	"go.wdy.de/nago/pkg/xstrings"
	"go.wdy.de/nago/presentation/core"
	icons "go.wdy.de/nago/presentation/icons/hero/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/form"
)

func (c TFormEditor) enabledFormDialog(presented *core.State[bool]) core.View {
	if !presented.Get() {
		return nil
	}

	code := core.StateOf[string](c.wnd, string(c.selected.Get())+"-enabled-expr-code").Init(func() string {
		view, ok := flow.GetView(c.ws, c.form.ID, c.selected.Get())
		if !ok {
			return ""
		}

		if view, ok := view.(flow.Enabler); ok {
			return string(view.EnabledExpr())
		}

		return ""
	})

	structType, ok := c.ws.Packages.StructTypeByID(c.form.Type())
	var fieldDetails core.View
	if ok {
		fieldDetails = c.viewHelpFieldsExpr(structType, code)
	}

	return alert.Dialog(
		"enabled condition",
		ui.VStack(
			ui.HStack(ui.TertiaryButton(func() {
				code.Set("")
			}).PreIcon(icons.XMark)).Alignment(ui.Trailing).FullWidth(),
			ui.CodeEditor(code.Get()).InputValue(code).Frame(ui.Frame{}.FullWidth()),
			ui.Text("example: field(\"MyFieldName\").Bool() == true").
				Color(ui.ColorIconsMuted).
				Font(ui.BodySmall),
			ui.Space(ui.L16),
			fieldDetails,
		).FullWidth().Alignment(ui.Leading),
		presented,
		alert.Closeable(),
		alert.Larger(),
		alert.Cancel(nil),
		alert.Save(func() (close bool) {
			if err := c.uc.HandleCommand(c.wnd.Subject(), flow.UpdateFormEnableExpr{
				Workspace:  c.ws.Identity(),
				Form:       c.form.Identity(),
				ID:         c.selected.Get(),
				Expression: flow.Expression(code.Get()),
			}); err != nil {
				alert.ShowBannerError(c.wnd, err)
			}

			return true
		}),
	)
}

func (c TFormEditor) conditionalFormDialog(presented *core.State[bool]) core.View {
	if !presented.Get() {
		return nil
	}

	code := core.StateOf[string](c.wnd, string(c.selected.Get())+"-visibility-expr-code").Init(func() string {
		view, ok := flow.GetView(c.ws, c.form.ID, c.selected.Get())
		if !ok {
			return ""
		}
		return string(view.VisibleExpr())
	})

	structType, ok := c.ws.Packages.StructTypeByID(c.form.Type())
	var fieldDetails core.View
	if ok {
		fieldDetails = c.viewHelpFieldsExpr(structType, code)
	}

	return alert.Dialog(
		"conditional visibility",
		ui.VStack(
			ui.HStack(ui.TertiaryButton(func() {
				code.Set("")
			}).PreIcon(icons.XMark)).Alignment(ui.Trailing).FullWidth(),
			ui.CodeEditor(code.Get()).InputValue(code).Frame(ui.Frame{}.FullWidth()),
			ui.Text("example: field(\"MyFieldName\").Bool() == true").
				Color(ui.ColorIconsMuted).
				Font(ui.BodySmall),
			ui.Space(ui.L16),
			fieldDetails,
		).FullWidth().Alignment(ui.Leading),
		presented,
		alert.Closeable(),
		alert.Larger(),
		alert.Cancel(nil),
		alert.Save(func() (close bool) {
			if err := c.uc.HandleCommand(c.wnd.Subject(), flow.UpdateFormVisibleExpr{
				Workspace:  c.ws.Identity(),
				Form:       c.form.Identity(),
				ID:         c.selected.Get(),
				Expression: flow.Expression(code.Get()),
			}); err != nil {
				alert.ShowBannerError(c.wnd, err)
			}

			return true
		}),
	)
}

func (c TFormEditor) formActionDialog(ctx RContext, btn flow.Actionable, presented *core.State[bool]) core.View {
	if !presented.Get() {
		return nil
	}

	code := core.StateOf[string](ctx.Window(), string(btn.Identity())+"-action-expr-code").Init(func() string {
		return string(xstrings.Join(btn.ActionExpr(), "\n"))
	})

	structType, ok := ctx.Workspace().Packages.StructTypeByID(ctx.Form().Type())
	var fieldDetails core.View
	if ok {
		fieldDetails = c.viewHelpFieldsStmt(structType, code)
	}

	expressions := func() []flow.Expression {
		var tmp []flow.Expression
		for _, s := range strings.Split(code.Get(), "\n") {
			tmp = append(tmp, flow.Expression(strings.TrimSpace(s)))
		}

		return tmp
	}

	return alert.Dialog(
		"action expressions",
		ui.VStack(
			ui.HStack(ui.TertiaryButton(func() {
				code.Set("")
			}).PreIcon(icons.XMark)).Alignment(ui.Trailing).FullWidth(),
			ui.CodeEditor(code.Get()).InputValue(code).Frame(ui.Frame{}.FullWidth()),
			ui.Text("example:\nput(\"MyFieldName\",\"my value\")\ndelete(\"OtherFieldName\")\ndeleteWithPrefix(\"MySubFormPrefix\")").
				Color(ui.ColorIconsMuted).
				Font(ui.BodySmall),
			ui.Space(ui.L16),
			fieldDetails,
		).FullWidth().Alignment(ui.Leading),
		presented,
		alert.Closeable(),
		alert.Larger(),
		alert.Cancel(nil),
		alert.Save(func() (close bool) {
			if err := ctx.HandleCommand(flow.UpdateFormActionExpr{
				Workspace: ctx.Workspace().Identity(),
				Form:      ctx.Form().Identity(),
				ID:        btn.Identity(),
				Action:    expressions(),
			}); err != nil {
				alert.ShowBannerError(ctx.Window(), err)
			}

			return true
		}),
	)
}

func (c TFormEditor) deleteFormDialog(presented *core.State[bool]) core.View {
	if !presented.Get() {
		return nil
	}

	return alert.Dialog(
		"Delete form",
		ui.Text("Are you sure you want to delete this form?"),
		presented,
		alert.Closeable(),
		alert.Cancel(nil),
		alert.Delete(func() {
			if err := c.uc.HandleCommand(c.wnd.Subject(), flow.DeleteFormCmd{
				Workspace: c.ws.Identity(),
				ID:        c.form.Identity(),
			}); err != nil {
				alert.ShowBannerError(c.wnd, err)
			}
		}),
	)
}

func (c TFormEditor) deleteViewDialog(presented *core.State[bool]) core.View {
	if !presented.Get() {
		return nil
	}

	return alert.Dialog(
		"Delete",
		ui.Text("Are you sure you want to delete this view?"),
		presented,
		alert.Closeable(),
		alert.Cancel(nil),
		alert.Delete(func() {
			if err := c.uc.HandleCommand(c.wnd.Subject(), flow.DeleteViewCmd{
				Workspace: c.ws.Identity(),
				Form:      c.form.Identity(),
				View:      c.selected.Get(),
			}); err != nil {
				alert.ShowBannerError(c.wnd, err)
			}
		}),
	)
}

func (c TFormEditor) shareFormDialog(presented *core.State[bool]) core.View {
	if !presented.Get() {
		return nil
	}

	state := core.DerivedState[flow.FormShare](presented, "share").Init(func() flow.FormShare {
		optShare, err := c.uc.FindFormShare(c.wnd.Subject(), c.ws.Identity(), c.form.Identity())
		if err != nil {
			alert.ShowBannerError(c.wnd, err)
		}

		return optShare.UnwrapOr(flow.FormShare{})
	})

	errState := core.DerivedState[error](presented, "share-error")

	link := c.opts.ContextPathURI("nago/flow/form/share", core.Values{"share": string(state.Get().ID)})
	return alert.Dialog(
		"Share",
		ui.VStack(
			form.Auto(form.AutoOptions{Errors: errState.Get()}, state),
			ui.IfFunc(state.Get().ID != "", func() core.View {
				return ui.VStack(
					ui.HLine(),
					ui.HStack(
						ui.ImageIcon(icons.Link),
						ui.Text(link).Hyphens(ui.HyphensAuto).LineBreak(true),
						ui.SecondaryButton(func() {
							_ = c.wnd.Clipboard().SetText(link)
						}).Title("Copy link").Frame(ui.Frame{Width: ui.L160}),
					).FullWidth().BackgroundColor(ui.M3).Border(ui.Border{}.Radius(ui.L8)).Padding(ui.Padding{}.All(ui.L8)),
				).FullWidth().Gap(ui.L8)
			}),
			ui.If(state.Get().ID == "", ui.Text("no share active")),
		).FullWidth(),
		presented,
		alert.Larger(),
		alert.Closeable(),
		alert.Cancel(nil),
		alert.Apply(func() (close bool) {
			_, err := c.uc.UpdateFormShare(c.wnd.Subject(), c.ws.Identity(), c.form.Identity(), flow.ShareFormOptions{
				AllowUnauthenticated: state.Get().AllowUnauthenticated,
				AllowedUsers:         state.Get().AllowedUsers,
			})
			if err != nil {
				alert.ShowBannerError(c.wnd, err)
				return false
			}

			return true
		}),
	)
}
