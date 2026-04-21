<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->
<template>
	<div v-if="ui.pages?.value.length" :id="id" ref="switcher" class="switcher" :class="classes">
		<div class="toggles-container">
			<div class="toggles">
				<template v-for="page in ui.pages.value" :key="`switcher_toggle_${id}_${page.id}`">
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
			<div class="content" :style="`height: ${contentHeight}px; transition-duration: ${transitionDuration}ms;`">
				<template v-for="page in ui.pages.value" :key="`switcher_page_${id}_${page.id}`">
					<SwitcherPage
						v-if="page.content && page.id"
						:page-id="page.id"
						:active-id="ui.value"
						:emit-height="!ui.dynamicHeight || page.id === ui.value"
						:transition-duration="transitionDuration"
						:img="page.img"
						:img-object-fit="ui.imageObjectFit"
						:vertical="ui.orientation === OrientationValues.Vertical"
						:fixed-height="
							!ui.dynamicHeight && ui.orientation !== OrientationValues.Vertical
								? contentHeight
								: undefined
						"
						@update:height="onUpdateHeight(page.id, $event)"
						@image-loaded="onImageLoad"
					>
						<UiGeneric :ui="page.content" />
					</SwitcherPage>
				</template>
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
import type { Switcher } from '@/shared/proto/nprotoc_gen';
import { OrientationValues, UpdateStateValueRequested } from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: Switcher;
}>();

const TRANSITION_DURATION = 400;

const id = props.ui.id || randomStr(16);
const serviceAdapter = useServiceAdapter();

const switcher = ref<HTMLDivElement>();
const activeToggle = ref<HTMLButtonElement>();

const initializing = ref(true);
const resizing = ref(false);
const resizingTimeout = ref();
const previousId = ref<string>();
const activeId = ref(props.ui.value);
const nextId = ref<string>();
const contentHeights = ref<Map<string, number>>(new Map());
const activeToggleBgStyles = ref('');
const loadedImages = ref<string[]>([]);

const transitionDuration = computed<number>(() =>
	initializing.value || resizing.value || !allImagesLoaded.value ? 0 : TRANSITION_DURATION
);

const contentHeight = computed<number>(() => {
	if (props.ui.dynamicHeight) {
		return contentHeights.value.get(props.ui.value || '') || 0;
	}

	let maxHeight = 0;
	contentHeights.value.forEach((height) => {
		if (height > maxHeight) maxHeight = height;
	});
	return maxHeight;
});

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

function onUpdateHeight(pageId: string, height: number) {
	contentHeights.value.set(pageId, height);
}

function onImageLoad(img: string) {
	setTimeout(() => {
		if (!loadedImages.value.includes(img)) loadedImages.value.push(img);
	}, 50); // TODO: Find better way to handle dimension changes by loaded images
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
	previousId.value = activeId.value;
	activeId.value = undefined;
	nextId.value = props.ui.value;
	calcActiveToggleStyles();

	setTimeout(() => {
		previousId.value = undefined;
		activeId.value = props.ui.value;
		nextId.value = undefined;
	}, transitionDuration.value / 2);
}

function onWindowResize() {
	if (resizingTimeout.value) clearTimeout(resizingTimeout.value);
	resizing.value = true;
	resizingTimeout.value = setTimeout(() => (resizing.value = false), 50);
	calcActiveToggleStyles();
}

onMounted(() => {
	calcActiveToggleStyles();
	setTimeout(() => {
		initializing.value = false;
	}, 50); // TODO: Find better way to determine completed initial rendering
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
