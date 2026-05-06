<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->
<template>
	<div v-if="ui.pages?.value.length" :id="id" class="switcher" :class="classes">
		<div ref="togglesContainer" class="toggles-container">
			<div class="toggles">
				<template v-for="page in pages" :key="`switcher_toggle_${id}_${page.id}`">
					<button
						v-if="page.toggle"
						:ref="
							(el) => {
								if (ui.value === page.id) activeToggle = el as HTMLButtonElement;
							}
						"
						class="toggle group"
						:class="{ active: ui.value === page.id }"
						:title="page.title"
						@click="toggle(page.id || '')"
					>
						<UiGeneric :ui="page.toggle" class="opacity-50 group-hover:opacity-100 group-active:scale-90" />
					</button>
				</template>
				<span v-if="activeToggle" class="active-toggle-bg" :style="activeToggleBgStyles"></span>
			</div>
		</div>
		<div class="content-container">
			<div class="content" :style="`height: ${switcherHeight}px; transition-duration: ${transitionDuration}ms;`">
				<SwitcherPage
					v-for="page in pages"
					:key="`switcher_page_${id}_${page.id}`"
					ref="switcherPages"
					:page-id="page.id as string"
					:active-id="ui.value"
					:transition-duration="transitionDuration"
					:img="page.img"
					:img-object-fit="ui.imageObjectFit"
					:vertical="ui.orientation === OrientationValues.Vertical"
					@image-loaded="onImageLoad"
				>
					<UiGeneric :ui="page.content as Component" />
				</SwitcherPage>
			</div>
		</div>
	</div>
</template>
<script lang="ts" setup>
import { computed, onMounted, ref, watch } from 'vue';
import UiGeneric from '@/components/UiGeneric.vue';
import { frameCSS } from '@/components/shared/frame';
import { randomStr } from '@/components/shared/util';
import SwitcherPage from '@/components/switcher/SwitcherPage.vue';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import { nextRID } from '@/eventhandling';
import { CssClasses } from '@/shared/cssClasses';
import type { Component, SwitcherPage as NagoSwitcherPage, Switcher } from '@/shared/proto/nprotoc_gen';
import { OrientationValues, UpdateStateValueRequested } from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: Switcher;
}>();

const TRANSITION_DURATION = 400;

const id = props.ui.id || randomStr(16);
const serviceAdapter = useServiceAdapter();

const togglesContainer = ref<HTMLDivElement>();
const activeToggle = ref<HTMLButtonElement>();
const switcherPages = ref<InstanceType<typeof SwitcherPage>[]>([]);

const resizing = ref(false);
const resizingTimeout = ref();
const activeId = ref(props.ui.value);
const activeToggleBgStyles = ref('');
const loadedImages = ref<string[]>([]);
const switcherHeight = ref(0);

const pages = computed<NagoSwitcherPage[]>(() => props.ui.pages?.value.filter((p) => !!p.content && !!p.id) ?? []);

const transitionDuration = computed<number>(() => (resizing.value || !allImagesLoaded.value ? 0 : TRANSITION_DURATION));

const allImagesLoaded = computed<boolean>(() => {
	const uniqueImages: string[] = [];
	props.ui.pages?.value.forEach((p) => {
		if (p.img && !uniqueImages.includes(p.img)) uniqueImages.push(p.img);
	});
	return loadedImages.value.length === uniqueImages.length;
});

const classes = computed<string[]>(() => {
	const classes: string[] = [];
	const styles = frameCSS(props.ui.frame);
	classes.push(CssClasses.getOrCreate(styles));
	classes.push(props.ui.orientation === OrientationValues.Horizontal ? 'horizontal' : 'vertical');
	if (props.ui.dynamicHeight) classes.push('dynamic-height');
	return classes;
});

function onImageLoad(img: string) {
	if (!loadedImages.value.includes(img)) loadedImages.value.push(img);
	calcSwitcherHeight();
}

function toggle(value: string) {
	serviceAdapter.sendEvent(new UpdateStateValueRequested(props.ui.inputValue, 0, nextRID(), value));
}

function calcActiveToggleStyles() {
	const duration = `transition-duration: ${transitionDuration.value}ms;`;
	if (!activeToggle.value) return duration;

	const at = activeToggle.value;
	activeToggleBgStyles.value = `${duration} left: ${at.offsetLeft}px; top: ${at.offsetTop}px; width: ${at.offsetWidth}px; height: ${at.offsetHeight}px;`;
}

function onValueChange() {
	activeId.value = props.ui.value;
	calcSwitcherHeight();
	calcActiveToggleStyles();
}

function onWindowResize() {
	if (resizingTimeout.value) clearTimeout(resizingTimeout.value);
	resizing.value = true;
	resizingTimeout.value = setTimeout(() => (resizing.value = false), 50);
	calcSwitcherHeight();
	calcActiveToggleStyles();
}

function calcSwitcherHeight() {
	if (props.ui.dynamicHeight) {
		switcherHeight.value = calcMaxHeight(activeId.value);
	} else {
		switcherHeight.value = calcMaxHeight();
	}
}

function calcMaxHeight(pageId?: string, minHeight = 0): number {
	const minHeightIncToggles = Math.max(minHeight, togglesContainer.value?.getBoundingClientRect().height || 0);

	if (pageId) {
		const activePage = switcherPages.value.find((p) => p.$props.pageId === activeId.value);
		return activePage?.calcPageHeight(minHeightIncToggles) || 0;
	}

	let maxHeight = 0;
	switcherPages.value.forEach((p) => {
		const height = p.calcPageHeight(minHeightIncToggles);
		if (height > maxHeight) maxHeight = height;
	});

	if (!minHeight) return calcMaxHeight(undefined, maxHeight);
	return maxHeight;
}

onMounted(() => {
	calcSwitcherHeight();
	calcActiveToggleStyles();
});

watch(() => props.ui.value, onValueChange);
watch(activeToggle, calcActiveToggleStyles);
addEventListener('resize', onWindowResize);
</script>
<style scoped>
.switcher {
	@apply bg-M2 rounded-3xl overflow-hidden flex items-end gap-8;

	.toggles-container {
		@apply flex flex-col justify-end pl-8 py-8;

		.toggles {
			@apply relative flex flex-col gap-2 bg-M4 p-2 rounded-2xl;

			.toggle {
				@apply relative flex items-center justify-center size-16 rounded-xl z-10 duration-100 outline-offset-2 outline-I0;

				& > * {
					@apply duration-100;
				}

				&:not(.active) {
					@apply hover:bg-M3;
				}

				&.active > * {
					@apply opacity-100;
				}
			}

			.active-toggle-bg {
				@apply absolute outline outline-2 -outline-offset-2 outline-M0 bg-M3 rounded-xl z-0;
			}
		}
	}

	.content-container {
		@apply flex flex-col justify-end grow;

		.content {
			@apply relative w-full min-h-full;
		}
	}

	&.vertical {
		@apply flex-col items-start;

		.toggles-container {
			@apply pb-0 pr-8 flex-row justify-start;

			.toggles {
				@apply flex-row;
			}
		}

		.content-container {
			@apply w-full;
		}
	}
}
</style>
