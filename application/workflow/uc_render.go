// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package workflow

import (
	"fmt"
	"os"
	"reflect"

	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/std/concurrent"
	"go.wdy.de/nago/presentation/core"
)

func NewRender(declarations *concurrent.RWMap[ID, *workflow]) Render {
	return func(subject user.Subject, wid ID, opts RenderOptions) (core.SVG, error) {
		wf, ok := declarations.Get(wid)
		if !ok {
			return nil, os.ErrNotExist
		}

		gv := newGvFlowModel()
		externalId := 1
		for _, r := range wf.uniqueEventTypes() {
			evt := gvEvent{
				ID:   makeValidID(r.String()),
				Name: nameOf(r),
			}

			if opts.ShowEventFields {
				evt.Fields = fieldsOfType(r)
			}

			gv.Events = append(gv.Events, evt)

			if wf.globalEvent(r) && opts.ShowExternalEventBus {
				actId := fmt.Sprintf("_external_%d", externalId)
				gv.Actions = append(gv.Actions, gvAction{
					ID:    actId,
					Label: "external",
					Type:  "external",
				})

				if wf.externalOutEvent(r) {
					gv.Transitions = append(gv.Transitions, gvTransition{
						From:  makeValidID(r.String()),
						Event: actId,
					})
				} else {
					gv.Transitions = append(gv.Transitions, gvTransition{
						To:    makeValidID(r.String()),
						Event: actId,
					})
				}

				externalId++
			}

		}

		// add manual transitions
		for _, transition := range wf.opts.Transitions {
			actId := fmt.Sprintf("_custom_transition_%d", externalId)
			gv.Actions = append(gv.Actions, gvAction{
				ID:    actId,
				Label: transition.Type.String(),
				Type:  "user",
			})

			gv.Transitions = append(gv.Transitions, gvTransition{
				From:  makeValidID(transition.InEvent.String()),
				Event: actId,
				To:    makeValidID(transition.OutEvent.String()),
			})
		}

		if opts.ShowTitle {
			gv.Label = wf.opts.Name
		}

		for _, n := range wf.nodes {
			typ := "normal"
			if len(n.cfg.startEvents) > 0 {
				typ = "start"
			}

			if len(n.cfg.stopEvents) > 0 {
				typ = "end"
			}

			gv.Actions = append(gv.Actions, gvAction{
				ID:    n.ID(),
				Label: n.Name(),
				Type:  typ,
			})

			for _, r := range wf.outgoingEventTypes(n) {
				gv.Transitions = append(gv.Transitions, gvTransition{
					Event: makeValidID(r.String()),
					From:  n.ID(),
				})
			}

			for _, r := range wf.incomingEventTypes(n) {
				gv.Transitions = append(gv.Transitions, gvTransition{
					Event: makeValidID(r.String()),
					To:    n.ID(),
				})
			}

			if n.cfg.description != "" {
				gv.Comments = append(gv.Comments, gvComment{
					ID:    "_comment_" + n.ID(),
					Label: n.cfg.description,
					To:    []string{n.ID()},
				})
			}
		}

		return renderSVG([]byte(gv.RenderDOT(opts)))
	}
}

func fieldsOfType(t reflect.Type) []gvField {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return nil
	}

	var res []gvField
	for i := 0; i < t.NumField(); i++ {
		tf := t.Field(i)
		if !tf.IsExported() {
			continue
		}
		name := tf.Name
		if n := tf.Tag.Get("label"); n != "" {
			name = n
		}

		res = append(res, gvField{
			Name: name,
			Type: makeValidID(tf.Type.String()),
		})
	}

	return res
}

func nameOf(t reflect.Type) string {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() == reflect.Struct {
		f, ok := t.FieldByName("_")
		if ok {
			if n, ok := f.Tag.Lookup("label"); ok {
				return n
			}
		}
	}

	return t.Name()
}
