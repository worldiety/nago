/**
 * Copyright (c) 2025 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */

export default interface DatepickerDay {
	dayOfWeek: number;
	dayOfMonth: number;
	monthIndex: number;
	year: number;
	selectedStart: boolean;
	selectedEnd: boolean;
	withinRange: boolean;
	selectable: boolean;
	otherMonth: boolean;
}
