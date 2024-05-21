package tmpfs

import (
	"io/fs"
	"time"
)

type FileInfo struct {
	FName         string    `json:"name"`
	FResourceName string    `json:"resourceName"`
	FSize         int64     `json:"size"`
	FHash         string    `json:"hash"`
	CreatedAt     time.Time `json:"createdAt"`
	SeqNum        int64     `json:"seqNum"`
}

// ResourceName is either empty or returns the most original file or resource name which addresses the actual binary
// data. See also [DirEntry].
func (s FileInfo) ResourceName() string {
	return s.FResourceName
}

func (s FileInfo) Hash() string {
	return s.FHash
}

func (s FileInfo) Name() string {
	return s.FName
}

func (s FileInfo) Size() int64 {
	return s.FSize
}

func (s FileInfo) Mode() fs.FileMode {
	return 0
}

func (s FileInfo) ModTime() time.Time {
	return s.CreatedAt
}

func (s FileInfo) IsDir() bool {
	return false
}

func (s FileInfo) Sys() any {
	return s
}

type fakeDir string

func (f fakeDir) Name() string       { return string(f) }
func (f fakeDir) Size() int64        { return 0 }
func (f fakeDir) Mode() fs.FileMode  { return fs.ModeDir | 0500 }
func (f fakeDir) ModTime() time.Time { return time.Unix(0, 0) }
func (f fakeDir) IsDir() bool        { return true }
func (f fakeDir) Sys() any           { return nil }

func (f fakeDir) String() string {
	return fs.FormatFileInfo(f)
}
