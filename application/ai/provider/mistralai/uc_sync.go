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
	"reflect"
	"sync"

	"go.wdy.de/nago/application/ai/agent"
	"go.wdy.de/nago/application/ai/workspace"
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/secret"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
)

func NewSync(mutex *sync.Mutex, repoWorkspaceName string, syncAgentRepo SyncAgentRepository, match secret.Match, findWorkspaces workspace.FindWorkspacesByPlatform, findAgent agent.FindByID) Sync {
	return func(subject auth.Subject) error {
		mutex.Lock()
		defer mutex.Unlock()

		for ws, err := range findWorkspaces(user.SU(), workspace.MistralAI) {
			if err != nil {
				return err
			}

			if err := subject.AuditResource(repoWorkspaceName, string(ws.ID), PermSync); err != nil {
				slog.Warn("cannot sync workspace: not allowed", "wid", ws.ID)
				continue
			}

			if err := syncWorkspace(syncAgentRepo, match, findAgent, ws); err != nil {
				return err
			}
		}

		return nil
	}
}

func syncWorkspace(syncAgentRepo SyncAgentRepository, match secret.Match, findAgent agent.FindByID, ws workspace.Workspace) error {
	optSec, err := match(user.SU(), reflect.TypeFor[Settings](), secret.MatchOptions{
		Hint:   ws.SecretHint,
		Group:  group.System,
		Expect: true,
	})

	if err != nil {
		return err
	}

	// collect agents
	var agents []agent.Agent
	for _, aid := range ws.Agents {
		optAg, err := findAgent(user.SU(), aid)
		if err != nil {
			return fmt.Errorf("cannot find agent %q: %w", aid, err)
		}

		if optAg.IsNone() {
			slog.Error("orphaned agent in workspace", "agent", aid, "workspace", ws.ID)
			continue
		}

		ag := optAg.Unwrap()
		agents = append(agents, ag)
	}

	// grab secret and connect
	sec := optSec.Unwrap().(Settings)
	cl := NewClient(sec.Token)

	for _, localAgent := range agents {
		optLastSync, err := syncAgentRepo.FindByID(localAgent.ID)
		if err != nil {
			return fmt.Errorf("cannot find last sync agent: %w", err)
		}

		if optLastSync.IsNone() {
			if err := createCloudAgent(syncAgentRepo, cl, localAgent); err != nil {
				return err
			}

			continue
		}

		lastSync := optLastSync.Unwrap()

		// check if remote was deleted by third-party
		_, err = cl.GetAgent(lastSync.CloudAgent)
		if err != nil {
			// TODO create just as if it does not exist, however we CANNOT test this, because they just did not implement any possibility to ever delete an agent at all. there is not even a secret api
			return fmt.Errorf("cannot get cloud agent: %w", err)
		}

		if lastSync.LastMod == localAgent.LastMod {
			slog.Info("local agent and cloud agent are unmodified (same last mod)", "local-agent", localAgent.Name, "cloud-agent", lastSync.CloudAgent)
			continue
		}

		// and finally update
		model := calculateBestMistralModel(localAgent)
		err = cl.UpdateAgent(lastSync.CloudAgent, UpdateAgentRequest{
			Instructions: &localAgent.Prompt,
			Model:        &model,
			Name:         &localAgent.Name,
			Description:  &localAgent.Description,
		})

		if err != nil {
			return fmt.Errorf("cannot update cloud agent: %w", err)
		}

		lastSync.LastMod = localAgent.LastMod
		if err := syncAgentRepo.Save(lastSync); err != nil {
			return fmt.Errorf("cannot save last sync agent: %w", err)
		}

		slog.Info("cloud agent updated", "local-agent", localAgent.Name, "cloud-agent", lastSync.CloudAgent)
	}

	return nil
}

func createCloudAgent(syncAgentRepo SyncAgentRepository, cl *Client, localAgent agent.Agent) error {
	cla, err := cl.CreateAgent(CreateAgentRequest{
		Model:        calculateBestMistralModel(localAgent),
		Name:         localAgent.Name,
		Description:  localAgent.Description,
		Instructions: localAgent.Prompt,
	})

	if err != nil {
		return fmt.Errorf("cannot create cloud agent %q: %w", localAgent.ID, err)
	}

	if err := syncAgentRepo.Save(SynchronizedAgent{
		ID:         localAgent.ID,
		CloudAgent: cla.Id,
		LastMod:    localAgent.LastMod,
	}); err != nil {
		return fmt.Errorf("cannot save sync meta data %q: %w", localAgent.ID, err)
	}

	slog.Info("created cloud agent", "local-agent", localAgent.Name, "cloud-agent", cla.Id)
	return nil
}

func calculateBestMistralModel(a agent.Agent) string {
	switch a.Model {
	case agent.Efficiency:
		return "mistral-small-latest"
	case agent.Quality:
		return "mistral-large-latest"
	default:
		return "mistral-medium-latest"
	}
}
