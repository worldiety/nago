package application

import (
	"context"
	"go.wdy.de/nago/persistence/kv"
	"go.wdy.de/nago/presentation/ui2"
	"io/fs"
	"log/slog"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
)

type Configurator struct {
	appName   string
	kvStores  map[string]kv.Store
	ctx       context.Context
	done      context.CancelFunc
	logger    *slog.Logger
	debug     bool
	pages     map[ui2.PageID]scaffoldPage
	endpoints []ui2.Endpoint
	auth      authProviders
	fsys      []fs.FS
	uiApp     *ui2.Application
}

func NewConfigurator() *Configurator {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	return &Configurator{
		kvStores: make(map[string]kv.Store),
		pages:    make(map[ui2.PageID]scaffoldPage),
		ctx:      ctx,
		done:     done,
		uiApp:    &ui2.Application{},
		debug:    strings.Contains(strings.ToLower(runtime.GOOS), "windows") || strings.Contains(strings.ToLower(runtime.GOOS), "darwin"),
	}
}

// Name sets the applications name.
func (c *Configurator) Name(name string) *Configurator {
	c.appName = name
	c.uiApp.Name = name
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
