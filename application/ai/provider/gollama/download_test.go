// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gollama

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

// A download runs once for all requests which wait for it, and a stopped request neither cancels it nor waits.
func TestAwaitDownloadOutlivesRequester(t *testing.T) {
	dir := t.TempDir()
	entry := catalogEntry{ID: "m", File: "m.gguf", HFRepo: "x/y"}

	release := make(chan struct{})
	var calls atomic.Int32
	cancelledDownload := make(chan bool, 1)
	e := newEngine(Settings{})
	e.download = func(ctx context.Context, entry catalogEntry, storageDir, token string) error {
		calls.Add(1)
		<-release
		cancelledDownload <- ctx.Err() != nil
		return os.WriteFile(filepath.Join(storageDir, entry.File), []byte("gguf"), 0o644)
	}

	// the first request starts the download and gives up waiting
	ctx, cancel := context.WithCancel(context.Background())
	first := make(chan error, 1)
	go func() { first <- e.awaitDownload(ctx, entry, dir) }()
	time.Sleep(20 * time.Millisecond)
	cancel()
	if err := <-first; !errors.Is(err, context.Canceled) {
		t.Fatalf("expected the first request to give up, got %v", err)
	}

	// the second request joins the running download
	second := make(chan error, 1)
	go func() { second <- e.awaitDownload(context.Background(), entry, dir) }()
	time.Sleep(20 * time.Millisecond)
	close(release)

	if err := <-second; err != nil {
		t.Fatal(err)
	}

	if calls.Load() != 1 {
		t.Fatalf("expected a single download, got %d", calls.Load())
	}
	if <-cancelledDownload {
		t.Fatal("the download has been cancelled by the stopped request")
	}
	if !isRegularFile(filepath.Join(dir, entry.File)) {
		t.Fatal("the file has not been downloaded")
	}

	// a present file is not downloaded again
	if err := e.awaitDownload(context.Background(), entry, dir); err != nil || calls.Load() != 1 {
		t.Fatalf("expected no further download: %v %d", err, calls.Load())
	}
}
