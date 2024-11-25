package application

import (
	"fmt"
	"log/slog"
	"runtime/debug"

	"github.com/joho/godotenv"
	"go.wdy.de/nago/internal/server"
)

type Application struct {
	cfg    *Configurator
	failed bool
}

func Configure(f func(cfg *Configurator)) *Application {

	a := &Application{}
	a.failed = !a.init(f)

	return a
}

func (a *Application) init(configure func(cfg *Configurator)) (success bool) {
	// Load environment variables from .env.local file
	_ = godotenv.Load()

	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
			fmt.Println(string(debug.Stack()))
			a.cfg.defaultLogger().Error("init failed, going into maintenance", slog.Any("err", r), slog.String("stacktrace", string(debug.Stack())))
			success = false
		}
	}()

	a.cfg = NewConfigurator()
	configure(a.cfg)

	a.cfg.LoadConfigFromEnv()

	return true
}

func (a *Application) Run() {
	if a.failed {
		panic("fix me: start setup/maintainence mode, check log for details")
	}

	defer func() {
		a.cfg.done()

		if r := recover(); r != nil {
			fmt.Println(r)
			fmt.Println(string(debug.Stack()))
		}
	}()

	err := a.runServer()
	a.cfg.done()

	logger := a.cfg.defaultLogger()
	if err != nil {
		logger.Error("application error", "err", err)
	}

	if app := a.cfg.app; app != nil {
		app.Destroy()
	}

	logger.Info("successful shutdown")

}

func (a *Application) runServer() error {
	host := a.cfg.getHost()
	port := a.cfg.getPort()
	a.cfg.defaultLogger().Info("launching server", slog.String("host", host), slog.Int("port", port))
	httpSrv, err := server.NewServer(host, port)
	if err != nil {
		return fmt.Errorf("server.New: %w", err)
	}

	return httpSrv.ServeHTTPHandler(a.cfg.defaultLogger(), a.cfg.Context(), a.cfg.newHandler())
}
