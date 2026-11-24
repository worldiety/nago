// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uionboarding

import (
	"log/slog"
	"strings"

	"go.wdy.de/nago/application/onboarding"
	"go.wdy.de/nago/application/theme"
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
	// StepProblem means, that the setup cannot be done: the address is missing or invalid, see
	// [onboarding.State.Problem], or no code has been requested yet and there is no way to send it, see
	// [onboarding.DeliveryNone]. A code, which has been sent, keeps the window in [StepCode].
	StepProblem Step = iota
	// StepWelcome means, that no code is valid right now and a code can be sent, see [onboarding.State.Delivery].
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

// FlowOptions configure a [Flow], see [NewFlow].
type FlowOptions struct {
	// Legal optionally replaces the links to the legal texts of the theme settings below the card of every step,
	// see [Flow.Legal] and [theme.Settings]. It replaces both links: an empty value has no link.
	Legal *Legal
	// EmailEnv is the name of the environment variable with the address of the first user for the hints to the
	// operators on the problem step, see [StepProblem]. Default is [onboarding.EnvInitialUserEmail].
	EmailEnv string
	// MailServiceEnv is the name of the environment variable, which turns the Nago Mail Service off, for the hints
	// to the operators on the problem step.
	MailServiceEnv string
}

// Legal are the links to the legal texts. An empty value has no link. A value with "://" is an external address,
// which opens in a new tab, any other value is a route of the application.
type Legal struct {
	Impress       string
	PrivacyPolicy string
}

// Flow binds the use cases to the session of a window. Custom setup pages build on it. Every action renders the
// window again, because the state of the setup is not a state of the window.
type Flow struct {
	wnd             core.Window
	uc              onboarding.UseCases
	subjectFromUser user.SubjectFromUser
	opts            FlowOptions
	changed         *core.State[int]
	method          *core.State[Method]
}

// NewFlow creates the flow of the window. The subjectFromUser function logs the window in after the setup. At most
// one FlowOptions is used.
func NewFlow(wnd core.Window, uc onboarding.UseCases, subjectFromUser user.SubjectFromUser, opts ...FlowOptions) *Flow {
	f := &Flow{
		wnd:             wnd,
		uc:              uc,
		subjectFromUser: subjectFromUser,
		changed:         core.StateOf[int](wnd, "nago.onboarding.flow"),
		method:          core.StateOf[Method](wnd, "nago.onboarding.method"),
	}

	if len(opts) > 0 {
		f.opts = opts[0]
	}

	return f
}

func (f *Flow) Window() core.Window {
	return f.wnd
}

func (f *Flow) State() onboarding.State {
	return f.uc.State()
}

// Step derives the current step.
func (f *Flow) Step() Step {
	return f.stepOf(f.uc.State())
}

// stepOf derives the step from the given state, so that a page shows the step and its reasons of the same state.
func (f *Flow) stepOf(st onboarding.State) Step {
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

	// the delivery only matters, before a code has been sent
	if st.Delivery.Level == onboarding.DeliveryNone {
		return StepProblem
	}

	return StepWelcome
}

// Recheck renders the window again, which reads the state of the setup anew, e.g. after the operators have added a
// way to send the code.
func (f *Flow) Recheck() {
	f.changed.Invalidate()
}

// CodeStatus tells, what has become of the mail with the valid code, see [onboarding.CurrentCodeStatus].
func (f *Flow) CodeStatus() onboarding.CodeStatus {
	if f.uc.CodeStatus == nil {
		return onboarding.CodeStatusUnknown
	}

	return f.uc.CodeStatus()
}

// Legal returns the links to the legal texts of [FlowOptions.Legal] or, if not set, of the theme settings.
func (f *Flow) Legal() Legal {
	legal := f.opts.Legal
	if legal == nil {
		themeSettings := core.GlobalSettings[theme.Settings](f.wnd)
		legal = &Legal{Impress: themeSettings.Impress, PrivacyPolicy: themeSettings.PrivacyPolicy}
	}

	return Legal{Impress: strings.TrimSpace(legal.Impress), PrivacyPolicy: strings.TrimSpace(legal.PrivacyPolicy)}
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
