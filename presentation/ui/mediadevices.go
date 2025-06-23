// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui

import (
	"go.wdy.de/nago/pkg/xmediadevice"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
)

type TMediaDevices struct {
	inputValue            *core.State[[]xmediadevice.MediaDevice]
	hasGrantedPermissions *core.State[bool]
	withAudio             bool
	withVideo             bool
}

func MediaDevices() TMediaDevices {
	return TMediaDevices{
		withAudio: true,
		withVideo: true,
	}
}

func (c TMediaDevices) InputValue(inputValue *core.State[[]xmediadevice.MediaDevice]) TMediaDevices {
	c.inputValue = inputValue
	return c
}

func (c TMediaDevices) HasGrantedPermissions(hasGrantedPermissions *core.State[bool]) TMediaDevices {
	c.hasGrantedPermissions = hasGrantedPermissions
	return c
}

func (c TMediaDevices) WithAudio(withAudio bool) TMediaDevices {
	c.withAudio = withAudio
	return c
}

func (c TMediaDevices) WithVideo(withVideo bool) TMediaDevices {
	c.withVideo = withVideo
	return c
}

func (c TMediaDevices) Render(ctx core.RenderContext) core.RenderNode {
	return &proto.MediaDevices{
		InputValue:            c.inputValue.Ptr(),
		WithAudio:             proto.Bool(c.withAudio),
		WithVideo:             proto.Bool(c.withVideo),
		HasGrantedPermissions: c.hasGrantedPermissions.Ptr(),
	}
}
