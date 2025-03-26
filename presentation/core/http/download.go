// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package http

import (
	"io"
	"mime/multipart"
	"net/http"
)

// HandleHTTP provides classic http inter-operation with this page. This is required e.g. for file uploads
// using multipart forms etc.
// deprecated: entirely the wrong place for this
func HandleHTTP(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	pageToken := r.Header.Get("x-page-token")
	if pageToken == "" {
		pageToken = query.Get("page")
	}

	//if pageToken != p.token.Get() {
	//	w.WriteHeader(http.StatusNotFound)
	//	return
	//}
	// TODO where and how to handle that???
	/*
		switch r.URL.Path {
		case "/api/v1/upload":
			if r.Method != http.MethodPost {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}

			uploadToken := UploadToken(r.Header.Get("x-upload-token"))
			handler := p.renderState.uploads[uploadToken]
			if handler == nil || handler.onUploadReceived == nil {
				logging.FromContext(r.Context()).Warn("upload received but have no handler", slog.String("upload-token", string(uploadToken)))
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			if err := r.ParseMultipartForm(p.maxMemory); err != nil {
				logging.FromContext(r.Context()).Warn("cannot parse multipart form", slog.Any("err", err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			var files []FileUpload
			for _, headers := range r.MultipartForm.File {
				for _, header := range headers {
					files = append(files, httpMultipartFile{header: header})
				}
			}

			handler.onUploadReceived(files)
			p.Invalidate() // TODO race condition?!?!
		case "/api/v1/download":
			if r.Method != http.MethodGet {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}

			downloadToken := DownloadToken(r.Header.Get("x-download-token"))
			if downloadToken == "" {
				downloadToken = DownloadToken(query.Get("download"))
			}

			opener := p.renderState.downloads[downloadToken]
			if opener == nil {
				logging.FromContext(r.Context()).Warn("download request received but have no handler", slog.String("download-token", string(downloadToken)))
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			reader, err := opener()
			if err != nil {
				logging.FromContext(r.Context()).Warn("download request received but cannot open stream", slog.String("download-token", string(downloadToken)), slog.Any("err", err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if _, err := io.Copy(w, reader); err != nil {
				logging.FromContext(r.Context()).Warn("download request received but cannot complete data transfer", slog.String("download-token", string(downloadToken)), slog.Any("err", err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

		}

	*/
}

type httpMultipartFile struct {
	header *multipart.FileHeader
}

func (h httpMultipartFile) Size() int64 {
	return h.header.Size
}

func (h httpMultipartFile) Name() string {
	return h.header.Filename
}

func (h httpMultipartFile) Open() (io.ReadSeekCloser, error) {
	return h.header.Open()
}

func (h httpMultipartFile) Sys() any {
	return h.header
}
