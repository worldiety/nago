// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package aam

import (
	"fmt"
	"log/slog"

	"go.wdy.de/nago/app/builder/app"
	"go.wdy.de/nago/app/builder/environment"
	"go.wdy.de/nago/auth"
)

func NewCreate(replay environment.Replay) Create {
	return func(subject auth.Subject, env environment.ID, app app.ID) (*App, error) {
		a := &App{}
		for box, err := range replay(subject, env, app) {
			if err != nil {
				return a, err
			}

			evt, ok := box.Unwrap()
			if !ok {
				return a, fmt.Errorf("unable to unwrap box: %v", box.ID)
			}

			switch evt := evt.(type) {
			case environment.NamespaceCreated:
				a.Namespaces.Append(&Namespace{
					Name: evt.Name,
				})
			case environment.NamespaceDeleted:
				a.Namespaces.DeleteFunc(func(namespace *Namespace) bool {
					return namespace.Name == evt.Name
				})
			case environment.AppIDUpdated:
				a.ID = evt.ID
			case environment.GitRepoUpdated:
				a.GitRepoURL = evt.URL
			case environment.StructCreated:
				if v, ok := a.Namespaces.Get(evt.Namespace); ok {
					v.Structs.Put(&Struct{
						Namespace: evt.Namespace,
						Name:      evt.Name,
					})
				}
			case environment.TypeDeleted:
				if v, ok := a.Namespaces.Get(evt.Namespace); ok {
					v.Structs.DeleteFunc(func(s *Struct) bool {
						return s.Name == evt.Name
					})
				}
			default:
				slog.Error("unknown event", "evt", fmt.Sprintf("%T", evt))
			}
		}

		return a, nil
	}
}
