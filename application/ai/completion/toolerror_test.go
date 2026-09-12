package completion

import (
	"errors"
	"strings"
	"testing"

	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/xerrors"
)

func TestToolErrorText(t *testing.T) {
	if got := toolErrorText(nil); got != "" {
		t.Errorf("nil = %q", got)
	}

	// unclassified errors keep their detail so the model can recover
	if got := toolErrorText(errors.New("row not locked")); got != "row not locked" {
		t.Errorf("unknown = %q", got)
	}

	denied := toolErrorText(user.PermissionDeniedErr)
	if !strings.Contains(denied, "authorization") && !strings.Contains(denied, "Berechtigung") {
		t.Errorf("denial must be labelled as such: %q", denied)
	}

	var errs xerrors.FieldBuilder
	errs.Add("Name", "must not be empty")
	errs.Add("Age", "too small")
	v := toolErrorText(errs.Error())

	for _, want := range []string{"Age: too small", "Name: must not be empty"} {
		if !strings.Contains(v, want) {
			t.Errorf("missing %q in %q", want, v)
		}
	}

	if strings.Contains(v, "map[") {
		t.Errorf("must not leak a Go map dump: %q", v)
	}
}
