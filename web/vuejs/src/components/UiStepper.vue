<template>
	<div class="stepper-container" :aria-label="ariaLabelStepper">
		<div v-if="isSimple" class="simple-text">
			<template v-if="(ui.value ?? 0) < (ui.steps?.value.length ?? 0)">
				{{ ui.simpleText }}
			</template>
			<template v-else>
				{{ ui.completedText }}
			</template>
		</div>
		<div
			v-if="ui.steps && ui.steps.value.length > 1"
			ref="stepper"
			class="stepper"
			:class="{
				'vertical': isVertical,
				'simple': isSimple,
				'simple-list': isSimpleList,
				'no-numbers': !ui.numbers,
				'no-lines': !ui.lines,
			}"
			:style="`grid-template-columns: repeat(${ui.steps.value.length - 1}, minmax(0, 1fr));`"
		>
			<template v-for="(step, index) in ui.steps.value" :key="`step_${index}`">
				<div
					class="step"
					:class="{
						active: (ui.value ?? 0) === index,
						complete: (ui.value ?? 0) > index,
					}"
					:style="stepStyles"
				>
					<div class="bubble" :style="bubbleStyles">
						<span v-if="ui.numbers" class="label">
							{{ index + 1 }}
						</span>
						<IconCheck />
					</div>
					<div v-if="isHorizontal || isVertical || isSimpleList" class="content" :style="contentStyles">
						<div v-if="step.title" class="title">
							{{ step.title }}
						</div>
						<div v-if="step.subtitle && !isSimpleList" class="subtitle">
							{{ step.subtitle }}
						</div>
					</div>
					<div v-if="index < ui.steps.value.length - 1" class="line" :style="lineStyles"></div>
					<div v-if="index < ui.steps.value.length - 1" class="line-active" :style="lineStyles"></div>
				</div>
			</template>
		</div>
	</div>
</template>
<script lang="ts" setup>
import { Stepper, StepperLayoutValues } from '@/shared/proto/nprotoc_gen';
import { computed, onMounted, ref } from 'vue';
import IconCheck from '@/assets/svg/check.svg';
import { useI18n } from 'vue-i18n';

interface Props {
	ui: Stepper;
}

const props = defineProps<Props>();
const { t } = useI18n();

const stepper = ref<HTMLDivElement>();
const lineLength = ref(0);

const isHorizontal = computed<boolean>(() => props.ui.layout === StepperLayoutValues.StepperLayoutHorizontal);
const isVertical = computed<boolean>(() => props.ui.layout === StepperLayoutValues.StepperLayoutVertical);
const isSimple = computed<boolean>(() => props.ui.layout === StepperLayoutValues.StepperLayoutSimple);
const isSimpleList = computed<boolean>(() => props.ui.layout === StepperLayoutValues.StepperLayoutSimpleList);

const ariaLabelStepper = computed<string>(() => {
	if (!props.ui.steps || !props.ui.steps.value.length) return t('stepper.aria.progressUnknown');
	if ((props.ui.value ?? 0) >= props.ui.steps.value.length) return t('stepper.aria.progressComplete');
	return t('stepper.aria.progress', { current: (props.ui.value ?? 0) + 1, total: props.ui.steps.value.length });
});

const stepStyles = computed<string>(() => {
	if (isHorizontal.value || isSimple.value) return `min-width: ${stepSize.value}px;`;
	if (isVertical.value || isSimpleList.value) return `min-height: ${stepSize.value}px;`;

	return '';
});

const bubbleStyles = computed<string>(() => {
	let styles = `width: ${bubbleSize.value}px; height: ${bubbleSize.value}px;`;
	if (isHorizontal.value || isSimple.value) {
		styles += ` transform: translateX(-50%);`;
	}
	return styles;
});

const contentStyles = computed<string>(() => {
	if (isHorizontal.value || isSimple.value) return '';

	return `top: ${bubbleSize.value / 2}px;`;
});

// line styles contain tiny offsets to prevent render errors
const lineStyles = computed<string>(() => {
	let styles;
	if (isVertical.value || isSimpleList.value) {
		styles = `height: ${lineLength.value + 2}px;`;
		if (bubbleSize.value) styles += ` left: ${bubbleSize.value / 2}px; top: ${bubbleSize.value - 1}px;`;
		return styles;
	}

	styles = `width: ${lineLength.value + 2}px;`;
	if (bubbleSize.value) styles += ` left: ${bubbleSize.value / 2 - 1}px; top: ${bubbleSize.value / 2}px;`;
	return styles;
});

const stepSize = computed<number>(() => {
	if (isHorizontal.value || isVertical.value) return 130;
	if (isSimple.value) return 20;
	if (isSimpleList.value) return 30;

	return 0;
});

const bubbleSize = computed<number>(() => {
	if (isHorizontal.value || isVertical.value) return 30;
	if (isSimple.value || isSimpleList.value) return 10;

	return 0;
});

function calcLineLength() {
	if (!stepper.value) return;

	const step = stepper.value.querySelector('.step');
	if (!step) return;

	const bubble = step.querySelector('.bubble');
	if (!bubble) return;

	if (isVertical.value || isSimpleList.value) lineLength.value = step.clientHeight - bubble.clientHeight;
	else lineLength.value = step.clientWidth - bubble.clientWidth;
}

onMounted(() => {
	calcLineLength();
	window.addEventListener('resize', calcLineLength);
});
</script>
<style scoped>
.stepper-container {
	@apply flex justify-center items-center gap-8;

	.stepper {
		@apply grid relative;

		.step {
			@apply relative flex flex-col;

			.bubble {
				@apply relative mb-2 flex justify-center items-center rounded-full text-lg z-[1];
				@apply outline outline-2 -outline-offset-2 outline-current;
				@apply text-DIS duration-200;

				svg {
					@apply hidden;
				}
			}

			.content {
				@apply pr-4 text-DIS duration-200 -translate-x-4 flex flex-col gap-1;

				.title {
					@apply font-semibold leading-none;
				}

				.subtitle {
					@apply font-light leading-none;
				}
			}

			.line,
			.line-active {
				@apply absolute -translate-y-1/2 h-0.5 bg-current;
			}

			.line {
				@apply bg-DIS;
			}

			.line-active {
				@apply duration-200;
			}

			&:last-child {
				@apply absolute left-full w-max;
			}

			&.active {
				.bubble,
				.content {
					@apply text-current;
				}
			}

			&.complete {
				.bubble,
				.content {
					@apply text-current;
				}

				.bubble {
					@apply bg-current;

					.label {
						@apply text-M1;
					}
				}
			}

			&:not(.complete) {
				.line-active {
					@apply !w-0;
				}
			}
		}

		&.vertical,
		&.simple-list {
			@apply !grid-cols-1 mb-8;

			.step {
				@apply flex-row gap-4;

				.bubble {
					@apply mb-0;
				}

				.content {
					@apply absolute left-full -translate-y-1/2 translate-x-0 pr-0 pl-3;

					.title,
					.subtitle {
						@apply w-max;
					}
				}

				.line,
				.line-active {
					@apply -translate-x-1/2 translate-y-0 w-0.5;
				}

				&:last-child {
					@apply left-auto top-full;
				}

				&:not(.complete) {
					.line-active {
						@apply !h-0;
					}
				}
			}
		}

		&.simple {
			.step {
				.bubble {
					@apply mb-0 outline-1 -outline-offset-1;

					.label {
						@apply hidden;
					}
				}

				.line,
				.line-active {
					@apply h-px;
				}
			}
		}

		&.simple-list {
			.step {
				.bubble {
					@apply mb-0 outline-1 -outline-offset-1;

					.label {
						@apply hidden;
					}
				}

				.line,
				.line-active {
					@apply w-px;
				}
			}
		}

		&.no-numbers:not(.simple, .simple-list) {
			.step {
				&.active {
					.bubble {
						&:after {
							content: '';
							@apply block size-1.5 rounded-full bg-current;
						}
					}
				}

				&.complete {
					.bubble {
						svg {
							@apply block size-3 *:fill-M1;
						}
					}
				}
			}
		}

		&.simple.no-lines,
		&.simple-list.no-lines {
			.step {
				.line,
				.line-active {
					@apply hidden;
				}
			}
		}
	}
}
</style>
