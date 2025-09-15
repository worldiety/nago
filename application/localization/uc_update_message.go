// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package localization

import (
	"fmt"
	"os"
	"time"

	"github.com/worldiety/i18n"
	"go.wdy.de/nago/auth"
	"golang.org/x/text/language"
)

func NewUpdateMessage(repo Repository, res *i18n.Resources) UpdateMessage {
	return func(subject auth.Subject, tag language.Tag, msg i18n.Message) error {
		if err := subject.Audit(PermUpdateMessage); err != nil {
			return err
		}

		bnd, ok := res.MatchBundle(tag)
		if !ok {
			return fmt.Errorf("no bundle found for language %v: %w", tag, os.ErrNotExist)
		}

		if err := bnd.Update(msg); err != nil {
			return fmt.Errorf("cannot update bundle: %w", err)
		}

		if msg.Kind == i18n.MessageUndefined {
			msg.Kind = bnd.MessageTypeByKey(msg.Key)
		}

		optData, err := repo.FindByID(msg.Key)
		if err != nil {
			return err
		}

		data := optData.UnwrapOr(StringData{
			Key:      msg.Key,
			Messages: map[language.Tag]i18n.Message{},
		})

		data.Messages[tag] = msg
		data.UpdatedAt = time.Now()
		data.UpdatedBy = subject.ID()

		return repo.Save(data)
	}
}
