package msgstore

import "go.wdy.de/nago/pkg/ndb"

// FilePool is the bounded, LRU file-descriptor pool. It now lives in package
// ndb so that a single pool can be shared across all engine instances of a
// [ndb.DB]. This alias is retained so existing msgstore code and callers that
// reference msgstore.FilePool keep compiling.
type FilePool = ndb.FilePool

// NewFilePool creates a standalone FilePool. Prefer the shared pool injected by
// ndb via the engine factory; this is used only when msgstore is opened
// standalone or a caller supplies an explicit pool via [Options.FilePool] or the
// filepool=N DSN option.
func NewFilePool(maxOpen int) *FilePool {
	return ndb.NewFilePool(maxOpen)
}
