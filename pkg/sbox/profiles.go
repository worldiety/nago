// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package sbox

import "time"

// The profiles below are ready-made [Profile] values for the initial use cases.
// They deliberately never bind the application's data directory, so secrets and
// the database remain structurally unreachable from the untrusted process.
//
// All profiles use RootMinimal so that dynamically linked tools and TLS work
// out of the box, enable landlock and the strict seccomp policy, and expose
// only the explicitly listed working directories as writable.

// GoBuild returns a profile suited to running the go compiler ("go build").
//
//   - goroot is the GOROOT (containing the toolchain), exposed read-only.
//   - gocache is the build/module cache directory, exposed read-write.
//   - workdir is the module/source directory, exposed read-write (the build may
//     write generated files); pass a read-only copy if you need hermeticity.
//
// Network access defaults to the host so that "go build" can fetch modules; set
// Net to NetNone on the returned profile for a fully offline build.
func GoBuild(goroot, gocache, workdir string) Profile {
	return Profile{
		RootFS: RootMinimal,
		Binds: []Bind{
			{Host: goroot, Writable: false},
			{Host: gocache, Writable: true},
			{Host: workdir, Writable: true},
		},
		Env: []string{
			"GOROOT=" + goroot,
			"GOCACHE=" + gocache,
			"GOMODCACHE=" + gocache + "/mod",
			"HOME=" + workdir,
			"PATH=/usr/bin:/bin:" + goroot + "/bin",
		},
		WorkDir:  workdir,
		Net:      NetHost,
		Seccomp:  SeccompStrict,
		Landlock: true,
		Limits: Limits{
			Wall:   10 * time.Minute,
			NoFile: 4096,
		},
	}
}

// GoTest returns a profile for "go test". It is like [GoBuild] but keeps the
// host network (tests may bind to localhost); switch Net to NetLoopback to
// allow localhost sockets without outbound connectivity.
func GoTest(goroot, gocache, workdir string) Profile {
	p := GoBuild(goroot, gocache, workdir)
	p.Limits.Wall = 15 * time.Minute
	return p
}

// GoVet returns a profile for "go vet". Vet is read-mostly, so the working
// directory is still writable (vet needs a build cache) but the timeout is
// shorter.
func GoVet(goroot, gocache, workdir string) Profile {
	p := GoBuild(goroot, gocache, workdir)
	p.Limits.Wall = 5 * time.Minute
	return p
}

// Git returns a profile for running git against a working directory.
//
//   - workdir is the repository/checkout directory, exposed read-write.
//
// Network access defaults to the host so that clone/fetch/push work. Add binds
// for a read-only git config or SSH known_hosts as needed via the returned
// profile.
func Git(workdir string) Profile {
	return Profile{
		RootFS: RootMinimal,
		Binds: []Bind{
			{Host: workdir, Writable: true},
		},
		Env: []string{
			"HOME=" + workdir,
			"PATH=/usr/bin:/bin",
			"GIT_TERMINAL_PROMPT=0",
		},
		WorkDir:  workdir,
		Net:      NetHost,
		Seccomp:  SeccompStrict,
		Landlock: true,
		Limits: Limits{
			Wall:   10 * time.Minute,
			NoFile: 4096,
		},
	}
}

// UntrustedServer returns a profile for running an untrusted go webserver that
// is to be inspected. The binary and its working directory are exposed, and the
// network defaults to NetLoopback so the server can bind to localhost inside
// its own network namespace without reaching the outside network. Set Net to
// NetHost if the server must be reachable from the host.
func UntrustedServer(workdir string) Profile {
	return Profile{
		RootFS: RootMinimal,
		Binds: []Bind{
			{Host: workdir, Writable: true},
		},
		Env: []string{
			"HOME=" + workdir,
			"PATH=/usr/bin:/bin",
		},
		WorkDir:  workdir,
		Net:      NetLoopback,
		Seccomp:  SeccompStrict,
		Landlock: true,
		Limits: Limits{
			Wall:   0, // servers are long-lived; caller controls lifetime via ctx
			NoFile: 8192,
		},
	}
}
