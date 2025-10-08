// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package drive

import (
	"os"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/xtime"
)

// LogEntry either contains one value or none. It is invalid to represent more than one activity at a time.
type LogEntry struct {
	Deleted      option.Ptr[Deleted]      `json:"deleted,omitzero"`
	Created      option.Ptr[Created]      `json:"created,omitzero"`
	GroupChanged option.Ptr[GroupChanged] `json:"gidChanged,omitzero"`
	OwnerChanged option.Ptr[OwnerChanged] `json:"oidChanged,omitzero"`
	ModeChanged  option.Ptr[ModeChanged]  `json:"modeChanged,omitzero"`
	Renamed      option.Ptr[Renamed]      `json:"renamed,omitzero"`
	Moved        option.Ptr[Moved]        `json:"moved,omitzero"`
	Added        option.Ptr[Added]        `json:"entryAdded,omitzero"`
	VersionAdded option.Ptr[VersionAdded] `json:"versionAdded,omitzero"`
}

// Unwrap assumes an enum-state and returns the first found non-nil pointer value.
func (e LogEntry) Unwrap() (Activity, bool) {
	switch {
	case e.Created.IsSome():
		return e.Created.Unwrap(), true
	case e.Added.IsSome():
		return e.Added.Unwrap(), true
	case e.Deleted.IsSome():
		return e.Deleted.Unwrap(), true
	case e.VersionAdded.IsSome():
		return e.VersionAdded.Unwrap(), true
	case e.Renamed.IsSome():
		return e.Renamed.Unwrap(), true
	}

	return nil, false
}

type Activity interface {
	Mod() xtime.UnixMilliseconds
	ModBy() user.ID
}

type Renamed struct {
	Name    string                 `json:"name"`
	ByUser  user.ID                `json:"uid"`
	ModTime xtime.UnixMilliseconds `json:"ts"`
}

func (r Renamed) Mod() xtime.UnixMilliseconds {
	return r.ModTime
}

func (r Renamed) ModBy() user.ID {
	return r.ByUser
}

type FileModeChanged struct {
	FileMode os.FileMode            `json:"mode,omitempty"`
	By       user.ID                `json:"uid"`
	Time     xtime.UnixMilliseconds `json:"ts"`
}

type GroupChanged struct {
}
type OwnerChanged struct {
}

type Created struct {
	Filename string                 `json:"name"`
	Owner    user.ID                `json:"oid,omitempty"`
	Group    group.ID               `json:"gid,omitempty"`
	FileMode os.FileMode            `json:"mode,omitempty"`
	Parent   FID                    `json:"parent,omitempty"`
	ByUser   user.ID                `json:"uid,omitempty"`
	Time     xtime.UnixMilliseconds `json:"ts,omitempty"`
}

func (d Created) Mod() xtime.UnixMilliseconds {
	return d.Time
}

func (d Created) ModBy() user.ID {
	return d.Owner
}

type Deleted struct {
	FID    FID                    `json:"fid"`
	ByUser user.ID                `json:"uid,omitempty"`
	Time   xtime.UnixMilliseconds `json:"ts,omitempty"`
}

func (d Deleted) Mod() xtime.UnixMilliseconds {
	return d.Time
}

func (d Deleted) ModBy() user.ID {
	return d.ByUser
}

type ModeChanged struct {
}

type Moved struct {
}

type Added struct {
	FID    FID                    `json:"fid"`
	ByUser user.ID                `json:"uid,omitempty"`
	Time   xtime.UnixMilliseconds `json:"ts,omitempty"`
}

func (d Added) Mod() xtime.UnixMilliseconds {
	return d.Time
}

func (d Added) ModBy() user.ID {
	return d.ByUser
}

type SourceHint int

const (
	Unknown SourceHint = iota
	Upload
	AI
)

type VersionAdded struct {
	SourceHint SourceHint             `json:"src"`
	FileInfo   FileInfo               `json:"info"`
	ByUser     user.ID                `json:"uid"`
	Time       xtime.UnixMilliseconds `json:"ts"`
}

func (v VersionAdded) Mod() xtime.UnixMilliseconds {
	return v.Time
}

func (v VersionAdded) ModBy() user.ID {
	return v.ByUser
}
