// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package session

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/settings"
	"go.wdy.de/nago/application/user"
)

func NewExchangeNLS(mutex *sync.Mutex, repo NLSNonceRepository, repoSession Repository, loadGlobal settings.LoadGlobal, refresh RefreshNLS) ExchangeNLS {
	return func(id ID, nonce NLSNonce) (string, error) {
		usrSettings := settings.ReadGlobal[user.Settings](loadGlobal)

		optEntry, err := repo.FindByID(nonce)
		if err != nil {
			return "", err
		}

		if optEntry.IsNone() {
			return "", fmt.Errorf("nonce unknown: %w", os.ErrNotExist)
		}

		entry := optEntry.Unwrap()
		url := strings.TrimSuffix(usrSettings.SSONLSServer, "/") + "/api/nago/v1/exchange"
		client := &http.Client{
			Timeout: 30 * time.Second,
		}

		type body struct {
			Nonce NLSNonce `json:"nonce"`
		}

		buf := option.Must(json.Marshal(body{Nonce: nonce}))
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(buf))
		if err != nil {
			return "", fmt.Errorf("invalid nls request: %s: %w", url, err)
		}

		resp, err := client.Do(req)
		if err != nil {
			return "", fmt.Errorf("failed sending request: %s: %w", url, err)
		}

		defer resp.Body.Close()

		type responseBody struct {
			Refresh NLSRefreshToken `json:"refresh"`
		}

		var response responseBody
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return "", fmt.Errorf("failed decoding response: %s: %w", url, err)
		}

		if err := repo.DeleteByID(nonce); err != nil {
			return "", err
		}

		if response.Refresh == "" {
			return "", fmt.Errorf("nls exchange response has no refresh token: %s: %w", url, os.ErrNotExist)
		}

		err = func() error {
			mutex.Lock()
			defer mutex.Unlock()

			optSession, err := repoSession.FindByID(entry.Session)
			if err != nil {
				return fmt.Errorf("failed finding session: %w", err)
			}

			if optSession.IsNone() {
				return fmt.Errorf("session unknown: %w", os.ErrNotExist)
			}

			session := optSession.Unwrap()
			session.RefreshToken = response.Refresh
			if err := repoSession.Save(session); err != nil {
				return fmt.Errorf("failed saving session: %w", err)
			}

			return nil
		}()

		if err != nil {
			return "", err
		}

		if err := refresh(id); err != nil {
			return "", fmt.Errorf("failed refreshing session: %w", err)
		}

		return entry.Redirect, nil
	}
}
