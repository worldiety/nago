package nethost

var (
	ForbiddenAddress     = forbiddenAddress
	ForbiddenPrivateOnly = forbiddenPrivateOnly
)

// OpenRequests returns the number of requests, which count against the limits.
func (h *Host) OpenRequests() int64 { return h.open.Load() }
