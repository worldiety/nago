// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"go.wdy.de/nago/application/adm"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/internal/server"
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
	a.cfg.LoadConfigFromEnv()
	configure(a.cfg)
}

func (a *Application) Stop() {
	a.cfg.done()
}

// Run runs the application until it is stopped. If the application cannot start, because a migration fails or the
// server cannot open its port (e.g. it is taken), Run logs the error, cleans up and ends the process with exit code 1,
// so that the operation (systemd, Kubernetes, a hub) sees the failure instead of a regular end. Every other end,
// including one by an error while serving or shutting down, returns as before.
func (a *Application) Run() {
	if err := a.run(); errors.Is(err, errStart) {
		os.Exit(1)
	}
}

// errStart marks an error, with which the application could not start.
var errStart = errors.New("application could not start")

// run is Run without the exit code; an error, with which the application could not start, wraps errStart.
func (a *Application) run() error {

	defer func() {
		a.cfg.done()
	}()

	logger := a.cfg.defaultLogger()

	if mg := a.cfg.migrations; mg != nil {
		if err := mg.Apply(a.cfg.Context()); err != nil {
			logger.Error("failed to apply migrations, exiting", "err", err.Error())
			return fmt.Errorf("%w: migrations: %w", errStart, err)
		}
	}

	// apply adm commands
	admDir := filepath.Join(a.cfg.DataDir(), "adm/once-after-cfg")
	slog.Info("checking adm once instructions", "dir", admDir)
	cmds := adm.ReadCommands(admDir, adm.ReadCommandsOptions{DeleteAfterRead: true})
	slog.Info("read adm once instructions complete", "found", len(cmds))
	for _, cmd := range cmds {
		switch cmd := cmd.(type) {
		case adm.EnableBootstrapAdmin:
			if users := a.cfg.userManagement; users != nil {
				if _, err := users.UseCases.EnableBootstrapAdmin(cmd.AliveUntil, user.Password(cmd.Password)); err != nil {
					slog.Error("failed to enable bootstrap admin", "err", err.Error())
				} else {
					slog.Warn("enabled bootstrap admin by cmd", "alive", cmd.AliveUntil)
				}
			}
		}
	}

	var runErr error
	if !a.cfg.oneShot {
		runErr = a.runServer()
		if runErr != nil {
			logger.Error("application error", "err", runErr)
		}
	} else {
		slog.Info("one shot mode enabled, exiting")
	}

	a.cfg.done()

	if app := a.cfg.app.Load(); app != nil {
		app.Destroy()
	}

	if errors.Is(runErr, errStart) {
		logger.Error("shutdown, the application could not start", "err", runErr)
		return runErr
	}

	logger.Info("successful shutdown")
	return runErr
}

func (a *Application) runServer() error {
	host := a.cfg.getHost()
	port := a.cfg.getPort()
	a.cfg.defaultLogger().Info("launching server", slog.String("host", host), slog.Int("port", port))
	httpSrv, err := server.NewServer(host, port)
	if err != nil {
		return fmt.Errorf("%w: server.New: %w", errStart, err)
	}

	return httpSrv.ServeHTTPHandler(a.cfg.defaultLogger(), a.cfg.Context(), a.cfg.newHandler())
}
