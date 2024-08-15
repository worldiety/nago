package application

import (
	"archive/zip"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/gorilla/websocket"
	"github.com/laher/mergefs"
	"github.com/vearutop/statigz"
	"go.wdy.de/nago/logging"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/core/http/gorilla"
	"go.wdy.de/nago/presentation/ora"
	"io"
	"io/fs"
	"log"
	"log/slog"
	"mime"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"runtime/debug"
	"strings"
	"sync"
	"time"
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
	id := ora.ComponentFactoryId(viewRootID)
	if !id.Valid() {
		panic(fmt.Errorf("invalid component factory id: %v", id))
	}

	if _, ok := c.factories[id]; ok {
		panic(fmt.Errorf("another factory with id %v has already been registered", id))
	}

	c.factories[id] = factory
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

	factories := map[ora.ComponentFactoryId]core.ComponentFactory{}
	for id, f := range c.factories {
		factories[id] = func(scope core.Window) core.View {
			return f(scope)
		}
	}

	downloadStreams := map[string]func() (io.Reader, error){}
	var downloadFilesMutex sync.Mutex

	tmpDir := filepath.Join(c.dataDir, "tmp")
	slog.Info("tmp directory updated", "dir", tmpDir)
	app2 := core.NewApplication(c.ctx, tmpDir, factories, c.onWindowCreatedObservers, c.fps)
	c.app = app2
	app2.SetID(c.applicationID)
	for scheme, m := range c.colorSets {
		for _, set := range m {
			app2.AddColorSet(scheme, set)
		}

	}

	app2.SetVersion(c.applicationVersion)
	app2.SetAppIcon(c.appIconUri)

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

		scope.Publish(ora.SendMultipleRequested{
			Type: ora.SendMultipleRequestedT,
			Resources: []ora.Resource{
				{
					Name:     name,
					URI:      ora.URI(fmt.Sprintf("/api/ora/v1/download?scope=%v&id=%v", scope.ID(), options.ID)),
					MimeType: mimetype,
				},
			},
		})

		return nil
	})

	app2.SetOnShareStream(func(scope *core.Scope, f func() (io.Reader, error)) (core.URI, error) {
		downloadFilesMutex.Lock()
		defer downloadFilesMutex.Unlock()

		token := string(ora.NewScopeID())

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

	if len(c.fsys) > 0 {
		c.defaultLogger().Info("serving fsys assets")
		assets := statigz.FileServer(mergefs.Merge(c.fsys...).(mergefs.MergedFS), statigz.EncodeOnInit)
		r.Mount("/", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			cookie, err := request.Cookie("wdy-ora-access")
			if err != nil {
				// TODO move me to the wire, which is called from the JS
				cookie = &http.Cookie{}
				cookie.Name = "wdy-ora-access"
				cookie.Value = string(ora.NewScopeID())
				cookie.Expires = time.Now().Add(365 * 24 * time.Hour)
				cookie.Secure = false //TODO in release-mode this must be true
				cookie.HttpOnly = true
				cookie.SameSite = http.SameSiteStrictMode //TODO CSRF protection however, do we actually suffer for this problem due to random addresses? if not, Lax is probably enough? => discuss with Fred
				// TODO can we make it more secure to do something like ASLR? how does that work? Is entropy large enough?
				// TODO alternative: use UUID + tree deltas to mitigate larger ids and avoid CSRF attacks
				cookie.Path = "/"
				http.SetCookie(writer, cookie)
			}

			dir := filepath.Dir(request.URL.Path)
			/*if strings.HasPrefix(base,"index"){
				request.URL.Path = "/"
			}*/
			if dir != "" && dir != "/assets" {
				request.URL.Path = "/"
				assets.ServeHTTP(writer, request)
				return
			}

			log.Println(request.URL.Path)

			assets.ServeHTTP(writer, request)
		}))

	}

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

		options, ok := app2.ExportFilesOptions(ora.ScopeID(scopeID), downloadID)
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
		scopeID := ora.ScopeID(r.Header.Get("x-scope"))
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
					files = append(files, mulitPartFileHeaderAdapter{header})
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

	r.Mount("/wire", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := logging.FromContext(r.Context())
		logger.Info("wire is called, before upgrade")
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

		logger.Info("wire upgrade to websocket success", "id", scopeID)

		// todo new
		defer func() {
			if r := recover(); r != nil {
				debug.PrintStack()
			}
		}()
		channel := gorilla.NewWebsocketChannel(conn)
		scope := app2.Connect(channel, ora.ScopeID(scopeID))
		_ = scope
		//defer scope.Destroy() we don't want that, the client cannot recover through a new channel otherwise

		cookie, _ := r.Cookie("wdy-ora-access")
		if err := channel.PublishLocal(ora.Marshal(ora.SessionAssigned{
			Type:      ora.SessionAssignedT,
			SessionID: cookie.Value,
		})); err != nil {
			slog.Error("cannot publish session assigned to local channel", slog.Any("err", err))
			return
		}

		if err := channel.Loop(); err != nil {
			slog.Error("websocket channel loop failed", slog.Any("err", err), "id", scopeID)
			scope.Connect(nil) // we cannot use that anymore, so clean it up
			return
		}

	}))

	for _, endpoint := range c.rawEndpoint {
		r.Mount(endpoint.pattern, endpoint.handler)
	}

	for _, route := range r.Routes() {
		slog.Info("routes", "route", route.Pattern)
	}

	return r
}

type mulitPartFileHeaderAdapter struct {
	header *multipart.FileHeader
}

func (m mulitPartFileHeaderAdapter) Transfer(dst io.Writer) (int64, error) {
	reader, err := m.header.Open()
	if err != nil {
		return 0, err
	}
	defer reader.Close()

	return io.Copy(dst, reader)
}

func (m mulitPartFileHeaderAdapter) MimeType() (string, bool) {
	return m.header.Header.Get("Content-Type"), true
}

func (m mulitPartFileHeaderAdapter) Size() (int64, bool) {
	return m.header.Size, true
}

func (m mulitPartFileHeaderAdapter) Open() (io.ReadCloser, error) {
	return m.header.Open()
}

func (m mulitPartFileHeaderAdapter) Name() string {
	return m.header.Filename
}
