// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application

import (
	"context"
	"fmt"
	"go.wdy.de/nago/application/admin"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/blob/tdb"
	"go.wdy.de/nago/pkg/events"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
	"io/fs"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
)

type dependency struct {
	name    string
	service any
}

type EntityStorageFactory interface {
	EntityStore(bucketName string) (blob.Store, error)
}

type BlobStorageFactory interface {
	BlobStore(bucketName string) (blob.Store, error)
}

type Configurator struct {
	fileStores                 map[string]blob.Store
	entityStores               map[string]blob.Store
	globalTDB                  *tdb.DB
	ctx                        context.Context
	done                       context.CancelFunc
	logger                     *slog.Logger
	debug                      bool
	fsys                       []fs.FS
	host                       string
	port                       int
	scheme                     string
	applicationID              core.ApplicationID
	applicationName            string
	applicationVersion         string
	applicationSemanticVersion string
	dataDir                    string
	factories                  map[proto.RootViewID]func(wnd core.Window) core.View
	onWindowCreatedObservers   []core.OnWindowCreatedObserver
	destructors                []func()
	app                        *core.Application // may be nil
	rawEndpoint                []rawEndpoint
	colorSets                  map[core.ColorScheme]map[core.NamespaceName]core.ColorSet
	appIconUri                 proto.URI
	fps                        int

	systemServices         []dependency
	mailManagement         *MailManagement
	mailManagementMutator  func(*MailManagement)
	userManagement         *UserManagement
	roleManagement         *RoleManagement
	settingsManagement     *SettingsManagement
	adminManagement        *AdminManagement
	adminManagementGroups  []func(uid auth.Subject) admin.Group
	adminManagementMutator func(m *AdminManagement)
	sessionManagement      *SessionManagement
	permissionManagement   *PermissionManagement
	groupManagement        *GroupManagement
	licenseManagement      *LicenseManagement
	billingManagement      *BillingManagement
	imageManagement        *ImageManagement
	backupManagement       *BackupManagement
	secretManagement       *SecretManagement
	templateManagement     *TemplateManagement
	themeManagement        *ThemeManagement
	tokenManagement        *TokenManagement
	decorator              Decorator
	eventBus               events.EventBus
	contextPath            atomic.Pointer[string]
}

func NewConfigurator() *Configurator {
	printEnv()

	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	var buildInfo string
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				buildInfo = setting.Value
				break
			}
		}
	}

	if buildInfo == "" {
		buildInfo = fmt.Sprintf("%s %s", runtime.GOOS, runtime.GOARCH)
	}

	cfg := &Configurator{
		colorSets: map[core.ColorScheme]map[core.NamespaceName]core.ColorSet{
			core.Dark:  {},
			core.Light: {},
		},
		applicationSemanticVersion: "0.0.0",
		fileStores:                 map[string]blob.Store{},
		entityStores:               map[string]blob.Store{},
		fps:                        10,
		ctx:                        ctx,
		done:                       done,
		factories:                  map[proto.RootViewID]func(wnd core.Window) core.View{},
		applicationName:            filepath.Base(os.Args[0]),
		applicationVersion:         buildInfo,
		debug:                      strings.Contains(strings.ToLower(runtime.GOOS), "windows") || strings.Contains(strings.ToLower(runtime.GOOS), "darwin"),
	}

	return cfg
}

func printEnv() {
	for _, s := range os.Environ() {
		slog.Info(s)
	}
}

func (c *Configurator) secureCookie() bool {
	if strV, ok := os.LookupEnv("NAGO_COOKIES_INSECURE"); ok {
		if ok, _ := strconv.ParseBool(strV); ok {
			return false
		} else {
			return true
		}
	}

	return !c.debug
}

type envVarConfig struct {
	key      string
	required bool
	cb       func(envVarConfig, string, *Configurator, *slog.Logger) error
}

var envConfig []envVarConfig = []envVarConfig{
	{
		key:      "NAGO_HOST",
		required: false,
		cb: func(env envVarConfig, s string, cfg *Configurator, logger *slog.Logger) error {
			cfg.SetHost(s)
			return nil
		}},
	{
		key:      "NAGO_PORT",
		required: false,
		cb: func(env envVarConfig, s string, cfg *Configurator, logger *slog.Logger) error {
			parsed, err := strconv.Atoi(s)
			if err != nil {
				return fmt.Errorf("Invalid port value %v in %s: %w", s, env.key, err)
			}
			cfg.port = parsed
			return nil
		}},
}

func (c *Configurator) LoadConfigFromEnv() {
	logger := c.defaultLogger()
	for _, envVar := range envConfig {
		readFromEnv := os.Getenv(envVar.key)
		if readFromEnv == "" {
			if envVar.required {
				panic(fmt.Sprintf("Required environment variable %s not found", envVar.key))
			}
			continue
		}

		if err := envVar.cb(envVar, readFromEnv, c, logger); err != nil {
			panic(fmt.Sprintf("failed to set config from env var %s: %v", envVar.key, err))
		}
	}
}

func (c *Configurator) AddSystemService(name string, service any) *Configurator {
	c.systemServices = append(c.systemServices, dependency{
		name:    name,
		service: service,
	})

	return c
}

func SystemServiceFor[T any](cfg *Configurator, name string) (T, bool) {
	for _, service := range cfg.systemServices {
		if v, ok := service.service.(T); ok {
			if name == "" || service.name == name {
				return v, ok
			}
		}
	}

	var zero T

	return zero, false
}

func (c *Configurator) AddOnWindowCreatedObserver(observer core.OnWindowCreatedObserver) *Configurator {
	c.onWindowCreatedObservers = append(c.onWindowCreatedObservers, observer)
	return c
}

func (c *Configurator) OnDestroy(f func()) {
	c.destructors = append(c.destructors, f)
}

func (c *Configurator) AppIcon(ico core.URI) *core.Application {
	c.appIconUri = proto.URI(ico)
	return c.app
}

// DataDir returns the most private data directory, which is accessible. If not manually set, initialize as follows:
//   - use STATE_DIRECTORY from systemd, if defined or
//   - use user home if available or
//   - use working dir if available or
//   - use temp dir if available
//   - append .nago/<application id>
//   - ensure directory with 0700 to only allow owner to access
func (c *Configurator) DataDir() string {
	if c.dataDir == "" {
		if sysdStateDir, ok := os.LookupEnv("STATE_DIRECTORY"); ok {
			c.dataDir = strings.Split(sysdStateDir, ":")[0]
		}
	}

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

		dataDir = filepath.Join(dataDir, ".nago", string(c.ApplicationID()))

		_ = os.MkdirAll(dataDir, 0700) // security: only owner can read,write,exec
		c.SetDataDir(dataDir)
	}

	return c.dataDir
}

func (c *Configurator) SetDataDir(dir string) {
	c.dataDir = dir
	slog.Info("data directory updated", slog.String("dir", c.dataDir))
}

// SetFPS sets the internal application-wide update rate, at which state-induced rendering shall at most happen.
// This will cause a linear overhead depending on the amount of active scopes and states. Default is 10fps, which
// means that the application will check in 100ms slices, if a state has changed within a window and trigger a render.
// The higher the fps, the more CPU cycles will be burned. Keep in mind, that you have a website with a
// high-latency and small throughput channel. If you are unsure, if you need to adjust this, you probably won't need
// it.
func (c *Configurator) SetFPS(fps int) {
	c.fps = fps
}

// ContextPath returns something like localhost:3000 or whatever has been set or autodetected.
// It should contain the primary DNS name and port and eventually a path postfix, whatever is necessary.
// This path may be validated against any requests and is used when generating links.
// If undefined, it is taken from the first http wire request and represents the host part.
func (c *Configurator) ContextPath() string {
	if s := c.contextPath.Load(); s != nil {
		return *s
	}

	return ""
}

func (c *Configurator) ContextPathURI(path string, query core.Values) string {
	p := c.ContextPath()
	if p == "" {
		p = fmt.Sprintf("http://localhost:%d", c.port)
	}

	if !strings.HasSuffix(p, "/") {
		p = p + "/"
	}

	p += path

	if query != nil {
		p += "?" + query.URLEncode()
	}

	if !strings.HasPrefix(p, "http") {
		p = "http://" + p
	}

	return p
}

func (c *Configurator) SetContextPath(path string) {
	c.contextPath.Store(&path)
}

func (c *Configurator) directory(name string) string {
	name = filepath.Clean(name) // security: avoid path traversal attacks here
	path := filepath.Join(c.DataDir(), name)

	return path
}

// Directory returns an allocated local directory underneath the data dir
func (c *Configurator) Directory(name string) string {
	path := c.directory(name)
	// security: only owner can read,write,exec
	if err := os.MkdirAll(path, 0700); err != nil {
		panic(fmt.Errorf("irrecoverable denied directory access: %w", err))
	}

	slog.Info("directory created", slog.String("path", path))
	return path
}

func (c *Configurator) ApplicationID() core.ApplicationID {
	if c.applicationID == "" {
		panic("application id has not been set")
	}
	return c.applicationID
}

// SetApplicationID should be something like com.example.myapp
func (c *Configurator) SetApplicationID(id core.ApplicationID) {
	if !id.Valid() {
		panic(fmt.Errorf("invalid application id: %v", id))
	}

	c.applicationID = id
	slog.Info("application id updated", slog.String("id", string(id)))
}

// SetName sets the applications name which is used to communicate with the user of the application.
// This is likely not the internal code name, but instead the advertising or marketing product name.
func (c *Configurator) SetName(name string) {
	c.applicationName = name
}

func (c *Configurator) Name() string {
	return c.applicationName
}

// SetVersion sets the applications version to something arbitrary. It is best practice to include information
// about the build environment and git commit hash. This is likely shown in error reports or logs.
// This must not be used as a marketing version for the customer, because a marketing version does not change
// when bug fixes are released. See also [Configurator.SemanticVersion].
func (c *Configurator) SetVersion(version string) {
	c.applicationVersion = version
}

func (c *Configurator) Version() string {
	return c.applicationVersion
}

// SetSemanticVersion updates the version which is communicated towards the customer and should be in SemVer style
// like 1.2.3
func (c *Configurator) SetSemanticVersion(version string) {
	c.applicationSemanticVersion = version
}

func (c *Configurator) SemanticVersion() string {
	return c.applicationSemanticVersion
}

func (c *Configurator) VCSVersion() string {
	return c.applicationVersion
}

// Host returns the host to which the binding should be made. This is different from [Configurator.ContextPath].
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

// Context returns the applications default context.
func (c *Configurator) Context() context.Context {
	return c.ctx
}

// Debug sets the debug flag.
func (c *Configurator) Debug(isDebug bool) *Configurator {
	c.debug = isDebug
	return c
}

func (c *Configurator) EventBus() events.EventBus {
	if c.eventBus == nil {
		c.eventBus = events.NewEventBus()
	}

	return c.eventBus
}

// StandardSystems enables the admin center all stuff like user, session, mail, self service etc. systems.
// If you want to customize that, just don't call this and instead configure use each individual system for itself.
func (c *Configurator) StandardSystems() error {
	if _, err := c.AdminManagement(); err != nil {
		return err
	}

	if _, err := c.UserManagement(); err != nil {
		return err
	}

	if _, err := c.BackupManagement(); err != nil {
		return err
	}

	if _, err := c.LicenseManagement(); err != nil {
		return err
	}

	if _, err := c.MailManagement(); err != nil {
		return err
	}

	if _, err := c.SecretManagement(); err != nil {
		return err
	}

	if _, err := c.TemplateManagement(); err != nil {
		return err
	}

	if _, err := c.SessionManagement(); err != nil {
		return err
	}

	return nil
}
