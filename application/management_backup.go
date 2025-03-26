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
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/blob/crypto"
	"go.wdy.de/nago/pkg/xiter"
	"go.wdy.de/nago/presentation/core"
	"iter"
	"maps"
)

type BackupManagement struct {
	UseCases backup.UseCases
	Pages    uibackup.Pages
}

func (c *Configurator) BackupManagement() (BackupManagement, error) {
	if c.backupManagement == nil {
		c.backupManagement = &BackupManagement{
			UseCases: backup.NewUseCases(
				&cfgPersistence{c},
				func() crypto.EncryptionKey {
					return option.Must(c.MasterKey())
				},
				func(key crypto.EncryptionKey) {
					option.MustZero(c.WriteMasterKey(key))
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

type cfgPersistence struct {
	cfg *Configurator
}

func (c *cfgPersistence) FileStores() iter.Seq2[string, error] {
	return xiter.Zero2[string, error](maps.Keys(c.cfg.fileStores))
}

func (c *cfgPersistence) EntityStores() iter.Seq2[string, error] {
	return xiter.Zero2[string, error](maps.Keys(c.cfg.entityStores))
}

func (c *cfgPersistence) FileStore(name string) (blob.Store, error) {
	return c.cfg.FileStore(name)
}

func (c *cfgPersistence) EntityStore(name string) (blob.Store, error) {
	return c.cfg.EntityStore(name)
}
