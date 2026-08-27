<!--
 Copyright (c) 2026 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<template v-if="props.ui.iv">
	<!-- UiScrollView -->
	<div
		class="scroll-view"
		:class="{ horizontal: ui.axis === ScrollViewAxisValues.ScrollViewAxisHorizontal }"
		:style="styles"
	>
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
	ScrollAlignmentValues,
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
	if (scrolling.value) clearTimeout(scrolling.value);
	scrolling.value = setTimeout(() => {
		clearTimeout(scrolling.value);
		scrolling.value = false;
	}, 100);

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

	const alignStart = props.ui.scrollAlignment === ScrollAlignmentValues.ScrollAlignmentStart;

	if (!props.ui.axis || props.ui.axis === ScrollViewAxisValues.ScrollViewAxisVertical) {
		if (isAtScrollEnd()) {
			wantToScroll.value = false;
			return;
		}

		const containerEdge = alignStart
			? outer.value.getBoundingClientRect().top
			: outer.value.getBoundingClientRect().bottom;
		const scrollToEdge = alignStart
			? scrollTo.getBoundingClientRect().top
			: scrollTo.getBoundingClientRect().bottom;
		wantToScroll.value = scrollToEdge - containerEdge > SCROLL_TRESHOLD;
		return;
	}

	if (props.ui.axis === ScrollViewAxisValues.ScrollViewAxisHorizontal) {
		if (isAtScrollEnd()) {
			wantToScroll.value = false;
			return;
		}

		const containerEdge = alignStart
			? outer.value.getBoundingClientRect().left
			: outer.value.getBoundingClientRect().right;
		const scrollToEdge = alignStart
			? scrollTo.getBoundingClientRect().left
			: scrollTo.getBoundingClientRect().right;
		wantToScroll.value = scrollToEdge - containerEdge > SCROLL_TRESHOLD;
		return;
	}

	wantToScroll.value = false;
}

function isAtScrollEnd(): boolean {
	if (!outer.value) return false;

	if (!props.ui.axis || props.ui.axis === ScrollViewAxisValues.ScrollViewAxisVertical) {
		return outer.value.scrollHeight - (outer.value.scrollTop + outer.value.clientHeight) <= SCROLL_TRESHOLD;
	}

	if (props.ui.axis === ScrollViewAxisValues.ScrollViewAxisHorizontal) {
		return outer.value.scrollWidth - (outer.value.scrollLeft + outer.value.clientWidth) <= SCROLL_TRESHOLD;
	}

	return true;
}

function scroll(force?: boolean): void {
	if (!props.ui.scrollIntoView) return;

	const scrollNow =
		force ||
		!props.ui.scrollBehavior ||
		props.ui.scrollBehavior === ScrollBehaviorValues.ScrollBehaviorAlways ||
		(props.ui.scrollBehavior === ScrollBehaviorValues.ScrollBehaviorAuto && isAtScrollEnd());
	if (!scrollNow) return;

	nextTick(() => {
		if (!outer.value) return;

		let id = props.ui.scrollIntoView;
		const child = outer.value.querySelector(`#${id}`);
		if (!child) return;

		const isVertical = !props.ui.axis || props.ui.axis === ScrollViewAxisValues.ScrollViewAxisVertical;
		if (isVertical) {
			if (child.getBoundingClientRect().top < outer.value.getBoundingClientRect().top) return;
		} else {
			if (child.getBoundingClientRect().left < outer.value.getBoundingClientRect().left) return;
		}

		const scrollAlign = props.ui.scrollAlignment === ScrollAlignmentValues.ScrollAlignmentStart ? 'start' : 'end';

		switch (props.ui.scrollAnimation) {
			case ScrollAnimationValues.Instant:
				child.scrollIntoView({ block: scrollAlign });
				break;
			default:
				child.scrollIntoView({ behavior: 'smooth', block: scrollAlign });
		}
	});
}

function observeContent(): void {
	if (!outer.value) return;
	contentObserver.observe(outer.value, { childList: true, subtree: true });
}

function onContentChanged(): void {
	if (props.ui.listLength) return;

	contentChanged.value = true;
	scroll();
}

function onListLengthChanged(): void {
	if (!props.ui.listLength) return;

	contentChanged.value = true;
	scroll();
}

watch(() => props.ui.content, onContentChanged, { deep: true });
watch(() => props.ui.listLength, onListLengthChanged);

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
	@apply relative flex min-h-0 max-h-full overflow-clip;

	.container-outer {
		@apply flex-1 min-h-0 max-h-full;
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
