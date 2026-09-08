// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uicompletion

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/presentation/core"
)

// rootViewPurposeLookup mirrors application.RootViewPurposeLookup. It is redeclared here because this package
// must not import the application package, and a context value is matched structurally.
type rootViewPurposeLookup func(path core.NavigationPath) (string, bool)

// ctxRootViewPurpose mirrors application.CtxRootViewPurpose.
const ctxRootViewPurpose = "nago.rootview.purpose"

// WindowContext renders the situation the user is currently in as a block of plain text, ready to be embedded
// into a system prompt via [Agent.SystemPromptFunc]:
//
//	SystemPromptFunc: func() string {
//		return domainPrompt + "\n\n" + uicompletion.WindowContext(wnd)
//	}
//
// It answers the three questions an assistant otherwise has to guess at, and gets wrong: where the user
// stands, who they are, and what they are allowed to do. All three are derived from what the framework
// already knows - the registered route and its stated purpose (see application.Purpose), the values that
// route was opened with, and the permissions of the acting subject - so it cannot drift away from the
// application the way a hand-maintained description of screens does.
//
// It deliberately contains no domain knowledge. Everything specific to the application belongs in the
// caller's own part of the prompt.
func WindowContext(wnd core.Window) string {
	var sb strings.Builder

	sb.WriteString("--- Where the user currently is ---\n")
	sb.WriteString(describeLocation(wnd))

	if sel := describeSelection(wnd); sel != "" {
		sb.WriteString(sel)
	}

	sb.WriteString("\n--- Who is asking ---\n")
	sb.WriteString(describeActor(wnd))

	return sb.String()
}

// describeLocation names the current route and, when the registration stated one, its purpose.
func describeLocation(wnd core.Window) string {
	path := wnd.Path()
	if path == "" {
		path = "."
	}

	purpose, ok := lookupPurpose(wnd, path)
	if !ok {
		// Saying nothing would invite the model to invent a meaning for the path. Saying that the screen is
		// undescribed is both true and useful: it tells the model to rely on its tools instead.
		return fmt.Sprintf("Route: %s (no description was registered for this screen; do not guess what it shows)\n", path)
	}

	return fmt.Sprintf("Route: %s\nPurpose of this screen: %s\n", path, purpose)
}

// lookupPurpose asks the application for the purpose registered with the route.
func lookupPurpose(wnd core.Window, path core.NavigationPath) (string, bool) {
	lookup, ok := core.FromContext[rootViewPurposeLookup](wnd.Context(), ctxRootViewPurpose)
	if !ok || lookup == nil {
		return "", false
	}

	return lookup(path)
}

// describeSelection lists the values the current route was opened with, which is what "the record the user is
// looking at" amounts to in nago.
func describeSelection(wnd core.Window) string {
	values := wnd.Values()
	if len(values) == 0 {
		return ""
	}

	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	sb.WriteString("The screen was opened with these parameters (they usually identify what the user is looking at):\n")
	for _, k := range keys {
		if values[k] == "" {
			continue
		}
		sb.WriteString(fmt.Sprintf("- %s: %s\n", k, values[k]))
	}

	return sb.String()
}

// describeActor names the user and enumerates what they may do, derived from the permissions they actually
// hold rather than from a hand-written list.
func describeActor(wnd core.Window) string {
	subject := wnd.Subject()
	if subject == nil || !subject.Valid() {
		return "The user is not signed in.\n"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Name: %s\nUser id: %s\n", subject.Name(), subject.ID()))

	var allowed []string
	for perm := range permission.All() {
		if subject.HasPermission(perm.ID) {
			allowed = append(allowed, permissionLabel(perm))
		}
	}

	slices.Sort(allowed)
	allowed = slices.Compact(allowed)

	if len(allowed) == 0 {
		sb.WriteString("This user holds no permissions at all, so every tool will refuse.\n")
		return sb.String()
	}

	sb.WriteString("Your permissions are exactly those of this user. What they may not do, you cannot do either - a tool then answers with a permission error, which is an expected outcome and not a malfunction.\n")
	sb.WriteString("They may:\n")
	for _, label := range allowed {
		sb.WriteString("- " + label + "\n")
	}

	return sb.String()
}

// permissionLabel prefers the human-readable name a permission was declared with and falls back to its id.
func permissionLabel(perm permission.Permission) string {
	if name := strings.TrimSpace(perm.Name); name != "" {
		return name
	}

	return string(perm.ID)
}
