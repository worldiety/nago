// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package xtime

import "time"

type UnixMilliseconds int64

// Time returns a correctly located stdlib time. Note that the original time call always uses the global timezone
// of the runtime, which is often the wrong thing, because it depends on the configuration and/or location
// of the cluster/node/pod. However, timezones are mostly always a domain thing.
func (u UnixMilliseconds) Time(tz *time.Location) time.Time {
	return time.UnixMilli(int64(u)).In(tz)
}

func (u UnixMilliseconds) Date(tz *time.Location) Date {
	year, month, day := u.Time(tz).Date()
	return Date{day, month, year}
}

func Now() UnixMilliseconds {
	return UnixMilliseconds(time.Now().UnixMilli())
}
