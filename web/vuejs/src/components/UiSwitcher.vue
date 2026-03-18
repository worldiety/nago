<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<script lang="ts" setup>
import { computed, onMounted, ref, watch } from 'vue';
import UiGeneric from '@/components/UiGeneric.vue';
import { frameCSS } from '@/components/shared/frame';
import { randomStr } from '@/components/shared/util';
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
const contentHeight = ref(0);
const activeToggleBgStyles = ref('');

const transitionDuration = computed<number>(() => (initializing.value || resizing.value ? 0 : TRANSITION_DURATION));

const classes = computed<string[]>(() => {
	const classes: string[] = [];
	const styles = frameCSS(props.ui.frame);
	classes.push(CssClasses.getOrCreate(styles));
	classes.push(props.ui.orientation === OrientationValues.Horizontal ? 'horizontal' : 'vertical');
	if (props.ui.dynamicHeight) classes.push('dynamic-height');
	return classes;
});

function toggle(value: string) {
	serviceAdapter.sendEvent(new UpdateStateValueRequested(props.ui.inputValue, 0, nextRID(), value));
}

function calcContentHeight(pageId?: string) {
	if (!switcher.value) return 0;

	if (props.ui.dynamicHeight) {
		let height = 0;

		const pageContentInner = switcher.value.querySelector(
			`.page[data-page-id='${pageId || activeId.value || previousId.value}'] .page-content-inner`
		);
		height += pageContentInner?.getBoundingClientRect().height || 0;

		if (props.ui.orientation === OrientationValues.Vertical) {
			const image = switcher.value.querySelector(
				`.page[data-page-id='${pageId || activeId.value || previousId.value}'] .image`
			);
			height += image?.getBoundingClientRect().height || 0;
		}

		contentHeight.value = height;
	} else {
		const pages = switcher.value.querySelectorAll('.page');
		let maxHeight = 0;
		pages.forEach((p) => {
			let height = 0;

			const inner = p.querySelector('.page-content-inner');
			height += inner?.getBoundingClientRect().height || 0;

			if (props.ui.orientation === OrientationValues.Vertical) {
				const image = p.querySelector('.image') as HTMLDivElement;
				if (image) {
					image.style.height = 'auto';
					height += image?.getBoundingClientRect().height || 0;
					image.style.height = '';
				}
			}
			maxHeight = Math.max(maxHeight, height);
		});
		contentHeight.value = maxHeight;
	}
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
	calcContentHeight(props.ui.value);
	calcActiveToggleStyles();

	setTimeout(() => {
		previousId.value = undefined;
		activeId.value = props.ui.value;
		nextId.value = undefined;
	}, transitionDuration.value / 2);
}

function observePageSizes() {
	if (!switcher.value) return;

	const pages = switcher.value.querySelectorAll('.page-content-inner');
	const observer = new ResizeObserver(() => calcContentHeight());
	pages.forEach((p) => observer.observe(p));
}

function onWindowResize() {
	if (resizingTimeout.value) clearTimeout(resizingTimeout.value);
	resizing.value = true;
	resizingTimeout.value = setTimeout(() => (resizing.value = false), 50);
	calcActiveToggleStyles();
}

onMounted(() => {
	observePageSizes();
	calcActiveToggleStyles();
	setTimeout(() => {
		initializing.value = false;
	}, 50); // TODO: Find better way to determine completed initial rendering
});

watch(() => props.ui.value, onValueChange);
watch(activeToggle, calcActiveToggleStyles);
addEventListener('resize', onWindowResize);
</script>

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
					<div
						v-if="page.content"
						class="page"
						:class="{
							'active': activeId === page.id,
							'has-image': !!page.img,
						}"
						:data-page-id="page.id"
						:style="`transition-duration: ${transitionDuration / 2}ms;`"
						:inert="activeId !== page.id"
						:aria-hidden="activeId !== page.id"
					>
						<div class="page-content">
							<div class="page-content-inner">
								<UiGeneric :ui="page.content" />
							</div>
						</div>
						<div v-if="page.img" class="image-container">
							<div class="image" :style="`background-image: url(${page.img});`"></div>
						</div>
					</div>
				</template>
			</div>
		</div>
	</div>
</template>

<style scoped>
.switcher {
	@apply bg-M2 rounded-3xl overflow-hidden flex gap-8;

	.toggles-container {
		@apply flex flex-col justify-end pl-8 py-8;

		.toggles {
			@apply relative flex flex-col gap-2 bg-M4 p-2 rounded-2xl;

			.toggle {
				@apply relative flex items-center justify-center size-16 rounded-xl z-10 duration-100;
				@apply focus:outline-offset-2 focus-visible:outline-offset-2;

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

			.page {
				@apply absolute left-0 bottom-0 size-full pr-8 grid grid-cols-1 gap-8 opacity-0 pointer-events-none;

				.page-content {
					@apply flex flex-col justify-end;

					.page-content-inner {
						@apply py-8 w-full;
					}
				}

				&.has-image {
					@apply md:grid-cols-2 pr-0;
				}

				&.active {
					@apply opacity-100 pointer-events-auto;
				}
			}

			.image-container {
				@apply relative;

				.image {
					@apply absolute left-0 top-0 size-full bg-center overflow-hidden bg-cover;
				}
			}
		}
	}

	&.vertical {
		@apply flex-col;

		.toggles-container {
			@apply pb-0 pr-8 flex-row justify-start;

			.toggles {
				@apply flex-row;
			}
		}

		.content-container {
			.content {
				.page {
					@apply pr-0 gap-0 flex flex-col;

					.page-content {
						.page-content-inner {
							@apply px-8 pt-0 pb-8;
						}
					}

					.image-container {
						@apply grow w-full;

						.image {
							@apply min-h-64;
						}
					}
				}
			}
		}
	}

	&.dynamic-height.vertical {
		.content-container {
			.content {
				.page {
					.image-container {
						.image {
							@apply h-64;
						}
					}
				}
			}
		}
	}
}
</style>
