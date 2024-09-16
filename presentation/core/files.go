package core

import (
	"context"
	"go.wdy.de/nago/pkg/blob"
	"io"
	"io/fs"
	"iter"
)

type blobIterFile struct {
	store blob.Store
	key   string
}

func (b blobIterFile) Transfer(dst io.Writer) (int64, error) {
	optR, err := b.store.NewReader(context.Background(), b.key)
	if err != nil {
		return 0, err
	}

	if optR.IsNone() {
		return 0, fs.ErrNotExist
	}

	reader := optR.Unwrap()
	defer reader.Close()

	return io.Copy(dst, reader)
}

func (b blobIterFile) MimeType() (string, bool) {
	return "", false
}

func (b blobIterFile) Size() (int64, bool) {
	return 0, false
}

func (b blobIterFile) Name() string {
	return b.key
}

// FilesIter opens a transaction and yields the entries. The iterator must consume each file exactly within
// the yield and must neither retain a File nor the io.ReadCloser from Open.
func FilesIter(src blob.Store) iter.Seq2[File, error] {
	return func(yield func(File, error) bool) {
		for key, err := range src.List(context.Background(), blob.ListOptions{}) {
			if err != nil {
				if !yield(nil, err) {
					return
				}
			}

			if !yield(blobIterFile{
				store: src,
				key:   key,
			}, nil) {
				return
			}
		}
	}
}
