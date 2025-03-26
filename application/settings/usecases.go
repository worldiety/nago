// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package settings

import (
	"crypto/sha512"
	"encoding/hex"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/pkg/data"
	"log/slog"
	"reflect"
)

// GlobalSettings sum type is global for all known members of the settings and there is application wide exact one
// instance. The zero value of a settings type must be valid.
type GlobalSettings interface {
	GlobalSettings() bool // open sum type which can be extended by anyone
}

// UserSettings sum type is also global in a way that all users share all globally known members of
// UserSettings. However, each user has its own instance. The zero value of a settings type must be valid.
type UserSettings interface {
	UserSettings() bool // open sum type which can be extended by anyone
}

type LoadGlobal func(subject permission.Auditable, t reflect.Type) (GlobalSettings, error)
type StoreGlobal func(subject permission.Auditable, settings GlobalSettings) error

type LoadMySettings func(subject permission.Auditable, t reflect.Type) (UserSettings, error)
type StoreMySettings func(subject permission.Auditable, settings UserSettings) error

type ID string

func MySettings[T UserSettings](subject permission.Auditable, settings LoadMySettings) T {
	typ := reflect.TypeFor[T]()
	s, err := settings(subject, typ)
	if err != nil {
		slog.Error("failed to load per user settings", "err", err)
		var zero T
		return zero
	}

	return s.(T)
}

// ReadGlobal avoids any permission check and directly reads global settings.
func ReadGlobal[T GlobalSettings](global LoadGlobal) T {
	typ := reflect.TypeFor[T]()
	s, err := global(permission.SU(), typ)
	if err != nil {
		slog.Error("failed to load global settings", "err", err)
		var zero T
		return zero
	}

	return s.(T)
}

type StoreBox[T any] struct {
	ID       ID
	Settings T
}

func (b StoreBox[T]) Identity() ID {
	return b.ID
}

func (b StoreBox[T]) WithIdentity(id ID) StoreBox[T] {
	b.ID = id
	return b
}

type UseCases struct {
	LoadGlobal      LoadGlobal
	StoreGlobal     StoreGlobal
	LoadMySettings  LoadMySettings
	StoreMySettings StoreMySettings
}

func NewUseCases(globalRepo data.Repository[StoreBox[GlobalSettings], ID], userRepo data.Repository[StoreBox[UserSettings], ID]) UseCases {

	return UseCases{
		LoadGlobal:  NewLoadGlobal(globalRepo),
		StoreGlobal: NewStoreGlobal(globalRepo),
	}
}

type MetaData struct {
	Title       string
	Description string
}

func ReadMetaData(variant reflect.Type) MetaData {
	title := variant.Name()
	description := ""
	field, ok := variant.FieldByName("_")
	if ok {
		if s, ok := field.Tag.Lookup("title"); ok {
			title = s
		}
		if s, ok := field.Tag.Lookup("description"); ok {
			description = s
		}
	}

	return MetaData{
		Title:       title,
		Description: description,
	}
}

func TypeIdent(t reflect.Type) string {
	tmp := sha512.Sum512_224([]byte(t.PkgPath() + "." + t.Name()))
	return hex.EncodeToString(tmp[:])
}
