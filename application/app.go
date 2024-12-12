package application

import (
	"fmt"
	"github.com/joho/godotenv"
	"go.wdy.de/nago/internal/server"
	"log/slog"
)

type Application struct {
	cfg *Configurator
}

func Configure(f func(cfg *Configurator)) *Application {

	a := &Application{}
	a.init(f)

	return a
}

func (a *Application) init(configure func(cfg *Configurator)) {
	// Load environment variables from .env.local file
	_ = godotenv.Load()

	a.cfg = NewConfigurator()
	configure(a.cfg)

	a.cfg.LoadConfigFromEnv()

}

func (a *Application) Run() {

	defer func() {
		a.cfg.done()
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
