package iter

import (
	"testing"
)

func TestReduce(t *testing.T) {
	tmp := []string{"a", "b", "c"}
	Reduce[string](Values(tmp))
}
