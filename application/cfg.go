package application

import (
	"context"
	"go.wdy.de/nago/persistence/kv"
	"go.wdy.de/nago/presentation/ui"
	"log/slog"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
)

type Configurator struct {
	appName  string
	kvStores map[string]kv.Store
	ctx      context.Context
	done     context.CancelFunc
	logger   *slog.Logger
	debug    bool
	pages    map[string]ui.PageHandler
	auth     authProviders
}

func NewConfigurator() *Configurator {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	return &Configurator{
		kvStores: make(map[string]kv.Store),
		pages:    make(map[string]ui.PageHandler),
		ctx:      ctx,
		done:     done,
		debug:    strings.Contains(strings.ToLower(runtime.GOOS), "windows") || strings.Contains(strings.ToLower(runtime.GOOS), "darwin"),
	}
}

// Name sets the applications name.
func (c *Configurator) Name(name string) *Configurator {
	c.appName = name
	return c
}

// Context returns the applications default context.
func (c *Configurator) Context() context.Context {
	return c.ctx
}

// Debug sets the debug flag.
func (c *Configurator) Debug(isDebug bool) *Configurator {
	c.debug = isDebug
	return c
}
