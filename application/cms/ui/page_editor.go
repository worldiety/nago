// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uicms

import (
	"go.wdy.de/nago/application/cms"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/presentation/core"
	flowbiteOutline "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/editor"
	"go.wdy.de/nago/presentation/ui/form"
	"golang.org/x/text/language"
	"iter"
	"os"
)

func PageEditor(wnd core.Window, uc cms.UseCases) core.View {
	pagePickerPresented := core.AutoState[bool](wnd)
	componentPickerPresented := core.AutoState[bool](wnd)
	showRightToolbar := core.AutoState[bool](wnd)

	createPagePresented := core.AutoState[bool](wnd)
	editDocDataPresented := core.AutoState[bool](wnd)

	selectedUpdateData := core.AutoState[cms.CreationData](wnd)
	selectedDoc := core.AutoState[*cms.Document](wnd).Observe(func(newValue *cms.Document) {
		selectedUpdateData.Set(cms.CreationData{
			Title:     newValue.Title.String(),
			Slug:      newValue.Slug,
			Published: newValue.Published,
		})
	})

	if !wnd.Subject().Valid() {
		return alert.BannerError(user.InvalidSubjectErr)
	}

	return editor.Screen("Seiten").
		Header(editor.Header(wnd).
			Leading(ui.Text("")).
			Center(ui.Text("Seite bearbeiten")),
		).
		LeadingToolWindows(
			// the page picker
			editor.ToolWindowList(
				wnd,
				editor.ToolWindowListConfig[*cms.Document, cms.ID]{
					Name:   "Seiten",
					List:   uc.FindAll(wnd.Subject()),
					Delete: uc.Delete,
					OnSelected: func(document *cms.Document) {
						selectedDoc.Set(document)
						selectedDoc.Notify()
					},
					OnOptions: func(document *cms.Document) {
						editDocDataPresented.Set(true)
					},
					CreateEmpty: func(subject auth.Subject) error {
						createPagePresented.Set(true)
						return nil
					},
				},
			).Visible(pagePickerPresented.Get()),

			// component picker
			editor.ToolWindowList(
				wnd,
				editor.ToolWindowListConfig[ComponentAction, ComponentAction]{
					Name: "Komponenten",
					ListIcon: func(action ComponentAction) core.SVG {
						switch action {
						case AddRichText:
							return flowbiteOutline.TextSize
						default:
							return flowbiteOutline.ObjectsColumn
						}
					},
					List: ComponentActionIter2(),
					OnAddToContent: func(add ComponentAction) {
						err := uc.AppendElement(wnd.Subject(), selectedDoc.Get().ID, selectedDoc.Get().Body.ID, &cms.RichText{
							Text: map[language.Tag]string{wnd.Locale(): "<h2>Meine Seite</h2><p>Mein Seiteninhalt.</p>"},
						})

						if err != nil {
							alert.ShowBannerError(wnd, err)
							return
						}

						optDoc, err := uc.FindByID(wnd.Subject(), selectedDoc.Get().ID)
						if err != nil {
							alert.ShowBannerError(wnd, err)
							return
						}

						if optDoc.IsNone() {
							alert.ShowBannerError(wnd, os.ErrNotExist)
							return
						}

						selectedDoc.Set(optDoc.Unwrap())
					},
				},
			).Visible(componentPickerPresented.Get()),
		).
		Content(editor.Content(
			RenderEditor(selectedDoc, func(elem cms.Element) {
				if err := uc.ReplaceElement(wnd.Subject(), selectedDoc.Get().ID, elem); err != nil {
					alert.ShowBannerError(wnd, err)
					return
				}

			}),
		).Style(editor.ContentPage)).
		Navbar(editor.Navbar().Top(
			ui.TertiaryButton(func() {
				pagePickerPresented.Set(!pagePickerPresented.Get())
				componentPickerPresented.Set(false)
			}).PreIcon(flowbiteOutline.File),
			ui.TertiaryButton(func() {
				componentPickerPresented.Set(!showRightToolbar.Get())
				pagePickerPresented.Set(false)
			}).PreIcon(flowbiteOutline.Plus).Enabled(selectedDoc.Get() != nil),
		)).
		Modals(
			form.DialogCreate[cms.CreationData](wnd, "Seite erstellen", createPagePresented, func(subject auth.Subject, value cms.CreationData) error {
				_, err := uc.Create(subject, value)
				return err
			}),

			form.DialogEdit[cms.CreationData](wnd, "Seite bearbeiten", editDocDataPresented, selectedUpdateData, func(subject auth.Subject) error {
				if err := uc.UpdateTitle(subject, selectedDoc.Get().ID, wnd.Locale(), selectedUpdateData.Get().Title); err != nil {
					return err
				}

				if err := uc.UpdateSlug(subject, selectedDoc.Get().ID, selectedUpdateData.Get().Slug); err != nil {
					return err
				}

				if err := uc.UpdatePublished(subject, selectedDoc.Get().ID, selectedUpdateData.Get().Published); err != nil {
					return err
				}

				return nil
			}),
		)
}

type ComponentAction string

func (c ComponentAction) String() string {
	switch c {
	case AddRichText:
		return "RichText"
	default:
		return string(c)
	}
}

func (c ComponentAction) Identity() ComponentAction {
	return c
}

const (
	AddRichText ComponentAction = "AddRichText"
)

func ComponentActionIter2() iter.Seq2[ComponentAction, error] {
	return func(yield func(ComponentAction, error) bool) {
		if !yield(AddRichText, nil) {
			return
		}
	}
}
