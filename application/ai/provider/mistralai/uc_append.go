// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package mistralai

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/ai/conversation"
	"go.wdy.de/nago/application/ai/message"
	"go.wdy.de/nago/application/ai/workspace"
	"go.wdy.de/nago/application/secret"
	"go.wdy.de/nago/pkg/events"
)

func appendMessage(bus events.Bus, match secret.Match, findConvByID conversation.FindByID, findWSByID workspace.FindByID, syncConvRepo SyncConversationRepository, evt conversation.HumanAppended) error {
	_, conv, cl, err := createClientForConversation(match, findConvByID, findWSByID, evt.Conversation)
	if err != nil {
		return err
	}

	optSyncConv, err := syncConvRepo.FindByID(conv.ID)
	if err != nil {
		return err
	}

	if optSyncConv.IsNone() {
		return fmt.Errorf("conversation is not present in mistral cloud: %q: %w", conv.ID, os.ErrNotExist)
	}

	syncConv := optSyncConv.Unwrap()

	resp, err := cl.AppendConversation(syncConv.CloudConversation, AppendConversationRequest{
		Store:  true,
		Stream: false,
		Inputs: convInputToMistralInput(evt.Content),
	})

	if err != nil {
		return fmt.Errorf("cannot append mistral conversation: %q: %w", syncConv.CloudConversation, err)
	}

	slog.Info("appended message to mistral conversation", "old-id", syncConv.CloudConversation, "new-id", resp.ConversationId)

	if resp.ConversationId != syncConv.CloudConversation {
		slog.Info("mistral changed its conversation logic")
		syncConv.CloudConversation = resp.ConversationId
		if err := syncConvRepo.Save(syncConv); err != nil {
			return fmt.Errorf("cannot save mistral sync cloud conversation id update: %w", err)
		}
	}

	for _, output := range resp.Outputs {
		bus.Publish(conversation.AgentAppended{
			Conversation: conv.ID,
			Content: []message.Content{{
				Text: option.Pointer(&output.Content),
			}},
		})
	}

	return nil
}
