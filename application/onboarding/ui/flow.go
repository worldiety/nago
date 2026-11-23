// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uionboarding

import (
	"log/slog"

	"go.wdy.de/nago/application/onboarding"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/presentation/core"
)

type Pages struct {
	// Setup is the dedicated route of the setup. As long as the setup is pending, every route shows it.
	Setup core.NavigationPath
}

// Step of the setup from the perspective of a window.
type Step int

const (
	// StepProblem means, that the setup cannot be done, see [onboarding.State.Problem].
	StepProblem Step = iota
	// StepWelcome means, that no code is valid right now.
	StepWelcome
	// StepCode means, that a code has been sent and waits to be entered.
	StepCode
	// StepProfile means, that the session has been confirmed and the account can be created.
	StepProfile
	// StepDone means, that the instance has already been set up.
	StepDone
	// StepMethod means, that the session has been confirmed and the user chooses between the single sign-on and a
	// password, see [onboarding.State.SSOServer].
	StepMethod
	// StepSSO means, that the user has chosen the single sign-on and decides who else may sign in by it.
	StepSSO
)

// Method of the sign-in of the first user.
type Method int

const (
	// MethodUndecided lets the user choose, if the setup offers the single sign-on.
	MethodUndecided Method = iota
	// MethodSSO signs in by the login service, see [Flow.CompleteWithSSO].
	MethodSSO
	// MethodPassword creates an account with a password, see [Flow.Complete].
	MethodPassword
)

// Flow binds the use cases to the session of a window. Custom setup pages build on it. Every action renders the
// window again, because the state of the setup is not a state of the window.
type Flow struct {
	wnd             core.Window
	uc              onboarding.UseCases
	subjectFromUser user.SubjectFromUser
	changed         *core.State[int]
	method          *core.State[Method]
}

// NewFlow creates the flow of the window. The subjectFromUser function logs the window in after the setup.
func NewFlow(wnd core.Window, uc onboarding.UseCases, subjectFromUser user.SubjectFromUser) *Flow {
	return &Flow{
		wnd:             wnd,
		uc:              uc,
		subjectFromUser: subjectFromUser,
		changed:         core.StateOf[int](wnd, "nago.onboarding.flow"),
		method:          core.StateOf[Method](wnd, "nago.onboarding.method"),
	}
}

func (f *Flow) Window() core.Window {
	return f.wnd
}

func (f *Flow) State() onboarding.State {
	return f.uc.State()
}

// Step derives the current step.
func (f *Flow) Step() Step {
	st := f.uc.State()
	switch {
	case !st.Pending:
		return StepDone
	case st.Problem != onboarding.NoProblem:
		return StepProblem
	case f.uc.Verified(f.wnd.Session().ID()):
		switch {
		case st.SSOServer == "" || f.method.Get() == MethodPassword:
			return StepProfile
		case f.method.Get() == MethodSSO:
			return StepSSO
		default:
			return StepMethod
		}
	}

	if _, ok := f.uc.CurrentCode(); ok {
		return StepCode
	}

	return StepWelcome
}

// Code returns the code, which has been sent last, if it is still valid.
func (f *Flow) Code() (onboarding.Code, bool) {
	return f.uc.CurrentCode()
}

// RequestCode sends a new code.
func (f *Flow) RequestCode() error {
	defer f.changed.Invalidate()
	_, err := f.uc.RequestCode(f.wnd.Session().ID())
	return err
}

// VerifyCode confirms the session of the window.
func (f *Flow) VerifyCode(code string) error {
	defer f.changed.Invalidate()
	return f.uc.VerifyCode(f.wnd.Session().ID(), code)
}

// Complete creates the account, logs the session and the window in and navigates to the index.
func (f *Flow) Complete(profile onboarding.Profile) error {
	if profile.PreferredLanguage == (onboarding.Profile{}).PreferredLanguage {
		profile.PreferredLanguage = f.wnd.Locale()
	}

	defer f.changed.Invalidate()
	usr, err := f.uc.Complete(f.wnd.Session().ID(), profile)
	if err != nil {
		return err
	}

	// the window shows the setup at any route, so navigating to the index may keep the window and with it the
	// anonymous subject, which is only resolved for new windows
	if f.subjectFromUser != nil {
		if optSubject, err := f.subjectFromUser(user.SU(), usr.ID); err != nil {
			slog.Error("onboarding: cannot resolve the subject of the first user", "err", err)
		} else if optSubject.IsSome() {
			f.wnd.UpdateSubject(optSubject.Unwrap())
		}
	}

	f.wnd.Navigation().ResetTo(".", nil)
	return nil
}

// Method returns the method, which the user has chosen.
func (f *Flow) Method() Method {
	return f.method.Get()
}

// ChooseMethod continues with the single sign-on or the password. MethodUndecided goes back to the choice.
func (f *Flow) ChooseMethod(m Method) {
	f.method.Set(m)
	f.method.Notify()
}

// CompleteWithSSO creates the account and opens the login service in the window, see [onboarding.CompleteWithSSO].
// The login service sends the browser back to the application afterward, which then shows the index.
func (f *Flow) CompleteWithSSO(access onboarding.SSOAccess) error {
	defer f.changed.Invalidate()
	loginURL, err := f.uc.CompleteWithSSO(f.wnd.Session().ID(), access, f.wnd.Locale())
	if err != nil {
		return err
	}

	core.HTTPOpen(f.wnd.Navigation(), core.URI(loginURL), "_self")
	return nil
}

// Reset forgets the confirmation of the session.
func (f *Flow) Reset() {
	defer f.changed.Invalidate()
	f.uc.Reset(f.wnd.Session().ID())
}
