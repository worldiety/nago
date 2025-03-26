// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package image

import (
	"fmt"
	"go.wdy.de/nago/pkg/data"
)

type Format string

const (
	FormatJpeg Format = "jpeg"
	FormatPng  Format = "png"
)

type ObjectFit int

func (f ObjectFit) String() string {
	switch f {
	case FitNone:
		return "none"
	case FitCover:
		return "cover"
	default:
		return fmt.Sprintf("unknown %d", int(f))
	}
}

const (
	FitNone ObjectFit = iota
	FitCover
)

// SrcSet represents a bunch of scaled image variants of the same source.
type SrcSet struct {
	ID     ID      `json:"id,omitempty"`
	Name   string  `json:"n,omitempty"` // the original name, if any
	Images []Image `json:"set,omitempty"`
	Format Format  `json:"fmt,omitempty"`
}

func (s SrcSet) Fit(kind ObjectFit, width, height int) (Image, bool) {
	switch kind {
	case FitNone:
		var c Image
		for _, img := range s.Images {
			if img.Width > c.Width {
				c = img
			}
		}

		return c, c.Data != ""
	case FitCover:
		return s.FitCover(width, height)
	default:
		return s.FitCover(width, height)
	}
}

// FitCover returns the best candidate of an available image which is zoomed and cropped
// inside the given dimensions. It is like a virtual CSS object-fit: cover attribute.
func (s SrcSet) FitCover(width, height int) (Image, bool) {
	if len(s.Images) == 0 {
		return Image{}, false
	}

	if len(s.Images) == 1 {
		return s.Images[0], true
	}

	srcAspect := float64(s.Images[0].Width) / float64(s.Images[0].Height)
	targetAspect := float64(width) / float64(height)

	var requiredWidth, requiredHeight int
	if srcAspect > targetAspect {
		// image width is larger than target width
		requiredWidth = int(float64(width) * targetAspect)
		requiredHeight = int(float64(requiredWidth) * srcAspect)
	} else {
		// and vice versa
		requiredHeight = int(float64(height) / targetAspect)
		requiredWidth = int(float64(requiredHeight) * srcAspect)
	}

	bestImg := s.Images[0]
	for _, image := range s.Images {
		if image.Width > bestImg.Width || image.Height > bestImg.Height && (image.Width < requiredWidth || image.Height < requiredHeight) {
			// pick always the largest image which is smaller than the actual target size
			bestImg = image
			continue
		}

		if image.Width > requiredWidth && image.Height > requiredHeight && (image.Width < bestImg.Width || image.Height < bestImg.Height) {
			// pick always the smaller image which is larger than the required size
			bestImg = image
			continue
		}
	}

	return bestImg, true
}

func (s SrcSet) Identity() ID {
	return s.ID
}

// Image within a src set of images. It describes a distinct resolution of an arbitary image.
type Image struct {
	Width  int `json:"w"`
	Height int `json:"h"`
	Data   ID  `json:"d"`
}

func (i Image) Dimensions() (width, height int) {
	return i.Width, i.Height
}

func (i Image) Identity() string {
	return string(i.Data)
}

// ID to an encoded (and compressed) image. Usually a png or jpeg.
// However, it may also refer an entire SrcSet.
// This allows the unification of the image handling in different resolutions.
type ID string

type Repository data.Repository[SrcSet, ID]
