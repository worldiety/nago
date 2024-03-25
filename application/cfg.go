package application

import (
	"context"
	"go.etcd.io/bbolt"
	"go.wdy.de/nago/persistence/kv"
	"go.wdy.de/nago/presentation/ui"
	"io/fs"
	"log/slog"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
)

type Configurator struct {
	appName string
	// deprecated
	kvStores   map[string]kv.Store
	boltStores map[string]*bbolt.DB
	ctx        context.Context
	done       context.CancelFunc
	logger     *slog.Logger
	debug      bool
	auth       authProviders
	fsys       []fs.FS
	uiApp      *ui.Application
	host       string
	port       int
	scheme     string
}

func NewConfigurator() *Configurator {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	return &Configurator{
		kvStores:   make(map[string]kv.Store),
		boltStores: map[string]*bbolt.DB{},
		ctx:        ctx,
		done:       done,
		uiApp: &ui.Application{
			LivePages: make(map[ui.PageID]func(wire ui.Wire) *ui.Page),
		},
		debug: strings.Contains(strings.ToLower(runtime.GOOS), "windows") || strings.Contains(strings.ToLower(runtime.GOOS), "darwin"),
	}
}

func (c *Configurator) getHost() string {
	if c.host != "" {
		return c.host
	}

	return "localhost"
}

func (c *Configurator) SetHost(host string) *Configurator {
	c.host = host
	return c
}

func (c *Configurator) getPort() int {
	if c.port != 0 {
		return c.port
	}

	return 3000
}

func (c *Configurator) getScheme() string {
	if c.scheme != "" {
		return c.scheme
	}

	return "http"
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
