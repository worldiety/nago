package application

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/gorilla/websocket"
	"github.com/laher/mergefs"
	"github.com/vearutop/statigz"
	"go.wdy.de/nago/logging"
	"go.wdy.de/nago/presentation/ui"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"path/filepath"
	"regexp"
)

var validPageIdRegex = regexp.MustCompile(`[a-z0-9_\-{/}]+`)

func (c *Configurator) Serve(fsys fs.FS) *Configurator {
	c.fsys = append(c.fsys, fsys)
	return c
}

func (c *Configurator) Page(p ui.Pager) *Configurator {
	if err := p.PageID().Validate(); err != nil {
		panic(err)
	}
	c.uiApp.Pages = c.uiApp.Pages.AppendAll(p)
	if err := p.Validate(); err != nil {
		panic(err)
	}
	return c
}

func (c *Configurator) Index(target ui.Target) *Configurator {
	c.uiApp.IndexTarget = target
	return c
}

func (c *Configurator) newHandler() http.Handler {

	r := chi.NewRouter()

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
		c.keycloakMiddleware,
	)

	c.uiApp.ConfigureRouter(r)

	if len(c.fsys) > 0 {
		c.defaultLogger().Info("serving fsys assets")
		assets := statigz.FileServer(mergefs.Merge(c.fsys...).(mergefs.MergedFS), statigz.EncodeOnInit)
		r.Mount("/", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
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

	r.Mount("/wire", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("wire?")
		var upgrader = websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true //TODO security implications?
			},
		} // use default options
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}
		defer conn.Close()

		queryParams := r.URL.Query()
		pageID := queryParams.Get("_pid")

		livePageFn := c.uiApp.LivePages[ui.PageID(pageID)]

		if livePageFn == nil {
			logging.FromContext(r.Context()).Warn("client requested unknown page", slog.String("_pid", pageID))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		livePage := livePageFn(&connWrapper{conn: conn, r: r})
		logging.FromContext(r.Context()).Info(fmt.Sprintf("spawned live page %p", livePage))
		livePage.Invalidate()
		for {
			if err := livePage.HandleMessage(); err != nil {
				livePage.Close()
				logging.FromContext(r.Context()).Error(fmt.Sprintf("livePage is dead now %p", livePage), slog.Any("err", err))
				break
			}
			/*conn.WriteMessage(1, []byte("kack from server"))
			log.Println("reading a meassge")
			mt, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				break
			}
			log.Printf("recv: %s", message)
			err = conn.WriteMessage(mt, message)
			if err != nil {
				log.Println("write:", err)
				break
			}*/

		}
	}))

	for _, route := range r.Routes() {
		fmt.Println(route.Pattern)
	}

	return r
}

type connWrapper struct {
	conn *websocket.Conn
	r    *http.Request
}

func (c *connWrapper) ReadMessage() (messageType int, p []byte, err error) {
	return c.conn.ReadMessage()
}

func (c *connWrapper) WriteMessage(messageType int, data []byte) error {
	return c.conn.WriteMessage(messageType, data)
}

func (c *connWrapper) Values() ui.Values {
	tmp := ui.Values{}
	for k, strings := range c.r.URL.Query() {
		v := ""
		if len(strings) > 0 {
			v = strings[0]
		}
		tmp[k] = v
	}
	return tmp
}
