// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License
//
//gostrict:header
package ui

import (
	"go.wdy.de/nago/pkg/gostrict/std/js"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui/video"
)

// TVideo is a video, like video.TVideo of nago. See Video.
type TVideo struct{ ref js.Ref }

func (TVideo) isView() {}

// Video returns a video of the source, like video.Video of nago.
//
//gostrict:host
func Video(src URI) TVideo {
	return TVideo{ref: wrap(video.Video(core.URI(src)))}
}

// Frame sets the frame of the video.
//
//gostrict:host
func (c TVideo) Frame(frame Frame) TVideo {
	return TVideo{ref: wrap(get[video.TVideo](c.ref).Frame(frame.nago()))}
}

// Src sets the source of the video.
//
//gostrict:host
func (c TVideo) Src(src URI) TVideo {
	return TVideo{ref: wrap(get[video.TVideo](c.ref).Src(core.URI(src)))}
}

// Controls shows or hides the controls of the video.
//
//gostrict:host
func (c TVideo) Controls(controls bool) TVideo {
	return TVideo{ref: wrap(get[video.TVideo](c.ref).Controls(controls))}
}

// Loop repeats the video.
//
//gostrict:host
func (c TVideo) Loop(loop bool) TVideo {
	return TVideo{ref: wrap(get[video.TVideo](c.ref).Loop(loop))}
}

// AutoPlay starts the video automatically.
//
//gostrict:host
func (c TVideo) AutoPlay(autoplay bool) TVideo {
	return TVideo{ref: wrap(get[video.TVideo](c.ref).AutoPlay(autoplay))}
}

// Poster sets the image, which is shown before the video plays.
//
//gostrict:host
func (c TVideo) Poster(poster URI) TVideo {
	return TVideo{ref: wrap(get[video.TVideo](c.ref).Poster(core.URI(poster)))}
}

// PlaysInline lets the video play inline instead of fullscreen.
//
//gostrict:host
func (c TVideo) PlaysInline(plays bool) TVideo {
	return TVideo{ref: wrap(get[video.TVideo](c.ref).PlaysInline(plays))}
}

// Muted mutes the video.
//
//gostrict:host
func (c TVideo) Muted(muted bool) TVideo {
	return TVideo{ref: wrap(get[video.TVideo](c.ref).Muted(muted))}
}
