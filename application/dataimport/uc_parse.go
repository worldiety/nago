// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package dataimport

import (
	"context"
	"fmt"
	"go.wdy.de/nago/application/dataimport/parser"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/std/concurrent"
	"io"
	"log/slog"
	"os"
)

func NewParse(repoStaging StagingRepository, entryRepo EntryRepository, parsers *concurrent.RWMap[parser.ID, parser.Parser]) Parse {
	return func(subject auth.Subject, dst SID, src parser.ID, opts parser.Options, reader io.Reader) (ParseStats, error) {
		var stats ParseStats
		if err := subject.Audit(PermParse); err != nil {
			return stats, err
		}

		p, ok := parsers.Get(src)
		if !ok {
			return stats, fmt.Errorf("parser does not exist: %w", os.ErrNotExist)
		}

		optStage, err := repoStaging.FindByID(dst)
		if err != nil {
			return stats, err
		}

		if optStage.IsNone() {
			return stats, fmt.Errorf("staging does not exist: %w", os.ErrNotExist)
		}

		for obj, err := range p.Parse(context.Background(), reader, opts) {
			if err != nil {
				return stats, err
			}

			key := NewKey(dst)
			if err := entryRepo.Save(Entry{
				ID: key,
				In: obj,
			}); err != nil {
				return stats, fmt.Errorf("cannot save entry: %w", err)
			}

			stats.Count++
		}

		slog.Info("parsed and stored entries as import candidates", "count", stats.Count, "parser", src, "stage", dst, "user", subject.ID())

		return stats, nil
	}
}
