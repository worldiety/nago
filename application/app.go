// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application

import (
	"fmt"
	"log/slog"
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

func (a *Application) Run() {

	defer func() {
		a.cfg.done()
	}()

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
