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
	"iter"

	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/xiter"
)

func NewReplayWithIndex[Primary ~string, Evt any](perms Permissions, loadEvent Load[Evt], wsIndex *StoreIndex[Primary, Evt]) ReplayWithIndex[Primary, Evt] {
	return func(subject auth.Subject, primary Primary, opts ReplayOptions) iter.Seq2[Envelope[Evt], error] {
		if err := subject.Audit(perms.Replay); err != nil {
			return xiter.WithError[Envelope[Evt]](err)
		}

		return func(yield func(Envelope[Evt], error) bool) {
			var zero Envelope[Evt]
			for key, err := range wsIndex.AllByPrimary(context.Background(), primary) {
				if err != nil {
					if !yield(zero, err) {
						return
					}

					continue
				}

				seqId, err := key.Secondary.Parse()
				if err != nil {
					if !yield(zero, fmt.Errorf("invalid sequence id: %w", err)) {
						return
					}

					continue
				}

				if opts.FromInc > 0 && seqId <= opts.FromInc {
					continue
				}

				if opts.ToInc > 0 && seqId >= opts.ToInc {
					continue
				}

				optEvt, err := loadEvent(user.SU(), seqId)
				if err != nil {
					if !yield(zero, fmt.Errorf("cannot load event %s: %w", key.Secondary, err)) {

					}

					continue
				}

				if optEvt.IsNone() {
					if !yield(zero, fmt.Errorf("event %s is missing", key.Secondary)) {
						return
					}

					continue
				}

				if !yield(optEvt.Unwrap(), nil) {
					return
				}
			}
		}
	}
}
