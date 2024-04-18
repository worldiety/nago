package application

import (
	"context"
	"fmt"
	"go.etcd.io/bbolt"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
	"go.wdy.de/nago/presentation/ui"
	"io/fs"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"syscall"
)

type Configurator struct {
	appName       string
	boltStore     *bbolt.DB
	ctx           context.Context
	done          context.CancelFunc
	logger        *slog.Logger
	debug         bool
	auth          authProviders
	fsys          []fs.FS
	uiApp         *ui.Application
	host          string
	port          int
	scheme        string
	applicationID ApplicationID
	dataDir       string
}

var appIdRegex = regexp.MustCompile(`^[a-z]\w*(\.[a-z]\w*)+$`)

type ApplicationID string

func (a ApplicationID) Valid() bool {
	return appIdRegex.FindString(string(a)) == string(a)
}

func NewConfigurator() *Configurator {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	return &Configurator{
		ctx:  ctx,
		done: done,
		uiApp: &ui.Application{
			Components: map[ora.ComponentFactoryId]func(realm core.Window) core.Component{},
		},
		debug: strings.Contains(strings.ToLower(runtime.GOOS), "windows") || strings.Contains(strings.ToLower(runtime.GOOS), "darwin"),
	}
}

func (c *Configurator) DataDir() string {
	if c.dataDir == "" {
		dataDir, err := os.UserHomeDir()
		if err != nil {
			slog.Error("cannot get user home dir", "err", err)
			cwd, err := os.Getwd()
			if err != nil {
				slog.Error("cannot get current working directory", "err", err)
				dataDir = os.TempDir()
			}

			dataDir = cwd
		}

		dataDir = filepath.Join(dataDir, fmt.Sprintf(".%s", c.ApplicationID()))
		_ = os.MkdirAll(dataDir, 0700) // security: only owner can read,write,exec
		c.SetDataDir(dataDir)
	}

	return c.dataDir
}

func (c *Configurator) SetDataDir(dir string) {
	c.dataDir = dir
	slog.Info("data directory updated", slog.String("dir", c.dataDir))
}

// Directory returns an allocated local directory underneath the
func (c *Configurator) Directory(name string) string {
	name = filepath.Clean(name) // security: avoid path traversal attacks here
	path := filepath.Join(c.DataDir(), name)
	_ = os.MkdirAll(path, 0700) // security: only owner can read,write,exec
	return path
}

func (c *Configurator) ApplicationID() ApplicationID {
	if c.applicationID == "" {
		panic("application id has not been set")
	}
	return c.applicationID
}

// SetApplicationID should be something like com.example.myapp
func (c *Configurator) SetApplicationID(id ApplicationID) {
	if !id.Valid() {
		panic(fmt.Errorf("invalid application id: %v", id))
	}

	c.applicationID = id
	slog.Info("application id updated", slog.String("id", string(id)))
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
