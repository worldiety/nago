// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package drive

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/xslices"
)

// A File is similar to a unix file (or somewhat to an inode). It supports the same semantics for the
// unix owner, group and permission bits but in addition to a conventional file system a huge amount of metadata
// can be attached and also each binary can keep its version history.
// Also note, that most of the use cases also implement resource based permissions to allow fine-grained access control
// similar to the ACL pattern (access control lists).
type File struct {
	ID FID `json:"id"`
	// the following fields are a snapshot based on the activity audit log
	Filename string               `json:"name"`
	Entries  xslices.Slice[FID]   `json:"entries,omitzero"` // Entries are only valid if [os.FileMode.IsDir]
	Owner    user.ID              `json:"oid,omitempty"`
	Group    group.ID             `json:"gid,omitempty"`
	FileMode os.FileMode          `json:"mode,omitempty"`
	FileInfo option.Opt[FileInfo] `json:"info,omitempty"`   // FileInfo is only valid if ![os.FileMode.IsDir]
	Parent   FID                  `json:"parent,omitempty"` // every file (besides root) has a backward reference to its parent. We will never support hardlinks, because that would break this assumption, but it saves us from other headaches too.
	Shares   xslices.Slice[Share] `json:"shares,omitempty"`

	// AuditLog contains all changes to this file as kind of audit trail. The latest changes are always appended and the
	// fields above are updated to reflect the current state.
	AuditLog xslices.Slice[LogEntry] `json:"log,omitempty"`

	// the repo is used to provide some internal helpers and is only used internally. It is not intended to create a File from the outside.
	repo Repository
}

func (f File) Name() string {
	return f.Filename
}

// AbsolutePath tries to assemble a human readable (but not unique) name of the nested file structure in which
// this file can be found.
func (f File) AbsolutePath() (string, error) {
	var names []string
	leaf := f
	for leaf.Parent != "" {
		names = append(names, leaf.Filename)
		optParent, err := readFileStat(f.repo, leaf.Parent)
		if err != nil {
			return "", err
		}

		if optParent.IsNone() {
			return "", fmt.Errorf("parent is gone: %s: %w", leaf.Parent, os.ErrNotExist)
		}

		leaf = optParent.Unwrap()
	}

	slices.Reverse(names)
	return strings.Join(names, "/"), nil
}

// EntryByName walks over each entry and stats each linked file to inspect its name.
func (f File) EntryByName(name string) (option.Opt[File], error) {
	if f.repo == nil {
		return option.None[File](), fmt.Errorf("file detached: repository is nil")
	}

	for fid := range f.Entries.All() {
		optFile, err := readFileStat(f.repo, fid)
		if err != nil {
			return option.None[File](), err
		}

		if optFile.IsSome() && optFile.Unwrap().Filename == name {
			return optFile, nil
		}
	}

	return option.None[File](), nil
}

func (f File) Size() int64 {
	if f.FileInfo.IsSome() {
		return f.FileInfo.Unwrap().Size
	}

	return 0
}

func (f File) Mode() fs.FileMode {
	return f.FileMode
}

func (f File) CanRead(subject auth.Subject) bool {
	if f.ID == "" {
		return false
	}

	if user.IsSU(subject) {
		return true
	}

	if f.FileMode&OtherRead != 0 {
		// file is world readable
		return true
	}

	if f.Owner != "" && f.Owner == subject.ID() {
		// owners can always read
		return true
	}

	if f.Group != "" && subject.HasGroup(f.Group) && f.FileMode&GroupRead != 0 {
		// subject is group member and group is allowed to read
		return true
	}

	if f.repo != nil {
		if subject.HasResourcePermission(rebac.Namespace(f.repo.Name()), rebac.Instance(f.ID), PermOpenFile) {
			// file object has nago permission
			return true
		}
	}

	// check if it has been shared with the subject
	for share := range f.Shares.All() {
		if xslices.Contains(share.Users, subject.ID()) {
			return true
		}

		// TODO what about a share with token?
	}

	return false
}

const (
	OtherWrite os.FileMode = 0002
	OtherRead  os.FileMode = 0004
	GroupWrite os.FileMode = 0020
	GroupRead  os.FileMode = 0040
	OwnerWrite os.FileMode = 0200
	OwnerRead  os.FileMode = 0400
)

func (f File) CanDelete(subject auth.Subject) bool {
	if f.repo == nil {
		return false
	}

	if user.IsSU(subject) {
		return true
	}

	canDelete := subject.HasResourcePermission(rebac.Namespace(f.repo.Name()), rebac.Instance(f.ID), PermDelete)
	optParent, err := f.repo.FindByID(f.ID)
	if err != nil {
		slog.Error("cannot determine deletability", "fid", f.ID, "err", err)
		return false
	}

	if optParent.IsSome() {
		parent := optParent.Unwrap()
		if parent.CanWrite(subject) {
			canDelete = true
		}
	}

	return canDelete
}

func (f File) CanRename(subject auth.Subject) bool {
	if f.CanWrite(subject) {
		return true
	}

	if user.IsSU(subject) {
		return true
	}

	if subject.HasResourcePermission(rebac.Namespace(f.repo.Name()), rebac.Instance(f.ID), PermRename) {
		return true
	}

	return false
}

func (f File) CanWrite(subject auth.Subject) bool {
	if f.ID == "" {
		return false
	}

	if user.IsSU(subject) {
		return true
	}

	if f.FileMode&OtherWrite != 0 {
		// file is world writeable
		return true
	}

	if f.Owner != "" && f.Owner == subject.ID() {
		// owners can always write
		return true
	}

	if f.Group != "" && subject.HasGroup(f.Group) && f.FileMode&GroupWrite != 0 {
		// subject is group member and group is allowed to write
		return true
	}

	if f.repo != nil {
		if subject.HasResourcePermission(rebac.Namespace(f.repo.Name()), rebac.Instance(f.ID), PermPut) {
			// file object has nago permission
			return true
		}
	}

	// check if it has been shared with the subject
	for share := range f.Shares.All() {
		if share.CanWrite && xslices.Contains(share.Users, subject.ID()) {
			return true
		}

		// TODO what about a share with token?
	}

	return false
}

func (f File) ModTime() time.Time {
	if f.AuditLog.Len() > 0 {
		v, ok := f.AuditLog.Last()
		if !ok {
			panic("audit log is empty")
		}

		ac, ok := v.Unwrap()
		if !ok {
			panic("audit log entry is empty")
		}

		return time.UnixMilli(int64(ac.Mod()))
	}

	return time.Time{}
}

func (f File) IsDir() bool {
	return f.FileMode.IsDir()
}

func (f File) Sys() any {
	return f
}

func (f File) Identity() FID {
	return f.ID
}

// Versions collects all audit events from oldest to newest which added another version. This is only valid for
// file data version if this file does not represent a directory.
func (f File) Versions() []VersionAdded {
	var versions []VersionAdded
	for entry := range f.AuditLog.All() {
		v, ok := entry.Unwrap()
		if !ok {
			continue
		}

		if vadd, ok := v.(VersionAdded); ok {
			versions = append(versions, vadd)
		}
	}

	return versions
}
