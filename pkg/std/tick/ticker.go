package tick

import (
	"sync"
	"sync/atomic"
	"time"
)

type Granularity int

func (g Granularity) Duration() time.Duration {
	switch g {
	case Minute:
		return time.Minute
	default:
		return 0
	}
}

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

// EveryOnce uses a memoization based on the given granularity and returns an according getter, which only delegates
// to the given func, if the granularity has been exceeded between calls. The memoization itself is thread safe
// and lock-free. However, concurrent executions can be seen during execution of the given fn, but the tupel return
// is always consistent.
func EveryOnce[T any](granularity Granularity, fn func() (T, error)) func() (T, error) {
	type tupleResult struct {
		v T
		e error
	}
	var lastRequested atomic.Pointer[time.Time]
	var lastValue atomic.Pointer[tupleResult]
	return func() (T, error) {
		if t := lastRequested.Load(); t == nil || t.Before(Now(granularity).Add(granularity.Duration())) {
			v, err := fn()

			lastValue.Store(&tupleResult{v: v, e: err})
			now := Now(granularity)
			// update time after the value has been updated, thus we can see values and timestamps can
			// race logically against each other, however per definition, that must be fine.
			lastRequested.Store(&now)
		}

		res := lastValue.Load()
		if res == nil {
			// cannot happen, but not sure if I'm to clever regarding happens-before on atomic values
			var zero T
			return zero, nil
		}

		return res.v, res.e
	}
}
