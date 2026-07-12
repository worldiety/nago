// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package sbox provides a lightweight way to execute untrusted processes from a
// NAGO application with strong isolation, using only Linux kernel primitives
// (user/mount/pid/ipc/uts network namespaces, pivot_root, landlock and
// seccomp-bpf) and no external programs such as bubblewrap.
//
// # Threat model
//
// A NAGO application typically runs inside a systemd sandbox, but its data
// directory still contains secrets and the database. When such an application
// spawns untrusted processes (the go compiler, git, go test/vet, or untrusted
// go webservers that are to be inspected), a plain exec would let those
// processes read the secrets or attack the host. sbox adds a second, tighter
// isolation layer inside the application's own privileges so that:
//
//   - the data directory (secrets, database) is structurally unreachable,
//     both via mount isolation (the path does not exist inside the namespace)
//     and via landlock (no access right is granted to it);
//   - privilege escalation is blocked (no_new_privs + seccomp);
//   - dangerous syscalls (ptrace, bpf, keyctl, kexec, mount, ...) are denied.
//
// # Required initialization
//
// Applying mount/pivot_root/landlock/seccomp after fork and before execve is
// not possible from the multi-threaded Go runtime. sbox therefore uses a
// re-exec trampoline, like nsjail and bubblewrap: [Run] re-executes the current
// binary into a hidden sandbox-init mode. For this to work the application MUST
// call [Init] as the very first statement in main:
//
//	func main() {
//		sbox.Init() // must be first; takes over and never returns in sandbox mode
//		// ... normal application startup ...
//	}
//
// Without this call, [Run] fails with [ErrNotInitialized].
//
// # Platform support
//
// A real sandbox is only provided on linux/amd64 and linux/arm64 running on a
// recent kernel (Ubuntu 24.04+, kernel 6.8, landlock ABI v4+). On other
// platforms [Run] uses an UNSANDBOXED passthrough that logs a loud warning on
// every invocation. Because production always runs on linux, the insecure stub
// can never be compiled into a production binary. Setting the environment
// variable SBOX_REQUIRE_ISOLATION=1 makes the stub refuse to run at all.
//
// # Isolation modes
//
// [Profile.Isolation] selects how the sandbox is built:
//
//   - [IsolationNamespaces] (default): the strong mode. Fresh
//     user/mount/pid/ipc/uts (and optionally net) namespaces, a private root
//     filesystem via pivot_root, plus landlock and seccomp. The data directory
//     is protected redundantly (it is absent from the mount namespace AND
//     denied by landlock). This mode needs unprivileged user namespaces and
//     mount to be available.
//
//   - [IsolationLandlockOnly]: a privilege-free mode that applies only
//     landlock, seccomp, no_new_privs and rlimits before exec — no namespaces,
//     no mount, no pivot_root. Use it when the calling process runs under a
//     confinement that forbids namespace creation and mount. Here the data
//     directory is protected by landlock ALONE (landlock governs only
//     filesystem access), there is no PID/network namespace, and Net is treated
//     as NetHost.
//
// # systemd compatibility
//
// A hardened systemd unit commonly forbids exactly what [IsolationNamespaces]
// needs, causing clone(2) to fail with EPERM ("operation not permitted") when
// [Run] starts the trampoline. The relevant directives are RestrictNamespaces=
// (blocklisting user/mnt/pid/... namespaces), DynamicUser=/PrivateUsers= (only
// one uid is mapped, so a nested user namespace is impossible),
// SystemCallFilter= denying @mount, and CapabilityBoundingSet= dropping
// CAP_SYS_ADMIN/CAP_SETUID/CAP_SETGID.
//
// Such units can keep all of that hardening and use [IsolationLandlockOnly],
// which requires none of those capabilities: the landlock and seccomp syscalls
// are not in the denied @mount/@privileged groups and prctl(PR_SET_NO_NEW_PRIVS)
// is permitted. It is recommended to complement it with ProtectProc=,
// PrivatePIDs=, PrivateIPC= and PrivateTmp= to cover the vectors that landlock
// alone does not (process inspection, IPC, shared /tmp).
package sbox
