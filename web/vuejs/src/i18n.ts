/**
 * Copyright (c) 2025 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */
import { createI18n } from 'vue-i18n';
import de from '@/locales/de.json';
import en from '@/locales/en.json';

type MessageSchemaDe = typeof de;
type MessageSchemaEn = typeof en;

export const activeLocale = navigator.language;

const i18n = createI18n<[MessageSchemaDe | MessageSchemaEn], 'de' | 'en'>({
	legacy: false, // set `false`, to use Composition API
	locale: activeLocale, // set locale to the language determined by the browser
	fallbackLocale: 'de',
	messages: {
		de: de,
		en: en,
	},
});

export default i18n;
