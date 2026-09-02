<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<script lang="ts" setup>
import { computed, inject, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import CloseIcon from '@/assets/svg/close.svg';
import UiGeneric from '@/components/UiGeneric.vue';
import InputWrapper from '@/components/shared/InputWrapper.vue';
import { frameCSS } from '@/components/shared/frame';
import { inputWrapperStyleFrom } from '@/components/shared/inputWrapperStyle';
import { randomStr } from '@/components/shared/util';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import { nextRID } from '@/eventhandling';
import type { TextField } from '@/shared/proto/nprotoc_gen';
import {
	KeyboardTypeValues,
	TextAlignmentValues,
	TextFieldStyleValues,
	UpdateStateValueRequested,
} from '@/shared/proto/nprotoc_gen';
import { PRE_SUBMIT_GROUP, PreSubmitGroup } from '@/components/form/preSubmitGroup';

const props = defineProps<{
	ui: TextField;
}>();

const id = props.ui.id || randomStr(16);

const serviceAdapter = useServiceAdapter();

const preSubmitGroup = inject<PreSubmitGroup | null>(PRE_SUBMIT_GROUP, null);

const leadingElement = ref<HTMLDivElement>();
const trailingElement = ref<HTMLDivElement>();
const clearButton = ref<HTMLButtonElement>();
const leadingWidth = ref(0);
const trailingWidth = ref(0);

const showZero = !!props.ui.showZero;
const step = props.ui.step || 0;
const inputValue = ref<string>(props.ui.value ? formatValue(props.ui.value) : '');
let timer: number = 0;
// unmounting tells a focusout caused by our own destruction apart from the user really leaving the field.
// See onFocusOut for why this matters.
let unmounting = false;
// lastSentValue is the (parsed) value we last transmitted to the server. It is our synchronization anchor:
// as long as the local input still equals it, the user has not typed anything the server does not know about,
// and therefore any incoming server value is newer than our local state and must win.
let lastSentValue: string = props.ui.value ?? '';

const textarea = ref<HTMLTextAreaElement>();
const textareaHeight = ref('auto');

const frameStyles = computed<string>(() => {
	const styles = frameCSS(props.ui.frame);

	return styles.join(';');
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

const clearButtonVisible = computed<boolean>(() => {
	return (
		!!inputValue.value &&
		!props.ui.disabled &&
		!props.ui.lines &&
		props.ui.style != TextFieldStyleValues.TextFieldBasic &&
		!!props.ui.clearButton
	);
});

function parseFloat(input: string) {
	if (input === '') {
		return '';
	}

	const negative = input.startsWith('-');

	// Remove any non digits, superfluous separators and leading zeros
	const parts = input.split(/[,.]/);
	for (let i = 0; i < parts.length; i++) {
		parts[i] = parts[i].replaceAll(/\D/g, '');
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
	finalValue = /^0+$/g.test(finalValue) ? '0' : finalValue;

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

	const paddingLeft = leadingWidth.value ? `calc(${leadingWidth.value}px + 0.5rem)` : 'auto';
	const paddingRight = trailingWidth.value
		? `calc(${trailingWidth.value}px + 0.5rem)`
		: clearButtonVisible.value
			? '2.5rem'
			: 'auto';
	styles.push('padding-left:' + paddingLeft, 'padding-right:' + paddingRight);

	if (props.ui.lines) {
		styles.push(`height: ${textareaHeight.value}`);
	}

	return styles.join(';');
});

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
	lastSentValue = parsedValue;
	clearTimeout(timer); // cancel any debounced follow up event
}

function onTextareaInput(force: boolean) {
	resizeTextarea();
	submitInputValue(force);
}

function submitInputValue(force: boolean, functionPointer: number = 0): void {
	putValueInRange();

	// Any pending debounce is obsolete from here on: either we send right now ourselves, or the value already
	// matches the server. Would it fire later, it would push back a value the server has meanwhile discarded.
	clearTimeout(timer);

	const parsedValue = parseValue(inputValue.value);
	// note the ?? '': an empty value is not transmitted by the protocol, thus it arrives as undefined.
	if (parsedValue == (props.ui.value ?? '')) {
		return;
	}

	// Note, that the sendEvent may have a huge latency, causing ghost updates for the user input.
	// Thus, immediately increase the request id, so that everybody knows, that any older responses are outdated.
	nextRID();

	if (force || props.ui.disableDebounce) {
		serviceAdapter.sendEvent(
			new UpdateStateValueRequested(props.ui.inputValue, functionPointer, nextRID(), parsedValue)
		);
		lastSentValue = parsedValue;
		return;
	}

	debouncedInput();
}

function putValueInRange() {
	if (!isNumerical() || (!props.ui.min && !props.ui.max)) return;

	let numberVal = getNumberValue();
	numberVal = Math.max(props.ui.min || 0, numberVal);
	if (props.ui.max) numberVal = Math.min(props.ui.max, numberVal);

	const formatted = formatValue(`${numberVal}`);
	if (formatted === inputValue.value) return;

	inputValue.value = formatted;
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

	const input = document.getElementById(id) as HTMLInputElement;
	if (input) input.select();
}

/**
 * Handles the focusout driven submit.
 *
 * Removing a focused element from the DOM makes the browser emit focusout, which is indistinguishable from the
 * user tabbing away: relatedTarget is null and the activeElement falls back to the body, exactly like a click on
 * an inert area. The element is also still connected at that point, so isConnected cannot be used either.
 *
 * Submitting in that situation writes the local text back into the bound server state and resurrects a value the
 * server has just discarded. The visible symptom is a chat input which refuses to clear after sending: the submit
 * handler empties the state, the re-rendering recreates this component, and the dying instance pushes its old
 * text back in, so the next rendering shows it again.
 *
 * onBeforeUnmount is guaranteed to run before the DOM node is detached and therefore before that focusout, which
 * makes it a reliable discriminator.
 */
function onFocusOut(): void {
	if (unmounting) {
		return;
	}

	inputValue.value = formatValue(inputValue.value);
	submitInputValue(true);
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
		const parsedValue = parseValue(inputValue.value);
		if (parsedValue == (props.ui.value ?? '')) {
			return;
		}
		serviceAdapter.sendEvent(new UpdateStateValueRequested(props.ui.inputValue, 0, nextRID(), parsedValue));
		lastSentValue = parsedValue;
	}, debounceTime);
}

function calcAdditionalElementsWidths() {
	leadingWidth.value = leadingElement.value?.getBoundingClientRect().width || 0;
	trailingWidth.value = trailingElement.value?.getBoundingClientRect().width || 0;
}

function observeAdditionalElements() {
	const observer = new ResizeObserver(calcAdditionalElementsWidths);
	if (leadingElement.value) observer.observe(leadingElement.value);
	if (trailingElement.value) observer.observe(trailingElement.value);
	if (clearButton.value) observer.observe(clearButton.value);
}

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

/**
 * Adopts values coming from the server.
 *
 * Outdated renderings (those which are older than our last interaction) are already discarded globally in App.vue
 * based on the request tracing id. What that mechanism cannot catch is a *current* rendering, triggered by some
 * other component, which still carries an older value for this text field - namely everything the user has typed
 * but which has not been transmitted yet due to the debounce. Therefore we only adopt the server value while the
 * local input still equals what we last sent: in that case the server knows everything we know and has decided.
 *
 * This is what makes a server side reset work while the field keeps the focus, e.g. a chat input which is emptied
 * by its own submit handler. Note that an empty value is not transmitted at all by the protocol and thus arrives
 * as undefined, which is why a plain watcher on props.ui.value cannot observe a reset to the empty string.
 */
watch(
	() => props.ui,
	(newValue) => {
		const serverValue = newValue.value ?? '';
		const parsedInputValue = parseValue(inputValue.value);
		const parsedServerValue = parseValue(serverValue);
		if (parsedInputValue !== lastSentValue || parsedInputValue === parsedServerValue) {
			return;
		}

		clearTimeout(timer);
		inputValue.value = formatValue(serverValue);
		lastSentValue = serverValue;
	}
);

watch(inputValue, () => nextTick(resizeTextarea));
onBeforeUnmount(() => {
	unmounting = true;
});

onMounted(() => {
	if (preSubmitGroup) preSubmitGroup.join(() => submitInputValue(true));
	resizeTextarea();
	calcAdditionalElementsWidths();
	observeAdditionalElements();
});

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
			:input-id="id"
			:optional="ui.optional"
		>
			<div class="relative flex flex-col">
				<!-- Leading view -->
				<div v-if="props.ui.leading" ref="leadingElement" class="additional-left">
					<UiGeneric :ui="props.ui.leading" />
				</div>

				<input
					v-if="!props.ui.lines"
					:id="id"
					v-model="inputValue"
					:autocomplete="ui.autocomplete"
					class="input-field"
					:style="inputStyle"
					:disabled="props.ui.disabled"
					type="text"
					:inputmode="inputMode"
					@keydown.enter="handleKeydownEnter"
					@keydown.up.prevent="onInputUp"
					@keydown.down.prevent="onInputDown"
					@focus="onInputFocus"
					@focusout="onFocusOut"
					@input="submitInputValue(false)"
					@wheel="onInputWheel"
				/>
				<textarea
					v-if="props.ui.lines"
					:id="id"
					ref="textarea"
					v-model="inputValue"
					:autocomplete="ui.autocomplete"
					class="input-field"
					:style="inputStyle"
					:disabled="props.ui.disabled"
					type="text"
					:rows="props.ui.lines"
					@keydown.enter="handleKeydownEnter"
					@focusout="onFocusOut"
					@input="onTextareaInput(false)"
				/>

				<!-- Trailing view -->
				<div v-if="props.ui.trailing" ref="trailingElement" class="additional-right">
					<UiGeneric :ui="props.ui.trailing" />
				</div>

				<!-- Clear button -->
				<button
					v-else-if="clearButtonVisible"
					ref="clearButton"
					type="button"
					class="button-tertiary square small additional-right clear-button"
					@click="clearInputValue"
				>
					<CloseIcon class="size-3" />
				</button>
			</div>
		</InputWrapper>
	</div>
</template>

<style scoped>
.additional-left,
.additional-right {
	@apply absolute top-1/2 -translate-y-1/2 pointer-events-auto;
}

.additional-left {
	@apply left-1.5;
}

.additional-right {
	@apply right-1.5;
}

.clear-button {
	@apply size-8 p-1;
}
</style>
