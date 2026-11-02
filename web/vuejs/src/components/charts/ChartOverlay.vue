<!--
 Copyright (c) 2026 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<script lang="ts" setup>
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import type { Callout } from '@/components/charts/chart';
import { downloadCSV, downloadPNG, downloadSVG } from '@/components/charts/export';

const props = defineProps<{
	width: number;
	height: number;
	callout?: Callout;
	// announcement is read out by screen readers, for an inspection made with the keyboard.
	announcement?: string;
	message?: string;
	downloadable: boolean;
	// table returns the data of the chart as rows, the first one being the header.
	table: () => string[][];
}>();

const { t } = useI18n();
const overlay = ref<HTMLElement>();
const button = ref<HTMLButtonElement>();
const menu = ref<HTMLElement>();
const menuOpen = ref(false);

const calloutStyle = computed(() => {
	const c = props.callout;
	if (!c) {
		return {};
	}

	// beside the inspected position, on the side with more room
	const style: Record<string, string> = {};
	const gap = (c.half ?? 0) + 12;
	if (c.x <= props.width / 2) {
		style.left = `${c.x + gap}px`;
	} else {
		style.right = `${props.width - c.x + gap}px`;
	}

	if (c.y === undefined) {
		style.top = `${c.top}px`;
	} else if (c.y <= props.height / 2) {
		style.top = `${c.y + 12}px`;
	} else {
		style.bottom = `${props.height - c.y + 12}px`;
	}

	return style;
});

function drawing(): SVGSVGElement | null {
	return overlay.value?.parentElement?.querySelector('svg.chart-drawing') ?? null;
}

async function save(format: 'svg' | 'png' | 'csv') {
	closeMenu(true);
	const svg = drawing();
	switch (format) {
		case 'svg':
			if (svg) downloadSVG(svg, 'chart');
			break;
		case 'png':
			if (svg) await downloadPNG(svg, 'chart');
			break;
		case 'csv':
			downloadCSV(props.table(), 'chart');
			break;
	}
}

async function toggleMenu() {
	if (menuOpen.value) {
		closeMenu(false);
		return;
	}

	menuOpen.value = true;
	await nextTick();
	menu.value?.querySelector<HTMLElement>('[role="menuitem"]')?.focus();
}

function closeMenu(refocus: boolean) {
	menuOpen.value = false;
	if (refocus) {
		button.value?.focus();
	}
}

function onMenuKeyDown(e: KeyboardEvent) {
	const items = [...(menu.value?.querySelectorAll<HTMLElement>('[role="menuitem"]') ?? [])];
	const at = items.indexOf(document.activeElement as HTMLElement);
	switch (e.key) {
		case 'ArrowDown':
			items[(at + 1) % items.length]?.focus();
			break;
		case 'ArrowUp':
			items[(at - 1 + items.length) % items.length]?.focus();
			break;
		case 'Home':
			items[0]?.focus();
			break;
		case 'End':
			items[items.length - 1]?.focus();
			break;
		case 'Escape':
			closeMenu(true);
			break;
		case 'Tab':
			closeMenu(false);
			return;
		default:
			return;
	}

	e.preventDefault();
}

function onDocumentPointerDown(e: PointerEvent) {
	const target = e.target as Node;
	if (menuOpen.value && !menu.value?.contains(target) && !button.value?.contains(target)) {
		closeMenu(false);
	}
}

onMounted(() => document.addEventListener('pointerdown', onDocumentPointerDown, true));
onBeforeUnmount(() => document.removeEventListener('pointerdown', onDocumentPointerDown, true));
</script>

<template>
	<div ref="overlay" class="absolute inset-0 pointer-events-none">
		<div
			v-if="message"
			class="absolute inset-0 flex items-center justify-center p-4 text-center text-sm opacity-70"
		>
			{{ message }}
		</div>

		<div
			v-if="callout"
			class="chart-callout absolute z-10 min-w-32 max-w-64 rounded-lg border border-M5 bg-M1 px-2.5 py-2 shadow-lg"
			:style="calloutStyle"
		>
			<div class="truncate text-[11px] opacity-70">{{ callout.title }}</div>
			<div v-for="(row, i) in callout.rows" :key="i" class="flex items-center gap-2 text-xs leading-5">
				<span class="chart-swatch shrink-0" :class="row.kind" :style="{ background: row.color }" />
				<span class="truncate opacity-70">{{ row.label }}</span>
				<span class="ml-auto pl-3 font-medium whitespace-nowrap">{{ row.value }}</span>
				<span v-if="row.detail" class="whitespace-nowrap opacity-70">{{ row.detail }}</span>
			</div>
		</div>

		<div class="sr-only" aria-live="polite">{{ announcement }}</div>

		<div v-if="downloadable" class="absolute top-0 right-0 pointer-events-auto" @pointerdown.stop @keydown.stop>
			<button
				ref="button"
				type="button"
				class="flex size-7 items-center justify-center rounded-md opacity-70 hover:bg-M7/25 hover:opacity-100 focus-visible:outline-2 focus-visible:outline-I0"
				:title="t('chart.download')"
				:aria-label="t('chart.download')"
				aria-haspopup="menu"
				:aria-expanded="menuOpen"
				@click="toggleMenu"
			>
				<svg viewBox="0 0 20 20" fill="none" class="size-4" aria-hidden="true">
					<path
						stroke="currentColor"
						stroke-width="1.75"
						stroke-linecap="round"
						stroke-linejoin="round"
						d="M10 3v9m0 0L6.5 8.5M10 12l3.5-3.5M4 13.5v1.5A1.5 1.5 0 0 0 5.5 16.5h9a1.5 1.5 0 0 0 1.5-1.5v-1.5"
					/>
				</svg>
			</button>
			<div
				v-if="menuOpen"
				ref="menu"
				role="menu"
				:aria-label="t('chart.download')"
				class="absolute top-8 right-0 z-20 min-w-40 rounded-lg border border-M5 bg-M1 py-1 text-sm shadow-lg"
				@keydown="onMenuKeyDown"
			>
				<button
					v-for="format in ['svg', 'png', 'csv'] as const"
					:key="format"
					type="button"
					role="menuitem"
					tabindex="-1"
					class="block w-full px-3 py-1.5 text-left whitespace-nowrap hover:bg-M7/25 focus-visible:bg-M7/25 focus-visible:outline-none"
					@click="save(format)"
				>
					{{ t(`chart.download_${format}`) }}
				</button>
			</div>
		</div>
	</div>
</template>

<style scoped>
@reference '@/assets/css/tailwind.css';

.chart-swatch {
	@apply size-2 rounded-full;

	&.square {
		@apply rounded-[2px];
	}

	&.line {
		@apply h-0.5 w-2.5 rounded-none;
	}
}
</style>
