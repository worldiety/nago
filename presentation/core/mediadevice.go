// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package core

import (
	"go.wdy.de/nago/pkg/std/async"
	"go.wdy.de/nago/presentation/proto"
)

type MediaDeviceKind uint64

const (
	AudioInput  MediaDeviceKind = 0
	AudioOutput MediaDeviceKind = 1
	VideoInput  MediaDeviceKind = 2
)

func (k MediaDeviceKind) String() string {
	switch k {
	case AudioInput:
		return "Audio-Input"
	case AudioOutput:
		return "Audio-Output"
	case VideoInput:
		return "Video-Input"
	}
	return "Unbekannt"
}

type MediaDeviceID string

type MediaDevice struct {
	id      MediaDeviceID
	groupID string
	label   string
	kind    MediaDeviceKind
}

func (m MediaDevice) ID() MediaDeviceID {
	return m.id
}

func (m MediaDevice) Label() string {
	return m.label
}

func (m MediaDevice) Kind() MediaDeviceKind {
	return m.kind
}

type MediaDeviceListOptions struct {
	WithAudio bool
	WithVideo bool
}

type MediaDevices struct {
	w *scopeWindow
}

func (m MediaDevices) List(opts MediaDeviceListOptions) *async.Future[[]MediaDevice] {
	var fut async.Future[[]MediaDevice]

	AsyncCall(m.w, &proto.CallMediaDevicesEnumerate{Keep: true, WithAudio: proto.Bool(opts.WithAudio), WithVideo: proto.Bool(opts.WithVideo)}, func(ret proto.CallRet) {
		switch ret := ret.(type) {
		case *proto.RetMediaDevicesEnumerate:
			var tmp []MediaDevice
			for _, device := range ret.Devices {
				tmp = append(tmp, MediaDevice{
					id:      MediaDeviceID(device.DeviceID),
					groupID: string(device.GroupID),
					label:   string(device.Label),
					kind:    MediaDeviceKind(device.Kind),
				})
			}

			fut.Set(tmp, nil)
		case *proto.RetError:
			fut.Set(nil, newAsyncError(ret))
		}
	})
	return &fut
}

func (s *scopeWindow) MediaDevices() MediaDevices {
	return MediaDevices{w: s}
}
