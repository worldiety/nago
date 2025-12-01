// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package workflow

import (
	"context"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"strings"

	"go.wdy.de/nago/pkg/xjson"
)

type workflow struct {
	opts        DeclareOptions
	nodes       []node
	startEvents []reflect.Type
	stopEvents  []reflect.Type
}

func newWorkflow(opts DeclareOptions) *workflow {
	wf := &workflow{
		opts: opts,
	}

	return wf
}

func (w *workflow) init() {
	for _, n := range w.nodes {
		w.startEvents = append(w.startEvents, n.cfg.startEvents...)
		w.stopEvents = append(w.stopEvents, n.cfg.stopEvents...)

		if n.evtType != nil {
			xjson.RegisterSelf(n.evtType)
		}
	}

	for _, event := range w.startEvents {
		xjson.RegisterSelf(event)
	}

	for _, event := range w.stopEvents {
		xjson.RegisterSelf(event)
	}

}

func (w *workflow) isStartEvent(t reflect.Type) bool {
	return slices.Contains(w.startEvents, t)
}

func (w *workflow) isStopEvent(t reflect.Type) bool {
	return slices.Contains(w.stopEvents, t)
}

func (w *workflow) uniqueEventTypes() []reflect.Type {
	tmp := map[reflect.Type]struct{}{}
	for _, n := range w.nodes {
		for _, evt := range n.cfg.startEvents {
			tmp[evt] = struct{}{}
		}

		for _, evt := range n.cfg.stopEvents {
			tmp[evt] = struct{}{}
		}

		for _, evt := range n.cfg.publishLocalEvents {
			tmp[evt] = struct{}{}
		}

		for _, evt := range n.cfg.publishGlobalEvents {
			tmp[evt] = struct{}{}
		}

		if r := n.eventType(); r != nil {
			tmp[r] = struct{}{}
		}
	}

	return slices.SortedFunc(maps.Keys(tmp), func(r reflect.Type, r2 reflect.Type) int {
		return strings.Compare(r.String(), r2.String())
	})
}

func (w *workflow) outgoingEventTypes(n node) []reflect.Type {
	var res []reflect.Type
	for _, nod := range w.nodes {
		if nod.cfg == n.cfg {
			res = append(res, n.cfg.publishLocalEvents...)
			res = append(res, n.cfg.publishGlobalEvents...)
		}
	}

	for _, event := range n.cfg.startEvents {
		res = append(res, event)
	}

	return res
}

func (w *workflow) incomingEventTypes(n node) []reflect.Type {
	var res []reflect.Type
	for _, event := range n.cfg.stopEvents {
		res = append(res, event)
	}

	if evt := n.eventType(); evt != nil {
		res = append(res, evt)
	}

	return res
}

func (w *workflow) externalOutEvent(t reflect.Type) bool {
	for _, n := range w.nodes {
		for _, event := range n.cfg.stopEvents {
			if event == t {
				return true
			}
		}

		for _, event := range n.cfg.publishGlobalEvents {
			if event == t {
				return true
			}
		}

	}

	return false
}

func (w *workflow) globalEvent(t reflect.Type) bool {
	for _, n := range w.nodes {
		for _, event := range n.cfg.publishGlobalEvents {
			if event == t {
				return true
			}
		}

		// by definition any start event must be global/external
		for _, event := range n.cfg.startEvents {
			if event == t {
				return true
			}
		}

		// stop events can just trigger an action without any publish just for side effect, so it must not be external
		// by definition
	}

	// any event, which is not produced by our own, must be external by definition
	produced := false
	for _, n := range w.nodes {
		for _, event := range n.cfg.publishLocalEvents {
			if event == t {
				produced = true
				break
			}
		}
	}

	if !produced {
		return true
	}

	return false
}

type node struct {
	cfg           *Configuration
	action        Action
	evtType       reflect.Type
	onEventMethod reflect.Method
}

func newNode(cfg *Configuration, action Action) node {
	n := node{
		cfg:    cfg,
		action: action,
	}

	method, ok := reflect.TypeOf(n.action).MethodByName("OnEvent")
	if ok {
		n.onEventMethod = method
		mType := method.Type
		numParams := mType.NumIn()
		if numParams > 0 {
			lastParamType := mType.In(numParams - 1)
			n.evtType = lastParamType
		}
	}

	return n
}

func (n node) Name() string {
	if n.cfg.name != "" {
		return n.cfg.name
	}

	return reflect.TypeOf(n.action).String()
}

func (n node) ID() string {
	t := reflect.TypeOf(n.action)
	tmp := t.PkgPath() + "." + t.Name()
	if tmp == "." {
		tmp = t.String() // happens e.g. for test packages
	}

	return makeValidID(tmp)
}

// eventType is the receiver event type
func (n node) eventType() reflect.Type {
	return n.evtType
}

func (n node) actionType() reflect.Type {
	return reflect.TypeOf(n.action)
}

func (n node) invokeWithEvt(ctx context.Context, evt any) error {
	if n.onEventMethod.Func.IsZero() {
		return fmt.Errorf("type %t does not implement OnEvent", n.action)
	}

	errv := n.onEventMethod.Func.Call([]reflect.Value{reflect.ValueOf(n.action), reflect.ValueOf(ctx), reflect.ValueOf(evt)})
	if len(errv) > 0 && !errv[0].IsNil() {
		return errv[0].Interface().(error)
	}

	return nil
}

func (n node) canInvoke() bool {
	return !n.onEventMethod.Func.IsZero()
}

type ID string

type Step interface {
	inAwaitAllTypes() []reflect.Type
	inAwaitOneOfTypes() []reflect.Type
	outBlockingTypes() []reflect.Type
	outAsyncTypes() []reflect.Type
}

type EventConsumer interface {
	Invoke(ctx context.Context, evt any) error
}

type Configuration struct {
	workflow            ID
	description         string
	name                string
	startEvents         []reflect.Type
	stopEvents          []reflect.Type
	publishLocalEvents  []reflect.Type
	publishGlobalEvents []reflect.Type
}

func newConfiguration(id ID) *Configuration {
	return &Configuration{
		workflow: id,
	}
}

// SetName updates the action name
func (c *Configuration) SetName(name string) {
	c.name = name
}

func (c *Configuration) SetDescription(description string) {
	c.description = description
}

func (c *Configuration) Workflow() ID {
	return c.workflow
}

type Publisher[T any] func(T)

func LocalEvent[T any](cfg *Configuration) Publisher[T] {
	cfg.publishLocalEvents = append(cfg.publishLocalEvents, reflect.TypeFor[T]())
	return func(t T) {
		fmt.Println("TODO publish local", cfg.name, t)
	}
}

func GlobalEvent[T any](cfg *Configuration) Publisher[T] {
	cfg.publishGlobalEvents = append(cfg.publishGlobalEvents, reflect.TypeFor[T]())
	return func(t T) {
		fmt.Println("TODO publish global", cfg.name, t)
	}
}

func StartEvent[T any](cfg *Configuration) {
	cfg.startEvents = append(cfg.startEvents, reflect.TypeFor[T]())
}

func StopEvent[T any](cfg *Configuration) {
	cfg.stopEvents = append(cfg.stopEvents, reflect.TypeFor[T]())
}

type Action interface {
	Configure(cfg *Configuration) error
}

type EventAction[T any] interface {
	Configure(cfg *Configuration) error
	OnEvent(ctx context.Context, evt T) error
}
