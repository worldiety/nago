// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package usercircle

import (
	"go.wdy.de/nago/auth"
	"iter"
	"slices"
)

func NewMyCircles(repoCircle Repository) MyCircles {
	return func(subject auth.Subject) iter.Seq2[Circle, error] {
		return func(yield func(Circle, error) bool) {
			for circle, err := range repoCircle.All() {
				if err != nil {
					if !yield(circle, err) {
						return
					}

					continue
				}

				if !slices.Contains(circle.Administrators, subject.ID()) {
					continue
				}

				if !yield(circle, nil) {
					return
				}
			}
		}
	}
}
