<!--
 Copyright (c) 2026 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->
<template>
	<div
		ref="container"
		class="select-ora"
		:class="{
			active: dropdownVisible,
			shifted: isShifted,
			flipped: isFlipped,
		}"
	>
		<div v-if="props.ui.leading" ref="leading" class="leading">
			<UiGeneric :ui="props.ui.leading" />
		</div>

		<input
			:id="id"
			:value="valueLabel"
			class="input-field !pr-8 cursor-pointer"
			:disabled="props.ui.disabled"
			:style="{ paddingLeft: paddingLeft ? `${paddingLeft}px` : undefined }"
			role="combobox"
			readonly
			@click="toggleDropdown(undefined)"
		/>
		<div class="chevron">
			<ArrowDownIcon class="size-3" />
		</div>

		<div
			ref="dropdown"
			class="dropdown"
			:class="{ visible: dropdownVisible }"
			:style="{ ...floatingStyles, minWidth: `${dropdownMinWidth}px` }"
		>
			<div v-if="ui.dropdownInfo || ui.searchable" class="adds">
				<div v-if="ui.dropdownInfo" class="info">
					{{ ui.dropdownInfo }}
				</div>
				<div v-if="ui.searchable" class="search">
					<InputWrapper :wrapper-style="InputWrapperStyle.REDUCED">
						<input v-model="filter" class="input-field" :placeholder="$t('dropdown.search.placeholder')" />
					</InputWrapper>
					<MagnifierIcon />
				</div>
			</div>
			<div ref="options" class="options overflow-y-auto">
				<template v-if="filteredOptions.length">
					<button
						v-for="option in filteredOptions"
						:key="`select_${id}_${option.value}`"
						:value="option.value"
						:disabled="option.disabled"
						role="option"
						@click="selectOption(option)"
					>
						<span>
							{{ option.label }}
						</span>
						<span v-if="option.description" class="description">
							{{ option.description }}
						</span>
					</button>
				</template>
				<template v-else>
					<div class="no-options">
						{{ $t('dropdown.search.no-results') }}
					</div>
				</template>
			</div>
		</div>
	</div>
</template>
<script lang="ts" setup>
import { computed, onMounted, onUnmounted, ref, useTemplateRef } from 'vue';
import { Select, SelectOption, TextFieldStyleValues } from '@/shared/proto/nprotoc_gen';
import { autoUpdate, flip, Middleware, offset, shift, useFloating } from '@floating-ui/vue';
import ArrowDownIcon from '@/assets/svg/arrowDown.svg';
import MagnifierIcon from '@/assets/svg/magnifier.svg';
import { InputWrapperStyle } from '@/components/shared/inputWrapperStyle';
import InputWrapper from '@/components/shared/InputWrapper.vue';
import UiGeneric from '@/components/UiGeneric.vue';

interface Props {
	ui: Select;
}

interface Emits {
	(e: 'update:modelValue', value: string | undefined): void;
}

const props = defineProps<Props>();
const emit = defineEmits<Emits>();

const container = ref();
const leading = useTemplateRef('leading');
const dropdown = ref();
const options = ref();
const dropdownVisible = ref(false);
const filter = ref('');
const dropdownMinWidth = ref<number>();

const { floatingStyles, middlewareData } = useFloating(container, dropdown, {
	placement: 'bottom-start',
	strategy: 'fixed',
	whileElementsMounted: autoUpdate,
	middleware: [
		offset(props.ui.style === TextFieldStyleValues.TextFieldReduced ? 0 : 8),
		flip(),
		props.ui.style !== TextFieldStyleValues.TextFieldReduced ? shift({ crossAxis: true }) : undefined,
	].filter((m) => !!m) as Middleware[],
});

const isFlipped = computed<boolean>(() => {
	return middlewareData.value.flip?.index === 2;
});

const isShifted = computed<boolean>(() => {
	return !!middlewareData.value.shift?.y;
});

const valueLabel = computed<string>(() => {
	if (!props.ui.value || !props.ui.options?.value.length) return '';
	return props.ui.options.value.find((o) => o.value === props.ui.value)?.label || '';
});

const paddingLeft = computed<number | undefined>(() => {
	return leading.value?.offsetWidth;
});

const id = computed<string>(() => {
	if (props.ui.id) {
		return props.ui.id;
	}

	return 'tf-' + props.ui.inputValue;
});

const filteredOptions = computed<SelectOption[]>(() => {
	return (
		props.ui.options?.value.filter((o) => {
			return o.label
				?.toLowerCase()
				.split(' ')
				.some((l) => {
					return filter.value
						.toLowerCase()
						.split(' ')
						.some((f) => l.includes(f));
				});
		}) ?? []
	);
});

function toggleDropdown(visible?: boolean) {
	dropdownVisible.value = visible ?? !dropdownVisible.value;
	if (!dropdownVisible.value) resetDropdown();
}

function selectOption(option: SelectOption) {
	toggleDropdown(false);
	emit('update:modelValue', option.value);
}

function resetDropdown() {
	filter.value = '';
	if (options.value) (options.value as HTMLElement).scrollTo(0, 0);
}

function calcDropdownMinWidth() {
	if (!container.value) return;

	let minWidth = container.value.clientWidth;

	if (props.ui.style === TextFieldStyleValues.TextFieldReduced) {
		const dropdown = container.value.querySelector('.dropdown');
		if (!dropdown) return minWidth;
		const negativeMargin = parseFloat(getComputedStyle(dropdown).marginLeft.replaceAll('px', ''));
		minWidth += 2 * Math.abs(negativeMargin);
	}

	dropdownMinWidth.value = minWidth;
}

function onWindowPointerDown(event: MouseEvent) {
	if (isTargetInSelect(event.target as HTMLElement)) return;
	toggleDropdown(false);
}

function isTargetInSelect(target?: HTMLElement): boolean {
	if (!target || !container.value) return false;

	let temp: HTMLElement | null = target;
	while (temp) {
		if (temp === container.value) return true;
		temp = temp.parentElement;
	}

	return false;
}

function onWindowBlur() {
	toggleDropdown(false);
}

onMounted(() => {
	calcDropdownMinWidth();
	const observer = new ResizeObserver(calcDropdownMinWidth);
	observer.observe(container.value);

	window.addEventListener('pointerdown', onWindowPointerDown);
	window.addEventListener('blur', onWindowBlur);
});

onUnmounted(() => {
	window.removeEventListener('pointerdown', onWindowPointerDown);
	window.removeEventListener('blur', onWindowBlur);
});
</script>
<style scoped>
.select-ora {
	@apply relative;

	.leading {
		@apply absolute inset-y-0 left-0 pl-2 pr-1 flex items-center pointer-events-none;
	}

	input {
		&:hover + .chevron {
			@apply text-I0;
		}
	}

	.chevron {
		@apply absolute inset-y-0 right-0 pr-3 pl-1 flex items-center pointer-events-none;
	}

	.dropdown {
		@apply hidden max-h-80 rounded-lg bg-M4 px-1.5 z-30 shadow-md overscroll-none;

		.adds {
			@apply flex flex-col gap-px px-1.5 pt-1.5;

			.info {
				@apply border-b border-I0 pb-2 pt-1;
			}

			.search {
				@apply relative;

				input {
					@apply pl-7 border-solid border-I0 shadow-none;
					@apply placeholder:text-sm placeholder:text-SI0 dark:placeholder:text-ST0;
				}

				svg {
					@apply absolute left-1 top-1/2 -translate-y-1/2 size-3.5 fill-current;
				}
			}
		}

		.options {
			@apply w-full grow py-2;

			button[role='option'] {
				@apply flex flex-col w-full text-left p-1.5 rounded;
				@apply hover:bg-I0/20 hover:text-I0;

				.description {
					@apply text-sm text-M7 leading-tight -mt-0.5;
				}

				&[disabled] {
					@apply opacity-50 pointer-events-none;
				}
			}

			.no-options {
				@apply py-4 text-sm text-ST0 text-center;
			}
		}

		&.visible {
			@apply flex flex-col;
		}
	}

	&.active {
		.leading {
			@apply z-40;
		}

		.chevron {
			@apply -scale-y-100;
		}
	}

	&.outlined {
		&.active {
			input {
				@apply shadow-md bg-M2;
			}
		}

		.dropdown {
			@apply pt-0;
		}
	}

	&.reduced {
		.chevron {
			@apply pr-0;
		}

		.dropdown {
			min-width: calc(100% + 1.5rem); /* 100% + Tailwind 6 */
			@apply -mx-3 -mt-10 pt-10;
		}

		&.active {
			& > input {
				@apply z-40;
			}

			.chevron {
				@apply z-40;
			}
		}

		&.flipped {
			& > input {
				@apply !border-b-transparent border-t -mt-px;
			}

			.dropdown {
				@apply mt-10 pt-0 pb-10;
			}
		}
	}
}
</style>
