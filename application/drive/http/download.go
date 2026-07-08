// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package drivehttp exposes an authenticated HTTP endpoint which streams a drive file's binary content so it
// can be referenced by URL-based UI components (e.g. ui.Image or the video.Video player) and for downloads.
package drivehttp

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"go.wdy.de/nago/application/drive"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	corehttp "go.wdy.de/nago/presentation/core/http"
)

// Endpoint is the path under which the drive file streaming handler is served.
const Endpoint = "/api/nago/v1/drive/file"

// URL builds the URL to stream the given drive file. It is usable as a source for ui.Image and video.Video.
func URL(fid drive.FID) string {
	return Endpoint + "?fid=" + url.QueryEscape(string(fid))
}

// DownloadURL builds the URL to download (attachment disposition) the given drive file.
func DownloadURL(fid drive.FID) string {
	return Endpoint + "?fid=" + url.QueryEscape(string(fid)) + "&dl=1"
}

// NewHandler returns an authenticated handler streaming the referenced drive file. Access is authorized by the
// use case itself (drive.Get performs the CanRead check for the resolved subject), so a user can only stream
// files they are allowed to read. When the query parameter "dl" is set the response is sent as an attachment.
func NewHandler(get drive.Get) corehttp.SubjectHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, subject auth.Subject) {
		fid := drive.FID(r.URL.Query().Get("fid"))
		if fid == "" {
			http.Error(w, "missing fid", http.StatusBadRequest)
			return
		}

		optFile, err := get(subject, fid, "")
		if err != nil {
			if errors.Is(err, user.PermissionDeniedErr) {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}

			slog.Error("drive preview: cannot open file", "fid", fid, "err", err.Error())
			http.Error(w, "cannot open file", http.StatusInternalServerError)
			return
		}

		if optFile.IsNone() {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		file := optFile.Unwrap()

		if mime, ok := file.MimeType(); ok && mime != "" {
			w.Header().Set("Content-Type", mime)
		}

		if r.URL.Query().Get("dl") != "" {
			w.Header().Set("Content-Disposition", "attachment; filename=\""+sanitizeFilename(file.Name())+"\"")
		}

		reader, err := file.Open()
		if err != nil {
			slog.Error("drive preview: cannot read file", "fid", fid, "err", err.Error())
			http.Error(w, "cannot read file", http.StatusInternalServerError)
			return
		}
		defer reader.Close()

		// If the underlying reader is seekable (e.g. the filesystem blob store returns *os.File), use
		// http.ServeContent so that HTTP range requests work - this is what lets the video player seek.
		if seeker, ok := reader.(io.ReadSeeker); ok {
			http.ServeContent(w, r, file.Name(), time.Time{}, seeker)
			return
		}

		if size, ok := file.Size(); ok && size > 0 && r.URL.Query().Get("dl") != "" {
			w.Header().Set("Content-Length", strconv.FormatInt(size, 10))
		}

		if _, err := io.Copy(w, reader); err != nil {
			slog.Error("drive preview: cannot stream file", "fid", fid, "err", err.Error())
		}
	}
}

// sanitizeFilename strips characters that could break the Content-Disposition header.
func sanitizeFilename(name string) string {
	replacer := func(r rune) rune {
		if r == '"' || r == '\\' || r == '\n' || r == '\r' {
			return '_'
		}
		return r
	}

	out := make([]rune, 0, len(name))
	for _, r := range name {
		out = append(out, replacer(r))
	}
	return string(out)
}
