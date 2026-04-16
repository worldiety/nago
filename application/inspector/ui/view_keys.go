// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiinspector

import (
	"encoding/json"
	"net/url"
	"unicode/utf8"

	"github.com/worldiety/i18n"
	"github.com/worldiety/option"
	"go.wdy.de/nago/application/backup"
	"go.wdy.de/nago/application/inspector"
	"go.wdy.de/nago/application/inspector/rest"
	"go.wdy.de/nago/application/localization/rstring"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/xstrings"
	"go.wdy.de/nago/presentation/core"
	icons "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/dataview"
	"go.wdy.de/nago/presentation/ui/pager"
	"golang.org/x/text/language"
)

var (
	StrInvalidJSON              = i18n.MustString("nago.inspector.invalid_json", i18n.Values{language.German: "Ungültiges JSON", language.English: "Invalid JSON"})
	StrValidJSON                = i18n.MustString("nago.inspector.valid_json", i18n.Values{language.German: "JSON", language.English: "JSON"})
	StrCreateTextEntry          = i18n.MustString("nago.inspector.create_text_entry", i18n.Values{language.German: "Neuer Text-Eintrag", language.English: "New Text Entry"})
	StrImportFromZipFile        = i18n.MustString("nago.inspector.import_from_zip_file", i18n.Values{language.German: "Aus ZIP-Datei importieren", language.English: "Import from ZIP file"})
	StrDownloadAsJSONArray      = i18n.MustString("nago.inspector.download_as_json_array", i18n.Values{language.German: "Als JSON Array herunterladen", language.English: "Download as JSON Array"})
	StrDownloadAsJSONObject     = i18n.MustString("nago.inspector.download_as_json_object", i18n.Values{language.German: "Als JSON Objekt herunterladen", language.English: "Download as JSON Object"})
	StrDownloadAsZipFile        = i18n.MustString("nago.inspector.download_as_zip_file", i18n.Values{language.German: "Als ZIP-Datei herunterladen", language.English: "Download as ZIP file"})
	StrObjectSize               = i18n.MustString("nago.inspector.object_size", i18n.Values{language.German: "Größe", language.English: "Size"})
	StrDownloadAllEntries       = i18n.MustString("nago.inspector.download_all_entries", i18n.Values{language.German: "Alle Einträge herunterladen", language.English: "Download all entries"})
	StrBinaryDataCannotBeEdited = i18n.MustString("nago.inspector.binary_data_cannot_be_edited", i18n.Values{language.German: "Binäre Daten können nicht bearbeitet werden", language.English: "Binary data cannot be edited"})
)

func viewKeys(wnd core.Window, store inspector.Store) core.View {
	if !wnd.Subject().HasPermission(inspector.PermDataInspector) {
		return nil
	}

	if store.Store == nil {
		return nil
	}

	lineHeight := 69.0
	const overheadLines = 4

	invalidate := core.AutoState[int](wnd)
	createNewPresented := core.AutoState[bool](wnd)
	editPresented := core.AutoState[bool](wnd)
	selectedKey := core.AutoState[string](wnd)

	return ui.VStack(
		dialogNewTextEntry(wnd, store, createNewPresented),
		dialogEditEntry(wnd, store, editPresented, selectedKey),
		dataview.FromData(wnd, dataview.Data[inspector.Entry, string]{
			FindAll: func(yield func(string, error) bool) {
				for id, err := range store.Store.List(wnd.Context(), blob.ListOptions{}) {
					if !yield(id, err) {
						return
					}
				}
			},
			FindByID: func(id string) (option.Opt[inspector.Entry], error) {
				var data string
				if store.Stereotype == backup.StereotypeDocument {
					optBuf, _ := blob.Get(store.Store, id)
					if optBuf.IsSome() {
						buf := optBuf.Unwrap()
						if utf8.Valid(buf) {
							data = string(buf)
						}
					}
				}

				optInfo, err := blob.Stat(wnd.Context(), store.Store, id)
				if err != nil {
					return option.None[inspector.Entry](), err
				}

				return option.Some(inspector.Entry{
					Store:      store,
					Key:        id,
					SearchData: data,
					Size:       optInfo.UnwrapOr(blob.Info{Size: -1}).Size,
				}), nil
			},
			Fields: []dataview.Field[inspector.Entry]{
				{
					ID:   "key",
					Name: rstring.LabelName.Get(wnd),
					Map: func(obj inspector.Entry) core.View {
						return ui.Text(obj.Key)
					},
				},
				{
					ID:   "size",
					Name: StrObjectSize.Get(wnd),
					Map: func(obj inspector.Entry) core.View {
						return ui.Text(xstrings.FormatByteSize(wnd.Locale(), obj.Size, 2))
					},
				},
			},
		}).ModelOptions(pager.ModelOptions{
			PageSize: max(1, int((float64(wnd.Info().Height)-(overheadLines*lineHeight)-96)/lineHeight)),
		}).SelectOptions(
			dataview.NewSelectOptionDelete(wnd, func(ids []string) error {
				if !wnd.Subject().HasPermission(inspector.PermDataInspector) {
					return nil
				}

				for _, id := range ids {
					if err := store.Store.Delete(wnd.Context(), id); err != nil {
						return err
					}
				}

				return nil
			}),
			dataview.SelectOption[string]{
				Icon: icons.Download,
				Name: StrDownloadAsJSONArray.Get(wnd),
				Action: func(selected []string) error {
					q := url.Values{}
					q.Set("store", store.Name)
					q.Set("id", rest.EncodeQuery(selected))
					core.HTTPOpen(wnd.Navigation(), rest.PathDownloadAsJSONArray+"?"+core.URI(rest.EncodeQuery(selected)), "")
					return nil
				},
			},
			dataview.SelectOption[string]{
				Name: StrDownloadAsJSONObject.Get(wnd),
				Action: func(selected []string) error {
					q := url.Values{}
					q.Set("store", store.Name)
					q.Set("id", rest.EncodeQuery(selected))
					core.HTTPOpen(wnd.Navigation(), rest.PathDownloadAsJSONObject+"?"+core.URI(q.Encode()), "")
					return nil
				},
			},
			dataview.SelectOption[string]{
				Name: StrDownloadAsZipFile.Get(wnd),
				Action: func(selected []string) error {
					q := url.Values{}
					q.Set("store", store.Name)
					q.Set("id", rest.EncodeQuery(selected))
					core.HTTPOpen(wnd.Navigation(), rest.PathDownloadAsZip+"?"+core.URI(q.Encode()), "")
					return nil
				},
			},
		).CreateOptions(
			dataview.CreateOption{
				Icon: icons.Plus,
				Name: StrCreateTextEntry.Get(wnd),
				Action: func() error {
					createNewPresented.Set(true)
					return nil
				},
			},
			dataview.CreateOption{
				Icon: icons.Upload,
				Name: StrImportFromZipFile.Get(wnd),
				Action: func() error {
					wnd.ImportFiles(core.ImportFilesOptions{
						AllowedMimeTypes: []string{"application/zip"},
						OnCompletion: func(files []core.File) {
							if len(files) == 0 {
								return
							}

							if err := rest.ImportFromZip(store.Store, files[0]); err != nil {
								alert.ShowBannerError(wnd, err)
							}

							invalidate.Invalidate()
						},
					})

					return nil
				},
			},
			dataview.CreateOption{
				Icon: icons.Upload,
				Name: rstring.ActionFileUpload.Get(wnd),
				Action: func() error {
					wnd.ImportFiles(core.ImportFilesOptions{
						Multiple: true,
						OnCompletion: func(files []core.File) {
							for _, f := range files {
								r, err := f.Open()
								if err != nil {
									alert.ShowBannerError(wnd, err)
									return
								}

								if _, err = blob.Write(store.Store, f.Name(), r); err != nil {
									alert.ShowBannerError(wnd, err)
									return
								}

								if err := r.Close(); err != nil {
									alert.ShowBannerError(wnd, err)
									return
								}

							}

							invalidate.Invalidate()
						},
					})

					return nil
				},
			},
		).
			Action(func(e inspector.Entry) {
				selectedKey.Set(e.Key)
				editPresented.Set(true)
			}).
			Search(true).
			NextActionIndicator(true),
	).FullWidth()
}

func dialogNewTextEntry(wnd core.Window, store inspector.Store, presented *core.State[bool]) core.View {
	if !presented.Get() {
		return nil
	}

	key := core.AutoState[string](wnd)
	value := core.AutoState[string](wnd)

	return alert.Dialog(
		StrCreateTextEntry.Get(wnd),
		ui.VStack(
			ui.TextField("Key", key.Get()).InputValue(key).FullWidth(),
			ui.CodeEditor(value.Get()).Language("json").InputValue(value).Frame(ui.Frame{Height: ui.L320}.FullWidth()),
		).FullWidth().Gap(ui.L16),
		presented,
		alert.Closeable(),
		alert.Cancel(nil),
		alert.Create(func() (close bool) {
			if err := blob.Put(store.Store, key.Get(), []byte(value.Get())); err != nil {
				alert.ShowBannerError(wnd, err)
				return false
			}

			return true
		}),
	)
}

func dialogEditEntry(wnd core.Window, store inspector.Store, presented *core.State[bool], key *core.State[string]) core.View {
	if !presented.Get() {
		return nil
	}

	optStat, _ := blob.Stat(wnd.Context(), store.Store, key.Get())
	validUtf8 := true
	validJson := false
	textValue := core.AutoState[string](wnd).Init(func() string {
		if optStat.IsNone() {
			return ""
		}

		if optStat.Unwrap().Size < 1024*1024*1 {
			buf, _ := blob.Get(store.Store, key.Get())
			validUtf8 = utf8.Valid(buf.Unwrap())
			validJson = validJSON(string(buf.Unwrap()))

			if buf.IsSome() && validUtf8 {
				return format(buf.Unwrap())
			}
		}

		return ""
	}).Observe(func(newValue string) {
		validJson = validJSON(newValue)
	})

	q := url.Values{}
	q.Set("store", store.Name)
	q.Set("id", rest.EncodeQuery([]string{key.Get()}))
	downloadUrl := rest.PathDownloadAsRaw + "?" + core.URI(q.Encode())

	return alert.Dialog(
		rstring.ActionEdit.Get(wnd),
		ui.VStack(
			ui.HStack(
				ui.TertiaryButton(nil).HRef(downloadUrl).PreIcon(icons.Download).AccessibilityLabel(rstring.ActionDownload.Get(wnd)).Target("_blank"),
				ui.IfFunc(validUtf8, func() core.View {
					if validJson {
						return ui.ImageIcon(icons.Check).AccessibilityLabel(StrValidJSON.Get(wnd))
					}

					return ui.ImageIcon(icons.ExclamationCircle).AccessibilityLabel(StrInvalidJSON.Get(wnd))
				}),
			).FullWidth().Alignment(ui.Trailing),
			ui.If(validUtf8, ui.CodeEditor(textValue.Get()).Language("json").InputValue(textValue).FullWidth()),
			ui.If(!validUtf8, ui.Text(StrBinaryDataCannotBeEdited.Get(wnd))),
		).FullWidth().Gap(ui.L16),
		presented,
		alert.Closeable(),
		alert.Larger(),
		alert.Cancel(nil),
		alert.Save(func() (close bool) {
			if err := blob.Put(store.Store, key.Get(), []byte(textValue.Get())); err != nil {
				alert.ShowBannerError(wnd, err)
				return false
			}

			return true
		}),
	)
}

func validJSON(text string) bool {
	var obj any
	if err := json.Unmarshal([]byte(text), &obj); err != nil {
		return false
	}

	return true
}

func format(buf []byte) string {
	var obj any
	if err := json.Unmarshal(buf, &obj); err != nil {
		return string(buf)
	}

	return string(option.Must(json.MarshalIndent(obj, "", "  ")))
}
