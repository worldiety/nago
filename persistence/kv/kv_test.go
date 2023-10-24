package kv

import (
	"testing"
)

type Blub struct {
}

func (Blub) Identity() string {
	return ""
}

func TestCollection_Get(t *testing.T) {
	_ = Collection[string, Blub]{}
}
