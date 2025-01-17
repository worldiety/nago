package tick

import (
	"sync"
	"time"
)

type Granularity int

const (
	Minute Granularity = iota
)

func init() {
	lastTickTime = time.Now()
	ticker := time.NewTicker(time.Minute)
	go func() {
		select {
		case t := <-ticker.C:
			tickMutex.Lock()
			lastTickTime = t
			tickMutex.Unlock()
		}
	}()
}

var tickMutex sync.Mutex
var lastTickTime time.Time

// Now returns the current ticker time, which has a very low granularity, because we expect a massive
// scaled load of calls here, just doing mostly nothing. E.g. checking 100 user properties on 1000 active
// users, would require alone 2 seconds (if kernel uses its slow fallback code).
//
// https://github.com/golang/go/issues/57749
func Now(granularity Granularity) time.Time {
	tickMutex.Lock()
	defer tickMutex.Unlock()

	switch granularity {
	default:
		return lastTickTime
	}
}
