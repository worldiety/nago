package application

import (
	"fmt"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/presentation/core"
	"io"
	"iter"
	"os"
)

func copyFile(f core.File, dstFname string) error {
	file, err := f.Open()
	if err != nil {
		return fmt.Errorf("src file %s open error: %v", f.Name(), err)
	}

	defer file.Close()

	dst, err := os.Create(dstFname)
	if err != nil {
		return fmt.Errorf("dst file %s create error: %v", dstFname, err)
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		return fmt.Errorf("dst file %s copy error: %v", dstFname, err)
	}

	return nil
}

func BlobReceiver(dst blob.Store) func(iter.Seq2[core.File, error]) error {
	return func(it iter.Seq2[core.File, error]) error {
		return dst.Update(func(tx blob.Tx) error {
			var err error
			it(func(file core.File, e error) bool {
				if e != nil {
					err = e
					return false
				}

				e = tx.Put(blob.Entry{
					Key:  file.Name(),
					Open: file.Open,
				})

				if e != nil {
					err = e
					return false
				}

				return true
			})

			return err
		})
	}
}
