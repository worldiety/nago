// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package evs

import (
	"context"
	"iter"

	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/blob"
)

func NewReplay[Evt any](perms Permissions, eventStore blob.Store, loadEvt Load[Evt]) Replay[Evt] {
	return func(subject auth.Subject, fromInc, toInc SeqID) iter.Seq2[Envelope[Evt], error] {
		return func(yield func(Envelope[Evt], error) bool) {
			var zero Envelope[Evt]
			if err := subject.Audit(perms.Replay); err != nil {
				yield(zero, err)
				return
			}

			fromSeq, err := NewSeqKey(fromInc)
			if err != nil {
				yield(zero, err)
				return
			}

			toSeq, err := NewSeqKey(toInc)
			if err != nil {
				yield(zero, err)
				return
			}

			for key, err := range eventStore.List(context.Background(), blob.ListOptions{
				MinInc: string(fromSeq),
				MaxInc: string(toSeq),
			}) {
				if err != nil {
					if !yield(zero, err) {
						return
					}

					continue
				}

				seqId, err := SeqKey(key).Parse()
				if err != nil {
					if !yield(zero, err) {
						return
					}

					continue
				}

				optEvt, err := loadEvt(subject, seqId)
				if err != nil {
					if !yield(zero, err) {
						return
					}

					continue
				}

				if optEvt.IsNone() {
					continue
				}

				if !yield(optEvt.Unwrap(), nil) {
					return
				}
			}
		}
	}
}
