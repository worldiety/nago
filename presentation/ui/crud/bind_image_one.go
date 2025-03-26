// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package crud

import (
	"fmt"
	"go.wdy.de/nago/application/image"
	http_image "go.wdy.de/nago/application/image/http"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/avatar"
	"go.wdy.de/nago/presentation/ui/form"
)

type PickOneImageStyle int

const (
	PickOneImageStyleSingle PickOneImageStyle = iota
	PickOneImageStyleAvatar
)

type PickOneImageOptions[E, T any] struct {
	Label        string
	Style        PickOneImageStyle // Default is PickOneImageStyleSingle
	Paraphe      func(E) string
	CreateSrcSet image.CreateSrcSet
	LoadSrcSet   image.LoadSrcSet
	LoadBestFit  image.LoadBestFit
}

// PickOneImage binds a single field of an arbitrary string type (which will be semantically
// a [image.ID]) to an upload option.
func PickOneImage[E any, T ~string](opts PickOneImageOptions[E, T], property Property[E, std.Option[T]]) Field[E] {
	if opts.Paraphe == nil {
		opts.Paraphe = func(e E) string {
			type paraphe interface {
				Paraphe() string
			}

			if p, ok := any(e).(paraphe); ok {
				return p.Paraphe()
			}

			return fmt.Sprintf("%v", e)
		}
	}

	return Field[E]{
		Label: opts.Label,
		RenderFormElement: func(self Field[E], entity *core.State[E]) ui.DecoredView {
			// here we create a copy for the local form field
			state := core.StateOf[image.ID](self.Window, self.ID+"-form.local").Init(func() image.ID {
				var tmp E
				tmp = entity.Get()
				return image.ID(property.Get(&tmp).UnwrapOr(""))
			})

			errState := core.StateOf[string](self.Window, self.ID+".err")

			// if the local field changes, we push our stuff into the given state (whatever that is)
			state.Observe(func(newValue image.ID) {
				var tmp E
				tmp = entity.Get()
				oldValue := property.Get(&tmp)
				if newValue == "" {
					property.Set(&tmp, std.None[T]())
				} else {
					property.Set(&tmp, std.Some[T](T(newValue)))
				}

				entity.Set(tmp)
				if image.ID(oldValue.UnwrapOr("")) != newValue {
					entity.Notify()
				}

				handleValidation(self, entity, errState)
			})

			entity.Observe(func(newValue E) {
				tmp := entity.Get()
				v := property.Get(&tmp).UnwrapOr("")
				state.Set(image.ID(v))
				state.Notify()
			})

			if self.requiresValidation() {
				state.Notify()
			}

			switch opts.Style {
			case PickOneImageStyleAvatar:
				return form.AvatarPicker(self.Window, opts.CreateSrcSet, self.ID, state.Get(), state, opts.Paraphe(entity.Get()))
			default:
				return form.SingleImagePicker(self.Window, opts.CreateSrcSet, opts.LoadSrcSet, opts.LoadBestFit, self.ID, state.Get(), state)
			}

		},
		RenderTableCell: func(self Field[E], entity *core.State[E]) ui.TTableCell {
			tmp := entity.Get()
			v := property.Get(&tmp)
			id := image.ID(v.UnwrapOr(""))
			var img core.View

			switch opts.Style {
			case PickOneImageStyleAvatar:
				if id != "" {
					// TODO replace me with source set due to different density problem
					uri := core.URI(http_image.NewURL(http_image.Endpoint, id, image.FitCover, 32, 32))
					img = avatar.URI(uri).Size(ui.L32)
				} else {
					img = avatar.Text(opts.Paraphe(tmp)).Size(ui.L32)
				}
			default:
				// TODO replace me with source set due to different density problem
				if id == "" {
					// TODO placeholder image?
					img = ui.HStack()
				} else {
					uri := core.URI(http_image.NewURL(http_image.Endpoint, id, image.FitCover, 32, 32))
					img = ui.Image().URI(uri).Frame(ui.Frame{}.Size(ui.L48, ui.L32)).Border(ui.Border{}.Radius(ui.L4)) // ca. 16:9
				}

			}

			return ui.TableCell(img)
		},
		RenderCardElement: func(self Field[E], entity *core.State[E]) ui.DecoredView {
			var tmp E
			tmp = entity.Get()
			v := property.Get(&tmp)
			id := image.ID(v.UnwrapOr(""))
			var img core.View
			if id != "" {
				// TODO replace me with source set due to different density problem
				uri := core.URI(http_image.NewURL(http_image.Endpoint, id, image.FitCover, 144, 144))
				img = avatar.URI(uri).Size(ui.L144)
			} else {
				img = avatar.Text(opts.Paraphe(tmp)).Size(ui.L144)
			}

			return ui.VStack(
				ui.VStack(ui.Text(self.Label).Font(ui.SubTitle)).
					Alignment(ui.Leading).
					Frame(ui.Frame{}.FullWidth()),
				img,
			).Alignment(ui.Trailing)
		},
		Comparator: nil,
		Stringer: func(e E) string {
			return string(property.Get(&e).UnwrapOr(""))
		},
	}
}
