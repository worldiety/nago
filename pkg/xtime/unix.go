package xtime

import "time"

type UnixMilliseconds int64

// Time returns a correctly located stdlib time. Note that the original time call always uses the global timezone
// of the runtime, which is often the wrong thing, because it depends on the configuration and/or location
// of the cluster/node/pod. However, timezones are mostly always a domain thing.
func (u UnixMilliseconds) Time(tz *time.Location) time.Time {
	return time.UnixMilli(int64(u)).In(tz)
}

func Now() UnixMilliseconds {
	return UnixMilliseconds(time.Now().UnixMilli())
}
