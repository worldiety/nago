// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiai

import (
	"fmt"
	"os"

	"github.com/worldiety/i18n"
	"github.com/worldiety/i18n/date"
	"go.wdy.de/nago/application/ai"
	"go.wdy.de/nago/application/ai/document"
	"go.wdy.de/nago/application/ai/library"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/presentation/core"
	icons "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/breadcrumb"
	"go.wdy.de/nago/presentation/ui/markdown"
)

func PageDocument(wnd core.Window, uc ai.UseCases) core.View {
	optProv, err := uc.FindProviderByID(wnd.Subject(), provider.ID(wnd.Values()["provider"]))
	if err != nil {
		return alert.BannerError(err)
	}

	if optProv.IsNone() {
		return alert.BannerError(fmt.Errorf("provider not found: %s: %w", wnd.Values()["provider"], os.ErrNotExist))
	}

	prov := optProv.Unwrap()

	optLibs := prov.Libraries()
	if optLibs.IsNone() {
		return alert.BannerError(fmt.Errorf("provider does not support libraries: %s", wnd.Values()["provider"]))
	}

	libs := optLibs.Unwrap()
	libID := library.ID(wnd.Values()["library"])

	optLibInfo, err := libs.FindByID(wnd.Subject(), libID)
	if err != nil {
		return alert.BannerError(err)
	}

	if optLibInfo.IsNone() {
		return alert.BannerError(fmt.Errorf("library not found: %s: %w", libID, os.ErrNotExist))
	}

	libInfo := optLibInfo.Unwrap()

	docID := document.ID(wnd.Values()["document"])
	optDoc, err := libs.Library(libID).FindByID(wnd.Subject(), docID)
	if err != nil {
		return alert.BannerError(err)
	}

	if optDoc.IsNone() {
		return alert.BannerError(fmt.Errorf("document not found: %s: %w", docID, os.ErrNotExist))
	}

	doc := optDoc.Unwrap()
	var docText string
	if doc.ProcessingStatus == document.ProcessingCompleted {
		optText, err := libs.Library(libID).TextContentByID(wnd.Subject(), docID)
		if err != nil {
			return alert.BannerError(err)
		}

		if optText.IsSome() {
			docText = optText.Unwrap()
		}
	}

	return ui.VStack(
		breadcrumb.Breadcrumbs(
			ui.TertiaryButton(func() {
				wnd.Navigation().BackwardTo("admin/ai/provider", wnd.Values())
			}).Title(StrLibraries.Get(wnd)),
			ui.TertiaryButton(func() {
				wnd.Navigation().BackwardTo("admin/ai/library", wnd.Values())
			}).Title(libInfo.Name),
		),
		ui.H1(StrDocument.Get(wnd)),

		renderDocInfo(wnd, doc, docText),
	).Alignment(ui.Leading).Frame(ui.Frame{}.Larger())
}

func renderDocInfo(wnd core.Window, doc document.Document, text string) core.View {
	return ui.VStack(
		ui.TextField("", doc.Name).Disabled(true).FullWidth(),
		ui.TextField("", doc.Summary).Lines(10).Disabled(true).FullWidth(),
		ui.TextField("", doc.MimeType).Disabled(true).FullWidth(),
		ui.TextField("", i18n.FormatFloat(wnd.Locale(), float64(doc.Size)/1024/1024, 2, "MiB")).Disabled(true).FullWidth(),
		ui.TextField("", doc.Hash).Disabled(true).FullWidth(),
		ui.TextField("", string(doc.ProcessingStatus)).Disabled(true).FullWidth(),
		ui.TextField("", date.Format(wnd.Subject().Language(), date.TimeMinute, doc.CreatedAt.Time(wnd.Location()))).Disabled(true).FullWidth(),
		ui.Space(ui.L48),
		ui.HStack(
			ui.SecondaryButton(func() {
				_ = wnd.Clipboard().SetText(text)
			}).PreIcon(icons.Clipboard),
		).FullWidth().Alignment(ui.Trailing).Visible(len(text) > 0),
		markdown.Render(markdown.Options{RichText: true, Window: wnd, TrimParagraph: true}, []byte(text)),
	).FullWidth().Gap(ui.L16)
}
