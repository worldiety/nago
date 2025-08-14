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
	"log/slog"
	"reflect"
	"runtime/debug"

	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/std/concurrent"
)

func NewProcessEvent(declarations *concurrent.RWMap[ID, *workflow], createInstance CreateInstance, saveEvent SaveEvent, getStatus GetStatus, findInstances FindInstances) ProcessEvent {
	return func(subject auth.Subject, evt any) error {
		if err := subject.Audit(PermProcessEvent); err != nil {
			return err
		}

		// keep in mind, that our event bus spawns a go routine per invocation
		evtType := reflect.TypeOf(evt)
		slog.Info("workflow engine received bus event", "type", evtType)
		ctx := context.Background()
		for _, wf := range declarations.All() {
			for _, n := range wf.nodes {
				if n.eventType() == evtType {

					// start events must allocate a new instance
					if wf.isStartEvent(evtType) {
						id, err := createInstance(user.SU(), wf.opts.ID)
						if err != nil {
							slog.Error("failed to create instance", "err", err.Error())
						} else {
							slog.Info("new workflow instance spawned", "id", id, "workflow", wf.opts.ID)
						}

						if err := saveEvent(user.SU(), id, InstanceCreated{
							Workflow: wf.opts.ID,
							Instance: id,
						}); err != nil {
							slog.Error("failed to save InstanceCreated event", "err", err)
						}

						if err := saveEvent(user.SU(), id, evt); err != nil {
							slog.Error("failed to save event", "err", err)
						}
					}

					// anyway, every instance, which accepts it or is addressed, must be notified
					var instances []Instance
					switch evt := evt.(type) {
					case InstanceEventEnvelope:
						// this event is specific to an instance
						instances = append(instances, evt.Instance)
					default:
						// everything else is a global event and must be passed to every instance
						for instance, err := range findInstances(user.SU(), wf.opts.ID) {
							if err != nil {
								slog.Error("failed to find instance", "err", err)
								continue
							}

							instances = append(instances, instance)
						}
					}

					for _, instance := range instances {
						// only process if the instance is actually alive
						optStatus, err := getStatus(user.SU(), instance)
						if err != nil {
							slog.Error("failed to get status", "err", err.Error(), "instance", instance)
							continue
						}

						status := optStatus.UnwrapOr(Status{})

						if status.State != InstanceRunning {
							slog.Error("failed to step workflow instance: not running", "instance", instance, "status", status)
							continue
						}

						// save the actual event
						if err := saveEvent(user.SU(), instance, evt); err != nil {
							slog.Error("failed to save event", "err", err)
						}

						if n.canInvoke() {
							// save that we want to execute something
							if err := saveEvent(user.SU(), instance, ActionInvoked{Instance: instance, Action: NewTypename(n.actionType())}); err != nil {
								slog.Error("failed to save action invoke event", "err", err)
								continue
							}

							// actually invoke that thing
							slog.Info("found workflow node to invoke with event", "workflow", wf.opts.ID, "type", evtType)
							if err := guardPanic(func() error {
								return n.invokeWithEvt(ctx, evt)
							}); err != nil {
								slog.Error("failed to invoke workflow event", "err", err.Error())
								continue
							}

							// save that we succeeded to execute something
							if err := saveEvent(user.SU(), instance, ActionCompletedSuccessfully{Instance: instance, Action: NewTypename(n.actionType())}); err != nil {
								slog.Error("failed to save action complete event", "err", err)
								continue
							}
						}

						// trigger instance stop state
						for _, stopEvtType := range n.cfg.stopEvents {
							if stopEvtType == evtType {
								if err := saveEvent(user.SU(), instance, InstanceStopped{Instance: instance}); err != nil {
									slog.Error("failed to save stop event", "err", err)
								}
							}
						}

					}

				}
			}
		}

		return nil
	}
}

func guardPanic(fn func() error) (e error) {
	defer func() {
		if err := recover(); err != nil {
			slog.Error("panic recovered", "err", err, "stack", string(debug.Stack()))
			e = fmt.Errorf("recovered from panic: %v", err)
		}
	}()

	return fn()
}
