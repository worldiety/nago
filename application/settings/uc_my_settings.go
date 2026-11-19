// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package settings

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"

	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/pkg/data"
)

// AccountOf tells the account of a subject, whose settings [LoadMySettings] and [StoreMySettings] read and write.
// It reports false for a subject without an account, e.g. an anonymous one. The wiring in package application
// provides it, because this package must not depend on the user package.
type AccountOf func(subject permission.Auditable) (account string, ok bool)

// ErrNoAccount is returned by [StoreMySettings] for a subject without an account, see [AccountOf].
var ErrNoAccount = errors.New("the subject has no account to keep settings for")

// accountSeparator separates the account from the type in the key of the settings of a person. An account which
// contains it is refused, otherwise the settings of one account could be taken for those of another.
const accountSeparator = "/"

var userSettingsType = reflect.TypeFor[UserSettings]()

// makeUserID returns the key of the settings of type t of the account: one record per account and type.
func makeUserID(account string, t reflect.Type) ID {
	return ID(account + accountSeparator + string(makeGlobalID(t)))
}

// ownAccount returns the account of the subject, whose settings it may read and write.
func ownAccount(subject permission.Auditable, accountOf AccountOf) (string, bool) {
	if subject == nil || accountOf == nil {
		return "", false
	}

	account, ok := accountOf(subject)
	if !ok || account == "" || strings.Contains(account, accountSeparator) {
		return "", false
	}

	return account, true
}

// NewLoadMySettings returns a [LoadMySettings], which reads the settings of type t of the account of the subject,
// see [AccountOf], and never those of another account. Without a record, and for a subject without an account, it
// returns the zero value of t, which must therefore be valid. A record is (de)serialized like [GlobalSettings]: t
// must be declared as a variant of [UserSettings], e.g.
//
//	var _ = enum.Variant[settings.UserSettings, MySettings](enum.Rename[MySettings]("my.settings"))
func NewLoadMySettings(repo data.Repository[StoreBox[UserSettings], ID], accountOf AccountOf) LoadMySettings {
	return func(subject permission.Auditable, t reflect.Type) (UserSettings, error) {
		var zero UserSettings
		if t == nil || t.Kind() != reflect.Struct || !t.Implements(userSettingsType) {
			return zero, fmt.Errorf("type must be a struct type implementing UserSettings")
		}

		zero = reflect.New(t).Elem().Interface().(UserSettings)

		account, ok := ownAccount(subject, accountOf)
		if !ok {
			return zero, nil
		}

		optBox, err := repo.FindByID(makeUserID(account, t))
		if err != nil {
			return zero, err
		}

		if optBox.IsNone() {
			return zero, nil
		}

		s := optBox.Unwrap().Settings
		if reflect.TypeOf(s) != t {
			return zero, fmt.Errorf("stored settings of account %s have type %T instead of %v", account, s, t)
		}

		return s, nil
	}
}

// NewStoreMySettings returns a [StoreMySettings], which writes the settings as the record of their type for the
// account of the subject, see [AccountOf], and never for another account. It fails with [ErrNoAccount] for a
// subject without an account. Unchanged settings are not written again.
func NewStoreMySettings(repo data.Repository[StoreBox[UserSettings], ID], accountOf AccountOf) StoreMySettings {
	return func(subject permission.Auditable, settings UserSettings) error {
		if settings == nil {
			return fmt.Errorf("settings must not be nil")
		}

		t := reflect.TypeOf(settings)
		if t.Kind() != reflect.Struct {
			return fmt.Errorf("settings must be a struct type, not %v", t)
		}

		account, ok := ownAccount(subject, accountOf)
		if !ok {
			return ErrNoAccount
		}

		id := makeUserID(account, t)

		// like the global settings, an unchanged value is not written again
		optBox, err := repo.FindByID(id)
		if err == nil && optBox.IsSome() && reflect.DeepEqual(optBox.Unwrap().Settings, settings) {
			return nil
		}

		return repo.Save(StoreBox[UserSettings]{ID: id, Settings: settings})
	}
}

// DeleteAccount removes every settings record of the account, e.g. after the account has been deleted. An account
// which [LoadMySettings] would refuse has no records and is ignored.
func DeleteAccount(repo data.Repository[StoreBox[UserSettings], ID], account string) error {
	if account == "" || strings.Contains(account, accountSeparator) {
		return nil
	}

	var ids []ID
	for id, err := range repo.IdentifiersByPrefix(ID(account + accountSeparator)) {
		if err != nil {
			return fmt.Errorf("cannot list the settings of account %s: %w", account, err)
		}

		ids = append(ids, id)
	}

	if len(ids) == 0 {
		return nil
	}

	if err := repo.DeleteAllByID(slices.Values(ids)); err != nil {
		return fmt.Errorf("cannot delete the settings of account %s: %w", account, err)
	}

	return nil
}
