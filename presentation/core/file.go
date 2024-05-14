package core

import (
	"go.wdy.de/nago/presentation/ora"
	"io"
	"io/fs"
	"os"
	"time"
)

type tmpFile struct {
	meta tmpFileInfo
	file *os.File
}

func newTmpFile(meta tmpFileInfo, file *os.File) *tmpFile {
	return &tmpFile{meta: meta, file: file}
}

func (t *tmpFile) ReadAt(p []byte, off int64) (n int, err error) {
	return t.file.ReadAt(p, off)
}

func (t *tmpFile) Seek(offset int64, whence int) (int64, error) {
	return t.file.Seek(offset, whence)
}

func (t *tmpFile) Stat() (fs.FileInfo, error) {
	return t.meta, nil
}

func (t *tmpFile) Read(bytes []byte) (int, error) {
	return t.file.Read(bytes)
}

func (t *tmpFile) Close() error {
	return t.file.Close()
}

type StreamReader interface {
	io.Reader
	// Name of the stream, e.g. the original file name
	Name() string
	// Receiver component of the stream, e.g. if a component has requested a file upload it likely contains
	// additional callbacks. The security is ensured using the secret and unique ScopeID.
	Receiver() ora.Ptr
	ScopeID() ora.ScopeID
}

// FileReceiver must be implemented by components which requested file uploads.
// The receiver is called from the event loop, thus if you need to block for a long time, you must run that
// within a different executor.
// Small files and fast processing times are usually never a problem, because we don't need to invalidate within
// millisecond range as mobile apps itself.
// Note, that you must close the file, because the scope don't if you have spawned a concurrent go routine.
type FileReceiver interface {
	OnFileReceived(f fs.File)
}

type tmpFileInfo struct {
	AbsolutePath string      `json:"absolutePath"`
	FName        string      `json:"name"`
	FSize        int64       `json:"size"`
	Hash         string      `json:"hash"`
	CreatedAt    time.Time   `json:"createdAt"`
	SeqNum       int64       `json:"seqNum"`
	Scope        ora.ScopeID `json:"scope"`
	Receiver     ora.Ptr     `json:"receiver"`
}

func (s tmpFileInfo) Name() string {
	return s.FName
}

func (s tmpFileInfo) Size() int64 {
	return s.FSize
}

func (s tmpFileInfo) Mode() fs.FileMode {
	return 0
}

func (s tmpFileInfo) ModTime() time.Time {
	return s.CreatedAt
}

func (s tmpFileInfo) IsDir() bool {
	return false
}

func (s tmpFileInfo) Sys() any {
	return s
}
