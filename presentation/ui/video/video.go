// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package video

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/internal"
)

type TVideo struct {
	frame       ui.Frame
	src         core.URI
	controls    bool
	loop        bool
	muted       bool
	playsInline bool
	post        core.URI
	autoplay    bool
}

func Video(src core.URI) TVideo {
	return TVideo{
		src: src,
	}
}

func (c TVideo) Frame(frame ui.Frame) TVideo {
	c.frame = frame
	return c
}

func (c TVideo) Src(src core.URI) TVideo {
	c.src = src
	return c
}

func (c TVideo) Controls(controls bool) TVideo {
	c.controls = controls
	return c
}

func (c TVideo) Loop(loop bool) TVideo {
	c.loop = loop
	return c
}

func (c TVideo) AutoPlay(autoplay bool) TVideo {
	c.autoplay = autoplay
	return c
}

func (c TVideo) Poster(poster core.URI) TVideo {
	c.post = poster
	return c
}

func (c TVideo) PlaysInline(plays bool) TVideo {
	c.playsInline = plays
	return c
}

func (c TVideo) Muted(muted bool) TVideo {
	c.muted = muted
	return c
}

func (c TVideo) Render(ctx core.RenderContext) core.RenderNode {
	return &proto.Video{
		Src:         proto.URI(c.src),
		Frame:       internal.FrameToOra(c.frame),
		Controls:    proto.Bool(c.controls),
		Loop:        proto.Bool(c.loop),
		Muted:       proto.Bool(c.muted),
		PlaysInline: proto.Bool(c.playsInline),
		Poster:      proto.URI(c.post),
		Autoplay:    proto.Bool(c.autoplay),
	}
}
