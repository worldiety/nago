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
			:error="ui.errorText"
			:help="ui.supportingText"
			:disabled="ui.disabled"
		>
			<div
				class="slider"
				:class="{ disabled: ui.disabled, error: ui.errorText }"
				:aria-valuemin="(ui.min ?? 0) + (ui.unit ?? '')"
				:aria-valuemax="(ui.max ?? 0) + (ui.unit ?? '')"
				:aria-valuenow="value + (ui.unit ?? '')"
				role="slider"
			>
				<div ref="bar" class="bar">
					<template v-if="ui.showMarkers">
						<div
							v-for="i in stepsBetween"
							:key="`slider_step_${i}`"
							class="step"
							:class="{ active: value >= (ui.min ?? 0) + step * i }"
							:style="`left: ${i * stepWidth * 100}%;`"
						></div>
					</template>
					<div class="bar-left" :style="barLeftStyles"></div>
					<div
						class="grabber"
						:style="grabberStyles"
						:class="{ dragging: isDragging }"
						:tabindex="ui.disabled ? -1 : 0"
						@pointerdown.prevent="onPointerDown"
						@keydown.left="prevStep"
						@keydown.right="nextStep"
					>
						<span ref="grabber" class="dot"></span>
						<span class="value"> {{ value }}{{ ui.unit }} </span>
					</div>
					<div class="bar-right" :style="barRightStyles"></div>
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
import { useServiceAdapter } from '@/composables/serviceAdapter';
import { nextRID } from '@/eventhandling';
import { Slider, UpdateStateValueRequested } from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: Slider;
}>();

const serviceAdapter = useServiceAdapter();

const bar = ref<HTMLElement>();
const grabber = ref<HTMLElement>();

const step = ref(props.ui.step ?? 1);
const value = ref(props.ui.value ?? 0);

const isDragging = ref(false);
let lastSnappedValue: number | null = null;

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

const frameStyles = computed<string>(() => {
	const styles = frameCSS(props.ui.frame);

	return styles.join(';');
});

const grabberStyles = computed<string>(() => {
	return `left: ${valueRatio.value * 100}%;`;
});

const valueRatio = computed<number>(() => {
	const min = props.ui.min ?? 0;
	const max = props.ui.max ?? 0;
	return (value.value - min) / (max - min);
});

const grabberWidth = computed<number>(() => {
	if (!grabber.value) return 0;
	return grabber.value.clientWidth;
});

const barLeftStyles = computed<string>(() => {
	return `width: calc(${valueRatio.value * 100}% - ${grabberWidth.value / 2}px);`;
});

const barRightStyles = computed<string>(() => {
	return `width: calc(${(1 - valueRatio.value) * 100}% - ${grabberWidth.value / 2}px);`;
});

function submitValue(value: number) {
	serviceAdapter.sendEvent(new UpdateStateValueRequested(props.ui.inputValue, 0, nextRID(), `${value}`));
}

function prevStep() {
	if (props.ui.disabled || value.value === props.ui.min) return;

	const startVal = value.value ?? 0;
	let prevVal = props.ui.min ?? 0;
	while (prevVal + step.value < startVal) {
		prevVal += step.value;
	}

	value.value = prevVal;
	submitValue(prevVal);
}

function nextStep() {
	if (props.ui.disabled || value.value === props.ui.max) return;

	const startVal = value.value ?? 0;
	let nextVal = props.ui.min ?? 0;
	while (!(nextVal > startVal)) {
		nextVal += step.value;
	}

	value.value = nextVal;
	submitValue(nextVal);
}

function snapToStep(clientX: number): number {
	const rect = bar.value!.getBoundingClientRect();
	const ratio = Math.max(0, Math.min(1, (clientX - rect.left) / rect.width));
	const min = props.ui.min ?? 0;
	const max = props.ui.max ?? 0;
	const totalSteps = (max - min) / step.value;
	const snappedStep = Math.round(ratio * totalSteps);
	return parseFloat((min + snappedStep * step.value).toFixed(getDecimalPlacesByStepSize()));
}

function getDecimalPlacesByStepSize(): number {
	const split = `${step.value}`.split('.');
	if (split.length === 1) return 0;
	return split[1].length;
}

function onPointerMove(event: PointerEvent) {
	const snapped = snapToStep(event.clientX);
	if (snapped !== lastSnappedValue) {
		lastSnappedValue = snapped;
		value.value = snapped;
	}
}

function onPointerUp() {
	if (lastSnappedValue !== null && lastSnappedValue !== (props.ui.value ?? 0)) {
		submitValue(lastSnappedValue);
	}
	isDragging.value = false;
	document.removeEventListener('pointermove', onPointerMove);
	document.removeEventListener('pointerup', onPointerUp);
}

function onPointerDown() {
	if (props.ui.disabled) return;
	isDragging.value = true;
	lastSnappedValue = value.value ?? null;
	document.addEventListener('pointermove', onPointerMove);
	document.addEventListener('pointerup', onPointerUp);
}

onUnmounted(() => {
	document.removeEventListener('pointermove', onPointerMove);
	document.removeEventListener('pointerup', onPointerUp);
});

watch(
	() => props.ui.value,
	(nextValue) => {
		value.value = nextValue ?? 0;
	}
);
</script>

<style scoped>
.slider {
	@apply pt-8 pb-4 min-w-32;

	.bar {
		@apply relative w-full h-px;

		.bar-left {
			@apply absolute left-0 top-0 h-full bg-I0;
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

		.grabber {
			@apply absolute top-1/2 -translate-x-1/2 -translate-y-1/2 size-8 !rounded-full flex justify-center items-center cursor-grab;

			.dot {
				content: '';
				@apply block size-3.5 rounded-full bg-I0;
			}

			.value {
				@apply absolute left-1/2 bottom-full -translate-x-1/2 text-I0 text-xs pb-px;
			}

			&:focus,
			&:hover,
			&.dragging {
				@apply bg-I0 bg-opacity-20;
			}

			&.dragging {
				@apply cursor-grabbing;
			}
		}
	}

	&.error {
		.bar {
			.bar-left {
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
					@apply bg-SE0 bg-opacity-20;
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
					@apply bg-ST0 bg-opacity-20;
				}
			}
		}
	}
}
</style>
