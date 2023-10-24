package application

import (
	"fmt"
	"go.wdy.de/nago/container/slice"
	"go.wdy.de/nago/internal/server"
	"go.wdy.de/nago/presentation/rest"
	"go.wdy.de/nago/presentation/ui"
)

type Application struct {
	cfg *Configurator
}

func Configure(f func(cfg *Configurator)) *Application {
	a := &Application{}
	a.init(f)

	return a
}

func (a *Application) init(configure func(cfg *Configurator)) bool {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("init failed, going into maintenance", r)
		}
	}()

	a.cfg = NewConfigurator()
	configure(a.cfg)

	return true
}

func (a *Application) Run() {

	defer func() {
		a.cfg.done()

		if r := recover(); r != nil {
			a.cfg.defaultLogger().Error("application panic", fmt.Errorf("panic: %v", r))
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
	httpSrv, err := server.NewServer("0.0.0.0", 8080)
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
