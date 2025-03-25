package swagger

import (
	"io/fs"
)

func Dist() fs.FS {
	return files
}
