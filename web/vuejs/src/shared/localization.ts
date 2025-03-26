/**
 * Copyright (c) 2025 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */
import { activeLocale } from '@/i18n';

export function localizeNumber(rawNumber: number, options: Intl.NumberFormatOptions): string {
	return rawNumber.toLocaleString(activeLocale, options);
}
