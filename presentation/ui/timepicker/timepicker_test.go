package timepicker

import (
	"testing"
	"time"
)

func TestMinutes(t *testing.T) {
	d := time.Millisecond*12 + time.Second*42 + time.Minute*63

	if m := Minutes(d); m != 3 {
		t.Fatal(m)
	}

	if m := Hours(d); m != 1 {
		t.Fatal(m)
	}
}
