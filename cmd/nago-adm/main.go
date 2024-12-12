package main

import (
	"flag"
	"fmt"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	"log/slog"
	"os"
	"time"
)

func main() {
	appId := flag.String("app", "", "application id which shall be modified")
	cmd := flag.String("cmd", "", "command which shall be executed, one of [admin-reset]")
	pwd := flag.String("pwd", "", "password which shall be used for admin-reset")

	flag.Parse()

	switch *cmd {
	case "admin-reset":
		adminReset(*appId, *pwd)
	default:
		fmt.Printf("Unknown command: %s\n", *cmd)
		os.Exit(1)
	}

}

// adminReset is invoked as follows:
//
//	nago-adm -app=de.worldiety.tutorial -cmd=admin-reset -pwd=<my super secret>
func adminReset(appId string, pwd string) {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID(core.ApplicationID(appId))
		users := std.Must(cfg.UserManagement())
		uid := std.Must(users.UseCases.EnableBootstrapAdmin(time.Now().Add(time.Hour), user.Password(pwd)))
		slog.Info("password for admin account has been updated", "uid", uid, "login", "admin@localhost")
	})
}
