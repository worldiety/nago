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
	Deleted      option.Ptr[Deleted]      `json:"d,omitzero"`
	Created      option.Ptr[Created]      `json:"c,omitzero"`
	GroupChanged option.Ptr[GroupChanged] `json:"g,omitzero"`
	OwnerChanged option.Ptr[OwnerChanged] `json:"o,omitzero"`
	ModeChanged  option.Ptr[ModeChanged]  `json:"m,omitzero"`
	Renamed      option.Ptr[Renamed]      `json:"r,omitzero"`
	Moved        option.Ptr[Moved]        `json:"v,omitzero"`
	Added        option.Ptr[Added]        `json:"a,omitzero"`
	VersionAdded option.Ptr[VersionAdded] `json:"e,omitzero"`
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
	Name    string                 `json:"n"`
	ByUser  user.ID                `json:"b"`
	ModTime xtime.UnixMilliseconds `json:"t"`
}

func (r Renamed) Mod() xtime.UnixMilliseconds {
	return r.ModTime
}

func (r Renamed) ModBy() user.ID {
	return r.ByUser
}

type FileModeChanged struct {
	FileMode os.FileMode            `json:"m,omitempty"`
	By       user.ID                `json:"b"`
	Time     xtime.UnixMilliseconds `json:"t"`
}

type GroupChanged struct {
}
type OwnerChanged struct {
}

type Created struct {
	Filename string                 `json:"n"`
	Owner    user.ID                `json:"o,omitempty"`
	Group    group.ID               `json:"g,omitempty"`
	FileMode os.FileMode            `json:"m,omitempty"`
	Parent   FID                    `json:"p,omitempty"`
	ByUser   user.ID                `json:"b,omitempty"`
	Time     xtime.UnixMilliseconds `json:"t,omitempty"`
}

func (d Created) Mod() xtime.UnixMilliseconds {
	return d.Time
}

func (d Created) ModBy() user.ID {
	return d.Owner
}

type Deleted struct {
	FID    FID                    `json:"f"`
	ByUser user.ID                `json:"b,omitempty"`
	Time   xtime.UnixMilliseconds `json:"t,omitempty"`
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
	FID    FID                    `json:"f"`
	ByUser user.ID                `json:"b,omitempty"`
	Time   xtime.UnixMilliseconds `json:"t,omitempty"`
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
	SourceHint SourceHint             `json:"s"`
	FileInfo   FileInfo               `json:"f"`
	ByUser     user.ID                `json:"b"`
	Time       xtime.UnixMilliseconds `json:"t"`
}

func (v VersionAdded) Mod() xtime.UnixMilliseconds {
	return v.Time
}

func (v VersionAdded) ModBy() user.ID {
	return v.ByUser
}
