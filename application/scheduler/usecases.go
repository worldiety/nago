package scheduler

import (
	"context"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/std"
	"iter"
	"time"
)

type ID string

type State int

const (
	Stopped State = iota
	Running
	Disabled
	Paused
)

type Kind int

const (
	OneShot Kind = iota
	Schedule
	Manual
)

type Options struct {
	ID          ID
	Name        string
	Description string
	Kind        Kind
	Defaults    Settings
	Runner      func(context.Context) error
}

type StatusResult struct {
	State           State
	LastStartedAt   time.Time
	LastCompletedAt time.Time
	NextPlannedAt   time.Time
	LastError       error
	Options         Options
}
type Status func(subject auth.Subject, id ID) (StatusResult, error)

// Configure introduces a new system level service. It is not intended, that these services are configured by
// end users. Usually, a developer defines the schedulers at build time.
type Configure func(subject auth.Subject, opts Options) error

type ViewLogs func(subject auth.Subject, id ID) iter.Seq2[LogEntry, error]
type ExecuteNow func(subject auth.Subject, id ID) error

type ListSchedulers func(subject auth.Subject) iter.Seq2[Options, error]

type Stop func(subject auth.Subject, id ID) error
type Start func(subject auth.Subject, id ID) error

type FindSettingsByID func(subject auth.Subject, id ID) (std.Option[Settings], error)
type UpdateSettings func(subject auth.Subject, settings Settings) error
type DeleteSettingsByID func(subject auth.Subject, id ID) error

type UseCases struct {
	Configure          Configure
	ViewLogs           ViewLogs
	Status             Status
	ExecuteNow         ExecuteNow
	ListSchedulers     ListSchedulers
	Stop               Stop
	Start              Start
	FindSettingsByID   FindSettingsByID
	UpdateSettings     UpdateSettings
	DeleteSettingsByID DeleteSettingsByID
}

func NewUseCases(ctx context.Context, settingsRepo SettingsRepository) UseCases {
	m := NewManager(ctx, settingsRepo)
	return UseCases{
		Configure:          NewConfigure(m),
		ViewLogs:           NewViewLogs(m),
		Status:             NewStatus(m),
		ExecuteNow:         NewExecuteNow(m),
		ListSchedulers:     NewListSchedulers(m),
		Stop:               NewStop(m),
		Start:              NewStart(m),
		FindSettingsByID:   NewFindSettingsByID(settingsRepo),
		UpdateSettings:     NewUpdateSettings(settingsRepo),
		DeleteSettingsByID: NewDeleteSettingsByID(settingsRepo),
	}
}
