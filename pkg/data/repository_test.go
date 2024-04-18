package data

import "testing"

func TestRandIdent(t *testing.T) {
	t.Log(RandIdent[string]())
}
