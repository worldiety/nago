// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package drive

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"regexp"
	"sync"
	"time"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/xslices"
	"go.wdy.de/nago/pkg/xtime"
)

type FID string

type FileInfo struct {
	OriginalFilename string                 `json:"n,omitempty"` // OriginalFilename contains optionally the filename from the source, e.g. the name from the uploaded file
	Blob             string                 `json:"b"`
	Sha3H256         Sha3H256               `json:"h"`
	CreatedAt        xtime.UnixMilliseconds `json:"t"`
	Size             int64                  `json:"s"`           // in bytes
	MimeType         string                 `json:"m,omitempty"` //e.g. video/mp4
	Comment          string                 `json:"c,omitempty"`
	CreatedBy        user.ID                `json:"u,omitempty"`
}

// A File is similar to a unix file (or somewhat to an inode). It supports the same semantics for the
// unix owner, group and permission bits but in addition to a conventional file system a huge amount of metadata
// can be attached and also each binary can keep its version history.
// Also note, that most of the use cases also implement resource based permissions to allow fine-grained access control
// similar to the ACL pattern (access control lists).
type File struct {
	ID        FID                     `json:"id"`
	Filename  string                  `json:"n"`
	Entries   xslices.Slice[FID]      `json:"e,omitempty,omitzero"` // Entries are only valid if [os.FileMode.IsDir]
	Owner     user.ID                 `json:"o,omitempty"`
	Group     group.ID                `json:"g,omitempty"`
	DeletedAt xtime.UnixMilliseconds  `json:"d,omitempty"`
	FileMode  os.FileMode             `json:"m,omitempty"`
	History   xslices.Slice[FileInfo] `json:"h,omitempty,omitzero"` // History is only valid if [os.FileMode.IsFile]. The most current data is the last element in the slice. The history is append-only
	Shares    xslices.Slice[Share]    `json:"s,omitempty,omitzero"` // TODO create an inverse share index for lookup
	repo      Repository
}

func (f File) Name() string {
	return f.Filename
}

func (f File) Size() int64 {
	if f.History.Len() > 0 {
		return f.History.At(f.History.Len() - 1).Size
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

	if f.FileMode&OtherRead != 0 {
		// file is world readable
		return true
	}

	if f.Owner != "" && f.Owner == subject.ID() {
		// owners can always read
		return true
	}

	if f.Group != "" && subject.HasGroup(f.Group) && f.FileMode&0040 != 0 {
		// subject is group member and group is allowed to read
		return true
	}

	if f.repo != nil {
		if subject.HasResourcePermission(f.repo.Name(), string(f.ID), PermOpenFile) {
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
)

func (f File) CanWrite(subject auth.Subject) bool {
	if f.ID == "" {
		return false
	}

	if f.FileMode&OtherWrite != 0 {
		// file is world writeable
		return true
	}

	if f.Owner != "" && f.Owner == subject.ID() {
		// owners can always write
		return true
	}

	if f.Group != "" && subject.HasGroup(f.Group) && f.FileMode&0020 != 0 {
		// subject is group member and group is allowed to write
		return true
	}

	if f.repo != nil {
		if subject.HasResourcePermission(f.repo.Name(), string(f.ID), PermOpenFile) {
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
	if f.History.Len() > 0 {
		return time.Unix(int64(f.History.At(f.History.Len()-1).CreatedAt), 0)
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

type ShareID string
type Share struct {
	ID          ShareID                `json:"id"`
	SharedUntil xtime.UnixMilliseconds `json:"sharedUntil"` // zero value means unlimited
	Algorithm   user.HashAlgorithm     `json:"algorithm,omitempty"`
	TokenHash   []byte                 `json:"tokenHash,omitempty"` // TokenHash is the derives password, the same limits apply as the for the normal token usage. See also [user.Password.TokenHash]
	Users       xslices.Slice[user.ID] `json:"users,omitempty"`     // may be empty, but if not the user must be authenticated and one of the denoted ones
	File        FID                    `json:"file,omitempty"`      // File refers to the shared object
	CanWrite    bool                   `json:"canWrite,omitempty"`  // ByDefault shares a read-only but can be changed to be mutated by others
}

func (s Share) Identity() ShareID {
	return s.ID
}

type ShareRepository data.Repository[Share, ShareID]

type Repository data.Repository[File, FID]

const (
	// FSDrive is the default name for a filesystem root. This shall be used for a users' private drive filesystem
	// or for a commonly shared filesystem.
	FSDrive = "nago.drive"
)

type UserRootRepository data.Repository[UserRoots, user.ID]
type UserRoots struct {
	ID    user.ID
	Roots map[string]FID
}

func (u UserRoots) Identity() user.ID {
	return u.ID
}

type NamedRootRepository data.Repository[NamedRoot, string]

type NamedRoot struct {
	ID   string
	Root FID
}

func (n NamedRoot) Identity() string {
	return n.ID
}

type OpenRootOptions struct {
	User   user.ID     // if empty, uses the global named lookup. If the user is set, it is assigned as the owner and this denotes a private drive.
	Group  group.ID    // if Create is true and Group is set, this group is set as the associated group.
	Name   string      // if empty, uses [FSDrive]
	Create bool        // if true, the root is created automatically. Otherwise [os.ErrNotExists] is returned if no such root exists.
	Mode   os.FileMode // if Create use this Mode for the root element (if not 0). Only the Perm bits are used.
}

// OpenRoot either opens an existing root or creates a new one. The newly created root directory
// has an empty name and the owner is the given uid, if the permission allows it.
type OpenRoot func(subject auth.Subject, opts OpenRootOptions) (File, error)

type Stat func(subject auth.Subject, fid FID) (option.Opt[File], error)

type DeleteOptions struct {
	// If Retention is not 0 the file will be kept at least for the given duration before deletion.
	// Note, that this is the minimum time to keep and not the maximum time, due to scheduler timings or
	// if the service is just not active.
	Retention time.Duration

	// Recursive is only applied if the file denotes a directory.
	Recursive bool
}

// Delete removes the denoted file. It is not an error to remove a non-existing file.
type Delete func(subject auth.Subject, fid FID, opts DeleteOptions) error

// Store adds a blob stream as a new version for the given file.
type Store func(subject auth.Subject, fid FID, src io.Reader) error

type PutOptions struct {
	// If KeepVersion the old file content will be kept in the files` history.
	KeepVersion bool

	// Comment is stored with the version, if [PutOptions.KeepVersion] is true.
	Comment string
}

// Put either creates a new file entry or re-uses an existing one and stores a new version inside the given
// parent directory.
type Put func(subject auth.Subject, parent FID, name string, src io.Reader, opts PutOptions) error

type MkDirOptions struct {
	User  user.ID     // if empty, uses the global named lookup. If the user is set, it is assigned as the owner and this denotes a private drive.
	Group group.ID    // if Create is true and Group is set, this group is set as the associated group.
	Mode  os.FileMode // if Create use this Mode for the root element (if not 0). Only the Perm bits are used.
}

// MkDir creates a new directory within the given parent, if it does not yet exist. If it already exists, it returns
// the existing directory. If a file already exists, an [os.ErrExist] is returned.
type MkDir func(subject auth.Subject, parent FID, name string, opts MkDirOptions) (File, error)

// Load opens the desired version to read.
// An index of -1 will always open the latest blob from history. The returned ReadCloser may also be a
// [io.ReadSeekCloser] but that depends on the given blob store implementation.
type Load func(subject auth.Subject, fid FID, versionIndex int) (io.ReadCloser, error)

// LoadMetaInfo returns the currently available meta information. The access rights are checked against the
// given file identifier.
type LoadMetaInfo func(subject auth.Subject, fid FID, key Sha3H256) (option.Opt[MetaInfo], error)

// ScrapeMetaInfo inspects the given file and specific blob and tries to read, extract, parse and prepare
// as much as it can. This may take a lot of time, as it may involve a lot of time. Any existing [MetaInfo] is
// replaced on success.
type ScrapeMetaInfo func(ctx context.Context, subject auth.Subject, fid FID, key Sha3H256) (option.Opt[MetaInfo], error)

// OpenFS represents the underlying data using the file system contract.
type OpenFS func(subject auth.Subject, parent FID) (fs.FS, error)

// Chown changes the owner of the given file.
type Chown func(subject auth.Subject, uid user.ID, fid FID) error

// Chgrp changes the group of the given file.
type Chgrp func(subject auth.Subject, gid group.ID, fid FID) error

// Chmod sets the virtual and portable permission bits.
type Chmod func(subject auth.Subject, mode os.FileMode, fid FID) error

type Drives struct {
	Private map[string]FID
	Global  map[string]FID
	Shared  []FID
}

// ReadDrives returns those file roots which are either defined as a drive root or declared by a share.
type ReadDrives func(subject auth.Subject, uid user.ID) (Drives, error)

type UseCases struct {
	OpenRoot   OpenRoot
	Stat       Stat
	ReadDrives ReadDrives
	MkDir      MkDir
}

func NewUseCases(repo Repository, globalRootRepo NamedRootRepository, userRootRepo UserRootRepository, fileBlobs blob.Store) UseCases {
	// IMPORTANT: we must ensure that no evil locks occur. No (huge) payload use case call must be stalled or at least must stall other concurrent calls
	var mutex sync.Mutex

	return UseCases{
		OpenRoot:   NewOpenRoot(&mutex, repo, globalRootRepo, userRootRepo),
		Stat:       NewStat(repo),
		ReadDrives: NewReadDrives(globalRootRepo, userRootRepo),
		MkDir:      NewMkDir(&mutex, repo),
	}
}

var validFilename = regexp.MustCompile(`^[^<>:"/\\|?*\x00-\x1F]+$`)

func ValidateName(name string) error {
	if len(name) > 255 || !validFilename.MatchString(name) {
		return fmt.Errorf("invalid file name: %q: %w", name, os.ErrInvalid)
	}

	return nil
}
