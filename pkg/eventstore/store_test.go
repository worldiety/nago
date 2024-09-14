package eventstore

import (
	"testing"
	"time"
)

func TestNewID(t *testing.T) {
	for range 5000 {
		now := time.Now()
		id := timeIntoID(now)
		ti, err := id.Time(time.Local)
		if err != nil {
			t.Fatal(id, ti, err)
		}

		if ti.UnixMilli() != now.UnixMilli() {
			t.Fatal(id, ti, ti.UnixMilli())
		}
	}
}
