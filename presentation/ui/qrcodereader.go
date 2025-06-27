// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
)

type TQrCodeReader struct {
	inputValue           *core.State[[]string]
	mediaDevice          core.MediaDevice
	showTracker          bool
	trackerColor         Color
	trackerLineWidth     int
	activatedTorch       bool
	noMediaDeviceContent core.View
	onCameraReady        func()
	frame                Frame
}

func QrCodeReader(mediaDevice core.MediaDevice) TQrCodeReader {
	return TQrCodeReader{
		mediaDevice:      mediaDevice,
		showTracker:      true,
		trackerColor:     M0,
		trackerLineWidth: 2,
		activatedTorch:   false,
		onCameraReady:    func() {},
	}
}

func (c TQrCodeReader) InputValue(inputValue *core.State[[]string]) TQrCodeReader {
	c.inputValue = inputValue
	return c
}

func (c TQrCodeReader) ShowTracker(showTracker bool) TQrCodeReader {
	c.showTracker = showTracker
	return c
}

func (c TQrCodeReader) TrackerColor(trackerColor Color) TQrCodeReader {
	c.trackerColor = trackerColor
	return c
}

func (c TQrCodeReader) TrackerLineWidth(trackerLineWidth int) TQrCodeReader {
	c.trackerLineWidth = trackerLineWidth
	return c
}

func (c TQrCodeReader) ActivatedTorch(activatedTorch bool) TQrCodeReader {
	c.activatedTorch = activatedTorch
	return c
}

func (c TQrCodeReader) NoMediaDeviceContent(noMediaDeviceContent core.View) TQrCodeReader {
	c.noMediaDeviceContent = noMediaDeviceContent
	return c
}

func (c TQrCodeReader) OnCameraReady(onCameraReady func()) TQrCodeReader {
	c.onCameraReady = onCameraReady
	return c
}

func (c TQrCodeReader) Frame(frame Frame) TQrCodeReader {
	c.frame = frame
	return c
}

func (c TQrCodeReader) Render(ctx core.RenderContext) core.RenderNode {
	return &proto.QrCodeReader{
		InputValue: c.inputValue.Ptr(),
		MediaDevice: proto.MediaDevice{
			DeviceID: proto.Str(c.mediaDevice.ID()),
			GroupID:  proto.Str(c.mediaDevice.GroupID()),
			Label:    proto.Str(c.mediaDevice.Label()),
			Kind:     proto.MediaDeviceKind(c.mediaDevice.Kind()),
		},
		ShowTracker:          proto.Bool(c.showTracker),
		TrackerColor:         proto.Color(c.trackerColor),
		TrackerLineWidth:     proto.Uint(c.trackerLineWidth),
		ActivatedTorch:       proto.Bool(c.activatedTorch),
		NoMediaDeviceContent: render(ctx, c.noMediaDeviceContent),
		OnCameraReady:        ctx.MountCallback(c.onCameraReady),
		Frame:                c.frame.ora(),
	}
}
