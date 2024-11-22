package license

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

// FromEnv inspects environment variables as shown in the following examples:
//   - NAGO_LICENSES_APP=de.worldiety.example.license.app.jira:de.worldiety.example.license.app.sharepoint
//   - NAGO_LICENSES_USER=de.worldiety.example.license.user.chat=10:de.worldiety.example.license.user.chat=10
//
// It returns a slice of all declared Licenses in the given order. Known license types from the declaration
// are compared to the environment variables and enabled if defined, otherwise not enabled. Unknown types
// are just passed through.
// Invalid IDs are ignored and logged.
func FromEnv(declared ...License) []License {
	appLics := map[ID]bool{}
	userLics := map[ID]int{}
	if str := os.Getenv("NAGO_LICENSES_APP"); str != "" {
		for _, sid := range strings.Split(str, ":") {
			id := ID(strings.TrimSpace(sid))
			if !id.Valid() {
				slog.Error("invalid license identifier", "id", id)
				continue
			}

			if _, ok := appLics[id]; ok {
				slog.Error("duplicate license identifier", "id", id)
			}

			appLics[id] = true
		}
	}

	if str := os.Getenv("NAGO_LICENSES_USER"); str != "" {
		for _, idOrValue := range strings.Split(str, ":") {
			keyValue := strings.Split(idOrValue, "=")
			if len(keyValue) > 2 {
				slog.Error("invalid license user identifier", "id", idOrValue)
				continue
			}

			id := ID(keyValue[0])
			if !id.Valid() {
				slog.Error("invalid license user identifier", "id", id)
				continue
			}

			value := 0
			if len(keyValue) == 2 {
				v, err := strconv.Atoi(keyValue[1])
				if err != nil {
					slog.Error("invalid license user amount", "id", id)
					continue
				}

				value = v
			}

			if _, ok := userLics[id]; ok {
				slog.Error("duplicate license identifier", "id", id)
			}

			userLics[id] = value
		}
	}

	for i := range declared {
		if !declared[i].Identity().Valid() {
			// this is a programming error
			panic(fmt.Sprintf("invalid license identifier: %s", declared[i].Identity()))
		}

		switch lic := declared[i].(type) {
		case UserLicense:
			_, enabled := userLics[declared[i].Identity()]
			lic.IsEnabled = enabled
			declared[i] = lic
		case AppLicense:
			_, enabled := appLics[declared[i].Identity()]
			lic.IsEnabled = enabled
			declared[i] = lic
		}
	}

	return declared
}
