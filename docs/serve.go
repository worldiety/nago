// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"embed"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"os"
)

//go:embed nago.dev/public/*
var embeddedFiles embed.FS

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := "0.0.0.0:" + port

	subFS, err := fs.Sub(embeddedFiles, "nago.dev/public")
	if err != nil {
		log.Fatalf("failed to create sub FS: %v", err)
	}

	fsHandler := http.FileServer(http.FS(subFS))
	http.Handle("/", fsHandler)

	slog.Info("serving doc", "addr", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
