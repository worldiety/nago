package vuejs

import (
	"embed"
	"io/fs"
)

//go:embed dist/favicon.ico dist/index.html dist/assets/*
var Frontend embed.FS

func Dist() fs.FS {
	fsys, err := fs.Sub(Frontend, "dist")
	if err != nil {
		panic(err)
	}
	return fsys
}
