// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package evs

import (
	"context"
	"fmt"

	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
)

func NewReplayWithIndex[Primary ~string, Evt any](perms Permissions, loadEvent Load[Evt], wsIndex *StoreIndex[Primary, Evt]) ReplayWithIndex[Primary, Evt] {
	return func(subject auth.Subject, primary Primary, apply func(Envelope[Evt]) error, opts ReplayOptions) error {
		if err := subject.Audit(perms.Replay); err != nil {
			return err
		}

		for key, err := range wsIndex.AllByPrimary(context.Background(), primary) {
			if err != nil {
				return err
			}

			seqId, err := key.Secondary.Parse()
			if err != nil {
				return fmt.Errorf("invalid sequence id: %w", err)
			}

			if opts.FromInc > 0 && seqId <= opts.FromInc {
				continue
			}

			if opts.ToInc > 0 && seqId >= opts.ToInc {
				continue
			}

			optEvt, err := loadEvent(user.SU(), seqId)
			if err != nil {
				return fmt.Errorf("cannot load event %s: %w", key.Secondary, err)
			}

			if optEvt.IsNone() {
				return fmt.Errorf("event %s is missing", key.Secondary)
			}

			if err := apply(optEvt.Unwrap()); err != nil {
				return fmt.Errorf("apply failed: %w", err)
			}
		}

		return nil
	}
}
