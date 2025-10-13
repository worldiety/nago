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

// TQrCodeReader is a composite component (QR Code Reader).
// It uses a media device (e.g., camera) to scan QR codes in real-time.
// The component supports visual trackers, torch (flashlight) activation,
// custom UI when no media device is available, and a callback when the camera is ready.
type TQrCodeReader struct {
	inputValue           *core.State[[]string] // stores scanned QR code values
	mediaDevice          core.MediaDevice      // media device (camera) used for scanning
	showTracker          bool                  // whether to display a tracker overlay
	trackerColor         Color                 // color of the tracker overlay
	trackerLineWidth     int                   // thickness of the tracker overlay lines
	activatedTorch       bool                  // whether the device torch (flashlight) is active
	noMediaDeviceContent core.View             // fallback view when no media device is available
	onCameraReady        func()                // callback invoked when the camera is ready
	frame                Frame                 // layout frame for size and positioning
}

// QrCodeReader creates a new QR code reader using the given media device (camera).
// By default, it enables the tracker, sets the tracker color to M0, line width to 2,
// torch off, and initializes an empty onCameraReady callback.
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

// InputValue binds the QR code reader to a state, which will be updated
// with the scanned QR code values.
func (c TQrCodeReader) InputValue(inputValue *core.State[[]string]) TQrCodeReader {
	c.inputValue = inputValue
	return c
}

// ShowTracker toggles the visibility of the tracker overlay on the camera preview.
func (c TQrCodeReader) ShowTracker(showTracker bool) TQrCodeReader {
	c.showTracker = showTracker
	return c
}

// TrackerColor sets the color of the tracker overlay.
func (c TQrCodeReader) TrackerColor(trackerColor Color) TQrCodeReader {
	c.trackerColor = trackerColor
	return c
}

// TrackerLineWidth sets the thickness of the tracker overlay lines.
func (c TQrCodeReader) TrackerLineWidth(trackerLineWidth int) TQrCodeReader {
	c.trackerLineWidth = trackerLineWidth
	return c
}

// ActivatedTorch enables or disables the camera torch (flashlight).
func (c TQrCodeReader) ActivatedTorch(activatedTorch bool) TQrCodeReader {
	c.activatedTorch = activatedTorch
	return c
}

// NoMediaDeviceContent sets the fallback view shown when no media device (camera) is available.
func (c TQrCodeReader) NoMediaDeviceContent(noMediaDeviceContent core.View) TQrCodeReader {
	c.noMediaDeviceContent = noMediaDeviceContent
	return c
}

// OnCameraReady sets the callback function to be executed when the camera is ready for scanning.
func (c TQrCodeReader) OnCameraReady(onCameraReady func()) TQrCodeReader {
	c.onCameraReady = onCameraReady
	return c
}

// Frame sets the layout frame for the QR code reader component.
func (c TQrCodeReader) Frame(frame Frame) TQrCodeReader {
	c.frame = frame
	return c
}

// Render builds and returns the protocol representation of the QR code reader.
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
