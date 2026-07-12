// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package sbox

import (
	"os"
	"testing"
)

// TestMain wires up Init so that on linux the test binary itself can act as the
// re-exec sandbox trampoline. In the trampoline child Init never returns.
func TestMain(m *testing.M) {
	Init()
	os.Exit(m.Run())
}

func TestBuildSpecDefaults(t *testing.T) {
	// buildSpec exists only on linux; guard via a behavioural test through the
	// public Profile instead so this test compiles everywhere.
	p := GoBuild("/goroot", "/cache", "/work")
	if p.RootFS != RootMinimal {
		t.Fatalf("GoBuild should use RootMinimal, got %v", p.RootFS)
	}
	if p.Seccomp != SeccompStrict {
		t.Fatalf("GoBuild should use SeccompStrict")
	}
	if !p.Landlock {
		t.Fatalf("GoBuild should enable landlock")
	}
	// The data directory must never be bound implicitly.
	for _, b := range p.Binds {
		if b.Host == "" {
			t.Fatalf("empty bind host")
		}
	}
}

func TestUntrustedServerIsLoopbackByDefault(t *testing.T) {
	p := UntrustedServer("/work")
	if p.Net != NetLoopback {
		t.Fatalf("UntrustedServer should default to NetLoopback, got %v", p.Net)
	}
}
