package tmpfs

import (
	"io/fs"
	"os"
)

type File struct {
	meta FileInfo
	file *os.File
}

func NewFile(meta FileInfo, file *os.File) *File {
	return &File{meta: meta, file: file}
}

func (t *File) ReadAt(p []byte, off int64) (n int, err error) {
	return t.file.ReadAt(p, off)
}

func (t *File) Seek(offset int64, whence int) (int64, error) {
	return t.file.Seek(offset, whence)
}

func (t *File) Stat() (fs.FileInfo, error) {
	return t.meta, nil
}

func (t *File) Read(bytes []byte) (int, error) {
	return t.file.Read(bytes)
}

func (t *File) Close() error {
	return t.file.Close()
}
