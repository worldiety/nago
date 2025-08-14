// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package workflow

type Start[E any] struct {
}

func (a Start[E]) Configure(cfg *Configuration) error {
	StartEvent[E](cfg)
	cfg.SetName("Start")
	return nil
}

type Stop[E any] struct {
}

func (a Stop[E]) Configure(cfg *Configuration) error {
	StopEvent[E](cfg)
	cfg.SetName("End")
	return nil
}
