package application

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/gorilla/websocket"
	"github.com/laher/mergefs"
	"github.com/vearutop/statigz"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/logging"
	"go.wdy.de/nago/presentation/ui"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"path/filepath"
	"regexp"
	"sync"
)

var validPageIdRegex = regexp.MustCompile(`[a-z0-9_\-{/}]+`)

func (c *Configurator) Serve(fsys fs.FS) *Configurator {
	c.fsys = append(c.fsys, fsys)
	return c
}

func (c *Configurator) Index(target string) *Configurator {
	c.uiApp.IndexTarget = target
	return c
}

func (c *Configurator) newHandler() http.Handler {

	appSrv := newApplicationServer()
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
		logger := logging.FromContext(r.Context())

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
			logger.Warn("client requested unknown page", slog.String("_pid", pageID))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		wire := newConnWrapper(conn, r, c.auth)
		_, helloBuf, err := wire.ReadMessage()
		if err != nil {
			logger.Error("failed to read clients hello message", slog.Any("err", err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		tx := txMsg{}
		if err := json.Unmarshal(helloBuf, &tx); err != nil {
			logger.Error("failed to parse client tx hello message", slog.Any("err", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if len(tx.TX) == 0 {
			logger.Error("hello tx is empty")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var cHello clientHello
		if err := json.Unmarshal(tx.TX[0], &cHello); err != nil {
			logger.Error("failed to parse client hello message", slog.Any("err", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if cHello.Type != "hello" {
			logger.Error("invalid client hello message", slog.Any("hello", string(helloBuf)))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if cHello.Auth.Keycloak != "" && c.auth.keycloak != nil {
			wire.updateJWT(updJWT{
				Token:    cHello.Auth.Keycloak,
				OIDCName: OIDC_KEYCLOAK,
			})
		}

		livePage := livePageFn(wire)
		logger.Info(fmt.Sprintf("spawned live page %v", livePage.Token()))
		appSrv.putPage(livePage)
		// TODO we better wait what the client actually wants, instead of blindly render something?
		// this allows e.g. that the client can send update message like user authentication details => hello request
		livePage.Invalidate()
		for {
			if err := livePage.HandleMessage(); err != nil {
				logging.FromContext(r.Context()).Error(fmt.Sprintf("livePage is dead now %v", livePage.Token()), slog.Any("err", err))
				appSrv.removePage(livePage.Token())
				break
			}

		}
	}))

	r.Mount("/api/v1/upload", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pageToken := r.Header.Get("x-page-token")
		page := appSrv.getPage(ui.PageInstanceToken(pageToken))
		if page == nil {
			logging.FromContext(r.Context()).Error("invalid page token for upload") //, slog.String("token", pageToken))
			w.WriteHeader(http.StatusNotFound)
			return
		}

		page.HandleHTTP(w, r)
	}))

	r.Mount("/api/v1/download", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pageToken := r.Header.Get("x-page-token")
		if pageToken == "" {
			pageToken = r.URL.Query().Get("page")
		}
		page := appSrv.getPage(ui.PageInstanceToken(pageToken))
		if page == nil {
			logging.FromContext(r.Context()).Error("invalid page token for upload", slog.String("token", pageToken))
			w.WriteHeader(http.StatusNotFound)
			return
		}

		page.HandleHTTP(w, r)
	}))

	for _, route := range r.Routes() {
		fmt.Println(route.Pattern)
	}

	return r
}

type connWrapper struct {
	conn          *websocket.Conn
	r             *http.Request
	authProviders authProviders
	ctx           context.Context
	user          auth.User
}

func newConnWrapper(conn *websocket.Conn, req *http.Request, providers authProviders) *connWrapper {
	return &connWrapper{
		conn:          conn,
		r:             req,
		authProviders: providers,
		ctx:           req.Context(),
	}
}

type txMsg struct {
	TX []json.RawMessage `json:"tx"`
}

func (c *connWrapper) ReadMessage() (messageType int, p []byte, err error) {
	type msg struct {
		Type string `json:"type"`
	}

	t, buf, err := c.conn.ReadMessage()

	if err != nil {
		return t, buf, err
	}

	tx := txMsg{}
	if err := json.Unmarshal(buf, &tx); err != nil {
		slog.Default().Error("cannot decode ws batch message", slog.Any("err", err))
		return 0, nil, err
	}

	for _, buf := range tx.TX {
		var m msg
		if err := json.Unmarshal(buf, &m); err != nil {
			slog.Default().Error("cannot decode ws message", slog.Any("err", err))
			return 0, nil, err
		}

		switch m.Type {
		case "updateJWT":
			var jwt updJWT
			if err := json.Unmarshal(buf, &jwt); err != nil {
				panic(fmt.Errorf("cannot happen: %w", err))
			}

			c.updateJWT(jwt)
		}
	}

	return t, buf, err
}

func (c *connWrapper) updateJWT(jwt updJWT) {
	if jwt.OIDCName != OIDC_KEYCLOAK {
		logging.FromContext(c.r.Context()).Error("cannot update jwt user: oidc name is not implemented", slog.String("name", jwt.OIDCName))
		return
	}

	user, err := validateToken(c.authProviders.keycloak, c.r.Context(), jwt.Token)
	if err != nil {
		c.user = nil
		logging.FromContext(c.r.Context()).Error("cannot validate token", slog.Any("err", err))
	} else {
		ctx := auth.WithContext(c.r.Context(), user)
		c.ctx = ctx
		c.user = user
		//TODO do we have data races here?
		//TODO we have an async logic update problem here: if a token expires or updates while the page is open, we are not notified
		logging.FromContext(c.r.Context()).Info("updated authenticated user credentials")
	}
}

func (c *connWrapper) WriteMessage(messageType int, data []byte) error {
	return c.conn.WriteMessage(messageType, data)
}

func (c *connWrapper) Context() context.Context {
	return c.ctx
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

func (c *connWrapper) User() auth.User {
	if c.user == nil {
		return invalidUser{}
	}

	return c.user
}

func (c *connWrapper) Remote() ui.Remote {
	return &remoteImpl{
		req: c.r,
	}
}

type applicationServer struct {
	activePages map[ui.PageInstanceToken]*ui.Page
	mutex       sync.RWMutex
}

func newApplicationServer() *applicationServer {
	return &applicationServer{
		activePages: make(map[ui.PageInstanceToken]*ui.Page),
	}
}

func (a *applicationServer) getPage(token ui.PageInstanceToken) *ui.Page {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	return a.activePages[token]
}

func (a *applicationServer) putPage(page *ui.Page) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.activePages[page.Token()] = page
}

func (a *applicationServer) removePage(token ui.PageInstanceToken) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	page := a.activePages[token]
	if page != nil {
		if err := page.Close(); err != nil {
			slog.Error("cannot close page", slog.Any("err", err))
		}
	}
	delete(a.activePages, token)

}

type updJWT struct {
	Token    string `json:"token,omitempty"`
	OIDCName string `json:"OIDCName"`
}

type remoteImpl struct {
	req *http.Request
}

func (r *remoteImpl) Addr() string {
	return r.req.RemoteAddr
}

func (r *remoteImpl) ForwardedFor() string {
	if s := r.req.Header.Get("X-Forwarded-For"); s != "" {
		return s
	}

	if s := r.req.Header.Get("X-Real-IP"); s != "" {
		return s
	}

	if s := r.req.Header.Get("CF-Connecting-IP"); s != "" {
		return s
	}

	return ""
}

// clientHello must be the first message from the client.
type clientHello struct {
	Type string `json:"type,omitempty"`
	Auth struct {
		Keycloak string `json:"keycloak,omitempty"`
	} `json:"auth" json:"auth"`
}
