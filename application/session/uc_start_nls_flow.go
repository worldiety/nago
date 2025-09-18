// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package session

import (
	"fmt"
	"net/url"
	"os"
	"sync"

	"go.wdy.de/nago/application/settings"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/data"
)

func NewStartNLSFlow(mutex *sync.Mutex, redirect string, repoNonce NLSNonceRepository, loadGlobal settings.LoadGlobal) StartNLSFlow {
	return func(id ID) (string, error) {
		if id == "" {
			return "", fmt.Errorf("id must not be empty")
		}

		usrSettings := settings.ReadGlobal[user.Settings](loadGlobal)
		if !usrSettings.HasSSO() {
			return "", fmt.Errorf("sso NLS url has not been set")
		}

		mutex.Lock()
		defer mutex.Unlock()

		nonce := data.RandIdent[NLSNonce]()
		if optEntry, err := repoNonce.FindByID(nonce); err != nil || optEntry.IsSome() {
			if err != nil {
				return "", fmt.Errorf("error finding NLS nonce: %v", err)
			}

			return "", fmt.Errorf("nonce collision: %w", os.ErrExist)
		}

		if err := repoNonce.Save(NLSNonceEntry{
			ID:      nonce,
			Session: id,
		}); err != nil {
			return "", fmt.Errorf("error saving NLS nonce: %v", err)
		}

		params := url.Values{
			"redirect": []string{redirect},
			"nonce":    []string{string(nonce)},
		}

		nls := usrSettings.SSONLSServer + "?" + params.Encode()

		return nls, nil
	}
}
