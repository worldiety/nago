// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uimail

import "go.wdy.de/nago/presentation/core"

type Pages struct {
	// Dashboard shows statistics, problems and recent failures.
	Dashboard core.NavigationPath
	// OutgoingMailQueue lists the outgoing mails. Supports the query parameters status, server and stuck.
	OutgoingMailQueue core.NavigationPath
	// OutgoingMail shows the details of a single outgoing mail identified by the query parameter id.
	OutgoingMail core.NavigationPath
	// SmtpServers shows the configured smtp servers and their health.
	SmtpServers core.NavigationPath
	// Deprecated: there is no scheduler page.
	MailScheduler core.NavigationPath
	SendMailTest  core.NavigationPath

	// the following pages are optional and owned by other systems

	Templates   core.NavigationPath // template projects, filtered by tag=mail
	SecretVault core.NavigationPath
	SecretEdit  core.NavigationPath // requires id parameter
}
