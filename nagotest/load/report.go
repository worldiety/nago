// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package load

import (
	"cmp"
	"fmt"
	"io"
	"regexp"
	"slices"
	"strings"
	"sync"
	"text/tabwriter"
	"time"

	"go.wdy.de/nago/nagotest"
)

// A Report summarizes a run.
type Report struct {
	Started    time.Time       `json:"started"`
	Elapsed    time.Duration   `json:"elapsedNs"`
	Users      int             `json:"users"`
	Scenarios  []ScenarioStats `json:"scenarios"`
	Actions    []ActionStats   `json:"actions"`
	Errors     []ErrorStats    `json:"errors"`
	Violations []string        `json:"violations"`
}

// ScenarioStats counts the iterations of a scenario.
type ScenarioStats struct {
	Name       string  `json:"name"`
	Iterations int     `json:"iterations"`
	Failed     int     `json:"failed"`
	ErrorRate  float64 `json:"errorRate"`
}

// ActionStats are the latencies of an action, see [nagotest.Action]. Only settled actions are measured,
// thus a failed action is counted as an error of its scenario instead.
type ActionStats struct {
	Scenario string        `json:"scenario"`
	Kind     string        `json:"kind"`
	Target   string        `json:"target"`
	Count    int           `json:"count"`
	PerSec   float64       `json:"perSec"`
	Min      time.Duration `json:"minNs"`
	Mean     time.Duration `json:"meanNs"`
	P50      time.Duration `json:"p50Ns"`
	P95      time.Duration `json:"p95Ns"`
	P99      time.Duration `json:"p99Ns"`
	Max      time.Duration `json:"maxNs"`
}

// ErrorStats counts identical failure messages.
type ErrorStats struct {
	Scenario string `json:"scenario"`
	Message  string `json:"message"`
	Count    int    `json:"count"`
}

// Iterations returns the total amount of iterations.
func (r *Report) Iterations() (total, failed int) {
	for _, s := range r.Scenarios {
		total += s.Iterations
		failed += s.Failed
	}

	return total, failed
}

// WriteText writes a human-readable summary.
func (r *Report) WriteText(dst io.Writer) error {
	tw := tabwriter.NewWriter(dst, 0, 0, 2, ' ', 0)
	total, failed := r.Iterations()
	fmt.Fprintf(tw, "users: %d, elapsed: %v, iterations: %d, failed: %d\n\n", r.Users, r.Elapsed.Round(time.Millisecond), total, failed)

	fmt.Fprintln(tw, "SCENARIO\tITERATIONS\tFAILED\tERROR RATE")
	for _, s := range r.Scenarios {
		fmt.Fprintf(tw, "%s\t%d\t%d\t%.2f%%\n", s.Name, s.Iterations, s.Failed, s.ErrorRate*100)
	}

	fmt.Fprintln(tw)
	fmt.Fprintln(tw, "SCENARIO\tACTION\tCOUNT\tPER SEC\tMIN\tMEAN\tP50\tP95\tP99\tMAX")
	for _, a := range r.Actions {
		fmt.Fprintf(tw, "%s\t%s %s\t%d\t%.1f\t%v\t%v\t%v\t%v\t%v\t%v\n", a.Scenario, a.Kind, a.Target, a.Count, a.PerSec,
			ms(a.Min), ms(a.Mean), ms(a.P50), ms(a.P95), ms(a.P99), ms(a.Max))
	}

	if len(r.Errors) > 0 {
		fmt.Fprintln(tw)
		fmt.Fprintln(tw, "SCENARIO\tCOUNT\tERROR")
		for _, e := range r.Errors {
			fmt.Fprintf(tw, "%s\t%d\t%s\n", e.Scenario, e.Count, e.Message)
		}
	}

	if len(r.Violations) > 0 {
		fmt.Fprintln(tw)
		fmt.Fprintln(tw, "THRESHOLD VIOLATIONS")
		for _, v := range r.Violations {
			fmt.Fprintln(tw, v)
		}
	}

	return tw.Flush()
}

func ms(d time.Duration) time.Duration {
	if d >= time.Second {
		return d.Round(time.Millisecond)
	}

	return d.Round(10 * time.Microsecond)
}

type actionKey struct {
	scenario, kind, target string
}

type errorKey struct {
	scenario, message string
}

type recorder struct {
	mutex      sync.Mutex
	started    time.Time
	actions    map[actionKey][]time.Duration
	order      map[actionKey]int // first occurrence, to keep the order of the journey
	iterations map[string]*ScenarioStats
	errors     map[errorKey]int
}

func newRecorder() *recorder {
	return &recorder{
		started:    time.Now(),
		actions:    map[actionKey][]time.Duration{},
		order:      map[actionKey]int{},
		iterations: map[string]*ScenarioStats{},
		errors:     map[errorKey]int{},
	}
}

func (r *recorder) action(scenario string, a nagotest.Action) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	key := actionKey{scenario: scenario, kind: a.Kind, target: a.Target}
	if _, ok := r.order[key]; !ok {
		r.order[key] = len(r.order)
	}
	r.actions[key] = append(r.actions[key], a.Duration)
}

func (r *recorder) iteration(scenario string, failures []string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	s := r.iterations[scenario]
	if s == nil {
		s = &ScenarioStats{Name: scenario}
		r.iterations[scenario] = s
	}

	s.Iterations++
	if len(failures) > 0 {
		s.Failed++
	}

	for _, f := range failures {
		r.errors[errorKey{scenario: scenario, message: normalize(f)}]++
	}
}

// randomIDs are hex identifiers like scope or session ids.
var randomIDs = regexp.MustCompile(`[0-9a-f]{16,}`)

// normalize removes the varying parts of messages, so that identical failures are counted together.
func normalize(msg string) string {
	msg = strings.TrimSpace(msg)
	if i := strings.Index(msg, "\n"); i >= 0 {
		msg = msg[:i]
	}

	return randomIDs.ReplaceAllString(msg, "…")
}

func (r *recorder) report(cfg Config) *Report {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	rep := &Report{Started: r.started, Elapsed: time.Since(r.started), Users: cfg.Users}
	seconds := rep.Elapsed.Seconds()

	for _, s := range r.iterations {
		if s.Iterations > 0 {
			s.ErrorRate = float64(s.Failed) / float64(s.Iterations)
		}

		rep.Scenarios = append(rep.Scenarios, *s)
		if cfg.Thresholds.MaxErrorRate > 0 && s.ErrorRate > cfg.Thresholds.MaxErrorRate {
			rep.Violations = append(rep.Violations, fmt.Sprintf("%s: error rate %.2f%% exceeds %.2f%%", s.Name, s.ErrorRate*100, cfg.Thresholds.MaxErrorRate*100))
		}
	}

	for key, durations := range r.actions {
		slices.Sort(durations)
		var sum time.Duration
		for _, d := range durations {
			sum += d
		}

		a := ActionStats{
			Scenario: key.scenario,
			Kind:     key.kind,
			Target:   key.target,
			Count:    len(durations),
			PerSec:   float64(len(durations)) / seconds,
			Min:      durations[0],
			Mean:     sum / time.Duration(len(durations)),
			P50:      percentile(durations, 0.50),
			P95:      percentile(durations, 0.95),
			P99:      percentile(durations, 0.99),
			Max:      durations[len(durations)-1],
		}

		rep.Actions = append(rep.Actions, a)
		if cfg.Thresholds.P95 > 0 && a.P95 > cfg.Thresholds.P95 {
			rep.Violations = append(rep.Violations, fmt.Sprintf("%s: %s %s: p95 %v exceeds %v", a.Scenario, a.Kind, a.Target, ms(a.P95), cfg.Thresholds.P95))
		}
	}

	for key, count := range r.errors {
		rep.Errors = append(rep.Errors, ErrorStats{Scenario: key.scenario, Message: key.message, Count: count})
	}

	if len(rep.Scenarios) == 0 {
		rep.Violations = append(rep.Violations, "no iteration has been completed")
	}

	slices.SortFunc(rep.Scenarios, func(a, b ScenarioStats) int { return cmp.Compare(a.Name, b.Name) })
	slices.SortFunc(rep.Actions, func(a, b ActionStats) int {
		return cmp.Or(cmp.Compare(a.Scenario, b.Scenario), cmp.Compare(
			r.order[actionKey{a.Scenario, a.Kind, a.Target}],
			r.order[actionKey{b.Scenario, b.Kind, b.Target}],
		))
	})
	slices.SortFunc(rep.Errors, func(a, b ErrorStats) int {
		return cmp.Or(cmp.Compare(b.Count, a.Count), cmp.Compare(a.Message, b.Message))
	})
	slices.Sort(rep.Violations)

	return rep
}

// percentile uses the nearest-rank method on sorted durations.
func percentile(sorted []time.Duration, p float64) time.Duration {
	rank := int(float64(len(sorted))*p+0.5) - 1
	return sorted[max(0, min(rank, len(sorted)-1))]
}
