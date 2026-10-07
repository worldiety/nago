// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package anthropic

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// A file which is gone already counts as deleted for the cleanup of provider files, other failures do not.
func TestDeleteFileReportsAnUnknownFile(t *testing.T) {
	for status, wantNotExist := range map[int]bool{http.StatusNotFound: true, http.StatusInternalServerError: false} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(status)
		}))

		c := NewClient("token", "", 0, false)
		c.base = srv.URL + "/"
		c.retry = -1
		err := c.DeleteFile("file-1")
		srv.Close()

		if err == nil || errors.Is(err, os.ErrNotExist) != wantNotExist {
			t.Errorf("status %d: unexpected error %v", status, err)
		}
	}
}
