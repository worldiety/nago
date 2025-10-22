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

	"go.wdy.de/nago/application/ai/conversation"
	"go.wdy.de/nago/application/ai/workspace"
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/secret"
	"go.wdy.de/nago/application/user"
)

func createClientForConversation(match secret.Match, findConvByID conversation.FindByID, findWSByID workspace.FindByID, cid conversation.ID) (workspace.Workspace, conversation.Conversation, *Client, error) {
	optConv, err := findConvByID(user.SU(), cid)
	if err != nil {
		return workspace.Workspace{}, conversation.Conversation{}, nil, err
	}

	if optConv.IsNone() {
		return workspace.Workspace{}, conversation.Conversation{}, nil, fmt.Errorf("conversation is gone: %q: %w", cid, os.ErrNotExist)
	}

	conv := optConv.Unwrap()
	ws, cl, err := createClientForWorkspace(match, findWSByID, conv.Workspace)
	if err != nil {
		return workspace.Workspace{}, conversation.Conversation{}, nil, err
	}

	return ws, conv, cl, nil
}

func createClientForWorkspace(match secret.Match, findWSByID workspace.FindByID, wid workspace.ID) (workspace.Workspace, *Client, error) {

	optWS, err := findWSByID(user.SU(), wid)
	if err != nil {
		return workspace.Workspace{}, nil, err
	}

	if optWS.IsNone() {
		return workspace.Workspace{}, nil, fmt.Errorf("workspace is gone: %q: %w", wid, os.ErrNotExist)
	}

	ws := optWS.Unwrap()

	optSec, err := match(user.SU(), reflect.TypeFor[Settings](), secret.MatchOptions{
		Hint:   ws.SecretHint,
		Group:  group.System,
		Expect: true,
	})

	if err != nil {
		return workspace.Workspace{}, nil, err
	}

	if optSec.IsNone() {
		return workspace.Workspace{}, nil, fmt.Errorf("no secret for workspace platform: %w", os.ErrNotExist)
	}

	sec := optSec.Unwrap().(Settings)
	cl := NewClient(sec.Token)

	return ws, cl, nil
}
