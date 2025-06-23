// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package dataimport

import (
	"context"
	"errors"
	"fmt"
	"github.com/worldiety/jsonptr"
	"go.wdy.de/nago/application/dataimport/importer"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/pkg/std/concurrent"
	"log/slog"
	"os"
	"sync"
	"time"
)

func NewImport(mutex *sync.Mutex, entryRepo EntryRepository, stageRepo StagingRepository, imports *concurrent.RWMap[importer.ID, importer.Importer]) Import {
	return func(subject auth.Subject, stage SID, dst importer.ID, opts ImportOptions) error {
		if err := subject.AuditResource(stageRepo.Name(), string(stage), PermImport); err != nil {
			return err
		}

		optStageing, err := stageRepo.FindByID(stage)
		if err != nil {
			return fmt.Errorf("cannot find staging: %v", err)
		}

		if optStageing.IsNone() {
			return fmt.Errorf("staging %s not found: %w", stage, os.ErrNotExist)
		}

		staging := optStageing.Unwrap()

		imp, ok := imports.Get(dst)
		if !ok {
			return fmt.Errorf("import %s not found: %w", dst, os.ErrNotExist)
		}

		if opts.Context == nil {
			opts.Context = context.Background()
		}

		errCount := 0
		for entry, err := range entryRepo.FindAllByPrefix(Key(stage) + "/") {
			if err != nil {
				errCount++
				if !opts.ImporterOptions.ContinueOnError {
					return err
				}

				slog.Error("failed to find entry", "err", err.Error())
				continue
			}

			if entry.Imported {
				slog.Info("import entry ignored, already imported", "key", entry.ID)
				continue
			}

			if entry.Ignored {
				slog.Info("import entry ignored, ignored flag is set", "key", entry.ID)
				continue
			}

			obj := entry.Transform(staging.Transformation)
			if err := imp.Import(opts.Context, importer.Options{
				ContinueOnError: false, // intentionally always set to false, we will handle that accordingly, because we call it 1:1 for each entry
				MergeDuplicates: opts.ImporterOptions.MergeDuplicates,
			}, func(yield func(*jsonptr.Obj, error) bool) {
				yield(obj, nil)
			}); err != nil {
				errCount++
				if err2 := updateEntryOnImport(mutex, entryRepo, func() Entry {
					entry.Imported = false
					entry.Confirmed = false

					var locErr std.LocalizedError
					if errors.As(err, &locErr) {
						entry.ImportedError = locErr.Description()
					} else {
						entry.ImportedError = err.Error()
					}

					return entry
				}); err2 != nil {
					return err2
				}

				if !opts.ImporterOptions.ContinueOnError {
					return fmt.Errorf("cannot import entry %s: %w", entry.ID, err)
				}

				slog.Error("failed to import entry", "err", err.Error(), "key", entry.ID)
				continue
			}

			if err := updateEntryOnImport(mutex, entryRepo, func() Entry {
				entry.Imported = true
				entry.ImportedAt = time.Now()
				entry.ImportedError = ""
				return entry
			}); err != nil {
				return err
			}

		}

		if errCount > 0 {
			return std.NewLocalizedError("Import fehlgeschlagen", fmt.Sprintf("Der Import von %d Einträgen ist fehlgeschlagen.", errCount))
		}

		return nil
	}
}

func updateEntryOnImport(mutex *sync.Mutex, repo EntryRepository, fn func() Entry) error {
	mutex.Lock()
	defer mutex.Unlock()

	entry := fn()

	if err := repo.Save(entry); err != nil {
		return fmt.Errorf("cannot update entry state %s: %w", entry.ID, err)
	}

	return nil
}
