// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build linux

package sbox

// envInitMode is set by the parent on the re-executed child so that Init can
// detect that it must take over as the sandbox trampoline.
const envInitMode = "SBOX_INIT_MODE"

// envReady is set by the parent so that Run can detect (in the parent) whether
// Init has been wired up at all. Init clears/echoes it; see run().
const envReady = "SBOX_READY"

// specFD is the fixed file descriptor number on which the parent passes the
// JSON-encoded spec to the trampoline child. It is placed right after
// stdin/stdout/stderr. Using a pipe (instead of argv) keeps the spec out of
// the process table and avoids argv length limits.
const specFD = 3

// spec is the wire format handed from the parent to the trampoline child. It is
// an internal, versionless struct: both sides are the same binary, so the
// encoding never needs to be stable across versions.
type spec struct {
	RootFS         int
	Binds          []specBind
	Env            []string
	WorkDir        string
	Hostname       string
	Net            int
	AllowNewUserNS bool
	Seccomp        int
	Landlock       bool

	RLimits specRLimits

	Path string
	Args []string

	// resolvedBinds is populated inside the trampoline (not serialized) and
	// holds bind sources pre-opened as O_PATH fds.
	resolvedBinds []resolvedBind `json:"-"`
}

type specBind struct {
	Host     string
	Target   string
	Writable bool
	Optional bool
}

type specRLimits struct {
	CPUSeconds   uint64
	AddressSpace uint64
	NoFile       uint64
	NProc        uint64
	FileSize     uint64
}
