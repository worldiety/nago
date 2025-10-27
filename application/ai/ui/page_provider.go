// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiai

import (
	"fmt"
	"iter"
	"os"

	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application/ai"
	"go.wdy.de/nago/application/ai/agent"
	"go.wdy.de/nago/application/ai/conversation"
	"go.wdy.de/nago/application/ai/library"
	"go.wdy.de/nago/application/ai/model"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/application/localization/rstring"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/xslices"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/dataview"
	"go.wdy.de/nago/presentation/ui/form"
	"golang.org/x/text/language"
)

var (
	StrConversations = i18n.MustString("nago.ai.admin.conversations", i18n.Values{language.English: "Conversations", language.German: "Unterhaltungen"})
	StrLibraries     = i18n.MustString("nago.ai.admin.libraries", i18n.Values{language.English: "Libraries", language.German: "Bibliotheken"})
	StrLibrary       = i18n.MustString("nago.ai.admin.library", i18n.Values{language.English: "Library", language.German: "Bibliothek"})
	StrAgents        = i18n.MustString("nago.ai.admin.agents", i18n.Values{language.English: "Agents", language.German: "Agenten"})
)

func PageProvider(wnd core.Window, uc ai.UseCases) core.View {
	optProv, err := uc.FindProviderByID(wnd.Subject(), provider.ID(wnd.Values()["provider"]))
	if err != nil {
		return alert.BannerError(err)
	}

	if optProv.IsNone() {
		return alert.BannerError(fmt.Errorf("provider not found: %s: %w", wnd.Values()["provider"], os.ErrNotExist))
	}

	prov := optProv.Unwrap()

	return ui.VStack(
		ui.H1(prov.Name()),
		ui.Text(prov.Description()),
		ui.Space(ui.L48),
		agentTable(wnd, prov),
		ui.Space(ui.L48),
		libTable(wnd, prov),
		ui.Space(ui.L48),
		libConversations(wnd, prov),
	).FullWidth().Alignment(ui.Leading)
}

func agentTable(wnd core.Window, prov provider.Provider) core.View {
	optAgents := prov.Agents()
	if optAgents.IsNone() {
		return nil
	}

	agents := optAgents.Unwrap()

	loadedAgents := core.AutoState[[]agent.Agent](wnd).AsyncInit(func() []agent.Agent {
		v, err := xslices.Collect2(agents.All(wnd.Subject()))
		if err != nil {
			alert.ShowBannerError(wnd, err)
			return nil
		}

		return v
	})

	createPresented := core.AutoState[bool](wnd)

	return ui.VStack(
		dialogCreateAgent(wnd, prov, agents, createPresented, loadedAgents),
		ui.H2(StrAgents.Get(wnd)),
		ui.If(!loadedAgents.Valid(), ui.Text(rstring.LabelPleaseWait.Get(wnd))),

		ui.IfFunc(loadedAgents.Valid(), func() core.View {
			return dataview.FromSlice(wnd, loadedAgents.Get(), []dataview.Field[dataview.Element[agent.Agent]]{
				{
					Name: rstring.LabelName.Get(wnd),
					Map: func(obj dataview.Element[agent.Agent]) core.View {
						return ui.Text(obj.Value.Name)
					},
				},

				{
					Name: rstring.LabelDescription.Get(wnd),
					Map: func(obj dataview.Element[agent.Agent]) core.View {
						return ui.Text(obj.Value.Description)
					},
				},
			}).NextActionIndicator(true).
				NewAction(func() {
					createPresented.Set(true)
				}).SelectOptions(
				dataview.NewSelectOptionDelete(wnd, func(selected []dataview.Idx) error {
					for _, i := range selected {
						if idx, ok := i.Int(); ok {
							if err := agents.Delete(wnd.Subject(), loadedAgents.Get()[idx].ID); err != nil {
								return err
							}
						}
					}

					loadedAgents.Reset()

					return nil
				}),
			)
		}),
	).FullWidth().Alignment(ui.Leading)

}

func dialogCreateAgent(wnd core.Window, prov provider.Provider, ags provider.Agents, presented *core.State[bool], loadedAgents *core.State[[]agent.Agent]) core.View {
	models := core.AutoState[[]model.Model](wnd).AsyncInit(func() []model.Model {
		models, err := xslices.Collect2(prov.Models().All(wnd.Subject()))
		if err != nil {
			alert.ShowBannerError(wnd, err)
			return nil
		}

		return models
	})

	if !presented.Get() {
		return nil
	}

	ctx := core.WithContext(wnd.Context(), core.ContextValue("nago.ai.models", form.AnyUseCaseList(func(subject auth.Subject) iter.Seq2[model.Model, error] {
		return xslices.ValuesWithError(models.Get(), nil)
	})))

	type CreateForm struct {
		Name         string   `label:"nago.common.label.name"`
		Description  string   `label:"nago.common.label.description" lines:"3"`
		Model        model.ID `source:"nago.ai.models" dialogOptions:"larger"`
		Instructions string   `lines:"5"`
	}

	viewForm := core.AutoState[CreateForm](wnd)
	errModel := core.AutoState[error](wnd)

	return alert.Dialog(
		rstring.ActionNew.Get(wnd),
		form.Auto(form.AutoOptions{Window: wnd, Context: ctx, Errors: errModel.Get()}, viewForm),
		presented,
		alert.Closeable(),
		alert.Cancel(nil),
		alert.Large(),
		alert.Create(func() (close bool) {
			_, err := ags.Create(wnd.Subject(), agent.CreateOptions{
				Name:         viewForm.Get().Name,
				Description:  viewForm.Get().Description,
				Model:        viewForm.Get().Model,
				Instructions: viewForm.Get().Instructions,
			})

			loadedAgents.Reset()
			errModel.Set(err)
			return err == nil
		}),
	)
}

func libTable(wnd core.Window, prov provider.Provider) core.View {
	optLibs := prov.Libraries()
	if optLibs.IsNone() {
		return nil
	}

	libs := optLibs.Unwrap()

	loadedLibs := core.AutoState[[]library.Library](wnd).AsyncInit(func() []library.Library {
		v, err := xslices.Collect2(libs.All(wnd.Subject()))
		if err != nil {
			alert.ShowBannerError(wnd, err)
			return nil
		}

		return v
	})

	createPresented := core.AutoState[bool](wnd)

	return ui.VStack(
		ui.H2(StrLibraries.Get(wnd)),
		ui.If(!loadedLibs.Valid(), ui.Text(rstring.LabelPleaseWait.Get(wnd))),
		dialogNewLibrary(wnd, libs, createPresented, loadedLibs),

		ui.IfFunc(loadedLibs.Valid(), func() core.View {
			return dataview.FromSlice(wnd, loadedLibs.Get(), []dataview.Field[dataview.Element[library.Library]]{
				{
					Name: rstring.LabelName.Get(wnd),
					Map: func(obj dataview.Element[library.Library]) core.View {
						return ui.Text(obj.Value.Name)
					},
				},
				{
					Name: rstring.LabelDescription.Get(wnd),
					Map: func(obj dataview.Element[library.Library]) core.View {
						return ui.Text(obj.Value.Description)
					},
				},
			}).
				Action(func(e dataview.Element[library.Library]) {
					wnd.Navigation().ForwardTo("admin/ai/library", wnd.Values().Put("library", string(e.Value.ID)))
				}).
				NextActionIndicator(true).
				NewAction(func() {
					createPresented.Set(true)
				}).SelectOptions(
				dataview.NewSelectOptionDelete(wnd, func(selected []dataview.Idx) error {
					for _, i := range selected {
						if idx, ok := i.Int(); ok {
							if err := libs.Delete(wnd.Subject(), loadedLibs.Get()[idx].ID); err != nil {
								return err
							}
						}
					}

					loadedLibs.Reset()

					return nil
				}),
			)
		}),
	).FullWidth().Alignment(ui.Leading)

}

func dialogNewLibrary(wnd core.Window, libs provider.Libraries, presented *core.State[bool], loadedLibs *core.State[[]library.Library]) core.View {
	if !presented.Get() {
		return nil
	}

	type CreateForm struct {
		Name        string `label:"nago.common.label.name"`
		Description string `label:"nago.common.label.description" lines:"3"`
	}

	model := core.AutoState[CreateForm](wnd)
	errModel := core.AutoState[error](wnd)

	return alert.Dialog(
		rstring.ActionNew.Get(wnd),
		form.Auto(form.AutoOptions{Window: wnd, Errors: errModel.Get()}, model),
		presented,
		alert.Closeable(),
		alert.Cancel(nil),
		alert.Create(func() (close bool) {
			_, err := libs.Create(wnd.Subject(), library.CreateOptions{
				Name:        model.Get().Name,
				Description: model.Get().Description,
			})

			errModel.Set(err)
			loadedLibs.Reset()

			return err == nil
		}),
	)
}

func libConversations(wnd core.Window, prov provider.Provider) core.View {
	optConvs := prov.Conversations()
	if optConvs.IsNone() {
		return nil
	}

	convs := optConvs.Unwrap()

	loadedConvs := core.AutoState[[]conversation.Conversation](wnd).AsyncInit(func() []conversation.Conversation {
		v, err := xslices.Collect2(convs.All(wnd.Subject()))
		if err != nil {
			alert.ShowBannerError(wnd, err)
			return nil
		}

		return v
	})

	createPresented := core.AutoState[bool](wnd)

	return ui.VStack(
		ui.H2(StrConversations.Get(wnd)),
		ui.If(!loadedConvs.Valid(), ui.Text(rstring.LabelPleaseWait.Get(wnd))),

		ui.IfFunc(loadedConvs.Valid(), func() core.View {
			return dataview.FromSlice(wnd, loadedConvs.Get(), []dataview.Field[dataview.Element[conversation.Conversation]]{
				{
					Name: rstring.LabelName.Get(wnd),
					Map: func(obj dataview.Element[conversation.Conversation]) core.View {
						return ui.Text(obj.Value.Name)
					},
				},
				{
					Name: rstring.LabelDescription.Get(wnd),
					Map: func(obj dataview.Element[conversation.Conversation]) core.View {
						return ui.Text(obj.Value.Description)
					},
				},
			}).
				Action(func(e dataview.Element[conversation.Conversation]) {
					wnd.Navigation().ForwardTo("admin/ai/provider/conversation", wnd.Values().Put("conversation", string(e.Value.ID)))
				}).
				NextActionIndicator(true).
				NewAction(func() {
					createPresented.Set(true)
				}).SelectOptions(
				dataview.NewSelectOptionDelete(wnd, func(selected []dataview.Idx) error {
					for _, i := range selected {
						if idx, ok := i.Int(); ok {
							if err := convs.Delete(wnd.Subject(), loadedConvs.Get()[idx].ID); err != nil {
								return err
							}
						}
					}

					loadedConvs.Reset()

					return nil
				}),
			)
		}),
	).FullWidth().Alignment(ui.Leading)

}
