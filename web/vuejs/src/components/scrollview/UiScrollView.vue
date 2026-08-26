<!--
 Copyright (c) 2026 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<template v-if="props.ui.iv">
	<!-- UiScrollView -->
	<div class="scroll-view" :class="{ horizontal: ui.axis === ScrollViewAxisValues.ScrollViewAxisHorizontal }" :style="styles">
		<div ref="outer" class="container-outer" :class="classes" @scroll="onScroll">
			<div ref="inner" :style="innerStyles">
				<UiGeneric v-if="ui.content" :ui="ui.content" />
			</div>
		</div>
		<div v-if="askForScroll && wantToScroll && !scrolling && contentChanged" class="scroll-action">
			<button class="button-secondary" @click="scroll(true)">
				<IconArrowDown />
				<span>
					{{ ui.scrollButtonLabel }}
				</span>
				<IconArrowDown />
			</button>
		</div>
	</div>
</template>
<script lang="ts" setup>
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue';
import UiGeneric from '@/components/UiGeneric.vue';
import { borderCSS } from '@/components/shared/border';
import { colorValue } from '@/components/shared/colors';
import { frameCSS } from '@/components/shared/frame';
import { paddingCSS } from '@/components/shared/padding';
import { positionCSS } from '@/components/shared/position';
import {
	ScrollAnimationValues,
	ScrollBehaviorValues,
	ScrollView,
	ScrollViewAxisValues,
} from '@/shared/proto/nprotoc_gen';
import IconArrowDown from '@/assets/svg/arrow-down.svg';

const props = defineProps<{
	ui: ScrollView;
}>();

const SCROLL_TRESHOLD = 8;

const outer = ref<HTMLDivElement>();
const inner = ref<HTMLDivElement>();

const contentChanged = ref(false);
const contentObserver = new MutationObserver(calcWantToScroll);
const wantToScroll = ref(false);
const askForScroll = ref(props.ui.scrollBehavior !== ScrollBehaviorValues.ScrollBehaviorAlways);
const scrolling = ref<any>();

const styles = computed<string>(() => {
	let styles = borderCSS(props.ui.border);
	styles.push(...frameCSS(props.ui.frame));
	if (props.ui.backgroundColor) {
		styles.push(`background-color: ${colorValue(props.ui.backgroundColor)}`);
	}

	styles.push(...positionCSS(props.ui.position));
	styles.push(...borderCSS(props.ui.border));
	styles.push(...paddingCSS(props.ui.padding));

	return styles.join(';');
});

const classes = computed<string>(() => {
	const css: string[] = [];

	// note, that we defined its style in scrollbars.css
	switch (props.ui.axis) {
		case ScrollViewAxisValues.ScrollViewAxisHorizontal:
			css.push('overflow-x-auto', 'overflow-y-hidden');
			break;
		case ScrollViewAxisValues.ScrollViewAxisBoth:
			css.push('overflow-x-auto', 'overflow-y-auto');
			break;
		default:
			css.push('overflow-y-auto', 'overflow-x-hidden');
			break;
	}

	return css.join(' ');
});

const innerStyles = computed<string>(() => {
	const css: string[] = []; //borderCSS(props.ui.border);

	switch (props.ui.axis) {
		case ScrollViewAxisValues.ScrollViewAxisHorizontal:
			css.push('min-width: max-content');
			break;
		default:
			css.push('height: max-content');
			break;
	}

	return css.join(';');
});

function onScroll(): void {
	calcWantToScroll();
	if (!wantToScroll.value) contentChanged.value = false;
}

function calcWantToScroll(): void {
	if (!props.ui.scrollIntoView || !outer.value) {
		wantToScroll.value = false;
		return;
	}

	const scrollTo = outer.value.querySelector(`#${props.ui.scrollIntoView}`);
	if (!scrollTo) {
		wantToScroll.value = false;
		return;
	}

	if (!props.ui.axis || props.ui.axis === ScrollViewAxisValues.ScrollViewAxisVertical) {
		const containerBottom = outer.value.getBoundingClientRect().bottom;
		const scrollToBottom = scrollTo.getBoundingClientRect().bottom;
		wantToScroll.value = scrollToBottom - containerBottom > SCROLL_TRESHOLD;
		return;
	}

	if (props.ui.axis === ScrollViewAxisValues.ScrollViewAxisHorizontal) {
		const containerRight = outer.value.getBoundingClientRect().right;
		const scrollToRight = scrollTo.getBoundingClientRect().right;
		wantToScroll.value = scrollToRight - containerRight > SCROLL_TRESHOLD;
		return;
	}

	wantToScroll.value = false;
}

function scroll(force?: boolean): void {
	if (!props.ui.scrollIntoView || !outer.value) return;

	const scrollNow =
		force ||
		!props.ui.scrollBehavior ||
		props.ui.scrollBehavior === ScrollBehaviorValues.ScrollBehaviorAlways ||
		(props.ui.scrollBehavior === ScrollBehaviorValues.ScrollBehaviorAuto && !wantToScroll.value);
	if (!scrollNow) return;

	let id = props.ui.scrollIntoView;
	const child = outer.value.querySelector(`#${id}`);
	if (!child) return;

	nextTick(() => {
		switch (props.ui.scrollAnimation) {
			case ScrollAnimationValues.Instant:
				child?.scrollIntoView({});
				break;
			default:
				if (scrolling.value) clearTimeout(scrolling.value);
				child?.scrollIntoView({ behavior: 'smooth', block: 'end' });
				scrolling.value = setTimeout(() => {
					clearTimeout(scrolling.value);
					scrolling.value = false;
				}, 500);
		}
	});
}

function observeContent(): void {
	if (!outer.value) return;
	contentObserver.observe(outer.value, { childList: true, subtree: true });
}

function onContentChanged(): void {
	contentChanged.value = true;
	scroll();
}

watch(() => props.ui.content, onContentChanged, { deep: true });

onMounted(() => {
	calcWantToScroll();
	observeContent();
});

onUnmounted(() => {
	contentObserver.disconnect();
});

// note that we need the max-content hack, otherwise we get layout bugs at least for horizontal areas
</script>
<style scoped>
.scroll-view {
	@apply relative;

	.container-outer {
		@apply h-full;
	}

	.scroll-action {
		@apply absolute left-0 bottom-0 w-full px-8 pb-3;

		button {
			@apply w-full bg-I0/15 flex justify-center items-center gap-4;

			svg {
				@apply size-3;
			}
		}
	}

	&.horizontal {
		.scroll-action {
			@apply w-max h-fit left-auto right-0 top-1/2 translate-x-1/2 -translate-y-full origin-bottom -rotate-90;
		}
	}
}
</style>
