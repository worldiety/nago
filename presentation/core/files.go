package core

import (
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/iter"
	"io"
)

type blobIterFile struct {
	tx    blob.Tx
	entry blob.Entry
}

func (b blobIterFile) Open() (io.ReadCloser, error) {
	return b.entry.Open()
}

func (b blobIterFile) Name() string {
	return b.entry.Key
}

// FilesIter opens a transaction and yields the entries. The iterator must consume each file exactly within
// the yield and must neither retain a File nor the io.ReadCloser from Open.
func FilesIter(src blob.Store) iter.Seq2[File, error] {
	return func(yield func(File, error) bool) {
		var cancelled bool
		err := src.View(func(tx blob.Tx) error {
			var err error
			tx.Each(func(entry blob.Entry, e error) bool {
				if e != nil {
					cancelled = true
					err = e
					return false
				}

				if !yield(blobIterFile{
					tx:    tx,
					entry: entry,
				}, e) {
					cancelled = true
					return false
				}

				return true
			})

			return err
		})

		if err != nil && !cancelled {
			yield(nil, err)
		}
	}
}
