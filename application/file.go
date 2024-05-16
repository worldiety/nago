package application

import (
	"fmt"
	"io"
	"io/fs"
	"os"
)

func copyFile(fsys fs.FS, path string, dstFname string) error {
	file, err := fsys.Open(path)
	if err != nil {
		return fmt.Errorf("src file %s open error: %v", path, err)
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
