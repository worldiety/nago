package blob

import (
	"go.wdy.de/nago/pkg/blob"
	"io"
	"io/fs"
)

// NewFS wraps the given blob Store and provides all contained blobs as fs.File.
// Note, that the current implementation tries not to interpret directories from names,
// thus it may violate any naming rules of the fs.FS contract.
// WARNING: This may change in the future, to comply
func NewFS(store blob.Store) fs.FS {
	return &blobFS{store: store}
}

// Import copies and updates all blobs based on the given fs.FS.
func Import(dst blob.Store, src fs.FS) error {
	return dst.Update(func(tx blob.Tx) error {
		return fs.WalkDir(src, ".", func(path string, d fs.DirEntry, err error) error {
			if d.IsDir() || err != nil {
				return err
			}
			info, err := d.Info()
			if err != nil {
				return err
			}

			name := path
			if info, ok := info.(interface{ ResourceName() string }); ok {
				name = info.ResourceName()
			}

			return tx.Put(blob.Entry{
				Key: name,
				Open: func() (io.ReadCloser, error) {
					return src.Open(path)
				},
			})
		})
	})
}
