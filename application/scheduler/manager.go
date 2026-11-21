// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package scheduler

import (
	"context"
	"fmt"
	"go.wdy.de/nago/pkg/blob/mem"
	"go.wdy.de/nago/pkg/data/json"
	"iter"
	"log/slog"
	"slices"
	"strings"
	"sync"
	"time"
)

type Manager struct {
	ctx          context.Context
	mutex        sync.Mutex
	services     map[ID]*Scheduler
	settingsRepo SettingsRepository
	runs         *runStore
	clock        Clock
	sweeper      *sweeper
}

// ManagerOptions configure a [Manager], see [NewManagerWith].
type ManagerOptions struct {
	// Settings persists the settings of the schedulers. If nil, an in-memory repository is used.
	Settings SettingsRepository
	// Runs persists the runs and the plan state of durable schedulers. If nil, an in-memory repository is used.
	Runs RunRepository
	// LogDir is the directory of the log files of the runs. If empty, no log files are written.
	LogDir string
	// Clock provides the time and the timers. If nil, the system clock is used.
	Clock Clock
}

// NewManager creates a manager which keeps run statistics only in memory and writes no log files.
func NewManager(ctx context.Context, settingsRepo SettingsRepository) *Manager {
	return NewManagerWithPersistence(ctx, settingsRepo, nil, "")
}

// NewManagerWithPersistence creates a manager which persists run statistics into the given repository and
// writes the log of each run into its own file within logDir. If runRepo is nil, an in-memory repository is used.
// If logDir is empty, no log files are written.
func NewManagerWithPersistence(ctx context.Context, settingsRepo SettingsRepository, runRepo RunRepository, logDir string) *Manager {
	return NewManagerWith(ctx, ManagerOptions{Settings: settingsRepo, Runs: runRepo, LogDir: logDir})
}

// NewManagerWith creates a manager with the given options. The manager deletes runs older than [RunRetention] also
// without further runs, at the latest an hour after they expired.
func NewManagerWith(ctx context.Context, opts ManagerOptions) *Manager {
	settingsRepo := opts.Settings
	if settingsRepo == nil {
		settingsRepo = json.NewSloppyJSONRepository[Settings, ID](mem.NewBlobStore("nago.scheduler.settings"))
	}

	runRepo := opts.Runs
	if runRepo == nil {
		runRepo = json.NewSloppyJSONRepository[Run, RunID](mem.NewBlobStore("nago.scheduler.runs"))
	}

	clock := clockOrSystem(opts.Clock)
	m := &Manager{
		ctx:          ctx,
		services:     make(map[ID]*Scheduler),
		settingsRepo: settingsRepo,
		runs:         newRunStoreWith(runRepo, opts.LogDir, clock),
		clock:        clock,
	}

	m.sweeper = newSweeper(m)
	go m.sweeper.run(ctx)

	return m
}

func (m *Manager) Configure(opts Options) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, ok := m.services[opts.ID]; ok {
		return fmt.Errorf("service with id %s already exists", opts.ID)
	}

	s, err := m.newScheduler(opts)
	if err != nil {
		return err
	}

	m.services[opts.ID] = s
	s.Launch()

	return nil
}

// newScheduler validates the options and creates a scheduler, which restores the plan state of a durable one.
func (m *Manager) newScheduler(opts Options) (*Scheduler, error) {
	if opts.Runner == nil {
		return nil, fmt.Errorf("runner is required")
	}

	if err := validatePlan(opts, opts.Defaults); err != nil {
		return nil, err
	}

	s := newScheduler(m.ctx, opts, m.settingsRepo, m.clock, m.runs)
	s.onRunEnd = m.sweeper.offer

	if opts.Durable {
		if err := m.restoreDurable(s); err != nil {
			return nil, fmt.Errorf("cannot restore the state of %s: %w", opts.ID, err)
		}
	}

	if oldest, err := m.runs.oldestStart(opts.ID); err == nil {
		m.sweeper.offer(oldest)
	}

	return s, nil
}

// restoreDurable loads the plan state of a durable scheduler or creates it, ends its runs without an end as
// cancelled by a restart and deletes its expired runs.
func (m *Manager) restoreDurable(s *Scheduler) error {
	id := s.id
	state, ok, err := m.runs.loadState(id)
	if err != nil {
		return err
	}

	if !ok {
		// the first configuration is a reference of the plan
		state = DurableState{Reference: m.clock.Now()}
		if err := m.runs.saveState(id, state); err != nil {
			return err
		}
	}

	s.plan = state
	m.runs.setFloor(id, state.LogFloor)

	newest, found, err := m.runs.closeUnfinished(id)
	if err != nil {
		return err
	}

	if found {
		started := newest.StartedAt
		s.lastStartedAt.Store(&started)
	}

	ended := state.LastEnd
	s.lastCompletedAt.Store(&ended)

	s.pruneRuns()
	return nil
}

// Remove stops the scheduler and removes it, so that its ID can be configured again. Its persisted settings are
// deleted as well, while the history of its runs is kept, see [Manager.PurgeRuns]. A run in progress is cancelled
// but not awaited.
func (m *Manager) Remove(id ID) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	s, ok := m.services[id]
	if !ok {
		return fmt.Errorf("service with id %s not found", id)
	}

	s.markRemoved()
	s.Destroy()
	delete(m.services, id)

	if err := m.settingsRepo.DeleteByID(id); err != nil {
		return fmt.Errorf("cannot delete settings of %s: %w", id, err)
	}

	return nil
}

// Reconfigure replaces the options of an existing scheduler, e.g. its runner, and launches it again. Its persisted
// settings are kept. A run in progress is cancelled but not awaited. See [Manager.Update] to replace the options
// without cancelling a run.
func (m *Manager) Reconfigure(opts Options) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	old, ok := m.services[opts.ID]
	if !ok {
		return fmt.Errorf("service with id %s not found", opts.ID)
	}

	if opts.Runner == nil {
		return fmt.Errorf("runner is required")
	}

	if err := validatePlan(opts, opts.Defaults); err != nil {
		return err
	}

	old.markRemoved()
	old.Destroy()

	s, err := m.newScheduler(opts)
	if err != nil {
		return err
	}

	m.services[opts.ID] = s
	s.Launch()

	return nil
}

// Update replaces the options of an existing scheduler without cancelling a run in progress, which ends with the
// options it began with. A changed plan applies from the next due on and does not start a run immediately; its
// time becomes a reference of the plan, see [DurableState.Reference]. Durable cannot be changed.
func (m *Manager) Update(opts Options) error {
	s, ok := m.scheduler(opts.ID)
	if !ok {
		return fmt.Errorf("service with id %s not found", opts.ID)
	}

	if opts.Runner == nil {
		return fmt.Errorf("runner is required")
	}

	if opts.Durable != s.options().Durable {
		return fmt.Errorf("service with id %s: durable cannot be changed by an update", opts.ID)
	}

	if err := validatePlan(opts, opts.Defaults); err != nil {
		return err
	}

	s.update(opts)
	return nil
}

// Has returns true, if a scheduler with the given id has been configured.
func (m *Manager) Has(id ID) bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	_, ok := m.services[id]
	return ok
}

// Wake makes the scheduler apply its changed settings immediately instead of after its current delay.
func (m *Manager) Wake(id ID) {
	m.mutex.Lock()
	s, ok := m.services[id]
	m.mutex.Unlock()

	if ok {
		s.Wake()
	}
}

// settingsChanged applies changed settings of a scheduler: a changed plan of a durable scheduler becomes a
// reference of its plan. The loop is woken in any case.
func (m *Manager) settingsChanged(id ID, before Settings) {
	s, ok := m.scheduler(id)
	if !ok {
		return
	}

	opts := s.options()
	if after, err := s.loadSettings(); opts.Durable && (err != nil || planOf(opts, before) != planOf(opts, after)) {
		s.addReference(m.clock.Now())
		return
	}

	s.Wake()
}

// effectiveSettings returns the persisted settings or the defaults of a configured scheduler.
func (m *Manager) effectiveSettings(id ID) (Settings, Options, bool) {
	s, ok := m.scheduler(id)
	if !ok {
		return Settings{}, Options{}, false
	}

	settings, err := s.loadSettings()
	if err != nil {
		return Settings{}, Options{}, false
	}

	return settings, s.options(), true
}

func (m *Manager) Start(id ID) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	s, ok := m.services[id]
	if !ok {
		return fmt.Errorf("service with id %s not found", id)
	}

	if s.Looping() {
		slog.Info("service already started")
		return nil
	}

	s.Launch()
	return nil
}

func (m *Manager) Stop(id ID) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	s, ok := m.services[id]
	if !ok {
		return fmt.Errorf("service with id %s not found", id)
	}

	if !s.Looping() {
		slog.Info("service already stopped")
		return nil
	}

	s.Destroy()
	return nil
}

func (m *Manager) LastStartedAt(id ID) time.Time {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	s, ok := m.services[id]
	if !ok {
		return time.Time{}
	}

	return s.LastStartedAt()
}

func (m *Manager) LastCompletedAt(id ID) time.Time {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	s, ok := m.services[id]
	if !ok {
		return time.Time{}
	}

	return s.LastCompletedAt()
}

func (m *Manager) NextPlannedAt(id ID) time.Time {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	s, ok := m.services[id]
	if !ok {
		return time.Time{}
	}

	return s.NextPlannedAt()
}

func (m *Manager) Logs(id ID) []LogEntry {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	s, ok := m.services[id]
	if !ok {
		return nil
	}

	tmp := s.Logs()
	slices.Reverse(tmp)
	return tmp
}

func (m *Manager) LastError(id ID) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	s, ok := m.services[id]
	if !ok {
		return nil
	}

	return s.LastError()
}

func (m *Manager) State(id ID) State {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	s, ok := m.services[id]
	if !ok {
		return Stopped
	}

	return s.State()
}

// Suspended returns true between [Manager.Pause] and [Manager.Resume].
func (m *Manager) Suspended(id ID) bool {
	s, ok := m.scheduler(id)
	return ok && s.isSuspended()
}

// ExecuteNow runs the scheduler immediately and blocks until the run completes. It waits for a run in progress.
// See [Manager.StartNow] for a manual run without waiting.
func (m *Manager) ExecuteNow(id ID) error {
	m.mutex.Lock()

	s, ok := m.services[id]
	if !ok {
		m.mutex.Unlock()
		return fmt.Errorf("service with id %s not found", id)
	}

	m.mutex.Unlock()

	// ensure, that we execute without the manager lock
	return s.ExecuteNow()
}

// StartNow begins a manual run with by as [Run.By] and returns without waiting for it. While a run is in
// progress, it returns [ErrRunning] without a run. A due during the manual run is omitted, and its end is the end
// of the previous run of a [Schedule]. A paused or stopped scheduler can be started as well.
func (m *Manager) StartNow(id ID, by string) error {
	s, ok := m.scheduler(id)
	if !ok {
		return fmt.Errorf("service with id %s not found", id)
	}

	return s.StartNow(by)
}

// CancelRun cancels only the run in progress: it ends as [OutcomeCanceled] with by as [Run.CanceledBy], while the
// scheduler keeps its plan. Without a run in progress, it returns [ErrNotRunning].
func (m *Manager) CancelRun(id ID, by string) error {
	s, ok := m.scheduler(id)
	if !ok {
		return fmt.Errorf("service with id %s not found", id)
	}

	return s.CancelRun(by)
}

// Pause omits every due of the scheduler until [Manager.Resume]. A run in progress is not cancelled. A durable
// scheduler stays paused over restarts.
func (m *Manager) Pause(id ID) error {
	s, ok := m.scheduler(id)
	if !ok {
		return fmt.Errorf("service with id %s not found", id)
	}

	s.Pause()
	return nil
}

// Resume ends a pause. Its time becomes a reference of the plan (see [DurableState.Reference]), so that it does not
// start a run immediately. A scheduler which is not paused is not changed.
func (m *Manager) Resume(id ID) error {
	s, ok := m.scheduler(id)
	if !ok {
		return fmt.Errorf("service with id %s not found", id)
	}

	s.Resume()
	return nil
}

// SetReference adds the given time to the references of the plan (see [DurableState.Reference]), e.g. after the
// restore of a backup. A time before the current reference has no effect.
func (m *Manager) SetReference(id ID, t time.Time) error {
	s, ok := m.scheduler(id)
	if !ok {
		return fmt.Errorf("service with id %s not found", id)
	}

	s.addReference(t)
	return nil
}

// PurgeRuns deletes all runs, log files and the plan state of the scheduler. The id needs not to be configured, e.g.
// after [Manager.Remove]. A run in progress is not saved at its end. A configured scheduler continues in memory as
// if it had been configured now, keeping a pause; its next configuration after a restart is its first one.
func (m *Manager) PurgeRuns(id ID) error {
	if err := m.runs.purge(id); err != nil {
		return err
	}

	if s, ok := m.scheduler(id); ok {
		s.resetPlan(m.clock.Now())
	}

	return nil
}

// RunsPage returns a page of the runs of the scheduler, newest first, and the amount of all its runs. It reads
// only the runs of the page.
func (m *Manager) RunsPage(id ID, offset, limit int) ([]Run, int, error) {
	return m.runs.page(id, offset, limit)
}

func (m *Manager) Options(id ID) (Options, bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	s, ok := m.services[id]
	if !ok {
		return Options{}, false
	}

	return s.options(), true
}

// Scheduler returns the configured options for all schedulers sorted by name ascending.
func (m *Manager) Scheduler() []Options {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	tmp := make([]Options, 0, len(m.services))
	for _, s := range m.services {
		tmp = append(tmp, s.options())
	}

	slices.SortFunc(tmp, func(a, b Options) int {
		return strings.Compare(a.Name, b.Name)
	})

	return tmp
}

func (m *Manager) scheduler(id ID) (*Scheduler, bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	s, ok := m.services[id]
	return s, ok
}

func (m *Manager) schedulers() []*Scheduler {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	tmp := make([]*Scheduler, 0, len(m.services))
	for _, s := range m.services {
		tmp = append(tmp, s)
	}

	return tmp
}

// Runs returns all known runs of the given scheduler, newest first. See [Manager.RunsPage] to read only a page.
func (m *Manager) Runs(id ID) ([]Run, error) {
	if _, ok := m.scheduler(id); !ok {
		return nil, fmt.Errorf("service with id %s not found", id)
	}

	return m.runs.runs(id)
}

// Stats aggregates the runs of the last 24 hours. It returns these runs, newest first, and at least the latest
// completed run; it reads only the runs of the last 24 hours.
func (m *Manager) Stats(id ID) (RunStats, []Run, error) {
	s, ok := m.scheduler(id)
	if !ok {
		return RunStats{}, nil, fmt.Errorf("service with id %s not found", id)
	}

	runs, err := m.runs.recent(id, locationOf(s.options()))
	if err != nil {
		return RunStats{}, nil, err
	}

	return stats(runs, m.runs.now()), runs, nil
}

// RunLog returns a single page of the log of the given run.
func (m *Manager) RunLog(id ID, runID RunID, q LogQuery) (LogPage, error) {
	s, ok := m.scheduler(id)
	if !ok {
		return LogPage{}, fmt.Errorf("service with id %s not found", id)
	}

	optRun, err := m.runs.repo.FindByID(runID)
	if err != nil {
		return LogPage{}, err
	}

	if optRun.IsNone() || optRun.Unwrap().Scheduler != id || optRun.Unwrap().Durable != nil {
		return LogPage{}, fmt.Errorf("run %s: %w", runID, ErrRunNotFound)
	}

	run := optRun.Unwrap()

	// without a log file, only the current run is available from memory
	if run.LogFile == "" || m.runs.dir == "" {
		if cur, ok := s.CurrentRun(); ok && cur.ID == run.ID {
			logs := s.Logs()
			return pageEntries(func() iter.Seq[LogEntry] { return slices.Values(logs) }, q), nil
		}
	}

	return m.runs.pageLog(run, q)
}

// sweepAll prunes the runs of all schedulers and returns the start of the oldest remaining run.
func (m *Manager) sweepAll() time.Time {
	var oldest time.Time
	for _, s := range m.schedulers() {
		if t := s.pruneRuns(); !t.IsZero() && (oldest.IsZero() || t.Before(oldest)) {
			oldest = t
		}
	}

	return oldest
}

// pruneRuns deletes the expired runs and the log files beyond the kept ones. It returns the start of the oldest
// remaining run.
func (s *Scheduler) pruneRuns() time.Time {
	opts := s.options()
	settings, err := s.loadSettingsFor(opts)
	if err != nil {
		settings = opts.Defaults
	}

	oldest, floor, known := s.runs.pruneRuns(s.id, keepRunLogs(settings), locationOf(opts))
	if known {
		s.planMutex.Lock()
		if s.plan.LogFloor != floor {
			s.plan.LogFloor = floor
			s.persistPlanLocked()
		}
		s.planMutex.Unlock()
	}

	return oldest
}

func (s *Scheduler) markRemoved() {
	s.planMutex.Lock()
	defer s.planMutex.Unlock()

	s.removed = true
}

// resetPlan forgets the plan state after the runs have been purged, as if the scheduler had been configured now. It
// does not persist the plan state, which has been deleted.
func (s *Scheduler) resetPlan(now time.Time) {
	s.planMutex.Lock()
	s.plan = DurableState{Reference: now, Paused: s.plan.Paused}
	s.planMutex.Unlock()

	var zero time.Time
	s.lastCompletedAt.Store(&zero)
	s.Wake()
}

// sweepGap is the least time between two sweeps of the manager.
const sweepGap = time.Hour

// sweeper deletes expired runs without further runs. Its timer waits for the expiry of the oldest run, but sweeps
// at most once per sweepGap.
type sweeper struct {
	m     *Manager
	mu    sync.Mutex
	alarm time.Time
	last  time.Time
	kick  chan struct{}
}

func newSweeper(m *Manager) *sweeper {
	return &sweeper{m: m, kick: make(chan struct{}, 1)}
}

// offer tells the sweeper the start of the oldest run of a scheduler.
func (w *sweeper) offer(oldest time.Time) {
	if oldest.IsZero() {
		return
	}

	t := oldest.Add(RunRetention + time.Second)

	w.mu.Lock()
	if !w.last.IsZero() {
		t = later(t, w.last.Add(sweepGap))
	}

	lowered := w.alarm.IsZero() || t.Before(w.alarm)
	if lowered {
		w.alarm = t
	}
	w.mu.Unlock()

	if lowered {
		select {
		case w.kick <- struct{}{}:
		default:
		}
	}
}

func (w *sweeper) run(ctx context.Context) {
	for {
		w.mu.Lock()
		alarm := w.alarm
		w.mu.Unlock()

		if alarm.IsZero() {
			select {
			case <-ctx.Done():
				return
			case <-w.kick:
				continue
			}
		}

		if d := alarm.Sub(w.m.clock.Now()); d > 0 {
			timer := w.m.clock.NewTimer(d)
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-w.kick:
				timer.Stop()
				continue
			case <-timer.C():
			}
		}

		w.mu.Lock()
		w.alarm = time.Time{}
		w.last = w.m.clock.Now()
		w.mu.Unlock()

		w.offer(w.m.sweepAll())
	}
}
