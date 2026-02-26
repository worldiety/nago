// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application

import (
	"archive/zip"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"log/slog"
	"mime"
	"net/http"
	"path/filepath"
	"runtime/debug"
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
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/logging"
	"go.wdy.de/nago/pkg/blob/crypto"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/core/http/gorilla"
	"go.wdy.de/nago/presentation/proto"
)

// RootView registers a factory to create a [core.View] within a [core.Scope].
// For example, a web browser will create at least a single ViewRoot for each open tab.
// Note, that leading or succeeding slashes in the factory ids are not allowed, otherwise you can
// use them in arbitrary ways.
// Keep in mind, that web browsers will expose these ids to the user and they become part of your public
// API or contract with the user. A user may bookmark them.
//
// You cannot use path variables. Instead, use [core.Values] to transport a state from one ViewRoot
// (or window) to another.
func (c *Configurator) RootView(viewRootID core.NavigationPath, factory func(wnd core.Window) core.View) {
	id := proto.RootViewID(viewRootID)
	if !id.Valid() {
		panic(fmt.Errorf("invalid component factory id: %v", id))
	}

	if _, ok := c.factories[id]; ok {
		panic(fmt.Errorf("another factory with id %v has already been registered", id))
	}

	c.factories[id] = factory
}

func (c *Configurator) RootViewWithDecoration(viewRootID core.NavigationPath, factory func(wnd core.Window) core.View) {
	c.RootView(viewRootID, c.DecorateRootView(factory))
}

func (c *Configurator) Serve(fsys fs.FS) *Configurator {
	c.fsys = append(c.fsys, fsys)
	return c
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

func (c *Configurator) newHandler() http.Handler {

	factories := map[proto.RootViewID]core.ComponentFactory{}
	for id, f := range c.factories {
		factories[id] = func(scope core.Window) core.View {
			return f(scope)
		}
	}

	downloadStreams := map[string]func() (io.Reader, error){}
	var downloadFilesMutex sync.Mutex

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
		c.ctx,
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
	app2.SetDebug(c.debug)

	app2.SetContext(core.WithContext(app2.Context(),
		core.ContextValue("", option.Must(c.ImageManagement()).UseCases.CreateSrcSet),
		core.ContextValue("", option.Must(c.ImageManagement()).UseCases.LoadBestFit),
		core.ContextValue("", option.Must(c.ImageManagement()).UseCases.LoadSrcSet),
	))

	app2.SetContext(core.WithContext(app2.Context(),
		c.systemServices...,
	))

	c.app = app2
	app2.SetID(c.applicationID)
	for scheme, m := range c.colorSets {
		for _, set := range m {
			app2.UpdateColorSet(scheme, set)
		}
	}

	app2.SetName(c.applicationName)
	app2.SetVersion(c.applicationVersion)
	app2.SetAppIcon(core.URI(c.appIconUri))
	colors := option.Must(option.Must(c.ThemeManagement()).UseCases.ReadColors(user.SU()))
	app2.UpdateColorSet(core.Dark, colors.Dark)
	app2.UpdateColorSet(core.Light, colors.Light)

	fonts := option.Must(option.Must(c.ThemeManagement()).UseCases.ReadFonts(user.SU()))
	app2.UpdateFonts(fonts)

	// TODO we are in a weired order here
	for _, destructor := range c.destructors {
		app2.AddDestructor(destructor)
	}
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

	if c.debug {
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

	if len(c.fsys) > 0 {
		c.defaultLogger().Info("serving fsys assets")
		assets := statigz.FileServer(mergefs.Merge(c.fsys...).(mergefs.MergedFS), statigz.EncodeOnInit)
		r.Mount("/", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			cookie, err := request.Cookie("wdy-ora-access")
			if err != nil {
				// TODO move me to the wire, which is called from the JS
				cookie = &http.Cookie{}
				cookie.Name = "wdy-ora-access"
				cookie.Value = string(proto.NewScopeID())
				cookie.Expires = time.Now().Add(365 * 24 * time.Hour)
				cookie.Secure = c.secureCookie()
				cookie.HttpOnly = true
				// Security note: we used http.SameSiteStrictMode but it is at least broken in firefox as of today, even with our local-storage-restore process (perhaps due to browser bugs)
				// lets try to decrease security and see what the security review thinks. We could also improve testing and just enable Lax for firefox, if other browser work properly with our workaround.
				cookie.SameSite = http.SameSiteLaxMode
				cookie.Path = "/"
				http.SetCookie(writer, cookie)
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

	masterKey, err := c.MasterKey()
	if err != nil {
		slog.Error("error getting master key: %v", "err", err)
	}

	r.Mount("/api/nago/v1/session/restore", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			writer.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		hexBuf, err := io.ReadAll(request.Body)
		if err != nil {
			slog.Error("error reading request body: %v", "err", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		buf, err := hex.DecodeString(string(hexBuf))
		if err != nil {
			slog.Error("error decoding request body: %v", "err", err)
			writer.WriteHeader(http.StatusInternalServerError)
		}

		sidBuf, err := crypto.Decrypt(buf, masterKey)
		if err != nil {
			slog.Error("error decrypting request body: %v", "err", err)
			writer.WriteHeader(http.StatusInternalServerError)
		}

		cookie := &http.Cookie{}
		cookie.Name = "wdy-ora-access"
		cookie.Value = string(sidBuf)
		cookie.Expires = time.Now().Add(365 * 24 * time.Hour)
		cookie.Secure = c.secureCookie()
		cookie.HttpOnly = true
		// Security note: as above, we use the lax mode due to browser bugs and edge cases because our clever restore mechanics just not works reliable enough
		cookie.SameSite = http.SameSiteLaxMode
		cookie.Path = "/"
		http.SetCookie(writer, cookie)
	}))

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
			slog.Error("upload request has a weired x-scope id", "id", scopeID)
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
			CheckOrigin: func(r *http.Request) bool {
				return true //TODO security implications?
			},
			EnableCompression: true,
		} // use default options
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			slog.Info("http websocket upgrade failed", "err", err, "id", scopeID)
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
		_ = scope
		//defer scope.Destroy() we don't want that, the client cannot recover through a new channel otherwise

		cookie, _ := r.Cookie("wdy-ora-access")
		if cookie != nil {
			buf := bytes.NewBuffer(make([]byte, 0, 512))
			dst := proto.NewBinaryWriter(buf)
			err := proto.Marshal(dst, &proto.SessionAssigned{
				SessionID: proto.Str(cookie.Value),
			})

			if err != nil {
				panic(fmt.Errorf("unreachable: %w", err))
			}

			if err := channel.PublishLocal(buf.Bytes()); err != nil {
				slog.Error("cannot publish session assigned to local channel", slog.Any("err", err))
				return
			}
		} else {
			// if debug is false (e.g. on a linux) secure will switch to on but without https we will never get that cookie
			slog.Error("cookie is missing in /wire, maybe missing due to browser security constraints and wrong settings. If you want insecure cookies set env NAGO_COOKIES_INSECURE=true")
		}

		if err := channel.Loop(); err != nil {
			slog.Error("websocket channel loop failed", slog.Any("err", err), "id", scopeID)
			scope.Connect(nil) // we cannot use that anymore, so clean it up
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
