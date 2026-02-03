// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiflow

import (
	"fmt"
	"os"
	"reflect"

	"go.wdy.de/nago/application/flow"
	"go.wdy.de/nago/presentation/core"
	icons "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/treeview"
)

const (
	formEditorFlagInsertMode    = "form-editor-insertMode"
	formEditorFlagShowInspector = "form-editor-showInspector"
	formEditorFlagTest          = "form-editor-test"
)

type PageEditorOptions struct {
	UseCases  flow.UseCases
	Renderers map[reflect.Type]ViewRenderer
}

func PageEditor(wnd core.Window, opts PageEditorOptions) core.View {
	uc := opts.UseCases
	optWs, err := uc.LoadWorkspace(wnd.Subject(), flow.WorkspaceID(wnd.Values()["workspace"]))
	if err != nil {
		return alert.BannerError(err)
	}

	if optWs.IsNone() {
		return alert.BannerError(fmt.Errorf("workspace not found: %s: %w", wnd.Values()["workspace"], os.ErrNotExist))
	}

	ws := optWs.Unwrap()

	tree := createTree(ws)

	presentCreatePackage := core.AutoState[bool](wnd)
	presentCreateStringType := core.AutoState[bool](wnd)
	presentCreateStructType := core.AutoState[bool](wnd)

	presentAssignRepository := core.AutoState[bool](wnd)

	presentCreateForm := core.AutoState[bool](wnd)
	treeState := core.AutoState[treeview.TreeStateModel[string]](wnd)

	toolbarFormOptions := core.AutoState[bool](wnd)
	formEditorFlagInsertModeState := core.StateOf[bool](wnd, formEditorFlagInsertMode)
	formEditorFlagShowInspectorState := core.StateOf[bool](wnd, formEditorFlagShowInspector)
	formEditorFlagTestState := core.StateOf[bool](wnd, formEditorFlagTest)

	return ui.VStack(

		dialogCmd(wnd, ws, "Create new package", presentCreatePackage, uc.HandleCommand, func() flow.WorkspaceCommand {
			return flow.CreatePackageCmd{
				Workspace: ws.Identity(),
			}
		}),
		dialogCmd(wnd, ws, "Create new string type", presentCreateStringType, uc.HandleCommand, func() flow.WorkspaceCommand {
			return flow.CreateStringTypeCmd{
				Workspace: ws.Identity(),
			}
		}),
		dialogCmd(wnd, ws, "Create new struct type", presentCreateStructType, uc.HandleCommand, func() flow.WorkspaceCommand {
			return flow.CreateStructTypeCmd{
				Workspace: ws.Identity(),
			}
		}),

		dialogCmd(wnd, ws, "Assign Repository", presentAssignRepository, uc.HandleCommand, func() flow.WorkspaceCommand {
			return flow.AssignRepositoryCmd{
				Workspace: ws.Identity(),
			}
		}),

		dialogCmd(wnd, ws, "Create new Form", presentCreateForm, uc.HandleCommand, func() flow.WorkspaceCommand {
			return flow.CreateFormCmd{
				Workspace: ws.Identity(),
			}
		}),

		ui.HStack(
			ui.Text(string(ws.Name)),
			ui.Spacer(),

			ui.IfFunc(toolbarFormOptions.Get(), func() core.View {
				return ui.HStack(
					ui.Menu(ui.TertiaryButton(nil).PreIcon(icons.Eye).AccessibilityLabel("View Form Options"),
						ui.MenuGroup(
							ui.MenuItem(func() {
								formEditorFlagShowInspectorState.Set(!formEditorFlagShowInspectorState.Get())
								formEditorFlagTestState.Set(false)
							}, ui.HStack(ui.Checkbox(formEditorFlagShowInspectorState.Get()), ui.Text("Show Inspector")),
							),

							ui.MenuItem(func() {
								formEditorFlagInsertModeState.Set(!formEditorFlagInsertModeState.Get())
								formEditorFlagTestState.Set(false)
							}, ui.HStack(ui.Checkbox(formEditorFlagInsertModeState.Get()), ui.Text("Insert mode")),
							),
							ui.MenuItem(func() {
								formEditorFlagInsertModeState.Set(false)
								formEditorFlagShowInspectorState.Set(false)
								formEditorFlagTestState.Set(!formEditorFlagTestState.Get())
							}, ui.HStack(ui.Checkbox(formEditorFlagTestState.Get()), ui.Text("Test mode")),
							),
						),
					),
					ui.VLine().Frame(ui.Frame{Height: ui.L16}),
				)
			}),

			ui.TertiaryButton(func() {
				presentCreatePackage.Set(true)
			}).PreIcon(icons.Folder).AccessibilityLabel(StrActionCreatePackage.Get(wnd)),

			ui.Menu(
				ui.TertiaryButton(nil).PreIcon(icons.TableRow).AccessibilityLabel(StrActionCreateType.Get(wnd)),
				ui.MenuGroup(
					ui.MenuItem(func() {
						presentCreateStringType.Set(true)
					}, ui.Text("Create String Type")),

					ui.MenuItem(func() {
						presentCreateStructType.Set(true)
					}, ui.Text("Create Struct Type")),
				),
			),

			ui.TertiaryButton(func() {
				presentAssignRepository.Set(true)
			}).PreIcon(icons.Database).AccessibilityLabel(StrActionCreateRepository.Get(wnd)),

			ui.TertiaryButton(func() {
				presentCreateForm.Set(true)
			}).PreIcon(icons.CreditCardPlus).AccessibilityLabel(StrActionCreateForm.Get(wnd)),
		).FullWidth().Border(ui.Border{BottomColor: ui.ColorIconsMuted, BottomWidth: ui.L1}).Padding(ui.Padding{}.Vertical(ui.L2)),

		ui.HStack(
			ui.ScrollView(
				treeview.TreeView(tree, treeState)).Axis(ui.ScrollViewAxisBoth).Frame(ui.Frame{Width: ui.L200, MaxWidth: ui.L200}),
			ui.VLine().Frame(ui.Frame{}),
			ui.VStack(
				renderSelected(wnd, opts, ws, treeState, toolbarFormOptions),
			).Alignment(ui.Top).
				BackgroundColor(ui.ColorBackground).FullWidth(),
		).FullWidth().Alignment(ui.Stretch),
	).FullWidth().Alignment(ui.Leading)
}

func renderSelected(wnd core.Window, opts PageEditorOptions, ws *flow.Workspace, treeState *core.State[treeview.TreeStateModel[string]], toolbarFormOptions *core.State[bool]) core.View {
	selID, ok := treeState.Get().FirstSelected()
	if !ok {
		return ui.Text("Nothing selected")
	}

	selected, _ := ws.ByID(selID)
	if selected == nil {
		return ui.Text("Nothing selected")
	}

	switch t := selected.(type) {
	case *flow.StringType:
		toolbarFormOptions.Set(false)
		return viewTypeString(wnd, opts.UseCases, ws, t)
	case *flow.StructType:
		toolbarFormOptions.Set(false)
		return viewTypeStruct(wnd, opts.UseCases, ws, t)
	case *flow.Form:
		toolbarFormOptions.Set(true)
		return viewTypeForm(wnd, opts, ws, t)
	default:
		toolbarFormOptions.Set(false)
		return ui.Text(fmt.Sprintf("%T", selected))
	}

}
