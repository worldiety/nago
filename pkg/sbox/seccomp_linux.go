// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build linux

package sbox

import (
	seccomp "github.com/elastic/go-seccomp-bpf"
	"github.com/elastic/go-seccomp-bpf/arch"
)

// filterKnown drops syscall names that do not exist on the current
// architecture. The seccomp assembler rejects the entire policy if it contains
// a single unknown name (e.g. aarch64 has no "umount" or "_sysctl"), so we
// prune the list to what the running arch actually knows.
func filterKnown(names []string) []string {
	info, err := arch.GetInfo("")
	if err != nil {
		return names
	}
	out := names[:0:0]
	for _, n := range names {
		if _, ok := info.SyscallNames[n]; ok {
			out = append(out, n)
		}
	}
	return out
}

// applySeccomp installs a seccomp-bpf filter for the sandboxed process.
//
// The policy is a blocklist: the default action is Allow, and a curated set of
// dangerous syscalls is denied. A pure allowlist would be more principled, but
// it is brittle for arbitrary workloads such as the go toolchain, git and
// untrusted webservers, which use a very broad syscall surface. The blocklist
// removes the classes of syscalls that enable sandbox escapes, tracing, kernel
// manipulation and privilege games, while leaving normal operation intact. The
// mount, user-namespace, landlock and pivot_root isolation remain the primary
// containment; seccomp is defense in depth on top.
func applySeccomp(mode SeccompMode, allowNewUserNS bool) error {
	denied := filterKnown(deniedSyscalls(mode, allowNewUserNS))

	filter := seccomp.Filter{
		NoNewPrivs: true,
		Flag:       seccomp.FilterFlagTSync,
		Policy: seccomp.Policy{
			DefaultAction: seccomp.ActionAllow,
			Syscalls: []seccomp.SyscallGroup{
				{
					Action: seccomp.ActionErrno,
					Names:  denied,
				},
			},
		},
	}

	return seccomp.LoadFilter(filter)
}

// deniedSyscalls returns the syscalls blocked for the given mode. Unknown
// syscall names for the target architecture are silently dropped by the
// assembler-level lookup, so listing a name that does not exist on arm64/amd64
// is harmless.
func deniedSyscalls(mode SeccompMode, allowNewUserNS bool) []string {
	// Always denied: these have no legitimate use for the target workloads and
	// are classic escape / attack primitives.
	denied := []string{
		// Tracing / debugging other processes.
		"ptrace",
		"process_vm_readv",
		"process_vm_writev",
		"kcmp",
		// Kernel modules and kernel manipulation.
		"init_module",
		"finit_module",
		"delete_module",
		"kexec_load",
		"kexec_file_load",
		// Kernel keyring.
		"add_key",
		"request_key",
		"keyctl",
		// eBPF and perf.
		"bpf",
		"perf_event_open",
		// userfaultfd is a common exploitation primitive.
		"userfaultfd",
		// Swap and reboot.
		"swapon",
		"swapoff",
		"reboot",
		// Filesystem administration that could break out of the mount ns.
		"mount",
		"umount",
		"umount2",
		"pivot_root",
		"move_mount",
		"open_tree",
		"fsopen",
		"fsconfig",
		"fsmount",
		"mount_setattr",
		"chroot",
		// NUMA / memory policy that can affect the host.
		"mbind",
		// Obsolete / dangerous.
		"uselib",
		"personality",
		"acct",
		"quotactl",
		// Time and clock manipulation affects the whole host.
		"settimeofday",
		"clock_settime",
		"clock_adjtime",
		"adjtimex",
		// System-wide administration.
		"sethostname",
		"setdomainname",
		"nfsservctl",
		"_sysctl",
		// Namespace juggling beyond what we already set up.
		"setns",
		"unshare",
	}

	if !allowNewUserNS {
		// clone/clone3 remain allowed (needed for threads/processes) but new
		// user namespaces are dangerous; we can't easily filter clone flags by
		// name here, so unshare/setns above already block the common path.
		// Nothing extra to add.
		_ = allowNewUserNS
	}

	if mode == SeccompPermissive {
		// Permissive keeps only the most critical denials.
		return []string{
			"ptrace",
			"bpf",
			"kexec_load",
			"kexec_file_load",
			"init_module",
			"finit_module",
			"delete_module",
			"add_key",
			"request_key",
			"keyctl",
			"mount",
			"umount2",
			"pivot_root",
			"setns",
			"unshare",
			"reboot",
			"swapon",
			"swapoff",
		}
	}

	return denied
}
