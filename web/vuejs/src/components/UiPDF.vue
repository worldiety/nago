<!--
 Copyright (c) 2026 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->
<template>
	<PDFViewer class="pdf" :config="config" :style="frameStyles" @init="onInit" />
</template>

<script lang="ts" setup>
import { computed, ref } from 'vue';
import { frameCSS } from '@/components/shared/frame';
import { PDF } from '@/shared/proto/nprotoc_gen';
import PDFViewer, { EmbedPdfContainer, PDFViewerConfig } from '@embedpdf/vue-pdf-viewer';
import { ThemeKey, useThemeManager } from '@/shared/themeManager';
import { useI18n } from 'vue-i18n';

const props = defineProps<{
	ui: PDF;
}>();

const themeManager = useThemeManager();
const { locale } = useI18n();

const viewer = ref<EmbedPdfContainer>();

const config: PDFViewerConfig = {
	src: props.ui.src,
	i18n: {
		defaultLocale: locale.value.split('-')[0],
		fallbackLocale: 'de',
	},
	theme: {
		light: {
			accent: {
				primary: 'var(--I0)',
			},
			foreground: {
				secondary: 'currentColor',
				muted: 'currentColor',
				onAccent: 'currentColor',
			},
			background: {
				app: 'var(--M2)',
				surface: 'var(--M3)',
				surfaceAlt: 'var(--M4)',
				elevated: 'var(--M2)',
			},
			interactive: {
				hover: 'var(--M2)',
				selected: 'var(--M2)',
			},
			border: {
				default: 'transparent',
			},
		},
		dark: {
			accent: {
				primary: 'var(--I0)',
			},
			foreground: {
				secondary: 'currentColor',
				muted: 'currentColor',
			},
			background: {
				app: 'var(--M2)',
				surface: 'var(--M3)',
				surfaceAlt: 'var(--M4)',
				elevated: 'var(--M2)',
			},
			interactive: {
				hover: 'var(--M2)',
				selected: 'var(--M2)',
			},
			border: {
				default: 'transparent',
			},
		},
	},
	disabledCategories: ['annotation', 'document', 'form', 'insert', 'panel', 'redaction', 'tools'],
};

const frameStyles = computed<string>(() => {
	const styles = frameCSS(props.ui.frame);

	return styles.join(';');
});

function setTheme(themeKey?: ThemeKey) {
	if (!themeKey) themeKey = themeManager.getActiveThemeKey() ?? ThemeKey.SYSTEM;

	switch (themeKey) {
		case ThemeKey.LIGHT:
			config.theme.preference = 'light';
			break;
		case ThemeKey.DARK:
			config.theme.preference = 'dark';
			break;
		default:
			config.theme.preference = 'system';
			break;
	}
	viewer.value?.setTheme(config.theme);
}

function onInit(container: EmbedPdfContainer) {
	viewer.value = container;
	setTheme();
}

themeManager.observeTheme(setTheme);
</script>
<style scoped>
.pdf {
	@apply rounded-2xl overflow-hidden;
}
</style>
