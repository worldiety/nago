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

	"go.wdy.de/nago/pkg/xtime"
)

// tsKey is a timestamp encoded key. The Timestamp is encoded as a unix timestamp
// in milliseconds to avoid any time zone and offset confusion.
type tsKey string

func newTsKey(n xtime.UnixMilliseconds) (tsKey, error) {
	if n < 0 {
		return "", fmt.Errorf("invalid ts key: negative timestamp not allowed: %d", n)
	}
	if n > 9999999999999 {
		return "", fmt.Errorf("invalid ts key: timestamp exceeds 13-digit limit (max 9999999999999): %d", n)
	}
	
	return tsKey(fmt.Sprintf("%013d", n)), nil
}

func (s tsKey) Parse() (xtime.UnixMilliseconds, error) {

	num, err := strconv.ParseInt(string(s), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid ts key: invalid timestamp number: %s: %w", s, err)
	}

	return xtime.UnixMilliseconds(num), nil
}
