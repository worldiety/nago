// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application_test

import (
	"errors"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"

	"go.wdy.de/nago/application"
)

// envRunExit makes the test binary a nago application, which runs once; its value is "server" or "oneshot".
const envRunExit = "NAGO_TEST_RUN_EXIT"

// TestRunExitCode: An application, which cannot open its port, ends with exit code 1 and names the error in the log;
// the counter row, a regular end in one shot mode, ends with exit code 0.
func TestRunExitCode(t *testing.T) {
	if mode := os.Getenv(envRunExit); mode != "" {
		runOnce(mode)
		return
	}

	taken, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer taken.Close()
	port := strconv.Itoa(taken.Addr().(*net.TCPAddr).Port)

	tests := []struct {
		name     string
		mode     string
		wantExit int
		wantLog  string
	}{
		{"port taken", "server", 1, "bind"},
		{"one shot, regular end", "oneshot", 0, "successful shutdown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=^TestRunExitCode$")
			cmd.Env = append(os.Environ(), envRunExit+"="+tt.mode, "HOST=127.0.0.1", "PORT="+port, "NAGO_PORT=", "STATE_DIRECTORY="+t.TempDir())
			out, err := cmd.CombinedOutput()
			exit := 0
			var ee *exec.ExitError
			if errors.As(err, &ee) {
				exit = ee.ExitCode()
			} else if err != nil {
				t.Fatal(err)
			}
			if exit != tt.wantExit {
				t.Fatalf("exit %d, want %d; output:\n%s", exit, tt.wantExit, out)
			}
			if !strings.Contains(string(out), tt.wantLog) {
				t.Fatalf("output without %q:\n%s", tt.wantLog, out)
			}
		})
	}
}

func runOnce(mode string) {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.nago.test.exit")
		if mode == "oneshot" {
			cfg.OneShot()
		}
	}).Run()
}
