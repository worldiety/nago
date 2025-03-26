// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package httpimage

import (
	"go.wdy.de/nago/application/image"
	"go.wdy.de/nago/auth/iam"
	"go.wdy.de/nago/presentation/core"
	"io"
	"log/slog"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// Endpoint returns the default image endpoint.
const Endpoint = "/api/nago/v1/image"

// NewURL assembles a URL with the query url encoded parameters. See also [NewHandler] and keep in mind,
// that [image.ID] may resolve to either a [image.SrcSet] to pick from or a specific image. In either case,
// an actual image is resolved, but maybe in a different resolution.
func NewURL(apiPath string, imgOrSrcSet image.ID, fit image.ObjectFit, width, height int) string {
	values := url.Values{}
	values.Add("src", string(imgOrSrcSet))
	values.Set("fit", fit.String())
	values.Set("w", strconv.Itoa(width))
	values.Set("h", strconv.Itoa(height))

	return apiPath + "?" + values.Encode()
}

// URI uses the default Endpoint. See [NewURL].
func URI(imgOrSrcSet image.ID, fit image.ObjectFit, width, height int) core.URI {
	if imgOrSrcSet == "" {
		return ""
	}

	return core.URI(NewURL(Endpoint, imgOrSrcSet, fit, width, height))
}

// NewHandler uses the image src set loader use case and provides a http contract on it.
// The defined query parameters are:
//   - src: required string, either an image from a [blob.Store] or a [image.SrcSet] from a [image.Repository].
//   - fit: optional string enum, one of [image.FitCover]. If required, there is more to come.
//   - w: optional int, fit width in Pixel
//   - h: optional int, fit height in Pixel
func NewHandler(loadFit image.LoadBestFit) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		srcImgId := image.ID(query.Get("src"))
		wantedWidth, wantedHeight := query.Get("w"), query.Get("h")

		width, err := strconv.Atoi(wantedWidth)
		if err != nil {
			width = math.MaxInt
		}

		height, err := strconv.Atoi(wantedHeight)
		if err != nil {
			height = math.MaxInt
		}

		var fit image.ObjectFit
		switch query.Get("fit") {
		case image.FitCover.String():
			fit = image.FitCover
		case image.FitNone.String():
			fit = image.FitNone
		}

		optReader, err := loadFit(iam.FromContext(r.Context()), srcImgId, fit, width, height)
		if err != nil {
			http.Error(w, "blob error", http.StatusInternalServerError)
			slog.Error("cannot load image blob", "err", err, "src", srcImgId)
			return
		}

		if optReader.IsNone() {
			http.Error(w, "no such image available", http.StatusNotFound)
			return
		}

		// we can be this aggressive, because we assign each image a unique ID anyway
		w.Header().Set("Cache-Control", "public, max-age=31536000")
		// this mimetype is not specified, however it is explicitly allowed,
		// see also https://www.w3.org/Protocols/rfc1341/4_Content-Type.html
		w.Header().Set("Content-Type", "image/*")
		expires := time.Now().Add(365 * 24 * time.Hour)
		w.Header().Set("Expires", expires.Format(http.TimeFormat))

		reader := optReader.Unwrap()
		defer reader.Close()

		if _, err := io.Copy(w, reader); err != nil {
			slog.Error("cannot write image into response", "err", err, "src", srcImgId)
			return
		}
	}
}
