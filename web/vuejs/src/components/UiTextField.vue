<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<script lang="ts" setup>
import { computed, onMounted, ref, useTemplateRef, watch } from 'vue';
import CloseIcon from '@/assets/svg/close.svg';
import UiGeneric from '@/components/UiGeneric.vue';
import InputWrapper from '@/components/shared/InputWrapper.vue';
import { frameCSS } from '@/components/shared/frame';
import { inputWrapperStyleFrom } from '@/components/shared/inputWrapperStyle';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import { nextRID } from '@/eventhandling';
import type { TextField } from '@/shared/proto/nprotoc_gen';
import {
	KeyboardTypeValues,
	TextAlignmentValues,
	TextFieldStyleValues,
	UpdateStateValueRequested,
} from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: TextField;
}>();

const serviceAdapter = useServiceAdapter();
const leadingElement = useTemplateRef('leadingElement');
const trailingElement = useTemplateRef('trailingElement');
const clearButton = useTemplateRef('clearButton');
const showZero = !!props.ui.showZero;
const step = props.ui.step || 0;
const inputValue = ref<string>(props.ui.value ? formatValue(props.ui.value) : '');
let timer: number = 0;

const textarea = ref<HTMLTextAreaElement>();
const textareaHeight = ref('auto');

const frameStyles = computed<string>(() => {
	const styles = frameCSS(props.ui.frame);

	return styles.join(';');
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

function parseFloat(input: string) {
	if (input === '') {
		return '0';
	}

	const negative = input.lastIndexOf('-') >= 0;

	// Remove any non digits, superfluous separators and leading zeros
	const parts = input.split(/[,.]/);
	for (let i = 0; i < parts.length; i++) {
		parts[i] = parts[i].replaceAll(/\D/g, '');
		parts[i] = parts[i].replaceAll(/^0+/g, '');
	}
	const lastPart = parts.length > 1 ? parts.pop() : undefined;
	const firstPart = parts.join('');

	// Combine to final value
	let finalValue = firstPart === '' ? '0' : firstPart;
	finalValue += lastPart === undefined || lastPart === '' ? '' : '.' + lastPart;

	finalValue = negative ? '-' + finalValue : finalValue;
	return finalValue;
}

function formatFloat(input: string) {
	const negative = input.indexOf('-') >= 0 && input.indexOf('-') === input.lastIndexOf('-');
	const fractionSeparator = isLanguageGerman() ? ',' : '.';

	const parts = input.split(/[,.]/);
	const decimals = parts.length > 1 ? parts.pop()?.replaceAll(/\D/g, '') : undefined;
	let finalValue = parts.join('').replaceAll(/\D/g, '');
	finalValue = /^0+$/g.test(finalValue) ? '0' : finalValue.replaceAll(/^0+/g, '');

	if (decimals === '' && !showZero) {
		finalValue += fractionSeparator; // There is a tailing separator symbol
	} else if (decimals && decimals.length > 0) {
		finalValue += fractionSeparator + decimals;
	}

	if (finalValue === '') {
		finalValue = showZero ? '0' : '';
	} else {
		finalValue = negative ? '-' + finalValue : finalValue;
	}
	return finalValue;
}

function parseInt(input: string) {
	const negative = input.lastIndexOf('-') >= 0;
	let value = input.replace(/\D/g, '');
	if (value === '') {
		value = '0';
	} else {
		value = negative ? '-' + value : value;
	}
	return value;
}

function formatInt(input: string) {
	const negative = input.lastIndexOf('-') >= 0;
	const digits = input.split('');

	let finalValue = '';
	for (let i = 0; i < digits.length; i++) {
		const digit = digits[i];
		if (/\D/g.test(digit)) {
			continue;
		}
		if (finalValue !== '' || digit !== '0') {
			finalValue += digit;
		}
	}
	if (finalValue === '') {
		finalValue = showZero ? '0' : '';
	} else {
		finalValue = negative ? '-' + finalValue : finalValue;
	}

	return finalValue;
}

function parseValue(value: string) {
	switch (props.ui.keyboardOptions?.keyboardType) {
		case KeyboardTypeValues.KeyboardInteger:
			return parseInt(value);
		case KeyboardTypeValues.KeyboardFloat:
			return parseFloat(value);
		default:
			return value;
	}
}

function formatValue(value: string) {
	let formattedValue;
	switch (props.ui.keyboardOptions?.keyboardType) {
		case KeyboardTypeValues.KeyboardInteger:
			formattedValue = formatInt(value);
			break;
		case KeyboardTypeValues.KeyboardFloat:
			formattedValue = formatFloat(value);
			break;
		default:
			formattedValue = value;
	}
	return formattedValue;
}

const inputStyle = computed<string>(() => {
	const styles: string[] = [];

	switch (props.ui.textAlignment) {
		case TextAlignmentValues.TextAlignStart:
			styles.push('text-align: start');
			break;
		case TextAlignmentValues.TextAlignEnd:
			styles.push('text-align: end');
			break;
		case TextAlignmentValues.TextAlignCenter:
			styles.push('text-align: center');
			break;
		case TextAlignmentValues.TextAlignJustify:
			styles.push('text-align: justify', 'text-justify: inter-character'); // inter-character just looks so much better
			break;
	}

	if (props.ui.style == TextFieldStyleValues.TextFieldBasic) {
		styles.push('display:inline', 'background:unset');

		return styles.join(';');
	}

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

	if (props.ui.lines) {
		styles.push(`height: ${textareaHeight.value}`);
	}

	styles.push('padding-left:' + paddingLeft, 'padding-right:' + paddingRight);
	return styles.join(';');
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
	const formattedValue = formatValue(newValue);
	if (inputValue.value != formattedValue) {
		inputValue.value = formattedValue;
	}
});

watch(
	() => props.ui.value,
	(newValue) => {
		if (document.getElementById(id.value) !== document.activeElement) {
			inputValue.value = formatValue(newValue || '');
		}
	}
);

watch(
	() => props.ui,
	(newValue) => {
		if (document.getElementById(id.value) !== document.activeElement) {
			inputValue.value = formatValue(newValue.value || '');
		}
	}
);

function handleKeydownEnter(event: KeyboardEvent) {
	event.stopPropagation();

	// textarea
	if (props.ui.lines) {
		if (!props.ui.keydownEnter || event.shiftKey) return;
		event.preventDefault();
	}

	sendKeydownEnterEvent();
}

function onInputUp() {
	if (isNumerical()) changeValue(step);
}

function onInputDown() {
	if (isNumerical()) changeValue(-step);
}

function resizeTextarea() {
	if (!textarea.value) return;

	const computedStyle = getComputedStyle(textarea.value);
	const borderTop = window.parseFloat(parseFloat(computedStyle.getPropertyValue('border-top-width')));
	const borderBottom = window.parseFloat(parseFloat(computedStyle.getPropertyValue('border-bottom-width')));
	textarea.value.style.height = 'auto';
	const height = textarea.value.scrollHeight + borderTop + borderBottom;
	textarea.value.style.height = `${height}px`;
	textareaHeight.value = `${height}px`;
}

function sendKeydownEnterEvent() {
	if (!props.ui.keydownEnter) return;

	const parsedValue = parseValue(inputValue.value);
	// note that we must always issue the key-down event, even we did not change the text
	serviceAdapter.sendEvent(
		new UpdateStateValueRequested(props.ui.inputValue, props.ui.keydownEnter, nextRID(), parsedValue)
	);
	clearTimeout(timer); // cancel any debounced follow up event
}

function onTextareaInput(force: boolean) {
	resizeTextarea();
	submitInputValue(force);
}

function submitInputValue(force: boolean, functionPointer: number = 0): void {
	putValueInRange();

	const parsedValue = parseValue(inputValue.value);
	if (parsedValue == props.ui.value) {
		return;
	}

	// Note, that the sendEvent may have a huge latency, causing ghost updates for the user input.
	// Thus, immediately increase the request id, so that everybody knows, that any older responses are outdated.
	nextRID();

	if (force || props.ui.disableDebounce) {
		serviceAdapter.sendEvent(
			new UpdateStateValueRequested(props.ui.inputValue, functionPointer, nextRID(), parsedValue)
		);
		return;
	}

	debouncedInput();
}

function putValueInRange() {
	if (!isNumerical() || (!props.ui.min && !props.ui.max)) return;

	let numberVal = getNumberValue();
	numberVal = Math.max(props.ui.min || 0, numberVal);
	if (props.ui.max) numberVal = Math.min(props.ui.max, numberVal);

	inputValue.value = formatValue(`${numberVal}`);
}

function onInputWheel(e: WheelEvent) {
	if (!isNumerical()) return;

	const up = e.deltaY < 0;
	const down = e.deltaY > 0;
	if (!up && !down) return;

	changeValue(up ? step : -step);
}

function changeValue(amount: number) {
	if (!isNumerical()) return;

	let numberVal = getNumberValue();
	numberVal += amount;
	numberVal = Math.max(props.ui.min || 0, numberVal);
	if (props.ui.max) numberVal = Math.min(props.ui.max, numberVal);

	inputValue.value = formatValue(`${numberVal}`);
	submitInputValue(false);
}

function getNumberValue(): number {
	if (props.ui.keyboardOptions?.keyboardType === KeyboardTypeValues.KeyboardInteger) {
		return window.parseInt(parseInt(inputValue.value));
	} else if (props.ui.keyboardOptions?.keyboardType === KeyboardTypeValues.KeyboardFloat) {
		return window.parseFloat(parseFloat(inputValue.value));
	}

	return 0;
}

function isNumerical() {
	return (
		props.ui.keyboardOptions?.keyboardType == KeyboardTypeValues.KeyboardInteger ||
		props.ui.keyboardOptions?.keyboardType == KeyboardTypeValues.KeyboardFloat
	);
}

function isLanguageGerman() {
	return navigator.language.split('-')[0].toLowerCase() === 'de';
}

function onInputFocus() {
	if (!isNumerical()) return;

	const input = document.getElementById(id.value) as HTMLInputElement;
	if (input) input.select();
}

function leaveFocus(): void {
	inputValue.value = formatValue(inputValue.value);
	submitInputValue(true);
}

function clearInputValue(): void {
	inputValue.value = isNumerical() ? '0' : '';
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
		const parsedValue = parseValue(inputValue.value);
		if (parsedValue == props.ui.value) {
			return;
		}
		serviceAdapter.sendEvent(new UpdateStateValueRequested(props.ui.inputValue, 0, nextRID(), parsedValue));
	}, debounceTime);
}

onMounted(resizeTextarea);

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
					:id="id"
					v-model="inputValue"
					class="input-field"
					:style="inputStyle"
					:disabled="props.ui.disabled"
					type="text"
					:inputmode="inputMode"
					@keydown.enter="handleKeydownEnter"
					@keydown.up.prevent="onInputUp"
					@keydown.down.prevent="onInputDown"
					@focus="onInputFocus"
					@focusout="leaveFocus"
					@input="submitInputValue(false)"
					@wheel="onInputWheel"
				/>
				<textarea
					v-if="props.ui.lines"
					:id="id"
					ref="textarea"
					v-model="inputValue"
					class="input-field"
					:style="inputStyle"
					:disabled="props.ui.disabled"
					type="text"
					:rows="props.ui.lines"
					@keydown.enter="handleKeydownEnter"
					@focusout="submitInputValue(true)"
					@input="onTextareaInput(false)"
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
					v-else-if="
						inputValue &&
						!props.ui.disabled &&
						!props.ui.lines &&
						props.ui.style != TextFieldStyleValues.TextFieldBasic
					"
					ref="clearButton"
					class="absolute inset-y-0 right-0 pr-2 pl-1 flex items-center"
				>
					<CloseIcon class="w-4" tabindex="-1" @click="clearInputValue" @keydown.enter="clearInputValue" />
				</div>
			</div>
		</InputWrapper>
	</div>
</template>
