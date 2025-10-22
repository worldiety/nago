// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package mistralai

import (
	"fmt"
	"os"
	"reflect"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/ai/conversation"
	"go.wdy.de/nago/application/ai/message"
	"go.wdy.de/nago/application/ai/workspace"
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/secret"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/events"
)

func appendMessage(bus events.Bus, match secret.Match, findConvByID conversation.FindByID, findWSByID workspace.FindByID, syncConvRepo SyncConversationRepository, evt conversation.HumanAppended) error {
	optConv, err := findConvByID(user.SU(), evt.Conversation)
	if err != nil {
		return err
	}

	if optConv.IsNone() {
		return fmt.Errorf("conversation is gone: %q: %w", evt.Conversation, os.ErrNotExist)
	}

	conv := optConv.Unwrap()
	optWS, err := findWSByID(user.SU(), conv.Workspace)
	if err != nil {
		return err
	}

	if optWS.IsNone() {
		return fmt.Errorf("workspace is gone: %q: %w", conv.Workspace, os.ErrNotExist)
	}

	ws := optWS.Unwrap()

	optSec, err := match(user.SU(), reflect.TypeFor[Settings](), secret.MatchOptions{
		Hint:   ws.SecretHint,
		Group:  group.System,
		Expect: true,
	})

	if err != nil {
		return err
	}

	if optSec.IsNone() {
		return fmt.Errorf("no secret for workspace platform: %w", os.ErrNotExist)
	}

	optSyncConv, err := syncConvRepo.FindByID(conv.ID)
	if err != nil {
		return err
	}

	if optSyncConv.IsNone() {
		return fmt.Errorf("conversation is not present in mistral cloud: %q: %w", conv.ID, os.ErrNotExist)
	}

	syncConv := optSyncConv.Unwrap()
	sec := optSec.Unwrap().(Settings)
	cl := NewClient(sec.Token)

	resp, err := cl.AppendConversation(syncConv.CloudConversation, AppendConversationRequest{
		Inputs: convInputToMistralInput(evt.Content),
	})

	if err != nil {
		return fmt.Errorf("cannot append mistral conversation: %q: %w", syncConv.CloudConversation, err)
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
