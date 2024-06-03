package concurrent

import "testing"

func TestCompareAndSwap(t *testing.T) {
	var destroyed Value[bool]
	if !CompareAndSwap(&destroyed, false, true) {
		t.Fatal("unreachable")
	}

	if CompareAndSwap(&destroyed, false, true) {
		t.Fatal("unreachable")
	}
}
