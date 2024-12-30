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
	dir := flag.String("data-dir", "", "the path to the nago data dir. Only needed when it is not at ~/.nago/.")
	appId := flag.String("app", "", "application id which shall be modified")
	cmd := flag.String("cmd", "", "command which shall be executed, one of [admin-reset]")
	pwd := flag.String("pwd", "", "password which shall be used for admin-reset")
	lifetime := flag.Duration("lifetime", time.Hour, "how long is the lifetime, e.g. for the admin user. 0 means disabled lifetime. Default is 1 hour.")

	flag.Parse()

	switch *cmd {
	case "admin-reset":
		adminReset(*dir, *appId, *pwd, *lifetime)
	default:
		fmt.Printf("Unknown command: %s\n", *cmd)
		os.Exit(1)
	}

}

// adminReset is invoked as follows:
//
//	nago-adm -app=de.worldiety.tutorial -cmd=admin-reset -lifetime=0m -pwd=<my super secret>
func adminReset(dir string, appId string, pwd string, lifetime time.Duration) {
	application.Configure(func(cfg *application.Configurator) {
		if dir != "" {
			cfg.SetDataDir(dir)
		}
		cfg.SetApplicationID(core.ApplicationID(appId))
		users := std.Must(cfg.UserManagement())
		if lifetime == 0 {
			lifetime = time.Hour * 24 * 365 * 30 // 30 years = infinite
		}
		uid := std.Must(users.UseCases.EnableBootstrapAdmin(time.Now().Add(lifetime), user.Password(pwd)))
		slog.Info("password for admin account has been updated", "uid", uid, "login", "admin@localhost", "lifetime", lifetime)
	})
}
