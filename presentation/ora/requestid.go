package ora

// RequestId is usually used by the frontend to distinguish different concurrent requests to the server
// or backend. If the identifier is 0, it is considered as absent.
type RequestId int64
