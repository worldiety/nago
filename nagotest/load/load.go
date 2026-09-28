// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package load puts a running nago server under load. Virtual users connect through websockets like browsers
// and run scenarios, which are written against the same [nagotest.Window] API as in-process tests. The
// latency of each action is measured until the window settled, i.e. until the backend answered.
//
// Scenarios are either written in Go (see [Run]) or declared in a JSON step file (see [LoadFile]), which is
// executed by the nago-loadtest command.
package load

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"runtime"
	"sync"
	"time"

	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
)

// ErrThresholds is returned by [Run] if the report violates the configured thresholds.
var ErrThresholds = errors.New("load: thresholds violated")

// Config defines the load profile.
type Config struct {
	// URL of the nago server, e.g. http://localhost:3000.
	URL string
	// Users is the amount of concurrent virtual users.
	Users int
	// RampUp is the duration in which the users are started evenly.
	RampUp time.Duration
	// Duration limits the entire run including the ramp-up. Users complete their current iteration.
	// If zero, each user runs the configured Iterations.
	Duration time.Duration
	// Iterations limits the scenario runs per user. If zero, users iterate until the Duration elapsed.
	Iterations int
	// ThinkTime is the pause of a user between two iterations.
	ThinkTime time.Duration
	// Seed makes the random choices of the users reproducible.
	Seed uint64
	// Data is assigned round-robin to the users, e.g. their login credentials. See [User.Data].
	Data []map[string]string
	// Thresholds which cause [ErrThresholds], if violated.
	Thresholds Thresholds
	// Options for each window. Windows are always lean, session bound and observed.
	Options []nagotest.OpenOption
}

// Thresholds define the acceptance criteria of a run. Zero values are not checked.
type Thresholds struct {
	// MaxErrorRate is the maximum ratio of failed iterations, e.g. 0.01 for 1%.
	MaxErrorRate float64
	// P95 is the maximum 95th percentile latency of each action.
	P95 time.Duration
}

// A Scenario is a user journey starting at a route.
type Scenario struct {
	Name string
	// Weight is the relative probability to pick this scenario per iteration. Defaults to 1.
	Weight int
	Path   core.NavigationPath
	Values core.Values
	// Run executes the journey. Failed assertions abort the iteration.
	Run func(u *User, w *nagotest.Window)
}

// Run puts the server under load and returns the report. It returns [ErrThresholds] together with the report,
// if the thresholds are violated. Cancelling ctx stops starting new iterations.
func Run(ctx context.Context, cfg Config, scenarios ...Scenario) (*Report, error) {
	if cfg.URL == "" {
		return nil, errors.New("load: url is required")
	}

	if cfg.Users <= 0 {
		return nil, errors.New("load: at least one user is required")
	}

	if cfg.Duration <= 0 && cfg.Iterations <= 0 {
		return nil, errors.New("load: either a duration or iterations are required")
	}

	if len(scenarios) == 0 {
		return nil, errors.New("load: at least one scenario is required")
	}

	for i := range scenarios {
		if scenarios[i].Weight <= 0 {
			scenarios[i].Weight = 1
		}

		if scenarios[i].Run == nil {
			return nil, fmt.Errorf("load: scenario %q has no run function", scenarios[i].Name)
		}
	}

	if cfg.Duration > 0 {
		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, cfg.Duration)
		defer cancel()
	}

	rec := newRecorder()
	var wg sync.WaitGroup
	for id := range cfg.Users {
		var delay time.Duration
		if cfg.RampUp > 0 {
			delay = cfg.RampUp * time.Duration(id) / time.Duration(cfg.Users)
		}

		u := &User{
			ID:        id,
			Rand:      rand.New(rand.NewPCG(cfg.Seed, uint64(id))),
			ctx:       ctx,
			sessionID: string(proto.NewScopeID()),
		}

		if len(cfg.Data) > 0 {
			u.Data = cfg.Data[id%len(cfg.Data)]
		}

		wg.Go(func() {
			if !sleep(ctx, delay) {
				return
			}

			for i := 0; cfg.Iterations <= 0 || i < cfg.Iterations; i++ {
				if ctx.Err() != nil {
					return
				}

				if i > 0 && !sleep(ctx, cfg.ThinkTime) {
					return
				}

				u.Iteration = i
				sc := pick(u.Rand, scenarios)
				u.iterate(cfg, sc, rec)
			}
		})
	}

	wg.Wait()

	report := rec.report(cfg)
	if len(report.Violations) > 0 {
		return report, ErrThresholds
	}

	return report, nil
}

func pick(r *rand.Rand, scenarios []Scenario) Scenario {
	total := 0
	for _, sc := range scenarios {
		total += sc.Weight
	}

	n := r.IntN(total)
	for _, sc := range scenarios {
		if n < sc.Weight {
			return sc
		}
		n -= sc.Weight
	}

	return scenarios[len(scenarios)-1]
}

// sleep waits for d and returns false, if ctx has been cancelled.
func sleep(ctx context.Context, d time.Duration) bool {
	if d <= 0 {
		return ctx.Err() == nil
	}

	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

// A User is a virtual user. It keeps its session across all iterations, thus a login persists like in a
// browser. A User implements [nagotest.TB]: failed assertions abort the current iteration.
type User struct {
	// ID is the zero based number of the user.
	ID int
	// Iteration is the zero based number of the current iteration.
	Iteration int
	// Rand is the random source of this user.
	Rand *rand.Rand
	// Data is the entry of [Config.Data] assigned to this user.
	Data map[string]string

	ctx       context.Context
	sessionID string

	mutex    sync.Mutex
	failures []string
	cleanups []func()
}

// Context is cancelled when the run stops.
func (u *User) Context() context.Context {
	return u.ctx
}

// Think pauses the user, like reading the page. It returns early when the run stops.
func (u *User) Think(d time.Duration) {
	sleep(u.ctx, d)
}

func (u *User) Helper() {}

func (u *User) Logf(format string, args ...any) {}

func (u *User) Errorf(format string, args ...any) {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	u.failures = append(u.failures, fmt.Sprintf(format, args...))
}

// Fatalf records the failure and aborts the current iteration.
func (u *User) Fatalf(format string, args ...any) {
	u.Errorf(format, args...)
	runtime.Goexit()
}

func (u *User) Cleanup(f func()) {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	u.cleanups = append(u.cleanups, f)
}

// iterate runs a single iteration in its own goroutine, so that Fatalf can abort it.
func (u *User) iterate(cfg Config, sc Scenario, rec *recorder) {
	u.failures = nil
	done := make(chan struct{})

	go func() {
		defer close(done)
		defer u.runCleanups()
		defer func() {
			if r := recover(); r != nil {
				u.Errorf("panic: %v", r)
			}
		}()

		opts := append([]nagotest.OpenOption{
			nagotest.Values(sc.Values),
			nagotest.Session(u.sessionID),
			nagotest.Lean(),
			nagotest.Observe(func(a nagotest.Action) { rec.action(sc.Name, a) }),
		}, cfg.Options...)

		w := nagotest.Dial(u, cfg.URL, sc.Path, opts...)
		sc.Run(u, w)
	}()

	<-done
	rec.iteration(sc.Name, u.failures)
}

func (u *User) runCleanups() {
	u.mutex.Lock()
	cleanups := u.cleanups
	u.cleanups = nil
	u.mutex.Unlock()

	for i := len(cleanups) - 1; i >= 0; i-- {
		cleanups[i]()
	}
}
