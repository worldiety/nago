// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package evs

import (
	"fmt"
	"strconv"
)

type SeqKey string

func NewSeqKey(n SeqID) (SeqKey, error) {

	if n <= 0 {
		return "", fmt.Errorf("invalid sequence: must not be less or equal than 0: %d", n)
	}

	if n > 999_999_999_999 {
		return "", fmt.Errorf("sequence exceeds maximum supported value: %d", n)
	}

	return SeqKey(fmt.Sprintf("%012d", n)), nil
}

func (s SeqKey) Parse() (SeqID, error) {
	num, err := strconv.ParseInt(string(s), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid seq key: invalid seq number: %s: %w", s, err)
	}

	return SeqID(num), nil
}
