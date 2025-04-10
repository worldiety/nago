A ScopeID has at least a 32 byte entropy and must be generated using a secure random source.
It must be treated as a secret at the frontend (e.g. no exposing into URLs), because
it allows the hijacking of connections and allocated components.
These components may likely contain already authorized credentials, thus leaking the ScopeID
also means leaking the access rights.

If you know, that you are done, destroy the scope to release all associated backend resources.
Keep the lifetime of the scope small to trade resume comfort and security and resource usage.

Note that allocations of components inside a Scope are unrelated and must be managed explicitly.