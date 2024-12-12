package permission_test

import (
	"go.wdy.de/nago/application/permission"
	"testing"
)

func TestRegister(t *testing.T) {
	permission.Register[MakeStuff](permission.Permission{ID: "de.worldiety.test"})
	//permission.Make[MakeStuff]("de.worldiety.test")
}

type MakeStuff func()
