package core

import (
	"encoding/json"
	"fmt"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/proto"
	"log/slog"
)

// only for event loop
func (s *Scope) handleEvent(t proto.NagoEvent) {

	switch evt := t.(type) {
	case *proto.UpdateStateValueRequested:
		s.handleSetPropertyValueRequested(evt)
	case *proto.FunctionCallRequested:
		s.handleFunctionCallRequested(evt)
	case *proto.RootViewAllocationRequested:
		s.handleNewComponentRequested(evt)
	case *proto.RootViewRenderingRequested:
		s.handleComponentInvalidationRequested(evt)
	case *proto.ScopeConfigurationChangeRequested:
		s.handleConfigurationRequested(evt)
	case *proto.RootViewDestructionRequested:
		s.handleComponentDestructionRequested(evt)
	case *proto.ScopeDestructionRequested:
		s.handleScopeDestructionRequested(evt)
	case *proto.SessionAssigned:
		s.handleSessionAssigned(evt)
	case *proto.Ping:
	// do nothing, we already applied our keep-alive-tick
	case *proto.WindowInfoChanged:
		s.handleWindowInfoChanged(evt)
	default:
		slog.Error("unexpected event type in scope processing", slog.String("type", fmt.Sprintf("%T", evt)))
	}
}

func (s *Scope) handleWindowInfoChanged(evt *proto.WindowInfoChanged) {
	winfo := evt.WindowInfo
	s.updateWindowInfo(WindowInfo{
		Width:       DP(winfo.Width),
		Height:      DP(winfo.Height),
		Density:     Density(winfo.Density),
		SizeClass:   WindowSizeClass(winfo.SizeClass),
		ColorScheme: ColorScheme(winfo.ColorScheme),
	})
}

func (s *Scope) handleScopeDestructionRequested(evt *proto.ScopeDestructionRequested) {
	//s.destroy()
	//s.eventLoop.Destroy() // discards everything else queued
	s.Destroy()
}

func (s *Scope) handleSetPropertyValueRequested(evt *proto.UpdateStateValueRequested) {
	if s.allocatedRootView.IsNone() {
		s.Publish(&proto.ErrorRootViewAllocationRequired{RID: evt.GetRID()})
		return
	}

	alloc := s.allocatedRootView.Unwrap()

	if evt.StatePointer.IsZero() {
		return
	}

	state, ok := alloc.states[evt.StatePointer]
	if !ok {
		slog.Error("property not found", slog.Any("evt", evt))
		s.Publish(&proto.ErrorOccurred{
			Message: proto.Str(fmt.Sprintf("cannot set property: no such pointer found: %d", evt.StatePointer)),
		})
		return
	}

	if err := state.parse(string(evt.Value)); err != nil {
		slog.Error("invalid property value", slog.Any("evt", evt), slog.String("property-type", fmt.Sprintf("%T", state)))
		s.Publish(&proto.ErrorOccurred{
			RID:     evt.RID,
			Message: proto.Str(fmt.Sprintf("cannot set property: invalid property value: %d", evt.StatePointer)),
		})
	}

	if evt.FunctionPointer.IsZero() {
		return
	}

	s.handleFunctionCallRequested(&proto.FunctionCallRequested{
		RID: evt.RID,
		Ptr: evt.FunctionPointer,
	})
}

func (s *Scope) handleFunctionCallRequested(evt *proto.FunctionCallRequested) {
	if s.allocatedRootView.IsNone() {
		s.Publish(&proto.ErrorRootViewAllocationRequired{
			RID: evt.RID,
		})

		return
	}

	alloc := s.allocatedRootView.Unwrap()
	fn := alloc.callbacks[evt.Ptr]
	if fn == nil {
		s.Publish(&proto.ErrorOccurred{
			RID:     evt.RID,
			Message: proto.Str(fmt.Sprintf("cannot call function: no associated function found: %d", evt.Ptr)),
		})
		return
	}

	fn()
}

func (s *Scope) handleNewComponentRequested(evt *proto.RootViewAllocationRequested) {
	s.destroyView()
	s.updateWindowInfo(s.windowInfo)

	window := newScopeWindow(s, evt.Factory, newValuesFromProto(evt.Values))
	fac := s.factories[evt.Factory]
	if fac == nil {
		slog.Error("frontend requested unknown factory", slog.String("path", string(evt.Factory)), slog.Int("requestId", int(evt.RID)))
		fac = s.factories["_"]
		if fac == nil {
			s.Publish(&proto.ErrorOccurred{
				RID:     evt.RID,
				Message: proto.Str(fmt.Sprintf("factory %s not found", evt.Factory)),
			})
			return
		}
	}

	window.setFactory(fac)
	s.allocatedRootView = std.Some(window)

	// an allocation without rendering does not make sense, just perform in the same cycle
	renderTree := s.render(evt.RID, window)
	s.Publish(renderTree)

}

func (s *Scope) handleComponentInvalidationRequested(evt *proto.RootViewRenderingRequested) {
	if s.allocatedRootView.IsNone() {
		s.Publish(&proto.ErrorRootViewAllocationRequired{RID: evt.GetRID()})
		return
	}

	alloc := s.allocatedRootView.Unwrap()

	if alloc.destroyed {
		return
	}

	renderTree := s.render(evt.RID, alloc)
	s.Publish(renderTree)
}

func convertColorSetToMap(colorSet ColorSet) proto.NamedColors {
	// expensive but simple variant of going typesafe to arbitrary, but this only happens once per frontend instantiation
	var res proto.NamedColors
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

func (s *Scope) handleConfigurationRequested(evt *proto.ScopeConfigurationChangeRequested) {
	winfo := evt.WindowInfo
	s.windowInfo = WindowInfo{
		Width:       DP(winfo.Width),
		Height:      DP(winfo.Height),
		Density:     Density(winfo.Density),
		SizeClass:   WindowSizeClass(winfo.SizeClass),
		ColorScheme: ColorScheme(winfo.ColorScheme),
	}
	s.updateWindowInfo(s.windowInfo)

	themes := proto.Themes{
		Dark: proto.Theme{
			Colors: proto.NamespacedColors{},
		},
		Light: proto.Theme{
			Colors: proto.NamespacedColors{},
		},
	}

	for scheme, m := range s.app.colorSets {
		for name, set := range m {
			switch scheme {
			case Dark:
				themes.Dark.Colors[proto.NamespaceName(name)] = convertColorSetToMap(set)
			case Light:
				themes.Light.Colors[proto.NamespaceName(name)] = convertColorSetToMap(set)
			default:
				panic("implement me")
			}
		}
	}

	s.Publish(&proto.ScopeConfigurationChanged{
		ApplicationID:      proto.Str(s.app.id),
		ApplicationName:    proto.Str(s.app.name),
		ApplicationVersion: proto.Str(s.app.version),
		AppIcon:            proto.URI(s.app.appIcon),
		AvailableLocales:   proto.Locales{"de", "en"}, //TODO
		ActiveLocale:       proto.Locale(s.locale.String()),
		Themes:             themes,
		RID:                evt.RID,
	})
}

func (s *Scope) handleComponentDestructionRequested(evt *proto.RootViewDestructionRequested) {
	if s.allocatedRootView.IsNone() {
		// already destroyed, just ignore that
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
