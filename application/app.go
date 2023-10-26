package application

import (
	"fmt"
	"go.wdy.de/nago/container/slice"
	"go.wdy.de/nago/internal/server"
	"go.wdy.de/nago/presentation/rest"
	"go.wdy.de/nago/presentation/ui"
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
	httpSrv, err := server.NewServer("0.0.0.0", 3000)
	if err != nil {
		return fmt.Errorf("server.New: %w", err)
	}

	return httpSrv.ServeHTTPHandler(a.cfg.defaultLogger(), a.cfg.Context(), a.cfg.newHandler())
}

type PresentationLayer struct {
	Pages  slice.Slice[ui.Route]
	Events slice.Slice[ui.Event]
	Http   slice.Slice[rest.Route]
}
