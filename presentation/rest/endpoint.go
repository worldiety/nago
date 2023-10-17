package rest

import (
	"io"
	"net/http"
)

type Route struct {
	Pattern string
	Method  string
	Handler http.HandlerFunc
}

func HandleFileUpload(maxMemory int64, f func(name string, size int64, r io.ReaderAt) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(maxMemory); err != nil {
			panic("fix me" + err.Error())
		}

		form := r.MultipartForm
		for key := range form.File {
			file, fileHeader, err := r.FormFile(key)
			if err != nil {
				panic("fix me" + err.Error())
			}

			if err := f(fileHeader.Filename, fileHeader.Size, file); err != nil {
				_ = file.Close()
				panic("fix me" + err.Error())
			}

			_ = file.Close()
		}
	}
}
