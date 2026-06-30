package msgstore

const defaultMaxMessageSize int64 = 16 << 20 // 16 MiB

// Options configures the behaviour of the event store.
type Options struct {
	// MaxMessageSize is the upper limit for a single message payload in bytes.
	// Zero or negative values default to 16 MiB.
	MaxMessageSize int64

	// Compress decides per event type and payload how to compress.
	// nil defaults to DefaultCompression (S2 for payloads > 512 bytes).
	Compress CompressFunc

	// ShouldSplit decides before each append whether the current pending
	// segment should be finalized and a new one started.
	// nil defaults to split at 64 MiB or on day boundary.
	ShouldSplit SplitFunc

	// FilePool manages open file handles with LRU eviction.
	// nil defaults to NewFilePool(1024).
	FilePool *FilePool
}

// resolve fills nil/zero fields with sensible defaults.
func (o *Options) resolve() {
	if o.MaxMessageSize <= 0 {
		o.MaxMessageSize = defaultMaxMessageSize
	}
	if o.Compress == nil {
		o.Compress = DefaultCompression
	}
	if o.ShouldSplit == nil {
		o.ShouldSplit = defaultSplit
	}
	if o.FilePool == nil {
		o.FilePool = NewFilePool(1024)
	}
}
