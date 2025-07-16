// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application

import (
	"github.com/worldiety/option"
	"go.wdy.de/nago/application/backup"
	uibackup "go.wdy.de/nago/application/backup/ui"
	"go.wdy.de/nago/pkg/blob/crypto"
	"go.wdy.de/nago/presentation/core"
	"log/slog"
)

type BackupManagement struct {
	UseCases backup.UseCases
	Pages    uibackup.Pages
}

func (c *Configurator) BackupManagement() (BackupManagement, error) {
	if c.backupManagement == nil {
		stores, err := c.Stores()
		if err != nil {
			return BackupManagement{}, err
		}

		c.backupManagement = &BackupManagement{
			UseCases: backup.NewUseCases(
				stores,
				func() crypto.EncryptionKey {
					return option.Must(c.MasterKey())
				},
				func(key crypto.EncryptionKey) {
					option.MustZero(c.WriteMasterKey(key))
					if s := c.sessionManagement; s != nil {
						if err := s.UseCases.Clear(); err != nil {
							slog.Error("failed to clear session storage after setting masterkey", "err", err.Error())
							return
						}
					}
				},
			),
			Pages: uibackup.Pages{
				BackupAndRestore: "admin/backup-and-restore",
			},
		}

		c.RootViewWithDecoration(c.backupManagement.Pages.BackupAndRestore, func(wnd core.Window) core.View {
			return uibackup.BackupAndRestorePage(wnd, c.backupManagement.UseCases)
		})
	}

	return *c.backupManagement, nil
}
