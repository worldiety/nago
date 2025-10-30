<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<script lang="ts" setup>
import { computed, ref, useTemplateRef, watch } from 'vue';
import CloseIcon from '@/assets/svg/close.svg';
import UiGeneric from '@/components/UiGeneric.vue';
import InputWrapper from '@/components/shared/InputWrapper.vue';
import { frameCSS } from '@/components/shared/frame';
import { inputWrapperStyleFrom } from '@/components/shared/inputWrapperStyle';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import { nextRID } from '@/eventhandling';
import {
	FunctionCallRequested,
	KeyboardTypeValues,
	TextField,
	UpdateStateValueRequested,
} from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: TextField;
}>();

const serviceAdapter = useServiceAdapter();
const leadingElement = useTemplateRef('leadingElement');
const trailingElement = useTemplateRef('trailingElement');
const clearButton = useTemplateRef('clearButton');
const inputValue = ref<string>(props.ui.value ? props.ui.value : '');
let timer: number = 0;

const frameStyles = computed<string>(() => {
	return frameCSS(props.ui.frame).join(';');
});

const id = computed<string>(() => {
	if (props.ui.id) {
		return props.ui.id;
	}

	return 'tf-' + props.ui.inputValue;
});

const inputMode = computed<'numeric' | 'decimal' | 'email' | 'tel' | 'url' | 'search' | 'text' | 'none' | undefined>(
	() => {
		switch (props.ui.keyboardOptions?.keyboardType) {
			case KeyboardTypeValues.KeyboardInteger:
				return 'numeric';
			case KeyboardTypeValues.KeyboardFloat:
				return 'decimal';
			case KeyboardTypeValues.KeyboardEMail:
				return 'email';
			case KeyboardTypeValues.KeyboardPhone:
				return 'tel';
			case KeyboardTypeValues.KeyboardURL:
				return 'url';
			case KeyboardTypeValues.KeyboardSearch:
				return 'search';
		}

		return 'text';
	}
);

const inputStyle = computed<Record<string, string>>(() => {
	const leadingElementWidth = leadingElement.value?.offsetWidth;
	const paddingLeft = leadingElementWidth ? `${leadingElementWidth}px` : 'auto';

	let paddingRight: string;
	const trailingElementWidth = trailingElement.value?.offsetWidth;
	if (trailingElementWidth !== undefined) {
		paddingRight = `${trailingElementWidth}px`;
	} else {
		const clearButtonElementWidth = clearButton.value?.offsetWidth;
		paddingRight = clearButtonElementWidth ? `${clearButtonElementWidth}px` : 'auto';
	}

	return {
		'padding-left': paddingLeft,
		'padding-right': paddingRight,
	};
});

/**
 * Validates the input value and submits it, if it is valid.
 * The '-' sign and the empty string are treated as 0.
 * If the input value is invalid, the value gets reset to the last known valid value.
 */
watch(inputValue, (newValue, oldValue) => {
	if (newValue == oldValue) {
		return;
	}

	if (props.ui.keyboardOptions?.keyboardType == KeyboardTypeValues.KeyboardInteger) {
		if (newValue === '' || newValue === '-') {
			inputValue.value = '0';
		} else if (!newValue.match(/^-?[0-9]+$/)) {
			inputValue.value = oldValue;
		}

		return;
	}

	if (props.ui.keyboardOptions?.keyboardType == KeyboardTypeValues.KeyboardFloat) {
		if (newValue === '' || newValue === '-') {
			inputValue.value = '0';
		} else if (!newValue.match(/^[+-]?(\d+(\.\d*)?|\.\d+)$/)) {
			inputValue.value = oldValue;
		}

		return;
	}
});

watch(
	() => props.ui.value,
	(newValue) => {
		if (newValue) {
			inputValue.value = newValue;
		} else {
			inputValue.value = '';
		}
	}
);

watch(
	() => props.ui,
	(newValue) => {
		if (newValue.value) {
			inputValue.value = newValue.value;
		} else {
			inputValue.value = '';
		}
	}
);

function handleKeydownEnter(event: Event) {
	if (props.ui.keydownEnter) {
		event.stopPropagation();
		serviceAdapter.sendEvent(
			new UpdateStateValueRequested(props.ui.inputValue, props.ui.keydownEnter, nextRID(), inputValue.value)
		);
		clearTimeout(timer); // cancel any debounced follow up event
	}
}

function submitInputValue(force: boolean): void {
	if (inputValue.value == props.ui.value) {
		return;
	}

	// Note, that the sendEvent may have a huge latency, causing ghost updates for the user input.
	// Thus, immediately increase the request id, so that everybody knows, that any older responses are outdated.
	nextRID();

	if (force || props.ui.disableDebounce) {
		serviceAdapter.sendEvent(new UpdateStateValueRequested(props.ui.inputValue, 0, nextRID(), inputValue.value));

		return;
	}

	debouncedInput();
}

function clearInputValue(): void {
	inputValue.value = '';
	submitInputValue(true);
}

function deserializeGoDuration(durationInNanoseconds: number): number {
	return durationInNanoseconds / 1e6;
}

function debouncedInput() {
	let debounceTime = 500; // ms
	if (props.ui.debounceTime && props.ui.debounceTime > 0) {
		debounceTime = deserializeGoDuration(props.ui.debounceTime);
	}

	clearTimeout(timer);
	timer = window.setTimeout(() => {
		if (inputValue.value == props.ui.value) {
			return;
		}

		serviceAdapter.sendEvent(new UpdateStateValueRequested(props.ui.inputValue, 0, nextRID(), inputValue.value));
	}, debounceTime);
}

// TODO check :id="idPrefix + props.ui.id.toString()"
</script>

<template>
	<div v-if="!ui.invisible" :style="frameStyles">
		<InputWrapper
			:wrapper-style="inputWrapperStyleFrom(props.ui.style)"
			:label="props.ui.label"
			:error="props.ui.errorText"
			:help="props.ui.supportingText"
			:disabled="props.ui.disabled"
		>
			<div class="relative">
				<!-- Leading view -->
				<div
					v-if="props.ui.leading"
					ref="leadingElement"
					class="absolute inset-y-0 left-0 pl-2 pr-1 flex items-center pointer-events-none"
				>
					<UiGeneric :ui="props.ui.leading" />
				</div>

				<input
					v-if="!props.ui.lines"
					@keydown.enter="handleKeydownEnter"
					:id="id"
					v-model="inputValue"
					class="input-field"
					:style="inputStyle"
					:disabled="props.ui.disabled"
					type="text"
					:inputmode="inputMode"
					@focusout="submitInputValue(true)"
					@input="submitInputValue(false)"
				/>
				<textarea
					v-if="props.ui.lines"
					:id="id"
					v-model="inputValue"
					class="input-field"
					:style="inputStyle"
					:disabled="props.ui.disabled"
					type="text"
					:rows="props.ui.lines"
					@focusout="submitInputValue(true)"
					@input="submitInputValue(false)"
				/>

				<!-- Trailing view -->
				<div
					v-if="props.ui.trailing"
					ref="trailingElement"
					class="absolute inset-y-0 right-0 pr-2 pl-1 flex items-center pointer-events-none"
				>
					<UiGeneric :ui="props.ui.trailing" />
				</div>

				<!-- Clear button -->
				<div
					v-else-if="inputValue && !props.ui.disabled && !props.ui.lines"
					ref="clearButton"
					class="absolute inset-y-0 right-0 pr-2 pl-1 flex items-center"
				>
					<CloseIcon class="w-4" tabindex="-1" @click="clearInputValue" @keydown.enter="clearInputValue" />
				</div>
			</div>
		</InputWrapper>
	</div>
</template>
