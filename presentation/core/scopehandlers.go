package core

import (
	"fmt"
	"go.wdy.de/nago/presentation/ora"
	"log/slog"
)

// only for event loop
func (s *Scope) handleEvent(t ora.Event) {
	switch evt := t.(type) {
	case ora.EventsAggregated:
		s.handleEventsAggregated(evt)
	case ora.SetPropertyValueRequested:
		s.handleSetPropertyValueRequested(evt)
	case ora.FunctionCallRequested:
		s.handleFunctionCallRequested(evt)
	case ora.NewComponentRequested:
		s.handleNewComponentRequested(evt)
	case ora.ComponentInvalidationRequested:
		s.handleComponentInvalidationRequested(evt)
	case ora.ConfigurationRequested:
		s.handleConfigurationRequested(evt)
	case ora.ComponentDestructionRequested:
		s.handleComponentDestructionRequested(evt)
	case ora.ScopeDestructionRequested:
		s.handleScopeDestructionRequested(evt)
	case ora.SessionAssigned:
		s.handleSessionAssigned(evt)
	default:
		slog.Error("unexpected event type in scope processing", slog.String("type", fmt.Sprintf("%T", evt)))
	}
}

func (s *Scope) handleEventsAggregated(evt ora.EventsAggregated) {
	for _, event := range evt.Events {
		switch e := event.(type) {
		case ora.FunctionCallRequested:
			e.RequestId = evt.RequestId
			event = e
		case ora.SetPropertyValueRequested:
			e.RequestId = evt.RequestId
			event = e
		}

		s.handleEvent(event)
	}

	s.sendAck(evt.RequestId)
}

func (s *Scope) handleScopeDestructionRequested(evt ora.ScopeDestructionRequested) {
	s.sendAck(evt.RequestId)
	s.destroy()
	s.eventLoop.Destroy() // discards everything else queued
}

func (s *Scope) handleSetPropertyValueRequested(evt ora.SetPropertyValueRequested) {
	// we do not expect many components, usually only 1 or 2 (e.g. 2 open activities on mobile)
	for _, state := range s.allocatedComponents {
		prop := state.RenderState.props[evt.Ptr]
		if prop == nil {
			continue
		}

		if err := prop.Parse(evt.Value); err != nil {
			slog.Error("invalid property value", slog.Any("evt", evt), slog.String("property-type", fmt.Sprintf("%T", prop)))
			s.Publish(ora.ErrorOccurred{
				Type:      ora.ErrorOccurredT,
				RequestId: evt.RequestId,
				Message:   fmt.Sprintf("cannot set property: invalid property value: %d", evt.Ptr),
			})
		}

		return
	}

	slog.Error("property not found", slog.Any("evt", evt))
	s.Publish(ora.ErrorOccurred{
		Type:      ora.ErrorOccurredT,
		RequestId: evt.RequestId,
		Message:   fmt.Sprintf("cannot set property: no such pointer found: %d", evt.Ptr),
	})

	return
}

func (s *Scope) handleFunctionCallRequested(evt ora.FunctionCallRequested) {
	// we do not expect many components, usually only 1 or 2 (e.g. 2 open activities on mobile)
	for _, state := range s.allocatedComponents {
		fn := state.RenderState.funcs[evt.Ptr]
		if fn == nil {
			continue
		}

		fn.Invoke()
		return
	}

	slog.Error("function not found", slog.Any("evt", evt))
	s.Publish(ora.ErrorOccurred{
		Type:      ora.ErrorOccurredT,
		RequestId: evt.RequestId,
		Message:   fmt.Sprintf("cannot call function: no such pointer found: %d", evt.Ptr),
	})

	return
}

func (s *Scope) handleNewComponentRequested(evt ora.NewComponentRequested) {
	realm := newScopeRealm(s, evt.Factory, evt.Values)
	fac := s.factories[evt.Factory]
	var component Component
	if fac == nil {
		slog.Error("frontend requested unknown factory", slog.String("path", string(evt.Factory)), slog.Int("requestId", int(evt.RequestId)))
		fac = s.factories["_"]
		if fac != nil {
			notFoundComponent := fac(realm, evt)
			if notFoundComponent == nil {
				slog.Error("notFound factory returned a nil component which is not allowed", slog.String("id", "_"), slog.Int("requestId", int(evt.RequestId)))
				return
			}
			component = notFoundComponent
		} else {
			s.Publish(ora.ErrorOccurred{
				Type:      ora.ErrorOccurredT,
				RequestId: evt.RequestId,
				Message:   fmt.Sprintf("factory %s not found", evt.Factory),
			})
			return
		}

	} else {
		component = fac(realm, evt)
		if component == nil {
			slog.Error("factory returned a nil component which is not allowed", slog.String("id", string(evt.Factory)), slog.Int("requestId", int(evt.RequestId)))
			s.Publish(ora.ErrorOccurred{
				Type:      ora.ErrorOccurredT,
				RequestId: evt.RequestId,
				Message:   fmt.Sprintf("internal factory error: delivered null component"),
			})
			return
		}
	}

	s.allocatedComponents[component.ID()] = allocatedComponent{
		Realm:       realm,
		Component:   component,
		RenderState: NewRenderState(),
	}

	// an allocation without rendering does not make sense, just perform in the same cycle
	renderTree := s.render(evt.RequestId, component)
	s.Publish(renderTree)

}

func (s *Scope) handleComponentInvalidationRequested(evt ora.ComponentInvalidationRequested) {
	alloc, ok := s.allocatedComponents[evt.Component]
	if !ok {
		slog.Error("cannot invalidate: no such component in scope", slog.Any("evt", evt))
		s.Publish(ora.ErrorOccurred{
			Type:      ora.ErrorOccurredT,
			RequestId: evt.RequestId,
			Message:   fmt.Sprintf("cannot invalidate: no such component in scope: %d", evt.Component),
		})
		return
	}

	renderTree := s.render(evt.RequestId, alloc.Component)
	s.Publish(renderTree)
}

func (s *Scope) handleConfigurationRequested(evt ora.ConfigurationRequested) {
	// todo where is the configuration?
	s.Publish(ora.ConfigurationDefined{
		Type:             ora.ConfigurationDefinedT,
		ApplicationName:  "todo",
		AvailableLocales: []string{"de", "en"},
		ActiveLocale:     "de",
		Themes:           ora.Themes{},
		Resources:        ora.Resources{},
		RequestId:        evt.RequestId,
	})
}

func (s *Scope) handleComponentDestructionRequested(evt ora.ComponentDestructionRequested) {
	component, ok := s.allocatedComponents[evt.Component]
	if !ok {
		s.Publish(ora.ErrorOccurred{
			Type:      ora.ErrorOccurredT,
			RequestId: evt.RequestId,
			Message:   fmt.Sprintf("no such component: %d", evt.Component),
		})

		return
	}

	invokeDestructors(component)

	delete(s.allocatedComponents, evt.Component)

	s.sendAck(evt.RequestId)
}
