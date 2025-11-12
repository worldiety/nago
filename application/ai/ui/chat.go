// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiai

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/worldiety/i18n"
	"github.com/worldiety/option"
	"go.wdy.de/nago/application/ai/agent"
	"go.wdy.de/nago/application/ai/conversation"
	"go.wdy.de/nago/application/ai/file"
	"go.wdy.de/nago/application/ai/message"
	"go.wdy.de/nago/application/ai/model"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/application/localization/rstring"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/xslices"
	"go.wdy.de/nago/pkg/xsync"
	"go.wdy.de/nago/presentation/core"
	icons "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"golang.org/x/text/language"
)

var (
	StrAIDisclaimer = i18n.MustString("nago.ai.chat.disclaimer", i18n.Values{language.English: "Note: AI-generated content may contain errors. Please check your results.", language.German: "Hinweis: KI generierte Inhalte können fehlerhaft sein. Bitte überprüfen Sie Ihre Ergebnisse."})
)

type StartOptions struct {
	Agent      agent.ID
	AgentName  string   // alternative to Agent ID find the first agent with the given name
	Model      model.ID // alternative to Agent
	ModelName  string   // alternative to model.ID
	CloudStore bool
}

// resolve applies a bunch of search heuristics based on the developers (or users) parameters regarding
// various convenience combinations of agent and model names or identifiers.
func (opts StartOptions) resolve(subject auth.Subject, prov provider.Provider) (agent.ID, model.ID, error) {
	if opts.Agent != "" {
		return opts.Agent, "", nil
	}

	if opts.Model != "" {
		return "", opts.Model, nil
	}

	if opts.AgentName != "" && prov.Agents().IsNone() {
		return "", "", fmt.Errorf("agent name is required but provider has no agent support")
	}

	if opts.AgentName != "" {
		opt, err := xslices.Collect2(prov.Agents().Unwrap().FindByName(subject, opts.AgentName))
		if err != nil {
			return "", "", err
		}

		if len(opt) > 0 {
			return opt[0].ID, "", nil
		}

		return "", "", fmt.Errorf("agent by name not found %s: %w", opts.AgentName, os.ErrNotExist)
	}

	if opts.ModelName != "" {
		for m, err := range prov.Models().All(subject) {
			if err != nil {
				return "", "", err
			}

			if strings.ToLower(m.Name) == strings.ToLower(opts.ModelName) {
				return "", m.ID, nil
			}
		}

		return "", "", fmt.Errorf("model by name not found %s: %w", opts.ModelName, os.ErrNotExist)
	}

	// if nothing defined, just pick some agent first
	for ag, err := range prov.Agents().Unwrap().All(subject) {
		if err != nil {
			return "", "", err
		}

		return ag.ID, "", nil
	}

	// if no agent is defined, just pick some model
	for m, err := range prov.Models().All(subject) {
		if err != nil {
			return "", "", err
		}

		return "", m.ID, nil
	}

	return "", "", fmt.Errorf("neither a model nor an agent found to start a conversation")
}

type TChat struct {
	conv         *core.State[conversation.ID]
	text         *core.State[string]
	padding      ui.Padding
	provider     provider.Provider
	startOptions StartOptions
	frame        ui.Frame
	teaser       core.View
	more         core.View
}

func Chat(provider provider.Provider, conv *core.State[conversation.ID], text *core.State[string]) TChat {
	return TChat{conv: conv, text: text, provider: provider}
}

func (c TChat) Padding(padding ui.Padding) TChat {
	c.padding = padding
	return c
}

func (c TChat) Frame(frame ui.Frame) TChat {
	c.frame = frame
	return c
}

func (c TChat) StartOptions(opts StartOptions) TChat {
	c.startOptions = opts
	return c
}

// More is placed in-line or beneath the chat field, where the user enters his prompt.
func (c TChat) More(view core.View) TChat {
	c.more = view
	return c
}

// Teaser view is shown in empty chats above the input text
func (c TChat) Teaser(teaser core.View) TChat {
	c.teaser = teaser
	return c
}

func (c TChat) Render(ctx core.RenderContext) core.RenderNode {
	wnd := ctx.Window()

	if c.provider.Conversations().IsNone() {
		return alert.BannerError(fmt.Errorf("provider has no conversation support id: %s: %w", c.provider, os.ErrNotExist)).Render(ctx)
	}

	conversations := c.provider.Conversations().Unwrap()

	pendingUploadFiles := core.DerivedState[[]file.File](c.conv, "-upload-files")

	messages := core.DerivedState[[]message.Message](c.conv, "-messages").Init(func() []message.Message {
		if c.conv.Get() != "" {
			optConv, err := conversations.FindByID(wnd.Subject(), c.conv.Get())
			if err != nil {
				alert.ShowBannerError(wnd, fmt.Errorf("cannot find conversation %s: %w", c.conv.Get(), err))
				return nil
			}

			if optConv.IsNone() {
				alert.ShowBannerError(wnd, fmt.Errorf("cannot find conversation by id: %s: %w", c.provider, os.ErrNotExist))
				return nil
			}
			conv := optConv.Unwrap()
			msg, err := xslices.Collect2(conversations.Conversation(wnd.Subject(), conv.ID).All(wnd.Subject()))
			if err != nil {
				alert.ShowBannerError(wnd, fmt.Errorf("cannot collect messages from conv %s: %w", conv.ID, err))
				return nil
			}

			return msg
		}

		return nil
	})

	c.conv.Observe(func(newValue conversation.ID) {
		messages.Reset()
		c.text.Set("")
	})

	pleaseWaitPresented := core.DerivedState[bool](c.conv, "-pw-presented")

	return ui.VStack(
		// the actual messages
		ui.ForEach(messages.Get(), func(msg message.Message) core.View {

			stack := ui.VStack()
			style := MessageAgent
			align := ui.Leading
			if msg.Role == message.User {
				style = MessageHuman
				align = ui.Trailing
			}

			switch {
			case msg.MessageInput.IsSome():
				stack = stack.Append(ChatMessage().Markdown(msg.MessageInput.Unwrap()).Style(style))
			case msg.MessageOutput.IsSome():
				stack = stack.Append(ChatMessage().Markdown(msg.MessageOutput.Unwrap()).Style(style))
			case msg.File.IsSome():
				f := msg.File.Unwrap()
				switch f.MimeType {
				case file.PNG:
					fallthrough
				case file.JPEG:
					fallthrough
				case file.GIF:
					stack = stack.Append(ChatMessage().Image(f).Provider(c.provider))
				default:
					stack = stack.Append(ChatMessage().File(f).Provider(c.provider))
				}

			}

			return stack.Alignment(align).FullWidth()
		})...,
	).
		// greeting view if no messages are available
		Append(
			ui.If(len(messages.Get()) == 0, c.teaser),
		).
		// wait entertainment
		Append(
			ui.If(pleaseWaitPresented.Get(),
				ui.HStack(
					ChatMessage().Icon(icons.DotsHorizontal).Style(MessageAgent),
				).FullWidth().Alignment(ui.Leading),
			),
		).
		// area where uploaded but not yet sent files can be seen
		Append(
			ChatUploads(pendingUploadFiles),
		).
		// chat field
		Append(
			ui.Stack(
				ChatField(c.text).
					Enabled(!pleaseWaitPresented.Get()).
					Action(func() {
						pleaseWaitPresented.Set(true)
						xsync.Go(func() error {
							if c.conv.Get() == "" {
								agentID, modelID, err := c.startOptions.resolve(wnd.Subject(), c.provider)
								if err != nil {
									return err
								}

								// create a new conversation
								tmp := c.text.Get()
								inputs := []message.Input{
									{
										Text: option.Some(tmp),
									},
								}

								for _, f := range pendingUploadFiles.Get() {
									inputs = append(inputs, message.Input{
										File: option.Some(f),
									})
								}

								cv, msgs, err := conversations.Create(wnd.Subject(), conversation.CreateOptions{
									Model:      modelID,
									Agent:      agentID,
									Name:       tmp,
									Input:      inputs,
									CloudStore: c.startOptions.CloudStore,
								})
								if err != nil {
									return err
								}

								wnd.Post(func() {
									slog.Info("conversation created", "id", cv.ID)
									c.conv.Update(cv.ID)
									messages.Set(msgs)
								})
							} else {
								// append to existing conversation
								tmp := c.text.Get()
								inputs := []message.Input{
									{
										Text: option.Some(tmp),
									},
								}

								for _, f := range pendingUploadFiles.Get() {
									inputs = append(inputs, message.Input{
										File: option.Some(f),
									})
								}

								msgs, err := conversations.Conversation(wnd.Subject(), c.conv.Get()).Append(wnd.Subject(), message.AppendOptions{
									Input:      inputs,
									CloudStore: c.startOptions.CloudStore,
								})
								if err != nil {
									return err
								}

								slice := messages.Get()
								slice = append(slice, msgs...)
								messages.Set(slice)
								messages.Invalidate()
							}

							return nil
						}, func(err error) {
							wnd.PostDelayed(func() {
								pleaseWaitPresented.Set(false)

								// some post to work-around redraw cycles? that does not look very stable
								wnd.PostDelayed(func() {
									wnd.RequestFocus("ai-user-prompt")
								}, 100*time.Millisecond)

								if err != nil {
									alert.ShowBannerError(wnd, err)
									return
								} else {
									c.text.Set("")
									pendingUploadFiles.Set(nil)
								}

							}, time.Millisecond*500) // we got some state failures in practice, probably caused by the debounce time of the input text field which is 500ms by default

						})

					}),
				ui.PrimaryButton(func() {
					wnd.ImportFiles(core.ImportFilesOptions{
						Multiple: true,
						OnCompletion: func(files []core.File) {
							for _, f := range files {
								aiFile, err := c.provider.Files().Unwrap().Put(wnd.Subject(), file.CreateOptions{
									Name: f.Name(),
									Open: f.Open,
								})

								if err != nil {
									alert.ShowBannerError(wnd, err)
									return
								}

								wnd.Post(func() {
									slice := pendingUploadFiles.Get()
									slice = append(slice, aiFile)
									pendingUploadFiles.Set(slice) // equals will miss the update
									pendingUploadFiles.Invalidate()
								})

							}
						},
					})
				}).Title(rstring.ActionFileUpload.Get(wnd)).
					Enabled(c.provider.Files().IsSome()).
					PreIcon(icons.Upload).
					Frame(ui.Frame{MinWidth: "12rem"}),
				c.more,
			).Gap(ui.L8).
				FullWidth(),
		).
		// chat footer
		Append(
			ui.Space(ui.L16),
			ui.HStack(
				ui.Text(StrAIDisclaimer.Get(wnd)),
			).BackgroundColor(ui.M2).Border(ui.Border{}.Radius(ui.L16)).Padding(ui.Padding{}.Vertical(ui.L16).Horizontal(ui.L96)),
		).
		Append(ui.VStack().ID("end-of-chat")).
		Gap(ui.L16).
		Alignment(ui.Bottom).
		Padding(c.padding).
		Frame(c.frame).
		Render(ctx)
}
