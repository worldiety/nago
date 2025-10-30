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

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/ai/agent"
	"go.wdy.de/nago/application/ai/conversation"
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
}

func Chat(provider provider.Provider, conv *core.State[conversation.ID], text *core.State[string]) TChat {
	return TChat{conv: conv, text: text, provider: provider}
}

func (c TChat) Padding(padding ui.Padding) TChat {
	c.padding = padding
	return c
}

func (c TChat) StartOptions(opts StartOptions) TChat {
	c.startOptions = opts
	return c
}

func (c TChat) Render(ctx core.RenderContext) core.RenderNode {
	wnd := ctx.Window()

	if c.provider.Conversations().IsNone() {
		return alert.BannerError(fmt.Errorf("provider has no conversation support id: %s: %w", c.provider, os.ErrNotExist)).Render(ctx)
	}

	conversations := c.provider.Conversations().Unwrap()

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
	})

	pleaseWaitPresented := core.DerivedState[bool](c.conv, "-pw-presented")

	return ui.VStack(
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

			}

			return stack.Alignment(align).FullWidth()
		})...,
	).Append(
		ui.If(pleaseWaitPresented.Get(),
			ui.HStack(
				ChatMessage().Icon(icons.DotsHorizontal).Style(MessageAgent),
			).FullWidth().Alignment(ui.Leading),
		),
	).
		Append(
			ui.HStack(
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

								cv, msgs, err := conversations.Create(wnd.Subject(), conversation.CreateOptions{
									Model: modelID,
									Agent: agentID,
									Name:  tmp,
									Input: []message.Content{
										{
											Text: option.Pointer(&tmp),
										},
									},
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
								msgs, err := conversations.Conversation(wnd.Subject(), c.conv.Get()).Append(wnd.Subject(), message.AppendOptions{
									MessageInput: option.Pointer(&tmp),
									CloudStore:   c.startOptions.CloudStore,
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

								if err != nil {
									alert.ShowBannerError(wnd, err)
									return
								} else {
									c.text.Set("")
								}
							}, time.Millisecond*500) // we got some state failures in practice, probably caused by the debounce time of the input text field which is 500ms by default

						})

					}),
				ui.PrimaryButton(nil).Title(rstring.ActionFileUpload.Get(wnd)).PreIcon(icons.Upload).Frame(ui.Frame{MinWidth: "12rem"}),
				ui.SecondaryButton(nil).Title(rstring.LabelHowItWorks.Get(wnd)).Frame(ui.Frame{MinWidth: "12rem"}),
			).Gap(ui.L8).
				FullWidth(),
		).
		Append(ui.VStack().ID("end-of-chat")).
		FullWidth().
		Gap(ui.L16).
		Padding(c.padding).
		Render(ctx)
}
