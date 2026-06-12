<template>
	<div ref="splitView" class="split-view" :class="{ vertical: isVertical }" :style="splitViewStyles">
		<div v-if="ui.contentA" class="content a" :style="contentAStyles">
			<UiGeneric :ui="ui.contentA" />
		</div>
		<div
			v-if="ui.contentA && ui.contentB"
			ref="grabber"
			class="grabber"
			:class="{ dragging: isDragging }"
			:style="grabberStyles"
			@pointerdown.prevent="onGrabberPointerDown"
		></div>
		<div v-if="ui.contentB" class="content b" :style="contentBStyles">
			<UiGeneric :ui="ui.contentB" />
		</div>
	</div>
</template>
<script lang="ts" setup>
import { OrientationValues, SplitView, UpdateStateValueRequested } from '@/shared/proto/nprotoc_gen';
import UiGeneric from '@/components/UiGeneric.vue';
import { computed, onMounted, onUnmounted, ref, watch } from 'vue';
import { frameCSS } from '@/components/shared/frame';
import { nextRID } from '@/eventhandling';
import { useServiceAdapter } from '@/composables/serviceAdapter';

interface Props {
	ui: SplitView;
}

const props = defineProps<Props>();
const serviceAdapter = useServiceAdapter();

const splitView = ref<HTMLElement | null>(null);
const grabber = ref<HTMLElement | null>(null);
const ratio = ref(clampRatio(props.ui.value ?? 0));
const isDragging = ref(false);
const activePointerId = ref<number | null>(null);

const isVertical = computed<boolean>(() => {
	return props.ui.orientation === OrientationValues.Vertical;
});

const splitViewStyles = computed<string>(() => {
	return frameCSS(props.ui.frame).join(';');
});

const contentAStyles = computed<string>(() => {
	return isVertical.value ? `height: ${ratio.value * 100}%;` : `width: ${ratio.value * 100}%;`;
});

const contentBStyles = computed<string>(() => {
	return isVertical.value ? `height: ${(1 - ratio.value) * 100}%;` : `width: ${(1 - ratio.value) * 100}%;`;
});

const grabberStyles = computed<string>(() => {
	return isVertical.value ? `top: ${ratio.value * 100}%;` : `left: ${ratio.value * 100}%;`;
});

function clampRatio(value: number): number {
	const global = Math.min(1, Math.max(0, value));
	return Math.min(props.ui.maxRatio ?? 1, Math.max(props.ui.minRatio ?? 0, global));
}

function onGrabberPointerDown(event: PointerEvent): void {
	if (event.pointerType === 'mouse' && event.button !== 0) return;

	isDragging.value = true;
	activePointerId.value = event.pointerId;

	if (event.currentTarget instanceof HTMLElement) {
		event.currentTarget.setPointerCapture(event.pointerId);
	}

	updateRatioFromPointer(event);
}

function onDocumentPointerMove(event: PointerEvent): void {
	if (!isRelevantPointerEvent(event) || !isDragging.value) return;

	updateRatioFromPointer(event);
}

function onDocumentPointerUp(event: PointerEvent): void {
	if (!isRelevantPointerEvent(event) || !isDragging.value) return;

	updateRatioFromPointer(event);
	finishDrag(event);
}

function onDocumentPointerCancel(event: PointerEvent): void {
	if (!isRelevantPointerEvent(event) || !isDragging.value) return;

	finishDrag(event);
}

function isRelevantPointerEvent(event: PointerEvent): boolean {
	return activePointerId.value === event.pointerId;
}

function updateRatioFromPointer(event: PointerEvent): void {
	if (!splitView.value) return;

	const rect = splitView.value.getBoundingClientRect();
	const availableSpace = isVertical.value ? rect.height : rect.width;
	if (availableSpace <= 0) return;

	const nextRatio = isVertical.value
		? clampRatio((event.clientY - rect.top) / rect.height)
		: clampRatio((event.clientX - rect.left) / rect.width);

	if (Number.isNaN(nextRatio) || nextRatio === ratio.value) return;

	ratio.value = nextRatio;
}

function finishDrag(event: PointerEvent): void {
	if (grabber.value?.hasPointerCapture(event.pointerId)) {
		grabber.value.releasePointerCapture(event.pointerId);
	}

	isDragging.value = false;
	activePointerId.value = null;
	onGrabberDrop();
}

function onGrabberDrop() {
	emitNewRatio(ratio.value);
}

function emitNewRatio(ratio: number) {
	serviceAdapter.sendEvent(new UpdateStateValueRequested(props.ui.inputValue, 0, nextRID(), `${ratio}`));
}

onMounted(() => {
	document.addEventListener('pointermove', onDocumentPointerMove);
	document.addEventListener('pointerup', onDocumentPointerUp);
	document.addEventListener('pointercancel', onDocumentPointerCancel);
});

onUnmounted(() => {
	document.removeEventListener('pointermove', onDocumentPointerMove);
	document.removeEventListener('pointerup', onDocumentPointerUp);
	document.removeEventListener('pointercancel', onDocumentPointerCancel);
});

watch(
	() => props.ui.value,
	() => {
		if (!isDragging.value) {
			ratio.value = clampRatio(props.ui.value ?? 0);
		}
	}
);
</script>
<style scoped>
.split-view {
	@apply relative flex;

	.content {
		@apply max-h-full min-w-0 min-h-0;
	}

	.grabber {
		@apply absolute top-0 h-full w-4 -translate-x-1/2 cursor-ew-resize;
		touch-action: none;

		&:after {
			content: '';
			@apply absolute top-0 left-1/2 h-full w-0.5 -translate-x-1/2 bg-M7/50;
		}

		&:hover,
		&.dragging {
			&:after {
				@apply bg-I0;
			}
		}
	}

	&.vertical {
		@apply flex-col;

		.content {
			@apply max-w-full;
		}

		.grabber {
			@apply left-0 top-auto h-4 w-full translate-x-0 -translate-y-1/2 cursor-ns-resize;

			&:after {
				@apply top-1/2 left-0 h-0.5 w-full -translate-x-0 -translate-y-1/2;
			}
		}
	}
}
</style>
