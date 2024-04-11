package core

import (
	"fmt"
	"go.wdy.de/nago/presentation/protocol"
	"log/slog"
)

// only for event loop
func (s *Scope) handleEvent(t protocol.Event) {
	switch evt := t.(type) {
	case protocol.EventsAggregated:
		s.handleEventsAggregated(evt)
	case protocol.SetPropertyValueRequested:
		s.handleSetPropertyValueRequested(evt)
	case protocol.FunctionCallRequested:
		s.handleFunctionCallRequested(evt)
	case protocol.NewComponentRequested:
		s.handleNewComponentRequested(evt)
	case protocol.ComponentInvalidationRequested:
		s.handleComponentInvalidationRequested(evt)
	case protocol.ConfigurationRequested:
		s.handleConfigurationRequested(evt)
	case protocol.ComponentDestructionRequested:
		s.handleComponentDestructionRequested(evt)
	case protocol.ScopeDestructionRequested:
		s.handleScopeDestructionRequested(evt)
	default:
		slog.Error("unexpected event type in scope processing", slog.String("type", fmt.Sprintf("%T", evt)))
	}
}

func (s *Scope) handleEventsAggregated(evt protocol.EventsAggregated) {
	for _, event := range evt.Events {
		s.handleEvent(event)
	}

	s.sendAck(evt.RequestId)
}

func (s *Scope) handleScopeDestructionRequested(evt protocol.ScopeDestructionRequested) {
	s.sendAck(evt.RequestId)
	s.destroy()
	s.eventLoop.Destroy() // discards everything else queued
}

func (s *Scope) handleSetPropertyValueRequested(evt protocol.SetPropertyValueRequested) {
	// we do not expect many components, usually only 1 or 2 (e.g. 2 open activities on mobile)
	for _, state := range s.allocatedComponents {
		prop := state.RenderState.props[evt.Ptr]
		if prop == nil {
			continue
		}

		if err := prop.Parse(evt.Value); err != nil {
			slog.Error("invalid property value", slog.Any("evt", evt), slog.String("property-type", fmt.Sprintf("%T", prop)))
			s.Publish(protocol.ErrorOccurred{
				Type:      protocol.ErrorOccurredT,
				RequestId: evt.RequestId,
				Message:   fmt.Sprintf("cannot set property: invalid property value: %d", evt.Ptr),
			})
		}

		return
	}

	slog.Error("property not found", slog.Any("evt", evt))
	s.Publish(protocol.ErrorOccurred{
		Type:      protocol.ErrorOccurredT,
		RequestId: evt.RequestId,
		Message:   fmt.Sprintf("cannot set property: no such pointer found: %d", evt.Ptr),
	})

	return
}

func (s *Scope) handleFunctionCallRequested(evt protocol.FunctionCallRequested) {
	// we do not expect many components, usually only 1 or 2 (e.g. 2 open activities on mobile)
	for _, state := range s.allocatedComponents {
		fn := state.RenderState.funcs[evt.Ptr]
		if fn == nil {
			continue
		}

		fn.Invoke()
	}

	slog.Error("function not found", slog.Any("evt", evt))
	s.Publish(protocol.ErrorOccurred{
		Type:      protocol.ErrorOccurredT,
		RequestId: evt.RequestId,
		Message:   fmt.Sprintf("cannot call function: no such pointer found: %d", evt.Ptr),
	})

	return
}

func (s *Scope) handleNewComponentRequested(evt protocol.NewComponentRequested) {
	fac := s.factories[evt.Factory]
	var component Component
	if fac == nil {
		slog.Error("frontend requested unknown factory", slog.String("path", string(evt.Factory)), slog.Int("requestId", int(evt.RequestId)))
		fac = s.factories["_"]
		if fac != nil {
			notFoundComponent := fac(s, evt)
			if notFoundComponent == nil {
				slog.Error("notFound factory returned a nil component which is not allowed", slog.String("id", "_"), slog.Int("requestId", int(evt.RequestId)))
				return
			}
			component = notFoundComponent
		} else {
			s.Publish(protocol.ErrorOccurred{
				Type:      protocol.ErrorOccurredT,
				RequestId: evt.RequestId,
				Message:   fmt.Sprintf("factory %s not found", evt.Factory),
			})
			return
		}

	} else {
		component = fac(s, evt)
		if component == nil {
			slog.Error("factory returned a nil component which is not allowed", slog.String("id", string(evt.Factory)), slog.Int("requestId", int(evt.RequestId)))
			s.Publish(protocol.ErrorOccurred{
				Type:      protocol.ErrorOccurredT,
				RequestId: evt.RequestId,
				Message:   fmt.Sprintf("internal factory error: delivered null component"),
			})
			return
		}
	}

	s.allocatedComponents[component.ID()] = allocatedComponent{
		Component:   component,
		RenderState: NewRenderState(),
	}

	// an allocation without rendering does not make sense, just perform in the same cycle
	renderTree := s.render(evt.RequestId, component)
	s.Publish(renderTree)

}

func (s *Scope) handleComponentInvalidationRequested(evt protocol.ComponentInvalidationRequested) {
	alloc, ok := s.allocatedComponents[evt.Component]
	if !ok {
		slog.Error("cannot invalidate: no such component in scope", slog.Any("evt", evt))
		s.Publish(protocol.ErrorOccurred{
			Type:      protocol.ErrorOccurredT,
			RequestId: evt.RequestId,
			Message:   fmt.Sprintf("cannot invalidate: no such component in scope: %d", evt.Component),
		})
		return
	}

	renderTree := s.render(evt.RequestId, alloc.Component)
	s.Publish(renderTree)
}

func (s *Scope) handleConfigurationRequested(evt protocol.ConfigurationRequested) {
	// todo where is the configuration?
	s.Publish(protocol.ConfigurationDefined{
		Type:             protocol.ConfigurationDefinedT,
		ApplicationName:  "todo",
		AvailableLocales: []string{"de", "en"},
		ActiveLocale:     "de",
		Themes:           protocol.Themes{},
		Resources:        protocol.Resources{},
		RequestId:        evt.RequestId,
	})
}

func (s *Scope) handleComponentDestructionRequested(evt protocol.ComponentDestructionRequested) {
	component, ok := s.allocatedComponents[evt.Component]
	if !ok {
		s.Publish(protocol.ErrorOccurred{
			Type:      protocol.ErrorOccurredT,
			RequestId: evt.RequestId,
			Message:   fmt.Sprintf("no such component: %d", evt.Component),
		})

		return
	}

	invokeDestructors(component)

	delete(s.allocatedComponents, evt.Component)

	s.sendAck(evt.RequestId)
}
