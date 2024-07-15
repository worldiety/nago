package application

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/gorilla/websocket"
	"github.com/laher/mergefs"
	"github.com/vearutop/statigz"
	"go.wdy.de/nago/internal/incubator/tmpfiles"
	"go.wdy.de/nago/logging"
	"go.wdy.de/nago/pkg/iter"
	"go.wdy.de/nago/pkg/slices"
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
	"os"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Component registers a factory to create a [core.ViewRoot] within a [core.Scope].
// For example, a web browser will create at least a single ViewRoot for each open tab.
// Note, that leading or succeeding slashes in the factory ids are not allowed, otherwise you can
// use them in arbitrary ways.
// Keep in mind, that web browsers will expose these ids to the user and they become part of your public
// API or contract with the user. A user may bookmark them.
//
// You cannot use path variables. Instead, use [core.Window.Values] to transport a state from one ViewRoot
// (or window) to another.
func (c *Configurator) Component(id ora.ComponentFactoryId, factory func(wnd core.Window) core.View) {
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

type httpFileDownload struct {
	Token        string
	Name         string
	AbsolutePath string
	Mimetype     string
}

func (c *Configurator) newHandler() http.Handler {

	factories := map[ora.ComponentFactoryId]core.ComponentFactory{}
	for id, f := range c.factories {
		factories[id] = func(scope core.Window) core.View {
			return f(scope)
		}
	}

	downloadFiles := map[string]httpFileDownload{}
	downloadStreams := map[string]func() (io.Reader, error){}
	var downloadFilesMutex sync.Mutex

	tmpDir := filepath.Join(c.dataDir, "tmp")
	slog.Info("tmp directory updated", "dir", tmpDir)
	app2 := core.NewApplication(c.ctx, tmpDir, factories, c.onWindowCreatedObservers)
	c.app = app2
	app2.SetID(c.applicationID)
	app2.SetThemes(c.themes)
	app2.SetVersion(c.applicationVersion)

	// TODO we are in a weired order here
	for _, destructor := range c.destructors {
		app2.AddDestructor(destructor)
	}
	r := chi.NewRouter()
	app2.SetOnSendFiles(func(scope *core.Scope, it iter.Seq2[core.File, error]) error {
		var err error
		collectedFiles := slices.Collect(iter.BreakOnError(&err, it))
		if err != nil {
			return err
		}

		switch len(collectedFiles) {
		case 0:
			return fmt.Errorf("no files found in fsys: %v", it)
		case 1:
			// issue a direct link
			token := string(ora.NewScopeID())
			tmpFile := filepath.Join(c.Directory("download"), token)
			if err := copyFile(collectedFiles[0], tmpFile); err != nil {
				return fmt.Errorf("could not copy file %v: %v", tmpFile, err)
			}

			scope.AddOnDestroyObserver(func() {
				if err := os.Remove(tmpFile); err != nil {
					slog.Error("cannot remove download file", "file", tmpFile, "err", err)
				}

				downloadFilesMutex.Lock()
				defer downloadFilesMutex.Unlock()
				delete(downloadFiles, token)
			})

			download := httpFileDownload{
				Token:        token,
				Name:         collectedFiles[0].Name(),
				AbsolutePath: tmpFile,
			}

			mime := mime.TypeByExtension(filepath.Ext(download.Name))
			if mime == "" {
				mime = "application/octet-stream"
			}

			download.Mimetype = mime
			downloadFilesMutex.Lock()
			downloadFiles[token] = download
			downloadFilesMutex.Unlock()

			scope.Publish(ora.SendMultipleRequested{
				Type: ora.SendMultipleRequestedT,
				Resources: []ora.Resource{
					{
						Name:     download.Name,
						URI:      ora.URI("/api/ora/v1/download?token=" + token),
						MimeType: download.Mimetype,
					},
				},
			})

		default:
			// issue a zip file
			token := string(ora.NewScopeID())
			zipFile := filepath.Join(c.Directory("download"), token)
			err := makeZip(zipFile, it)
			if err != nil {
				return fmt.Errorf("cannot create zip file for multi download: %w", err)
			}

			scope.AddOnDestroyObserver(func() {
				if err := os.Remove(zipFile); err != nil {
					slog.Error("cannot remove zip file", "file", zipFile, "err", err)
				}
			})

			download := httpFileDownload{
				Token:        token,
				Name:         "files.zip",
				AbsolutePath: zipFile,
				Mimetype:     "application/zip",
			}

			downloadFilesMutex.Lock()
			downloadFiles[token] = download
			downloadFilesMutex.Unlock()

			scope.Publish(ora.SendMultipleRequested{
				Type: ora.SendMultipleRequestedT,
				Resources: []ora.Resource{
					{
						Name:     download.Name,
						URI:      ora.URI("/api/ora/v1/download?token=" + token),
						MimeType: download.Mimetype,
					},
				},
			})
		}

		scope.Tick()

		return nil
	})

	app2.SetOnShareStream(func(scope *core.Scope, f func() (io.Reader, error)) (ora.URI, error) {
		downloadFilesMutex.Lock()
		defer downloadFilesMutex.Unlock()

		token := string(ora.NewScopeID())

		scope.AddOnDestroyObserver(func() {
			downloadFilesMutex.Lock()
			defer downloadFilesMutex.Unlock()
			delete(downloadStreams, token)
		})

		uri := ora.URI("/api/ora/v1/share?token=" + token)
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
		token := r.URL.Query().Get("token")
		downloadFilesMutex.Lock()
		download, ok := downloadFiles[token]
		downloadFilesMutex.Unlock()

		if !ok {
			// TODO how to make DOS or id brute force attacks harder?
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		file, err := os.Open(download.AbsolutePath)
		if err != nil {
			slog.Error("cannot open file for download", "file", download.AbsolutePath, "err", err)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		defer file.Close()

		w.Header().Set("Content-Type", download.Mimetype)
		w.Header().Set("Content-Disposition", "attachment; filename=\""+download.Name+"\"")
		if _, err := io.Copy(w, file); err != nil {
			slog.Error("cannot write download file", "file", download.AbsolutePath, "err", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
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

		receiverPtr, err := strconv.Atoi(r.Header.Get("x-receiver"))
		if err != nil {
			slog.Error("upload request has no parseable x-receiver header", "err", err)
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

			var tmp [32]byte
			if _, err := rand.Read(tmp[:]); err != nil {
				panic(err)
			}
			uplTmpDir := c.Directory(filepath.Join("upload", hex.EncodeToString(tmp[:])))

			fsys, err := tmpfiles.New(uplTmpDir)
			if err != nil {
				slog.Error("cannot create tmpfs filesystem", "err", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			for _, headers := range r.MultipartForm.File {
				// we don't care about specific field names and instead just collect everything what looks like a file
				for _, header := range headers {
					file, err := header.Open()
					if err != nil {
						defer fsys.Close()
						slog.Error("cannot open multipart form file", "err", err)
						http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
						return
					}

					defer file.Close()

					if err := fsys.Import(header.Filename, file); err != nil {
						defer fsys.Close()
						slog.Error("cannot import multipart form file", "err", err)
						http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
						return
					}

					slog.Info("imported multipart form file", "file", header.Filename)
				}
			}

			if err := app2.OnFilesReceived(scopeID, ora.Ptr(receiverPtr), fsys.Each); err != nil {
				defer fsys.Close()
				slog.Error("cannot process received stream", "err", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

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

	for _, route := range r.Routes() {
		slog.Info("routes", "route", route.Pattern)
	}

	for _, endpoint := range c.rawEndpoint {
		r.Mount(endpoint.pattern, endpoint.handler)
	}

	return r
}

type httpFileStream struct {
	file     multipart.File
	header   *multipart.FileHeader
	scopeID  ora.ScopeID
	receiver ora.Ptr
}

func newHttpFileStream(file multipart.File, header *multipart.FileHeader, scopeID ora.ScopeID, receiver ora.Ptr) *httpFileStream {
	return &httpFileStream{file: file, header: header, scopeID: scopeID, receiver: receiver}
}

func (h *httpFileStream) Read(p []byte) (n int, err error) {
	return h.file.Read(p)
}

func (h *httpFileStream) Name() string {
	return h.header.Filename
}

func (h *httpFileStream) Receiver() ora.Ptr {
	return h.receiver
}

func (h *httpFileStream) ScopeID() ora.ScopeID {
	return h.scopeID
}
