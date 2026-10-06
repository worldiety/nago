// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application

import (
	"archive/zip"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/fs"
	"log"
	"log/slog"
	"maps"
	"mime"
	"net/http"
	"path/filepath"
	"runtime/debug"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/gorilla/websocket"
	"github.com/laher/mergefs"
	"github.com/vearutop/statigz"
	"github.com/worldiety/option"
	"go.wdy.de/nago/application/image/http"
	"go.wdy.de/nago/application/session"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/logging"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/core/http/gorilla"
	"go.wdy.de/nago/presentation/proto"
	"go.wdy.de/nago/presentation/ui/alert"
)

// RootViewMeta is the descriptive metadata attached to a root view at registration time. See [Purpose].
type RootViewMeta struct {
	// Purpose describes, in one sentence and in the user's terms, what a person does on this screen. Empty
	// when the registration did not state one.
	Purpose string
}

// RootViewOption augments a root view registration with metadata. See [Purpose].
type RootViewOption func(*RootViewMeta)

// Purpose states, in one sentence, what a user does on this screen - "review and settle open invoices", not
// "invoice list component".
//
// It is attached to the registration rather than to a menu entry because the registration is the only
// declaration every root view has: not every screen appears in a menu, and menus are rearranged freely.
//
// Its first consumer is the AI assistant, which tells the model where the user currently stands (see
// uicompletion.WindowContext). Applications that describe their screens here get that for free instead of
// maintaining a second, hand-written map of routes that inevitably drifts from the real ones.
func Purpose(text string) RootViewOption {
	return func(m *RootViewMeta) {
		m.Purpose = text
	}
}

// RootView registers a factory to create a [core.View] within a [core.Scope].
// For example, a web browser will create at least a single ViewRoot for each open tab.
// Note, that leading or succeeding slashes in the factory ids are not allowed, otherwise you can
// use them in arbitrary ways.
// Keep in mind, that web browsers will expose these ids to the user and they become part of your public
// API or contract with the user. A user may bookmark them.
//
// You cannot use path variables. Instead, use [core.Values] to transport a state from one ViewRoot
// (or window) to another.
//
// Optional [RootViewOption]s describe the view; see [Purpose].
func (c *Configurator) RootView(viewRootID core.NavigationPath, factory func(wnd core.Window) core.View, opts ...RootViewOption) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	id := proto.RootViewID(viewRootID)
	if !id.Valid() {
		panic(fmt.Errorf("invalid component factory id: %v", id))
	}

	if _, ok := c.factories[id]; ok {
		panic(fmt.Errorf("another factory with id %v has already been registered", id))
	}

	c.factories[id] = factory

	if len(opts) > 0 {
		var meta RootViewMeta
		for _, opt := range opts {
			opt(&meta)
		}

		if c.rootViewMeta == nil {
			c.rootViewMeta = map[proto.RootViewID]RootViewMeta{}
		}

		c.rootViewMeta[id] = meta
	}
}

func (c *Configurator) RootViewWithDecoration(viewRootID core.NavigationPath, factory func(wnd core.Window) core.View, opts ...RootViewOption) {
	c.RootView(viewRootID, c.DecorateRootView(factory), opts...)
}

// A RootViewInterceptor may replace the view of every root view, e.g. to show a setup page instead of any other
// page as long as an instance has not been set up. It returns false to render the requested root view. It is
// invoked on every render, so it must be cheap.
type RootViewInterceptor func(wnd core.Window) (core.View, bool)

// AddRootViewInterceptor installs an interceptor for all root views. Interceptors are asked in the order of their
// installation and the first one wins. They must be installed before the application starts.
func (c *Configurator) AddRootViewInterceptor(interceptor RootViewInterceptor) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.rootViewInterceptors = append(c.rootViewInterceptors, interceptor)
}

// RootViewMetaOf returns the metadata registered for the given root view. The zero value is returned for a
// view that was registered without any [RootViewOption], which is the normal case.
func (c *Configurator) RootViewMetaOf(viewRootID core.NavigationPath) RootViewMeta {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.rootViewMeta[proto.RootViewID(viewRootID)]
}

// RootViewMetas returns the metadata of all root views that were registered with one.
func (c *Configurator) RootViewMetas() map[proto.RootViewID]RootViewMeta {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return maps.Clone(c.rootViewMeta)
}

// RootViews returns the list of registered root views.
func (c *Configurator) RootViews() map[proto.RootViewID]func(wnd core.Window) core.View {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return maps.Clone(c.factories)
}

// Serve registers the given filesystem to be served by the application.
// See also [Configurator.Filesystems].
func (c *Configurator) Serve(fsys fs.FS) *Configurator {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.fsys = append(c.fsys, fsys)
	return c
}

// Filesystems returns the list of file systems that are served by the application at the point of calling.
// See also [Configurator.Serve].
func (c *Configurator) Filesystems() []fs.FS {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return slices.Clone(c.fsys)
}

func nameAndMime(options core.ExportFilesOptions) (name, mimetype string) {
	total := len(options.Files)
	if total == 0 {
		return
	}

	if total > 1 {
		mimetype = "application/zip"
		name = "files.zip"
		return
	}

	if len(options.Files) == 1 {
		name = options.Files[0].Name()
		mimetype, _ = options.Files[0].MimeType()
		if mimetype == "" {
			mimetype = mime.TypeByExtension(filepath.Ext(name))
		}

		return
	}

	return
}

// NewInProcessApplication applies the configured migrations and creates the core application without any HTTP
// server or websocket endpoints. Windows are connected directly using [core.Application.Connect] and an
// in-process [core.EventChannel]. This is intended for in-process drivers like the nagotest package.
// The returned application must be released using [core.Application.Destroy], which also releases this
// Configurator.
func (c *Configurator) NewInProcessApplication() (*core.Application, error) {
	if mg := c.migrations; mg != nil {
		if err := mg.Apply(c.Context()); err != nil {
			return nil, fmt.Errorf("cannot apply migrations: %w", err)
		}
	}

	app := c.newCoreApplication()
	app.AddDestructor(c.done)
	return app, nil
}

// NewHTTPHandler applies the configured migrations and creates the HTTP handler of the application including all
// endpoints and the websocket, without starting a server. This is intended for embedding or tests, e.g. using
// httptest. The returned application must be released using [core.Application.Destroy], which also releases
// this Configurator.
func (c *Configurator) NewHTTPHandler() (http.Handler, *core.Application, error) {
	if mg := c.migrations; mg != nil {
		if err := mg.Apply(c.Context()); err != nil {
			return nil, nil, fmt.Errorf("cannot apply migrations: %w", err)
		}
	}

	handler := c.newHandler()
	app := c.app.Load()
	app.AddDestructor(c.done)
	return handler, app, nil
}

// newCoreApplication creates the transport independent core application and assigns it to c.app.
func (c *Configurator) newCoreApplication() *core.Application {
	if _, ok := c.factories["_"]; !ok {
		c.RootView("_", func(wnd core.Window) core.View {
			return c.DecorateRootView(func(wnd core.Window) core.View {
				return alert.NotFound()
			})(wnd)
		})
	}

	interceptors := slices.Clone(c.rootViewInterceptors)
	factories := map[proto.RootViewID]core.ComponentFactory{}
	for id, f := range c.factories {
		factories[id] = func(scope core.Window) core.View {
			for _, intercept := range interceptors {
				if view, ok := intercept(scope); ok {
					return view
				}
			}

			return f(scope)
		}
	}

	sessionMgmt, err := c.SessionManagement()
	if err != nil {
		panic(fmt.Errorf("session management is not optional anymore: %v", err))
	}

	tmpDir := filepath.Join(c.dataDir, "tmp")
	slog.Info("tmp directory updated", "dir", tmpDir)
	key, err := c.MasterKey()
	if err != nil {
		panic(fmt.Errorf("could not get master key: %v", err))
	}

	getAnonUser := std.Must(c.UserManagement()).UseCases.GetAnonUser // user management is also not optional anymore

	app2 := core.NewApplication(
		c.Context(),
		tmpDir,
		factories,
		c.onWindowCreatedObservers,
		c.fps,
		sessionMgmt.UseCases.FindUserSessionByID,
		key,
		c.eventBus,
		getAnonUser,
		sessionMgmt.UseCases.Logout,
	)
	app2.SetDebug(c.IsDebug())
	app2.AddDestructor(func() {
		if err := c.stores.Close(); err != nil {
			slog.Error("cannot close stores", "err", err.Error())
		}
	})

	app2.SetID(c.applicationID)
	for scheme, m := range c.colorSets {
		for _, set := range m {
			app2.UpdateColorSet(scheme, set)
		}
	}

	app2.SetName(c.applicationName)
	app2.SetVersion(c.applicationVersion)
	app2.SetAppIcon(core.URI(c.appIconUri))

	// Goroutines of the application, e.g. a scheduler, may already update the theme. Publishing the application,
	// reading the stored theme and applying it happen under the same lock as such an update, so an update is
	// either read here or applied afterwards, and never lost.
	themes := option.Must(c.ThemeManagement())
	c.themeMutex.Lock()
	c.app.Store(app2)
	colors := option.Must(themes.UseCases.ReadColors(user.SU()))
	app2.UpdateColorSet(core.Dark, colors.Dark)
	app2.UpdateColorSet(core.Light, colors.Light)

	fonts := option.Must(themes.UseCases.ReadFonts(user.SU()))
	app2.UpdateFonts(fonts)
	c.themeMutex.Unlock()

	// TODO we are in a weired order here
	for _, destructor := range c.destructors {
		app2.AddDestructor(destructor)
	}

	return app2
}

func (c *Configurator) newHandler() http.Handler {
	app2 := c.newCoreApplication()

	downloadStreams := map[string]func() (io.Reader, error){}
	var downloadFilesMutex sync.Mutex

	r := chi.NewRouter()

	app2.SetOnSendFiles(func(scope *core.Scope, options core.ExportFilesOptions) error {
		if len(options.Files) == 0 {
			return fmt.Errorf("no files to send")
		}

		name, mimetype := nameAndMime(options)

		scope.Publish(&proto.SendMultipleRequested{
			Resources: []proto.Resource{
				{
					Name:     proto.Str(name),
					URI:      proto.URI(fmt.Sprintf("/api/ora/v1/download?scope=%v&id=%v", scope.ID(), options.ID)),
					MimeType: proto.Str(mimetype),
				},
			},
		})

		return nil
	})

	app2.SetOnShareStream(func(scope *core.Scope, f func() (io.Reader, error)) (core.URI, error) {
		downloadFilesMutex.Lock()
		defer downloadFilesMutex.Unlock()

		token := string(proto.NewScopeID())

		scope.AddOnDestroyObserver(func() {
			downloadFilesMutex.Lock()
			defer downloadFilesMutex.Unlock()
			delete(downloadStreams, token)
		})

		uri := core.URI("/api/ora/v1/share?token=" + token)
		downloadStreams[token] = f
		return uri, nil
	})

	if c.IsDebug() {
		r.Use(
			cors.Handler(cors.Options{
				AllowedOrigins:   []string{"http://*"},
				AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
				AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
				ExposedHeaders:   []string{"Link"},
				AllowCredentials: true,
				MaxAge:           300, // Maximum value not ignored by any of major browsers

			}),
		)
		c.defaultLogger().Warn("using debug cors settings")
	}
	r.Use(
		c.loggerMiddleware,
	)

	r.Mount("/api/nago/v1/instance", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		type instanceResponse struct {
			ID string `json:"id"`
		}

		v := instanceResponse{ID: app2.Instance()}
		if err := json.NewEncoder(writer).Encode(v); err != nil {
			slog.Error("cannot encode instance response", "err", err.Error())
			return
		}

	}))

	// Serve the configured application icon under the static favicon paths referenced by index.html.
	faviconHandler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if c.appIconUri == "" {
			http.NotFound(writer, request)
			return
		}

		http.Redirect(writer, request, string(c.appIconUri), http.StatusFound)
	})
	r.Mount("/favicon.svg", faviconHandler)
	r.Mount("/favicon.ico", faviconHandler)

	if len(c.fsys) > 0 {
		c.defaultLogger().Info("serving fsys assets")
		assets := statigz.FileServer(mergefs.Merge(c.fsys...).(mergefs.MergedFS), statigz.EncodeOnInit)
		r.Mount("/", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			if _, err := request.Cookie(sessionCookieName); err != nil && mayIssueSessionCookie(request) {
				http.SetCookie(writer, c.newSessionCookie(request, string(proto.NewScopeID())))
			}

			if strings.HasPrefix(request.URL.Path, "/api/doc") {
				assets.ServeHTTP(writer, request)
				return
			}

			dir := filepath.Dir(request.URL.Path)
			/*if strings.HasPrefix(base,"index"){
				request.URL.Path = "/"
			}*/

			if dir != "" &&
				!(strings.HasPrefix(dir, "/modern") || strings.HasPrefix(dir, "/legacy")) {
				request.URL.Path = "/"
				assets.ServeHTTP(writer, request)
				return
			}

			assets.ServeHTTP(writer, request)
		}))

	}

	// The former endpoint /api/nago/v1/session/restore set the session cookie to any id encrypted with the master
	// key, which the frontend kept in its local storage during an http flow. Since the cookie is SameSite=Lax, it is
	// sent on the redirect back from an identity provider, so the workaround is gone. It undermined the http-only
	// cookie and accepted ids without expiry.

	r.Mount("/api/nago/v1/manifest.json", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		type icon struct {
			Src   string `json:"src"`
			Sizes string `json:"sizes"`
			Type  string `json:"type"`
		}

		type manifest struct {
			Name            string `json:"name"`
			ShortName       string `json:"short_name"`
			StartUrl        string `json:"start_url"`
			Display         string `json:"display"`
			BackgroundColor string `json:"background_color"`
			ThemeColor      string `json:"theme_color"`

			Icons []icon `json:"icons"`
		}

		buf, err := json.Marshal(manifest{
			Name:      c.applicationName,
			ShortName: c.applicationName,
			StartUrl:  "/",
			Display:   "standalone",
			Icons: []icon{
				{
					Src:   string(c.pwaIcon),
					Sizes: "512x512",
					Type:  "image/png",
				},
			},
		})

		if err != nil {
			slog.Error("failed to marshal manifest", "err", err.Error())
		}

		writer.Header().Add("Content-Type", "application/manifest+json")
		writer.Write(buf)
	}))

	r.Mount("/api/ora/v1/share", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		downloadFilesMutex.Lock()
		download, ok := downloadStreams[token]
		downloadFilesMutex.Unlock()

		if !ok {
			// TODO how to make DOS or id brute force attacks harder?
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		reader, err := download()
		if err != nil {
			slog.Error("cannot open shared stream", "token", token, "err", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		defer core.Release(reader)

		if mt, ok := reader.(core.ReaderWithMimeType); ok {
			w.Header().Set("Content-Type", mt.MimeType())
		} else {
			w.Header().Set("Content-Type", "application/octet-stream")
		}

		w.Header().Set("Cache-Control", "No-Store")
		if _, err := io.Copy(w, reader); err != nil {
			slog.Error("cannot write shared stream", "token", token, "err", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}

	}))

	r.Mount("/api/ora/v1/download", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scopeID := r.URL.Query().Get("scope")
		downloadID := r.URL.Query().Get("id")

		options, ok := app2.ExportFilesOptions(proto.ScopeID(scopeID), downloadID)
		if !ok {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		now := time.Now()
		name, mimetype := nameAndMime(options)
		multiple := len(options.Files) > 1
		if multiple {
			w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, name))
			w.Header().Set("Content-Type", "application/zip")
			zipWriter := zip.NewWriter(w)
			defer zipWriter.Close()

			for _, file := range options.Files {
				header := &zip.FileHeader{
					Name:     file.Name(),
					Method:   zip.Deflate,
					Modified: now,
				}
				f, err := zipWriter.CreateHeader(header)
				if err != nil {
					slog.Error("failed to open create zip file entry", "file", file.Name(), "err", err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				if _, err := file.Transfer(f); err != nil {
					slog.Error("failed to write file body into zip entry", "file", file.Name(), "err", err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}

			return
		}

		// single file case, push file
		if len(options.Files) > 0 {
			file := options.Files[0]
			w.Header().Set("Content-Type", mimetype)
			w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, name))

			if _, err := file.Transfer(w); err != nil {
				slog.Error("cannot write pull file", "file", file.Name, "err", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}

			return
		}

	}))

	r.Mount("/api/ora/v1/upload", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Info("received upload request")
		// we support currently only multipart upload forms
		scopeID := proto.ScopeID(r.Header.Get("x-scope"))
		if len(scopeID) < 32 {
			slog.Error("upload request has a weired x-scope id", "id", logging.Secret(string(scopeID)))
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		uploadId := r.Header.Get("x-receiver")
		if uploadId == "" {
			slog.Error("upload request has no parseable x-receiver header")
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		isMultipart := strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data")
		if isMultipart {
			if err := r.ParseMultipartForm(1024 * 1024); err != nil {
				slog.Error("cannot parse multipart form", "err", err)
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}

			var files []core.File
			for _, headers := range r.MultipartForm.File {
				// we don't care about specific field names and instead just collect everything what looks like a file
				for _, header := range headers {
					files = append(files, core.NewMultipartFile(header))
				}
			}

			importer, ok := app2.ImportFilesOptions(scopeID, uploadId)
			if !ok {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}

			importer.OnCompletion(files)

			slog.Info("multipart upload complete")
			return
		} else {
			slog.Error("upload request must be multipart form", "content-type", r.Header.Get("Content-Type"))
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

	}))

	images := option.Must(c.ImageManagement())
	r.Mount(httpimage.Endpoint, httpimage.NewHandler(images.UseCases.LoadBestFit))

	if c.sitemap != nil {
		r.Mount("/sitemap.xml", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/xml; charset=utf-8")

			buf, err := xml.MarshalIndent(c.sitemap, "", "  ")
			if err != nil {
				slog.Error("failed to marshal sitemap", "err", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			if _, err := w.Write([]byte(xml.Header)); err != nil {
				slog.Error("failed to write sitemap xml header", "err", err)
				return
			}

			if _, err := w.Write(buf); err != nil {
				slog.Error("failed to write sitemap body", "err", err)
			}
		}))
	}

	r.Mount("/wire", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if c.contextPath.Load() == nil {
			c.contextPath.Store(&r.Host)
		}

		logger := logging.FromContext(r.Context())
		//logger.Info("wire is called, before upgrade")
		queryParams := r.URL.Query()
		scopeID := queryParams.Get("_sid")
		_ = logger
		var upgrader = websocket.Upgrader{
			CheckOrigin:       c.checkWireOrigin,
			EnableCompression: true,
		} // use default options
		// The page normally got its cookie already. If not, e.g. because all assets came from the browser cache after
		// the page has been reached by a link from another site, the handshake issues it: the wire is always called
		// by the page itself, so the cookie cannot be missing just due to SameSite.
		cookie, _ := r.Cookie(sessionCookieName)
		var responseHeader http.Header
		if cookie == nil {
			cookie = c.newSessionCookie(r, string(proto.NewScopeID()))
			responseHeader = http.Header{"Set-Cookie": {cookie.String()}}
		}

		conn, err := upgrader.Upgrade(w, r, responseHeader)
		if err != nil {
			log.Print("upgrade:", err)
			slog.Info("http websocket upgrade failed", "err", err, "id", logging.Secret(string(scopeID)))
			return
		}
		defer conn.Close()

		conn.EnableWriteCompression(true)

		//logger.Info("wire upgrade to websocket success", "id", scopeID)

		// todo new
		defer func() {
			if r := recover(); r != nil {
				fmt.Println(r)
				debug.PrintStack()
			}
		}()
		channel := gorilla.NewWebsocketChannel(conn)
		scope := app2.Connect(channel, proto.ScopeID(scopeID))
		//defer scope.Destroy() we don't want that, the client cannot recover through a new channel otherwise

		// the session comes from the http-only cookie, which the client cannot override over the wire
		scope.AssignSession(session.ID(cookie.Value))

		if err := channel.Loop(); err != nil {
			//slog.Error("websocket channel loop failed", slog.Any("err", err), "id", scopeID)
			scope.Disconnect(channel) // we cannot use that anymore, so clean it up, unless the client already reconnected
			return
		}

	}))

	for _, endpoint := range c.rawEndpoint {
		if endpoint.method == "" {
			r.Mount(endpoint.pattern, endpoint.handler)
		} else {
			r.Method(endpoint.method, endpoint.pattern, endpoint.handler)
		}
	}

	for _, route := range r.Routes() {
		slog.Info("routes", "route", route.Pattern)
	}

	return r
}

// CtxRootViewPurpose is the context value under which a lookup of registered [RootViewMeta] is published.
//
// It exists so that packages which must not import this one - the AI assistant UI, for instance - can still
// ask what a route is for, using core.FromContext[RootViewPurposeLookup](wnd.Context(), CtxRootViewPurpose).
const CtxRootViewPurpose = "nago.rootview.purpose"

// RootViewPurposeLookup answers what the given route is for. The second result is false for routes that were
// registered without a [Purpose], which is the normal case.
type RootViewPurposeLookup func(path core.NavigationPath) (string, bool)

// publishRootViewPurposes makes the registered purposes readable through the window context.
func (c *Configurator) publishRootViewPurposes() {
	c.AddContextValue(core.ContextValue(CtxRootViewPurpose, RootViewPurposeLookup(func(path core.NavigationPath) (string, bool) {
		meta := c.RootViewMetaOf(path)
		return meta.Purpose, meta.Purpose != ""
	})))
}
