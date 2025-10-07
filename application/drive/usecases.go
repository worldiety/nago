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

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/events"
	"go.wdy.de/nago/pkg/xslices"
	"go.wdy.de/nago/pkg/xtime"
	"go.wdy.de/nago/presentation/core"
)

// FID is a file identifier and unique through the entire drive.
type FID string

// BID is version identifier and identifies a specific blob.
type BID string
type FileInfo struct {
	OriginalFilename string   `json:"n,omitempty"` // OriginalFilename contains optionally the filename from the source, e.g. the name from the uploaded file
	Blob             BID      `json:"b"`
	Sha3H256         Sha3H256 `json:"h"`
	Size             int64    `json:"s"`  // in bytes
	MimeType         string   `json:"m,"` //e.g. video/mp4
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
	// Recursive is only applied if the file denotes a directory.
	Recursive bool
}

// Delete removes the denoted file. It is not an error to remove a non-existing file.
type Delete func(subject auth.Subject, fid FID, opts DeleteOptions) error

type PutOptions struct {
	OriginalFilename string
	SourceHint       SourceHint
	// If KeepVersion the old file content will be kept in the files` history.
	KeepVersion bool
	Mode        os.FileMode // only the perm bits are used. If zero, the perm bits from the parent is used.
	Owner       user.ID     // only used when created, otherwise use [Chown]. If empty, the parent Owner is used.
	Group       group.ID    // only used when created, otherwise use [Chgrp]. If empty, the parent Group is used.
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

// Get opens the desired version to read which is identified by the given blob identifier.
// An empty version identifier returns the latest blob from history. The returned File may also return a
// [io.ReadSeekCloser] but that depends on the given blob store implementation.
type Get func(subject auth.Subject, fid FID, version BID) (option.Opt[core.File], error)

// Zip takes the given options and returns a zip file containing all given files in their latest version.
// The implementation may optimize and create the file on the fly, thus browsers may not be able to show
// a download progress.
type Zip func(subject auth.Subject, fids []FID) (core.File, error)

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

// Rename tries to rename the given file. It ensures that the filename is unique and valid in the parent directory,
// even though we actually would not need that limitation.
type Rename func(subject auth.Subject, fid FID, newName string) error

type Drives struct {
	Private map[string]FID
	Global  map[string]FID
	Shared  []FID
}

// ReadDrives returns those file roots which are either defined as a drive root or declared by a share.
type ReadDrives func(subject auth.Subject, uid user.ID) (Drives, error)

type WalkDir func(subject auth.Subject, root FID, walker func(fid FID, file File, err error) error) error

type UseCases struct {
	OpenRoot   OpenRoot
	Stat       Stat
	ReadDrives ReadDrives
	MkDir      MkDir
	Delete     Delete
	WalkDir    WalkDir
	Put        Put
	Get        Get
	Zip        Zip
	Rename     Rename
}

func NewUseCases(bus events.Bus, repo Repository, globalRootRepo NamedRootRepository, userRootRepo UserRootRepository, fileBlobs blob.Store) UseCases {
	// IMPORTANT: we must ensure that no evil locks occur. No (huge) payload use case call must be stalled or at least must stall other concurrent calls
	var mutex sync.Mutex

	walkDirFn := NewWalkDir(repo)

	return UseCases{
		OpenRoot:   NewOpenRoot(&mutex, repo, globalRootRepo, userRootRepo),
		Stat:       NewStat(repo),
		ReadDrives: NewReadDrives(globalRootRepo, userRootRepo),
		MkDir:      NewMkDir(&mutex, bus, repo),
		Delete:     NewDelete(&mutex, bus, repo, walkDirFn, fileBlobs),
		WalkDir:    walkDirFn,
		Put:        NewPut(&mutex, bus, repo, fileBlobs),
		Get:        NewGet(repo, fileBlobs),
		Zip:        NewZip(repo, fileBlobs, walkDirFn),
		Rename:     NewRename(&mutex, bus, repo),
	}
}

var validFilename = regexp.MustCompile(`^[^<>:"/\\|?*\x00-\x1F]+$`)

// ValidateName checks if the given string can be safely used as a file name on common operating systems like
// windows, macos or linux. Even though we may support arbitrary strings, the user can't when downloading
// a file or within a zip file. This also lowers the attack surface at the users side.
func ValidateName(name string) error {
	if len(name) > 255 || !validFilename.MatchString(name) {
		return fmt.Errorf("invalid file name: %q: %w", name, os.ErrInvalid)
	}

	return nil
}
