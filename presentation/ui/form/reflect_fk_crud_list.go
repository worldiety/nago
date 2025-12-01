// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package form

/*
func (t TAuto[T]) renderFkCrudList(ctx core.RenderContext, field reflect.StructField) (core.View, bool) {
	if !(field.Type.Kind() == reflect.Slice && field.Type.Elem().Kind() == reflect.String) {
		return nil, false
	}

	sourceName, ok := field.Tag.Lookup("source")
	if !ok {
		return nil, false
	}

	findByName, ok := field.Tag.Lookup("findBy")
	if !ok {
		return nil, false
	}

	wnd := t.opts.Window
	if wnd == nil {
		wnd = ctx.Window()
	}

	listAllIdents, ok := core.FromContext[ent.FindAllIdentifiers[AnyEntity, string]](t.opts.context(), sourceName)
	if !ok {
		slog.Error("form.Auto expected to render []~string with source, but is undeclared in context", "source", sourceName, "type", reflect.TypeFor[ent.FindAllIdentifiers[AnyEntity, string]]())
		return nil, false
	}

	findByID, ok := core.FromContext[ent.FindByID[AnyEntity, string]](t.opts.context(), findByName)
	if !ok {
		slog.Error("form.Auto expected to render []~string with findBy, but is undeclared in context", "findBy", findByName, "type", reflect.TypeFor[ent.FindByID[AnyEntity, string]]())
		return nil, false
	}

	strState := core.DerivedState[[]AnyEntity](t.state, field.Name).Init(func() []AnyEntity {
		src := t.state.Get()
		slice := reflect.ValueOf(src).FieldByName(field.Name)
		tmp := make([]AnyEntity, 0, slice.Len())
		for _, id := range slice.Seq2() {
			id := id.String()

			optEntry, err := findByID(wnd.Subject(), id)
			if err != nil {
				slog.Error("renderFkCrudList failed to find entry by id", "id", id, "err", err.Error())
				continue
			}

			if optEntry.IsNone() {
				continue // just a stale reference: ignore it
			}

			tmp = append(tmp, optEntry.Unwrap())
		}

		return tmp
	})

}
*/
