package tmpfiles

import (
	"fmt"
	"go.wdy.de/nago/presentation/core"
	"io"
	"os"
	"path/filepath"
)

type tmpFile struct {
	absPath string
	name    string
	size    int64
}

func (d tmpFile) Open() (io.ReadCloser, error) {
	return os.Open(d.absPath)
}

func (d tmpFile) Name() string {
	return d.name
}

func (d tmpFile) Size() int64 {
	return d.size
}

type Files struct {
	scratchDir string
	nextFHnd   int
	files      []tmpFile
}

func New(scratchDir string) (*Files, error) {
	if err := os.MkdirAll(scratchDir, 0700); err != nil {
		return nil, err
	}

	return &Files{scratchDir: scratchDir, nextFHnd: 1}, nil
}

func (f *Files) Import(name string, r io.Reader) (e error) {
	fhnd := f.nextFHnd
	f.nextFHnd++
	tmpFileAbsPath := filepath.Join(f.scratchDir, fmt.Sprintf("%d.tmp", fhnd))
	file, err := os.OpenFile(tmpFileAbsPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("cannot open tmp file for write: %w", err)
	}

	defer func() {
		if err := file.Close(); err != nil && e == nil {
			e = err
		}
	}()

	size, err := io.Copy(file, r)
	if err != nil {
		return fmt.Errorf("cannot copy data: %w", err)
	}

	f.files = append(f.files, tmpFile{
		absPath: tmpFileAbsPath,
		name:    name,
		size:    size,
	})

	return nil
}

func (f *Files) Each(yield func(file core.File, err error) bool) {
	for _, file := range f.files {
		if !yield(file, nil) {
			return
		}
	}
}

// Close removes all files within the given scratchDir.
func (f *Files) Close() error {
	return os.RemoveAll(f.scratchDir)
}
