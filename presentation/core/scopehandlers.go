package core

import (
	"encoding/json"
	"fmt"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/ora"
	"log/slog"
)

// only for event loop
func (s *Scope) handleEvent(t ora.Event, ackRequired bool) {
	if ackRequired {
		defer s.sendAck(t.ReqID())
	}
	/*
		defer func() {
			if ackRequired {
				slog.Info(fmt.Sprintf("handleEvent eolDone: %d %T", t.ReqID(), t))
			}
		}*/
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
	case ora.Ping:
	// do nothing, we already applied our keep-alive-tick
	case ora.WindowInfoChanged:
		s.handleWindowInfoChanged(evt)
	default:
		slog.Error("unexpected event type in scope processing", slog.String("type", fmt.Sprintf("%T", evt)))
	}
}

func (s *Scope) handleWindowInfoChanged(evt ora.WindowInfoChanged) {
	winfo := evt.Info
	s.updateWindowInfo(WindowInfo{
		Width:       DP(winfo.Width),
		Height:      DP(winfo.Height),
		Density:     Density(winfo.Density),
		SizeClass:   WindowSizeClass(winfo.SizeClass),
		ColorScheme: ColorScheme(winfo.ColorScheme),
	})
}

func (s *Scope) handleEventsAggregated(evt ora.EventsAggregated) {
	for _, event := range evt.Events {
		s.handleEvent(event, false)
	}
}

func (s *Scope) handleScopeDestructionRequested(evt ora.ScopeDestructionRequested) {
	//s.destroy()
	//s.eventLoop.Destroy() // discards everything else queued
	s.Destroy()
}

func (s *Scope) handleSetPropertyValueRequested(evt ora.SetPropertyValueRequested) {
	alloc, err := s.allocatedRootView.Get()
	if err != nil {
		slog.Error("no component has been allocated in scope")
		s.Publish(ora.ErrorOccurred{
			Type:      ora.ErrorOccurredT,
			RequestId: evt.RequestId,
			Message:   fmt.Sprintf("no component has been allocated in scope: %d", evt.Ptr),
		})
		return
	}

	state, ok := alloc.states[evt.Ptr]
	if !ok {
		slog.Error("property not found", slog.Any("evt", evt))
		s.Publish(ora.ErrorOccurred{
			Type:      ora.ErrorOccurredT,
			RequestId: evt.RequestId,
			Message:   fmt.Sprintf("cannot set property: no such pointer found: %d", evt.Ptr),
		})
		return
	}

	if err := state.parse(evt.Value); err != nil {
		slog.Error("invalid property value", slog.Any("evt", evt), slog.String("property-type", fmt.Sprintf("%T", state)))
		s.Publish(ora.ErrorOccurred{
			Type:      ora.ErrorOccurredT,
			RequestId: evt.RequestId,
			Message:   fmt.Sprintf("cannot set property: invalid property value: %d", evt.Ptr),
		})
	}

}

func (s *Scope) handleFunctionCallRequested(evt ora.FunctionCallRequested) {
	alloc, err := s.allocatedRootView.Get()
	if err != nil {
		s.Publish(ora.ErrorOccurred{
			Type:      ora.ErrorOccurredT,
			RequestId: evt.RequestId,
			Message:   fmt.Sprintf("cannot call function: no view allocated: %d", evt.Ptr),
		})
		return
	}

	fn := alloc.callbacks[evt.Ptr]
	if fn == nil {
		s.Publish(ora.ErrorOccurred{
			Type:      ora.ErrorOccurredT,
			RequestId: evt.RequestId,
			Message:   fmt.Sprintf("cannot call function: no associated function found: %d", evt.Ptr),
		})
		return
	}

	fn()

}

func (s *Scope) handleNewComponentRequested(evt ora.NewComponentRequested) {
	s.destroyView()
	s.updateWindowInfo(s.windowInfo)

	window := newScopeWindow(s, evt.Factory, evt.Values)
	fac := s.factories[evt.Factory]
	if fac == nil {
		slog.Error("frontend requested unknown factory", slog.String("path", string(evt.Factory)), slog.Int("requestId", int(evt.RequestId)))
		fac = s.factories["_"]
		if fac == nil {
			s.Publish(ora.ErrorOccurred{
				Type:      ora.ErrorOccurredT,
				RequestId: evt.RequestId,
				Message:   fmt.Sprintf("factory %s not found", evt.Factory),
			})
			return
		}

	}

	window.setFactory(fac)
	s.allocatedRootView = std.Some(window)

	// an allocation without rendering does not make sense, just perform in the same cycle
	renderTree := s.render(evt.RequestId, window)
	s.Publish(renderTree)

}

func (s *Scope) handleComponentInvalidationRequested(evt ora.ComponentInvalidationRequested) {
	alloc, err := s.allocatedRootView.Get()
	if err != nil {
		slog.Error("cannot invalidate: no such component in scope", slog.Any("evt", evt))
		s.Publish(ora.ErrorOccurred{
			Type:      ora.ErrorOccurredT,
			RequestId: evt.RequestId,
			Message:   fmt.Sprintf("cannot invalidate: no such component in scope: %d", evt.Component),
		})
		return
	}

	if alloc.destroyed {
		return
	}

	renderTree := s.render(evt.RequestId, alloc)
	s.Publish(renderTree)
}

func convertColorSetToMap(colorSet ColorSet) map[string]ora.Color {
	// expensive but simple variant of going typesafe to arbitrary
	var res map[string]ora.Color
	buf, err := json.Marshal(colorSet)
	if err != nil {
		panic(fmt.Errorf("unreachable: %w", err))
	}

	err = json.Unmarshal(buf, &res)
	if err != nil {
		panic(fmt.Errorf("unreachable: %w", err))
	}

	return res
}

func (s *Scope) handleConfigurationRequested(evt ora.ConfigurationRequested) {
	winfo := evt.WindowInfo
	s.windowInfo = WindowInfo{
		Width:       DP(winfo.Width),
		Height:      DP(winfo.Height),
		Density:     Density(winfo.Density),
		SizeClass:   WindowSizeClass(winfo.SizeClass),
		ColorScheme: ColorScheme(winfo.ColorScheme),
	}
	s.updateWindowInfo(s.windowInfo)

	themes := ora.Themes{
		Dark: ora.Theme{
			Colors: map[ora.NamespaceName]map[string]ora.Color{},
		},
		Light: ora.Theme{
			Colors: map[ora.NamespaceName]map[string]ora.Color{},
		},
	}

	for scheme, m := range s.app.colorSets {
		for name, set := range m {
			switch scheme {
			case Dark:
				themes.Dark.Colors[ora.NamespaceName(name)] = convertColorSetToMap(set)
			case Light:
				themes.Light.Colors[ora.NamespaceName(name)] = convertColorSetToMap(set)
			default:
				panic("implement me")
			}
		}
	}

	s.Publish(ora.ConfigurationDefined{
		Type:               ora.ConfigurationDefinedT,
		ApplicationID:      string(s.app.id),
		ApplicationName:    s.app.name,
		ApplicationVersion: s.app.version,
		AppIcon:            s.app.appIcon,
		AvailableLocales:   []string{"de", "en"}, //TODO
		ActiveLocale:       s.locale.String(),
		Themes:             themes,
		RequestId:          evt.RequestId,
	})
}

func (s *Scope) handleComponentDestructionRequested(evt ora.ComponentDestructionRequested) {
	_, err := s.allocatedRootView.Get()
	if err != nil {
		//slog.Info("e1")
		s.Publish(ora.ErrorOccurred{
			Type:      ora.ErrorOccurredT,
			RequestId: evt.RequestId,
			Message:   fmt.Sprintf("no such component: %d", evt.Component),
		})
		//slog.Info("e2")

		return
	}

	s.destroyView()

}

func (s *Scope) destroyView() {
	alloc, err := s.allocatedRootView.Get()
	if err != nil {
		slog.Error("no root view to destroy, ignoring")
		return
	}

	alloc.destroy()
	s.allocatedRootView = std.None[*scopeWindow]()
}
