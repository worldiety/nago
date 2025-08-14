// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package workflow

import (
	"go.wdy.de/nago/application/permission"
)

var (
	PermFindDeclaredWorkflows = permission.Declare[FindDeclaredWorkflows]("nago.workflow.finddeclaredworkflows", "Finde deklarierte Workflows", "Träger dieser Berechtigung können definierte Workflows ansehen.")
	PermCreateInstance        = permission.Declare[FindDeclaredWorkflows]("nago.workflow.createinstance", "Erzeuge eine Workflow-Instanz", "Träger dieser Berechtigung können neue Workflow-Instanzen erstellen.")
	PermFindInstances         = permission.Declare[FindInstances]("nago.workflow.findinstances", "Finde Workflow-Instanzen", "Träger dieser Berechtigung können Instanzen finden.")
	PermAnalyze               = permission.Declare[Analyze]("nago.workflow.analyze", "Workflows statisch analysieren", "Träger dieser Berechtigung können Workflows statisch auswerten.")
	PermProcessEvent          = permission.Declare[ProcessEvent]("nago.workflow.processevent", "Ein Event verarbeiten", "Träger dieser Berechtigung können Workflow-Events zur Verarbeitung einreichen.")
	PermGetStatus             = permission.Declare[GetStatus]("nago.workflow.instance.getstatus", "Status einer Workflow-Instanz auslesen", "Träger dieser Berechtigung können den Status einer Workflow-Instanz auslesen.")
	PermFindInstanceEvents    = permission.Declare[FindInstanceEvents]("nago.workflow.instance.findevents", "Events einer Workflow-Instanz auflisten", "Träger dieser Berechtigung können die Events einer Instanz anzeigen.")
)
