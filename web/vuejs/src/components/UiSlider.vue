<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<template>
	<div :style="frameStyles">
		<InputWrapper
			:wrapper-style="InputWrapperStyle.BASIC"
			:label="ui.label"
			:label-id="labelId"
			:error="ui.errorText"
			:help="ui.supportingText"
			:disabled="ui.disabled"
		>
			<div
				class="slider"
				:class="{ 'disabled': ui.disabled, 'error': ui.errorText, 'range-mode': isRangeMode }"
				:role="isRangeMode ? 'group' : undefined"
				:aria-labelledby="isRangeMode && ui.label ? labelId : undefined"
			>
				<template v-if="isRangeMode">
					<span :id="fromLabelId" hidden>{{ $t('slider.from') }}</span>
					<span :id="toLabelId" hidden>{{ $t('slider.to') }}</span>
				</template>

				<div ref="bar" class="bar">
					<template v-if="ui.showMarkers">
						<div
							v-for="i in stepsBetween"
							:key="`slider_step_${i}`"
							class="step"
							:class="{ active: isStepActive(i) }"
							:style="`left: ${i * stepWidth * 100}%;`"
							aria-hidden="true"
						></div>
					</template>

					<div class="bar-left" :style="barLeftStyles" aria-hidden="true"></div>
					<div v-if="isRangeMode" class="bar-middle" :style="barMiddleStyles" aria-hidden="true"></div>
					<div class="bar-right" :style="barRightStyles" aria-hidden="true"></div>

					<!-- Enlarged hit area: clicking the rail moves the nearest thumb there and focuses it -->
					<div class="rail" aria-hidden="true" @pointerdown.prevent="onRailPointerDown"></div>

					<!-- From grabber (single grabber in non-range mode, left grabber in range mode) -->
					<div
						ref="grabberFrom"
						class="grabber custom-focus"
						:style="grabberFromStyles"
						:class="{ dragging: draggingGrabber === 'from' }"
						role="slider"
						:aria-labelledby="grabberFromLabelledBy"
						:aria-description="ui.errorText || ui.supportingText"
						:aria-valuemin="boundsFrom.min"
						:aria-valuemax="boundsFrom.max"
						:aria-valuenow="valueFrom"
						:aria-valuetext="formatValue(valueFrom)"
						:aria-disabled="ui.disabled"
						:aria-invalid="!!ui.errorText"
						:tabindex="ui.disabled ? -1 : 0"
						@pointerdown.prevent="onGrabberPointerDown($event, 'from')"
						@keydown="onKeyDown($event, 'from')"
					>
						<span ref="grabberDot" class="dot" aria-hidden="true"></span>
						<span class="value" aria-hidden="true">{{ formatValue(valueFrom) }}</span>
					</div>

					<!-- To grabber (only in range mode) -->
					<div
						v-if="isRangeMode"
						ref="grabberTo"
						class="grabber custom-focus"
						:style="grabberToStyles"
						:class="{ dragging: draggingGrabber === 'to' }"
						role="slider"
						:aria-labelledby="grabberToLabelledBy"
						:aria-description="ui.errorText || ui.supportingText"
						:aria-valuemin="boundsTo.min"
						:aria-valuemax="boundsTo.max"
						:aria-valuenow="valueTo"
						:aria-valuetext="formatValue(valueTo)"
						:aria-disabled="ui.disabled"
						:aria-invalid="!!ui.errorText"
						:tabindex="ui.disabled ? -1 : 0"
						@pointerdown.prevent="onGrabberPointerDown($event, 'to')"
						@keydown="onKeyDown($event, 'to')"
					>
						<span class="dot" aria-hidden="true"></span>
						<span class="value" aria-hidden="true">{{ formatValue(valueTo) }}</span>
					</div>
				</div>
			</div>
		</InputWrapper>
	</div>
</template>
<script lang="ts" setup>
import { computed, ref, onUnmounted, watch } from 'vue';
import InputWrapper from '@/components/shared/InputWrapper.vue';
import { frameCSS } from '@/components/shared/frame';
import { InputWrapperStyle } from '@/components/shared/inputWrapperStyle';
import { randomStr } from '@/components/shared/util';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import { nextRID } from '@/eventhandling';
import { Slider, UpdateStateValueRequested } from '@/shared/proto/nprotoc_gen';

type Grabber = 'from' | 'to';

// Number of steps for PageUp/PageDown, as in the WAI-ARIA APG slider examples.
const PAGE_STEPS = 10;

const props = defineProps<{
	ui: Slider;
}>();

const serviceAdapter = useServiceAdapter();

const idPrefix = `slider-${randomStr(16)}`;
const labelId = `${idPrefix}-label`;
const fromLabelId = `${idPrefix}-from`;
const toLabelId = `${idPrefix}-to`;

const bar = ref<HTMLElement>();
const grabberDot = ref<HTMLElement>();
const grabberFrom = ref<HTMLElement>();
const grabberTo = ref<HTMLElement>();

const step = ref(props.ui.step ?? 1);
const valueFrom = ref(props.ui.value?.from ?? 0);
const valueTo = ref(props.ui.value?.to ?? 0);

const isRangeMode = computed<boolean>(() => !!props.ui.rangeMode);

const draggingGrabber = ref<Grabber | null>(null);
let dragStartValue: number | null = null;

const stepsBetween = computed<number>(() => {
	const min = props.ui.min ?? 0;
	const max = props.ui.max ?? 0;
	return Math.floor((max - min) / step.value) - 1;
});

const stepWidth = computed<number>(() => {
	const min = props.ui.min ?? 0;
	const max = props.ui.max ?? 0;
	return step.value / (max - min);
});

// In range mode each thumb can only move up to its sibling, so the announced bounds follow the other thumb.
const boundsFrom = computed<{ min: number; max: number }>(() => ({
	min: props.ui.min ?? 0,
	max: isRangeMode.value ? fixDecimals(valueTo.value - step.value) : (props.ui.max ?? 0),
}));

const boundsTo = computed<{ min: number; max: number }>(() => ({
	min: fixDecimals(valueFrom.value + step.value),
	max: props.ui.max ?? 0,
}));

const grabberFromLabelledBy = computed<string | undefined>(() => {
	if (!isRangeMode.value) return props.ui.label ? labelId : undefined;
	return props.ui.label ? `${labelId} ${fromLabelId}` : fromLabelId;
});

const grabberToLabelledBy = computed<string>(() => {
	return props.ui.label ? `${labelId} ${toLabelId}` : toLabelId;
});

const frameStyles = computed<string>(() => {
	const styles = frameCSS(props.ui.frame);
	return styles.join(';');
});

const valueFromRatio = computed<number>(() => {
	const min = props.ui.min ?? 0;
	const max = props.ui.max ?? 0;
	return (valueFrom.value - min) / (max - min);
});

const valueToRatio = computed<number>(() => {
	const min = props.ui.min ?? 0;
	const max = props.ui.max ?? 0;
	return (valueTo.value - min) / (max - min);
});

const grabberHalfWidth = computed<number>(() => {
	if (!grabberDot.value) return 0;
	return grabberDot.value.clientWidth / 2;
});

const grabberFromStyles = computed<string>(() => {
	return `left: ${valueFromRatio.value * 100}%;`;
});

const grabberToStyles = computed<string>(() => {
	return `left: ${valueToRatio.value * 100}%;`;
});

const barLeftStyles = computed<string>(() => {
	return `width: calc(${valueFromRatio.value * 100}% - ${grabberHalfWidth.value}px);`;
});

const barMiddleStyles = computed<string>(() => {
	return `left: calc(${valueFromRatio.value * 100}% + ${grabberHalfWidth.value}px); right: calc(${(1 - valueToRatio.value) * 100}% + ${grabberHalfWidth.value}px);`;
});

const barRightStyles = computed<string>(() => {
	const ratio = isRangeMode.value ? valueToRatio.value : valueFromRatio.value;
	return `width: calc(${(1 - ratio) * 100}% - ${grabberHalfWidth.value}px);`;
});

function isStepActive(i: number): boolean {
	const pos = (props.ui.min ?? 0) + step.value * i;
	if (isRangeMode.value) {
		return pos >= valueFrom.value && pos <= valueTo.value;
	}
	return valueFrom.value >= pos;
}

function formatValue(value: number): string {
	return `${value}${props.ui.unit ?? ''}`;
}

function submitValue() {
	if (isRangeMode.value) {
		const json = JSON.stringify({ From: valueFrom.value, To: valueTo.value });
		serviceAdapter.sendEvent(new UpdateStateValueRequested(props.ui.inputValue, 0, nextRID(), json));
	} else {
		serviceAdapter.sendEvent(
			new UpdateStateValueRequested(props.ui.inputValue, 0, nextRID(), `${valueFrom.value}`)
		);
	}
}

function getDecimalPlacesByStepSize(): number {
	const split = `${step.value}`.split('.');
	if (split.length === 1) return 0;
	return split[1].length;
}

function fixDecimals(num: number): number {
	return parseFloat(num.toFixed(getDecimalPlacesByStepSize()));
}

function snapToStep(clientX: number): number {
	const rect = bar.value!.getBoundingClientRect();
	const ratio = Math.max(0, Math.min(1, (clientX - rect.left) / rect.width));
	const min = props.ui.min ?? 0;
	const max = props.ui.max ?? 0;
	const totalSteps = (max - min) / step.value;
	const snappedStep = Math.round(ratio * totalSteps);
	return fixDecimals(min + snappedStep * step.value);
}

function getValue(grabber: Grabber): number {
	return grabber === 'from' ? valueFrom.value : valueTo.value;
}

function getBounds(grabber: Grabber): { min: number; max: number } {
	return grabber === 'from' ? boundsFrom.value : boundsTo.value;
}

function getGrabberElement(grabber: Grabber): HTMLElement | undefined {
	return grabber === 'from' ? grabberFrom.value : grabberTo.value;
}

// Clamps the value to the grabber's bounds and applies it locally. Returns whether the value changed.
function moveGrabberTo(grabber: Grabber, value: number): boolean {
	const bounds = getBounds(grabber);
	const clamped = fixDecimals(Math.min(bounds.max, Math.max(bounds.min, value)));
	if (clamped === getValue(grabber)) return false;
	if (grabber === 'from') {
		valueFrom.value = clamped;
	} else {
		valueTo.value = clamped;
	}
	return true;
}

function onKeyDown(event: KeyboardEvent, grabber: Grabber) {
	if (props.ui.disabled) return;
	const current = getValue(grabber);
	const bounds = getBounds(grabber);

	let next: number;
	switch (event.key) {
		case 'ArrowLeft':
		case 'ArrowDown':
			next = current - step.value;
			break;
		case 'ArrowRight':
		case 'ArrowUp':
			next = current + step.value;
			break;
		case 'PageDown':
			next = current - PAGE_STEPS * step.value;
			break;
		case 'PageUp':
			next = current + PAGE_STEPS * step.value;
			break;
		case 'Home':
			next = bounds.min;
			break;
		case 'End':
			next = bounds.max;
			break;
		default:
			return;
	}

	event.preventDefault();
	event.stopPropagation();
	if (moveGrabberTo(grabber, next)) {
		submitValue();
	}
}

// --- Pointer handling ---
function onPointerMove(event: PointerEvent) {
	if (!draggingGrabber.value) return;
	moveGrabberTo(draggingGrabber.value, snapToStep(event.clientX));
}

function onPointerUp() {
	if (draggingGrabber.value && dragStartValue !== getValue(draggingGrabber.value)) {
		submitValue();
	}
	draggingGrabber.value = null;
	dragStartValue = null;
	document.removeEventListener('pointermove', onPointerMove);
	document.removeEventListener('pointerup', onPointerUp);
}

function startDragging(event: PointerEvent, grabber: Grabber, startValue: number) {
	draggingGrabber.value = grabber;
	dragStartValue = startValue;
	getGrabberElement(grabber)?.focus();
	(event.currentTarget as HTMLElement).setPointerCapture(event.pointerId);
	document.addEventListener('pointermove', onPointerMove);
	document.addEventListener('pointerup', onPointerUp);
}

function onGrabberPointerDown(event: PointerEvent, grabber: Grabber) {
	if (props.ui.disabled) return;
	startDragging(event, grabber, getValue(grabber));
}

function onRailPointerDown(event: PointerEvent) {
	if (props.ui.disabled) return;
	const target = snapToStep(event.clientX);
	let grabber: Grabber = 'from';
	if (isRangeMode.value && Math.abs(target - valueTo.value) < Math.abs(target - valueFrom.value)) {
		grabber = 'to';
	}
	const startValue = getValue(grabber);
	moveGrabberTo(grabber, target);
	startDragging(event, grabber, startValue);
}

onUnmounted(() => {
	document.removeEventListener('pointermove', onPointerMove);
	document.removeEventListener('pointerup', onPointerUp);
});

watch(
	() => props.ui.value,
	(nextValue) => {
		valueFrom.value = nextValue?.from ?? 0;
		valueTo.value = nextValue?.to ?? 0;
	}
);
</script>

<style scoped>
@reference '@/assets/css/tailwind.css';
.slider {
	@apply pt-8 pb-4 min-w-32;

	.bar {
		@apply relative w-full h-px;

		.bar-left {
			@apply absolute left-0 top-0 h-full bg-I0;
		}

		.bar-middle {
			@apply absolute top-0 h-full bg-I0;
		}

		.bar-right {
			@apply absolute right-0 top-0 h-full bg-current;
		}

		.step {
			@apply absolute top-1/2 -translate-y-1/2 -translate-x-1/2 w-px h-1 bg-current z-0;

			&.active {
				@apply bg-I0;
			}
		}

		.rail {
			@apply absolute inset-x-0 top-1/2 -translate-y-1/2 h-4 cursor-pointer z-1 touch-none;
		}

		.grabber {
			@apply absolute top-1/2 -translate-x-1/2 -translate-y-1/2 size-8 rounded-full flex justify-center items-center cursor-grab z-10 touch-none outline-none;

			.dot {
				@apply block size-3.5 rounded-full bg-I0;
			}

			.value {
				@apply absolute left-1/2 bottom-full -translate-x-1/2 text-I0 text-xs pb-px whitespace-nowrap;
			}

			&:focus,
			&:hover,
			&.dragging {
				@apply bg-I0/20;
			}

			&:focus-visible {
				@apply outline-2 outline-solid -outline-offset-2 outline-I0;
			}

			&.dragging {
				@apply cursor-grabbing;
			}
		}
	}

	&.range-mode {
		.bar {
			.bar-left {
				@apply bg-current;
			}
		}
	}

	&.error {
		.bar {
			.bar-left {
				@apply bg-SE0;
			}

			.bar-middle {
				@apply bg-SE0;
			}

			.grabber {
				.dot {
					@apply bg-SE0;
				}

				.value {
					@apply text-SE0;
				}

				&:focus,
				&:hover,
				&.dragging {
					@apply bg-SE0/20;
				}

				&:focus-visible {
					@apply outline-SE0;
				}
			}
		}
	}

	&.disabled {
		@apply text-SI0;

		.bar {
			.bar-left {
				@apply bg-ST0;
			}

			.bar-middle {
				@apply bg-ST0;
			}

			.step {
				@apply bg-SI0;

				&.active {
					@apply bg-ST0;
				}
			}

			.grabber {
				.dot {
					@apply bg-ST0;
				}

				.value {
					@apply text-ST0;
				}

				&:focus,
				&:hover,
				&.dragging {
					@apply bg-ST0/20;
				}

				&:focus-visible {
					@apply outline-ST0;
				}
			}

			.grabber,
			.rail {
				@apply cursor-default;
			}
		}
	}
}
</style>
