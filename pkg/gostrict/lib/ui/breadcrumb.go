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
	nui "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/breadcrumb"
)

// TBreadcrumbs is a horizontal trail of items, which shows the path of the navigation, see Breadcrumbs.
type TBreadcrumbs struct{ ref js.Ref }

func (TBreadcrumbs) isView() {}

// Breadcrumbs returns a trail of the items, which are separated by chevrons. Like breadcrumb.Breadcrumbs of nago.
//
//gostrict:host
func Breadcrumbs(items ...View) TBreadcrumbs {
	return TBreadcrumbs{ref: wrap(breadcrumb.Breadcrumbs(viewsOf(items)...))}
}

// Gap sets the spacing between the items.
//
//gostrict:host
func (c TBreadcrumbs) Gap(gap Length) TBreadcrumbs {
	return TBreadcrumbs{ref: wrap(get[breadcrumb.TBreadcrumbs](c.ref).Gap(nui.Length(gap)))}
}

// Frame sets the frame of the trail.
//
//gostrict:host
func (c TBreadcrumbs) Frame(frame Frame) TBreadcrumbs {
	return TBreadcrumbs{ref: wrap(get[breadcrumb.TBreadcrumbs](c.ref).Frame(frame.nago()))}
}

// Padding sets the inner padding around the trail.
//
//gostrict:host
func (c TBreadcrumbs) Padding(padding Padding) TBreadcrumbs {
	return TBreadcrumbs{ref: wrap(get[breadcrumb.TBreadcrumbs](c.ref).Padding(padding.nago()))}
}

// ClampLeading aligns the title of a first item with the leading edge of the trail.
//
//gostrict:host
func (c TBreadcrumbs) ClampLeading() TBreadcrumbs {
	return TBreadcrumbs{ref: wrap(get[breadcrumb.TBreadcrumbs](c.ref).ClampLeading())}
}

// Item appends a tertiary button with the title, which calls action, when the user clicks it.
//
//gostrict:host
func (c TBreadcrumbs) Item(title string, action func()) TBreadcrumbs {
	return TBreadcrumbs{ref: wrap(get[breadcrumb.TBreadcrumbs](c.ref).Item(title, action))}
}
