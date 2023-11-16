package application

import (
	"fmt"
	"github.com/joho/godotenv"
	"go.wdy.de/nago/internal/server"
	"log/slog"
	"runtime/debug"
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
	// Load environment variables from .env file
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

	return true
}

func (a *Application) Run() {
	if a.failed {
		panic("fix me: start setup/maintainence mode, check log for details")
	}

	defer func() {
		a.cfg.done()

		if r := recover(); r != nil {
			a.cfg.defaultLogger().Error("application panic", slog.Any("err", fmt.Errorf("panic: %v", r)))
			debug.PrintStack()
		}
	}()

	err := a.runServer()
	a.cfg.done()

	logger := a.cfg.defaultLogger()
	if err != nil {
		logger.Error("application error", err)
	}

	logger.Info("successful shutdown")

}

func (a *Application) runServer() error {
	host := "0.0.0.0"
	port := 3000
	a.cfg.defaultLogger().Info("launching server", slog.String("host", host), slog.Int("port", port))
	httpSrv, err := server.NewServer("0.0.0.0", 3000)
	if err != nil {
		return fmt.Errorf("server.New: %w", err)
	}

	return httpSrv.ServeHTTPHandler(a.cfg.defaultLogger(), a.cfg.Context(), a.cfg.newHandler())
}
